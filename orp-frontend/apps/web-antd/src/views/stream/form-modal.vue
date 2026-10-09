<script lang="ts" setup>
import type {
  OrpDirective,
  OrpIpGroup,
  OrpIpPolicy,
  StreamService,
  UpstreamGroup,
} from '#/api';

import { computed, reactive, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  AutoComplete,
  Button,
  Input,
  InputNumber,
  message,
  Popconfirm,
  RadioGroup,
  Select,
  Switch,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import {
  createStreamServiceApi,
  listCentersApi,
  listIpGroupsApi,
  updateStreamServiceApi,
} from '#/api';
import { useDirtyGuard } from '#/composables/use-dirty-guard';

defineOptions({ name: 'StreamFormModal' });

const emit = defineEmits<{ success: [] }>();

const isEdit = ref(false);
const editId = ref(0);

/** 已录入的 Upstream 组选项（可搜索下拉） */
const upstreamOptions = ref<{ label: string; value: string }[]>([]);

/** 受控后端行：组 + 端口 */
interface BackendRow {
  port: number | undefined;
  upstream?: string;
}

const backendRows = ref<BackendRow[]>([]);

/** Server 块指令编辑行 */
interface EditableDirective {
  name: string;
  value: string;
}

const editDirectives = ref<EditableDirective[]>([]);

/** IP 访问控制策略状态（stream server {} 块级 allow/deny 指令组） */
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

/** 常用 stream server {} 上下文指令预设（名称 · 说明） */
const STREAM_SERVER_DIRECTIVE_OPTIONS = [
  {
    label: 'proxy_connect_timeout · 上游连接超时（如 5s）',
    value: 'proxy_connect_timeout',
  },
  {
    label: 'proxy_timeout · 代理会话超时（如 600s）',
    value: 'proxy_timeout',
  },
  {
    label: 'proxy_protocol · PROXY 协议透传（on / off）',
    value: 'proxy_protocol',
  },
  {
    label: 'proxy_buffer_size · 代理缓冲区（如 16k）',
    value: 'proxy_buffer_size',
  },
  {
    label: 'proxy_next_upstream · 故障转移（on / off）',
    value: 'proxy_next_upstream',
  },
  {
    label: 'proxy_next_upstream_tries · 重试次数（如 2）',
    value: 'proxy_next_upstream_tries',
  },
  {
    label: 'proxy_next_upstream_timeout · 重试窗口（如 10s）',
    value: 'proxy_next_upstream_timeout',
  },
  {
    label: 'error_log · 错误日志（如 logs/stream.err warn）',
    value: 'error_log',
  },
  {
    label: 'resolver_timeout · DNS 解析超时（如 5s）',
    value: 'resolver_timeout',
  },
];

/** 指令名格式（nginx 命名规范） */
const NAME_PATTERN = /^[a-z_][a-z0-9_]*$/;
const TIME_PATTERN = /^\d{1,4}(ms|s|m|h)?$/;
const SIZE_PATTERN = /^\d{1,7}(k|m|g)?$/;

/** 特定 Stream Server 指令取值校验规则 */
const STREAM_VALUE_RULES: Record<string, { pattern: RegExp; tip: string }> = {
  proxy_protocol: { pattern: /^(on|off)$/, tip: '取值应为 on 或 off' },
  proxy_next_upstream: { pattern: /^(on|off)$/, tip: '取值应为 on 或 off' },
  proxy_connect_timeout: {
    pattern: TIME_PATTERN,
    tip: '取值应为时间格式（如 5s）',
  },
  proxy_timeout: { pattern: TIME_PATTERN, tip: '取值应为时间格式（如 600s）' },
  proxy_next_upstream_timeout: {
    pattern: TIME_PATTERN,
    tip: '取值应为时间格式（如 10s）',
  },
  proxy_buffer_size: {
    pattern: SIZE_PATTERN,
    tip: '取值应为容量格式（如 16k）',
  },
  proxy_next_upstream_tries: {
    pattern: /^\d{1,3}$/,
    tip: '取值应为正整数（如 2）',
  },
  resolver_timeout: {
    pattern: TIME_PATTERN,
    tip: '取值应为时间格式（如 5s）',
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
    const rule = STREAM_VALUE_RULES[name];
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
        showSearch: true,
        optionFilterProp: 'label',
      },
      fieldName: 'centerId',
      label: '所属中心',
      rules: z.number({ message: '请选择所属中心' }).min(1, { message: '请选择所属中心' }),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        options: [
          { label: 'TCP', value: 'tcp' },
          { label: 'UDP', value: 'udp' },
        ],
      },
      defaultValue: 'tcp',
      fieldName: 'protocol',
      label: '监听协议',
      rules: z.enum(['tcp', 'udp'], { message: '请选择监听协议' }),
    },
    {
      component: 'Input',
      componentProps: { placeholder: '默认 0.0.0.0' },
      defaultValue: '0.0.0.0',
      fieldName: 'listenAddress',
      label: '监听地址',
      rules: z
        .string()
        .regex(/^(\d{1,3}\.){3}\d{1,3}$/, { message: '必须是 IPv4 地址格式' }),
    },
    {
      component: 'InputNumber',
      componentProps: { max: 65_535, min: 1, placeholder: '例如：3306', precision: 0 },
      fieldName: 'listenPort',
      label: '监听端口',
      rules: z
        .number({ message: '请输入监听端口' })
        .int()
        .min(1, { message: '端口范围 1-65535' })
        .max(65_535, { message: '端口范围 1-65535' }),
    },
    {
      component: 'Input',
      componentProps: { maxLength: 100, placeholder: '服务用途备注（可选）' },
      fieldName: 'description',
      label: '服务描述',
    },
  ],
  showDefaultActions: false,
});

