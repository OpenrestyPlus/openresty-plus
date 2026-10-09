<script lang="ts" setup>
import type {
  HttpListener,
  OrpDirective,
  OrpIpGroup,
  OrpIpPolicy,
} from '#/api';

import { computed, reactive, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  AutoComplete,
  Button,
  Input,
  message,
  RadioGroup,
  Select,
  Switch,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import { createHttpListenerApi, listCentersApi, listIpGroupsApi, updateHttpListenerApi } from '#/api';
import { useDirtyGuard } from '#/composables/use-dirty-guard';

defineOptions({ name: 'HttpListenerFormModal' });

const emit = defineEmits<{ success: [] }>();

const isEdit = ref(false);
const editId = ref(0);

/** Server 块指令编辑行 */
interface EditableDirective {
  name: string;
  value: string;
}

const editDirectives = ref<EditableDirective[]>([]);
const errorPages = ref<Record<string, string>>({});
const activeErrorKey = ref('');
const errorStatus = ref('404');
const errorContentType = ref('');
const errorContent = ref('');

function selectErrorPage(key: string) {
  activeErrorKey.value = key;
  const [status = '404', ...typeParts] = key.split('|');
  errorStatus.value = status;
  errorContentType.value = typeParts.join('|');
  errorContent.value = errorPages.value[key] ?? '';
}

function saveErrorPage(): boolean {
  const status = errorStatus.value.trim();
  const contentType = errorContentType.value.trim();
  const key = contentType ? `${status}|${contentType}` : status;
  if (!/^([45]\d\d)$/.test(status) || !errorContent.value.trim()) {
    message.warning('请输入 400–599 的状态码和页面内容');
    return false;
  }
  if (new TextEncoder().encode(errorContent.value).length > 1024 * 1024) {
    message.warning('页面内容不能超过 1 MB');
    return false;
  }
  if (contentType && !/^[\w!#$&^.+-]+\/[\w!#$&^.+-]+(?:\s*;\s*[\w!#$&^.+-]+=(?:[\w!#$&^.+-]+|"[^"]*"))*$/.test(contentType)) {
    message.warning('Content-Type 格式无效');
    return false;
  }
  errorPages.value = { ...errorPages.value, [key]: errorContent.value };
  activeErrorKey.value = key;
  markDirty();
  return true;
}

function hasUnsavedErrorPage(): boolean {
  const status = errorStatus.value.trim();
  const type = errorContentType.value.trim();
  const key = type ? `${status}|${type}` : status;
  return Boolean(
    errorContent.value || status !== '404' || type ||
    (activeErrorKey.value && (key !== activeErrorKey.value || errorContent.value !== errorPages.value[activeErrorKey.value])),
  );
}

function deleteErrorPage() {
  if (!activeErrorKey.value) return;
  const { [activeErrorKey.value]: _, ...remaining } = errorPages.value;
  errorPages.value = remaining;
  const next = Object.keys(remaining).sort()[0] ?? '';
  if (next) selectErrorPage(next);
  else {
    activeErrorKey.value = '';
    errorStatus.value = '404';
    errorContentType.value = '';
    errorContent.value = '';
  }
}

/** IP 访问控制策略状态（server {} 块级 allow/deny 指令组） */
const ipPolicy = reactive<OrpIpPolicy>({
  enabled: false,
  priority: 'allow-first',
  allowList: [],
  denyList: [],
});

/** IP 策略优先模式选项（allow 组 / deny 组谁先生成） */
const IP_PRIORITY_OPTIONS = [
  { label: '白名单优先', value: 'allow-first' },
  { label: '黑名单优先', value: 'deny-first' },
];

/** IP / CIDR 格式（IPv4 段 0-255、掩码 /0-/32；兼容 IPv6 及 /0-/128） */
const IPV4_CIDR_PATTERN =
  /^(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.){3}(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\/(?:3[0-2]|[12]?\d))?$/;
const IPV6_CIDR_PATTERN =
  /^(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}(?:\/(?:12[0-8]|1[01]\d|[0-9]?\d))?$/;

/** IP 组导入：可选组列表与下拉选项（选中后展开合并至目标名单并去重） */
const ipGroups = ref<OrpIpGroup[]>([]);
const ipGroupOptions = computed(() =>
  ipGroups.value.map((group) => ({
    label: `${group.name}（${group.members.length} 项）`,
    value: group.id,
  })),
);

async function loadIpGroups() {
  try {
    const res = await listIpGroupsApi({ page: 1, pageSize: 100 });
    ipGroups.value = res.items;
  } catch {
    // IP 组加载失败不阻塞表单主流程
  }
}

function importFromGroup(groupId: number, target: 'allow' | 'deny') {
  const group = ipGroups.value.find((item) => item.id === groupId);
  if (!group) return;
  const current = target === 'allow' ? ipPolicy.allowList : ipPolicy.denyList;
  const merged = new Set(current);
  let added = 0;
  for (const member of group.members) {
    if (!merged.has(member)) {
      merged.add(member);
      added += 1;
    }
  }
  if (target === 'allow') {
    ipPolicy.allowList = [...merged];
  } else {
    ipPolicy.denyList = [...merged];
  }
  message.success(`已从 IP 组「${group.name}」导入 ${added} 项（自动去重）`);
}

/** 重置 IP 策略为默认关闭态 */
function resetIpPolicy() {
  ipPolicy.enabled = false;
  ipPolicy.priority = 'allow-first';
  ipPolicy.allowList = [];
  ipPolicy.denyList = [];
}

/** 校验 IP 策略列表格式，返回错误信息（null 表示通过） */
function validateIpPolicy(): null | string {
  if (!ipPolicy.enabled) return null;
  const checks: Array<[string[], string]> = [
    [ipPolicy.allowList, '白名单'],
    [ipPolicy.denyList, '黑名单'],
  ];
  for (const [list, name] of checks) {
    for (const item of list) {
      if (!IPV4_CIDR_PATTERN.test(item) && !IPV6_CIDR_PATTERN.test(item)) {
        return `IP ${name}中「${item}」不是合法的 IP 地址或 CIDR 网段（示例：192.168.1.10 / 10.0.0.0/8）`;
      }
    }
  }
  return null;
}

/** 常用 server {} 上下文指令预设（名称 · 说明） */
const SERVER_DIRECTIVE_OPTIONS = [
  {
    label: 'client_max_body_size · 请求体上限（如 10m / 100m）',
    value: 'client_max_body_size',
  },
  {
    label: 'keepalive_timeout · 长连接超时（如 65 / 65s）',
    value: 'keepalive_timeout',
  },
  { label: 'server_tokens · 响应头版本号隐藏（on / off）', value: 'server_tokens' },
  { label: 'ssl_protocols · TLS 协议版本（如 TLSv1.2 TLSv1.3）', value: 'ssl_protocols' },
  {
    label: 'ssl_ciphers · TLS 加密套件（如 HIGH:!aNULL:!MD5）',
    value: 'ssl_ciphers',
  },
  {
    label: 'ssl_session_cache · 会话缓存（如 shared:SSL:10m）',
    value: 'ssl_session_cache',
  },
  {
    label: 'ssl_session_timeout · 会话超时（如 10m）',
    value: 'ssl_session_timeout',
  },
  { label: 'access_log · 访问日志（如 logs/api.access.log main）', value: 'access_log' },
  { label: 'error_log · 错误日志（如 logs/api.error.log warn）', value: 'error_log' },
  { label: 'root · 站点根目录（如 /data/www/api）', value: 'root' },
  { label: 'index · 默认首页（如 index.html）', value: 'index' },
];

/** 指令名格式（nginx 命名规范） */
const NAME_PATTERN = /^[a-z_][a-z0-9_]*$/;
const BOOLEAN_PATTERN = /^(on|off)$/;
const TIME_PATTERN = /^\d{1,4}(ms|s|m|h)?$/;
const SIZE_PATTERN = /^\d{1,7}(k|m|g)?$/;

/** 特定 Server 指令取值校验规则 */
const SERVER_VALUE_RULES: Record<string, { pattern: RegExp; tip: string }> = {
  server_tokens: { pattern: BOOLEAN_PATTERN, tip: '取值应为 on 或 off' },
  keepalive_timeout: {
    pattern: TIME_PATTERN,
    tip: '取值应为时间格式（如 65、65s、2m）',
  },
  client_max_body_size: {
    pattern: SIZE_PATTERN,
    tip: '取值应为容量格式（如 10m、100m）',
  },
  ssl_session_timeout: {
    pattern: TIME_PATTERN,
    tip: '取值应为时间格式（如 10m）',
  },
  ssl_session_cache: {
    pattern: /^(off|none|builtin[:\d]*|shared:\S+:\d+[kmg]?)$/,
    tip: '取值应为会话缓存格式（如 shared:SSL:10m）',
  },
  ssl_protocols: {
    pattern: /^(SSLv2|SSLv3|TLSv1(\.[0-3])?)(\s+(SSLv2|SSLv3|TLSv1(\.[0-3])?))*$/,
    tip: '取值应为 TLS 协议列表（如 TLSv1.2 TLSv1.3）',
  },
};

function addDirective() {
  editDirectives.value.push({ name: '', value: '' });
}

function removeDirective(index: number) {
  editDirectives.value.splice(index, 1);
}

function filterOption(input: string, option: { label: string }) {
  return option.label.toLowerCase().includes(input.toLowerCase());
}

/** 校验 Server 块指令列表，返回错误信息（null 表示通过） */
function validateDirectives(): null | string {
  const seen = new Set<string>();
  for (const [index, item] of editDirectives.value.entries()) {
    const name = item.name.trim();
    const value = item.value.trim();
    if (!name || !value) {
      return `第 ${index + 1} 条 Server 指令的指令名与取值均为必填项`;
    }
    if (!NAME_PATTERN.test(name)) {
      return `第 ${index + 1} 条指令名「${name}」不合法：仅支持小写字母、数字与下划线`;
    }
    if (seen.has(name)) {
      return `指令 ${name} 重复定义，请合并或删除重复项`;
    }
    seen.add(name);
    const rule = SERVER_VALUE_RULES[name];
    if (rule && !rule.pattern.test(value)) {
      return `指令 ${name} ${rule.tip}`;
    }
  }
  return null;
}

const { markDirty, onBeforeClose, resetDirty } = useDirtyGuard();

const [Form, formApi] = useVbenForm({
  handleValuesChange: () => markDirty(),
  handleSubmit: onSubmit,
  schema: [
    {
      component: 'ApiSelect',
      componentProps: {
        allowClear: false,
        api: async () => {
          const res = await listCentersApi({ page: 1, pageSize: 200 });
          return res.items.map((item) => ({ label: item.name, value: item.id }));
        },
        placeholder: '请选择所属中心',
      },
      fieldName: 'centerId',
      label: '所属中心',
      rules: z.number({ message: '请选择所属中心' }).min(1, { message: '请选择所属中心' }),
    },
    {
      component: 'Input',
      componentProps: { placeholder: '例如：api.example.cn 或 *.example.cn' },
      fieldName: 'domain',
      label: '绑定域名',
      rules: z
        .string()
        .min(1, { message: '请输入绑定域名' })
        .regex(/^(\*\.)?([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/, {
          message: '域名格式不正确',
        }),
    },
    {
      component: 'InputNumber',
      componentProps: { max: 65_535, min: 1, placeholder: '例如：8443', precision: 0 },
      fieldName: 'port',
      label: '监听端口',
      rules: z
        .number({ message: '请输入监听端口' })
        .int()
        .min(1, { message: '端口范围 1-65535' })
        .max(65_535, { message: '端口范围 1-65535' }),
    },
    {
      component: 'ApiSelect',
      componentProps: {
        allowClear: true,
        showSearch: true,

        optionFilterProp: 'label',
        api: async () => {
          const { listCertificatesApi } = await import('#/api');
          const res = await listCertificatesApi({ page: 1, pageSize: 200 });
          return res.items.map((item) => ({ label: item.name, value: item.id }));
        },
        placeholder: '不绑定则为明文 HTTP',
      },
      fieldName: 'tlsCertId',
      label: 'TLS 证书',
    },
  ],
  showDefaultActions: false,
});

watch(editDirectives, () => markDirty(), { deep: true });
watch(ipPolicy, () => markDirty(), { deep: true });
watch(errorPages, () => markDirty(), { deep: true });
watch([errorStatus, errorContentType, errorContent], () => markDirty());

const [Modal, modalApi] = useVbenModal<{ record?: HttpListener }>({
  fullscreenButton: false,
  onBeforeClose,
  async onConfirm() {
    await formApi.validateAndSubmitForm();
  },
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      await formApi.resetForm();
      loadIpGroups();
      const record = modalApi.getData()?.record;
      isEdit.value = Boolean(record);
      editId.value = record?.id ?? 0;
      modalApi.setState({ title: record ? '编辑监听配置' : '新建监听配置' });
      if (record) {
        formApi.setValues({
          centerId: record.centerId,
          domain: record.domain,
          port: record.port,
          tlsCertId: record.tlsCertId ?? undefined,
        });
        editDirectives.value = (record.directives ?? []).map((item) => ({
          ...item,
        }));
        errorPages.value = { ...(record.errorPages ?? {}) };
        const firstErrorKey = Object.keys(errorPages.value).sort()[0];
        if (firstErrorKey) selectErrorPage(firstErrorKey);
        else {
          activeErrorKey.value = '';
          errorStatus.value = '404';
          errorContentType.value = '';
          errorContent.value = '';
        }
        // IP 策略回显（深拷贝，避免直接引用列表数据）
        if (record.ipPolicy) {
          ipPolicy.enabled = record.ipPolicy.enabled;
          ipPolicy.priority = record.ipPolicy.priority;
          ipPolicy.allowList = [...record.ipPolicy.allowList];
          ipPolicy.denyList = [...record.ipPolicy.denyList];
        } else {
          resetIpPolicy();
        }
      } else {
        editDirectives.value = [];
        errorPages.value = {};
        activeErrorKey.value = '';
        errorStatus.value = '404';
        errorContentType.value = '';
        errorContent.value = '';
        formApi.setValues({ centerId: undefined });
        resetIpPolicy();
      }
      resetDirty();
    }
  },
});

async function onSubmit(values: Record<string, any>) {
  if (hasUnsavedErrorPage() && !saveErrorPage()) return;
  const directiveError = validateDirectives();
  if (directiveError) {
    message.error(directiveError);
    return;
  }
  const ipError = validateIpPolicy();
  if (ipError) {
    message.error(ipError);
    return;
  }
  modalApi.lock();
  try {
    const payload = {
      centerId: Number(values.centerId),
      directives: editDirectives.value.map(
        (item) => ({ name: item.name.trim(), value: item.value.trim() }),
      ) as OrpDirective[],
      errorPages: { ...errorPages.value },
      domain: String(values.domain),
      ipPolicy: {
        enabled: ipPolicy.enabled,
        priority: ipPolicy.priority,
        allowList: [...ipPolicy.allowList],
        denyList: [...ipPolicy.denyList],
      },
      port: Number(values.port),
      tlsCertId: values.tlsCertId ? Number(values.tlsCertId) : null,
    };
    if (isEdit.value) {
      await updateHttpListenerApi(editId.value, payload);
      message.success('监听配置已更新');
    } else {
      await createHttpListenerApi(payload);
      message.success('监听配置创建成功');
    }
    resetDirty();
    modalApi.close();
    emit('success');
  } catch {
    // 错误提示由请求拦截器统一处理
  } finally {
    modalApi.unlock();
  }
}
</script>

<template>
  <Modal class="w-[680px]">
    <div class="pr-2">
      <Form />
    </div>

    <!-- Server 块指令（server {} 上下文） -->
    <div class="mt-4 px-1">
      <div class="mb-2 flex items-center justify-between">
        <Tooltip
          title="server {} 块内指令（如 client_max_body_size、ssl_protocols），仅对当前监听的 Server 生效。listen / server_name 已由上方表单托管，此处无需重复配置。"
        >
          <span class="text-sm font-medium">Server 块指令</span>
        </Tooltip>
        <Button size="small" type="dashed" @click="addDirective">+ 添加指令</Button>
      </div>
      <div class="max-h-60 space-y-2 overflow-y-auto pr-1">
        <div
          v-for="(directive, index) in editDirectives"
          :key="index"
          class="flex flex-wrap items-center gap-2 rounded border border-gray-200 bg-gray-50 p-2 dark:border-gray-700 dark:bg-gray-800"
        >
          <Tag class="shrink-0" color="blue">#{{ index + 1 }}</Tag>
          <AutoComplete
            v-model:value="directive.name"
            :filter-option="filterOption"
            :options="SERVER_DIRECTIVE_OPTIONS"
            class="!w-60"
            placeholder="指令名，如 client_max_body_size"
            size="small"
          />
          <Input
            v-model:value="directive.value"
            class="!w-56"
            placeholder="取值，如 20m / TLSv1.2 TLSv1.3"
            size="small"
          />
          <Button danger size="small" type="link" @click="removeDirective(index)">
            删除
          </Button>
        </div>
        <div
          v-if="editDirectives.length === 0"
          class="py-4 text-center text-xs text-gray-400"
        >
          暂无 Server 块指令，点击「+ 添加指令」开始配置
        </div>
      </div>
    </div>

    <div class="mt-4 px-1">
      <div class="mb-2 text-sm font-medium">Server 错误页面覆盖</div>
      <div class="mb-2 flex flex-wrap gap-1">
        <Tag
          v-for="key in Object.keys(errorPages).sort()"
          :key="key"
          class="cursor-pointer"
          :color="activeErrorKey === key ? 'blue' : 'default'"
          @click="selectErrorPage(key)"
        >
          {{ key }}
        </Tag>
        <span v-if="Object.keys(errorPages).length === 0" class="text-xs text-gray-400">
          未配置的状态码继承 HTTP 块默认页面。
        </span>
      </div>
      <div class="grid grid-cols-2 gap-2">
        <Input v-model:value="errorStatus" placeholder="状态码，如 404" />
        <Input v-model:value="errorContentType" placeholder="可选 Content-Type，如 application/json" />
      </div>
      <Input.TextArea
        v-model:value="errorContent"
        class="!mt-2"
        :rows="4"
        placeholder="错误页面内容（最多 1 MB）"
      />
      <div class="mt-2 flex justify-end gap-2">
        <Button v-if="activeErrorKey" danger size="small" @click="deleteErrorPage">删除当前页面</Button>
        <Button size="small" type="primary" @click="saveErrorPage">加入站点覆盖</Button>
      </div>
    </div>

    <!-- IP 访问控制策略（server {} 块级 allow/deny） -->
    <div class="mt-4 px-1">
      <div class="mb-2 flex items-center justify-between">
        <Tooltip
          title="server {} 块级 IP 访问控制（allow/deny 指令组）。优先模式决定生成顺序：白名单优先 = allow 规则在前，命中白名单直接放行；黑名单优先 = deny 规则在前，命中黑名单直接拒绝。关闭时不生成任何 allow/deny 指令。"
        >
          <span class="text-sm font-medium">IP 访问控制策略</span>
        </Tooltip>
        <div class="flex items-center gap-2">
          <span class="text-xs text-gray-500">启用</span>
          <Switch
            v-model:checked="ipPolicy.enabled"
            checked-children="开"
            size="small"
            un-checked-children="关"
          />
        </div>
      </div>
      <template v-if="ipPolicy.enabled">
        <div class="mb-2 flex flex-wrap items-center gap-2">
          <span class="w-24 shrink-0 text-xs text-gray-500">优先模式</span>
          <RadioGroup
            v-model:value="ipPolicy.priority"
            :options="IP_PRIORITY_OPTIONS"
            option-type="button"
            size="small"
          />
        </div>
        <div class="mb-2 flex items-center gap-2">
          <span class="w-24 shrink-0 text-xs text-gray-500">白名单 allow</span>
          <Select
            v-model:value="ipPolicy.allowList"
            :open="false"
            class="min-w-0 flex-1"
            mode="tags"
            placeholder="输入 IP 或 CIDR 网段后回车，如 192.168.1.10 / 10.0.0.0/8"
            size="small"
          />
          <Select
            :options="ipGroupOptions"
            :value="undefined"
            allow-clear
            class="w-40 shrink-0"
            option-filter-prop="label"
            placeholder="从 IP 组导入"
            show-search
            size="small"
            @select="(value: any) => importFromGroup(Number(value), 'allow')"
          />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-24 shrink-0 text-xs text-gray-500">黑名单 deny</span>
          <Select
            v-model:value="ipPolicy.denyList"
            :open="false"
            class="min-w-0 flex-1"
            mode="tags"
            placeholder="输入 IP 或 CIDR 网段后回车，如 172.16.0.0/16 / 192.168.2.100"
            size="small"
          />
          <Select
            :options="ipGroupOptions"
            :value="undefined"
            allow-clear
            class="w-40 shrink-0"
            option-filter-prop="label"
            placeholder="从 IP 组导入"
            show-search
            size="small"
            @select="(value: any) => importFromGroup(Number(value), 'deny')"
          />
        </div>
      </template>
      <div v-else class="py-1 text-center text-xs text-gray-400">
        未启用（不生成 allow / deny 指令）
      </div>
    </div>
  </Modal>
</template>