watch(backendRows, () => markDirty(), { deep: true });
watch(editDirectives, () => markDirty(), { deep: true });
watch(ipPolicy, () => markDirty(), { deep: true });

const [Modal, modalApi] = useVbenModal<{ record?: StreamService }>({
  fullscreenButton: false,
  onBeforeClose,
  async onConfirm() {
    await formApi.validateAndSubmitForm();
  },
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      await formApi.resetForm();
      await loadUpstreams();
      loadIpGroups();
      const record = modalApi.getData()?.record;
      isEdit.value = Boolean(record);
      editId.value = record?.id ?? 0;
      modalApi.setState({ title: record ? '编辑 Stream 服务' : '新建 Stream 服务' });
      if (record) {
        formApi.setValues({
          centerId: record.centerId,
          description: record.description,
          listenAddress: record.listenAddress,
          listenPort: record.listenPort,
          protocol: record.protocol,
        });
        // backends 形如 "session-sticky:8080" → 组 + 端口行
        backendRows.value = record.backends.map((line) => {
          const [group, port] = line.split(':');
          return { port: Number(port), upstream: group ?? undefined };
        });
        editDirectives.value = (record.directives ?? []).map((item) => ({
          ...item,
        }));
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
        backendRows.value = [{ port: undefined, upstream: undefined }];
        formApi.setValues({ centerId: undefined });
        editDirectives.value = [];
        resetIpPolicy();
      }
      resetDirty();
    }
  },
});

async function loadUpstreams() {
  try {
    const { listUpstreamGroupsApi } = await import('#/api');
    const res = await listUpstreamGroupsApi({ page: 1, pageSize: 200 });
    upstreamOptions.value = res.items.map((g: UpstreamGroup) => ({
      label: `${g.name}（${g.nodes.length} 节点 · ${g.lbPolicy}）`,
      value: g.name,
    }));
  } catch {
    upstreamOptions.value = [];
  }
}

function addRow() {
  backendRows.value.push({ port: undefined, upstream: undefined });
}

function removeRow(index: number) {
  backendRows.value.splice(index, 1);
}

async function onSubmit(values: Record<string, any>) {
  // 组装 backends：group:port
  const rows = backendRows.value.filter((row) => row.upstream);
  if (rows.length === 0) {
    message.warning('请至少选择一个 Upstream 组作为后端');
    return;
  }
  const invalidPort = rows.find(
    (row) => !row.port || row.port < 1 || row.port > 65_535,
  );
  if (invalidPort) {
    message.warning(
      `后端端口无效：${invalidPort.upstream}（端口范围 1-65535）`,
    );
    return;
  }
  const backends = rows.map((row) => `${row.upstream}:${row.port}`);
  if (new Set(backends).size !== backends.length) {
    message.warning('请勿重复添加相同的上游组和目标端口');
    return;
  }

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
      backends,
      centerId: Number(values.centerId),
      description: String(values.description ?? ''),
      directives: editDirectives.value.map(
        (item) => ({ name: item.name.trim(), value: item.value.trim() }),
      ) as OrpDirective[],
      ipPolicy: {
        enabled: ipPolicy.enabled,
        priority: ipPolicy.priority,
        allowList: [...ipPolicy.allowList],
        denyList: [...ipPolicy.denyList],
      },
      listenAddress: String(values.listenAddress ?? '0.0.0.0'),
      listenPort: Number(values.listenPort),
      protocol: values.protocol as 'tcp' | 'udp',
    };
    if (isEdit.value) {
      await updateStreamServiceApi(editId.value, payload);
      message.success('Stream 服务已更新');
    } else {
      await createStreamServiceApi(payload);
      message.success('Stream 服务创建成功');
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
    <Form />
    <div class="mt-2">
      <div class="mb-2 flex items-center justify-between">
        <span class="text-sm font-medium">后端上游组与目标端口</span>
        <Button size="small" @click="addRow">添加后端</Button>
      </div>
      <div class="mb-2 text-xs text-gray-500">
        发布时会展开所选组的节点，并以此处端口作为目标端口。多组组合采用轮询；组内若配置其他负载策略或原生指令，预检会拒绝。
      </div>
      <div
        v-for="(row, index) in backendRows"
        :key="index"
        class="mb-2 flex items-center gap-2"
      >
        <Select
          v-model:value="row.upstream"
          :options="upstreamOptions"
          :filter-option="(input, option) => String(option?.value ?? '').toLowerCase().includes(input.toLowerCase())"
          class="flex-1"
          placeholder="选择 Upstream 组"
          show-search
        />
        <InputNumber
          v-model:value="row.port"
          :max="65_535"
          :min="1"
          class="w-32"
          placeholder="端口"
          :precision="0"
        />
        <Popconfirm title="确认移除该后端？" @confirm="removeRow(index)">
          <Button danger size="small" type="link">移除</Button>
        </Popconfirm>
      </div>
      <div v-if="upstreamOptions.length === 0" class="text-xs text-orange-500">
        当前尚未录入任何 Upstream 组，请先前往「上游服务器组」模块创建。
      </div>
    </div>

    <!-- Server 块指令（stream server {} 上下文） -->
    <div class="mt-2 px-1">
      <div class="mb-2 flex items-center justify-between">
        <Tooltip
          title="stream server {} 块内指令（如 proxy_connect_timeout、proxy_timeout），仅对当前监听服务生效。listen 与 proxy_pass 已由上方表单托管，此处无需重复配置。"
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
            :options="STREAM_SERVER_DIRECTIVE_OPTIONS"
            class="!w-60"
            placeholder="指令名，如 proxy_connect_timeout"
            size="small"
          />
          <Input
            v-model:value="directive.value"
            class="!w-56"
            placeholder="取值，如 5s / 600s / on"
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

    <!-- IP 访问控制策略（stream server {} 块级 allow/deny） -->
    <div class="mt-2 px-1">
      <div class="mb-2 flex items-center justify-between">
        <Tooltip
          title="stream server {} 块级 IP 访问控制（allow/deny 指令组，作用于四层 TCP/UDP 代理会话）。优先模式决定生成顺序：白名单优先 = allow 规则在前，命中白名单直接放行；黑名单优先 = deny 规则在前，命中黑名单直接拒绝。关闭时不生成任何 allow/deny 指令。"
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
