<script lang="ts" setup>
import type { OrpDirective } from '#/api';

import { onMounted, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  AutoComplete,
  Button,
  Divider,
  Input,
  message,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import {
  getHttpDirectivesApi,
  getHttpErrorPagesApi,
  getHttpLogFormatApi,
  updateHttpDirectivesApi,
  updateHttpErrorPagesApi,
  updateHttpLogFormatApi,
} from '#/api';
import { useDirtyGuard } from '#/composables/use-dirty-guard';

defineOptions({ name: 'HttpDirectivesModal' });

const emit = defineEmits<{ success: [] }>();

interface EditableDirective {
  name: string;
  value: string;
}

const { markDirty, onBeforeClose, resetDirty } = useDirtyGuard();

const editDirectives = ref<EditableDirective[]>([]);

/** http {} 上下文 log_format 配置 */
const logFormatName = ref('');
const logFormatValue = ref('');
const savingLogFormat = ref(false);
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

/** 常用 http {} 上下文全局指令预设（名称 · 说明） */
const HTTP_DIRECTIVE_OPTIONS = [
  { label: 'sendfile · 零拷贝发送文件（on / off）', value: 'sendfile' },
  { label: 'tcp_nopush · TCP_NOPUSH 优化（on / off）', value: 'tcp_nopush' },
  { label: 'tcp_nodelay · 禁用 Nagle 算法（on / off）', value: 'tcp_nodelay' },
  {
    label: 'keepalive_timeout · 长连接超时（如 65 / 65s）',
    value: 'keepalive_timeout',
  },
  {
    label: 'keepalive_requests · 单连接最大请求数（如 1000）',
    value: 'keepalive_requests',
  },
  { label: 'gzip · Gzip 压缩开关（on / off）', value: 'gzip' },
  { label: 'gzip_comp_level · 压缩级别 1-9（如 5）', value: 'gzip_comp_level' },
  {
    label: 'client_max_body_size · 请求体上限（如 10m / 100m）',
    value: 'client_max_body_size',
  },
  {
    label: 'server_names_hash_bucket_size · 域名哈希桶容量（如 128）',
    value: 'server_names_hash_bucket_size',
  },
  {
    label: 'types_hash_max_size · 类型哈希表上限（如 2048）',
    value: 'types_hash_max_size',
  },
  {
    label: 'default_type · 默认 MIME（application/octet-stream）',
    value: 'default_type',
  },
  {
    label: 'proxy_connect_timeout · 上游连接超时（如 60s）',
    value: 'proxy_connect_timeout',
  },
];

/** 指令名格式（nginx 命名规范） */
const NAME_PATTERN = /^[a-z_][a-z0-9_]*$/;
const BOOLEAN_PATTERN = /^(on|off)$/;
const TIME_PATTERN = /^\d{1,4}(ms|s|m|h)?$/;
const SIZE_PATTERN = /^\d{1,7}(k|m|g)?$/;

/** 特定 HTTP 指令取值校验规则 */
const HTTP_VALUE_RULES: Record<string, { pattern: RegExp; tip: string }> = {
  sendfile: { pattern: BOOLEAN_PATTERN, tip: '取值应为 on 或 off' },
  tcp_nopush: { pattern: BOOLEAN_PATTERN, tip: '取值应为 on 或 off' },
  tcp_nodelay: { pattern: BOOLEAN_PATTERN, tip: '取值应为 on 或 off' },
  gzip: { pattern: BOOLEAN_PATTERN, tip: '取值应为 on 或 off' },
  gzip_comp_level: { pattern: /^[1-9]$/, tip: '取值应为 1-9 的整数' },
  keepalive_timeout: {
    pattern: TIME_PATTERN,
    tip: '取值应为时间格式（如 65、65s、2m）',
  },
  keepalive_requests: {
    pattern: /^\d{1,6}$/,
    tip: '取值应为正整数（如 1000）',
  },
  client_max_body_size: {
    pattern: SIZE_PATTERN,
    tip: '取值应为容量格式（如 10m、100m）',
  },
  server_names_hash_bucket_size: {
    pattern: SIZE_PATTERN,
    tip: '取值应为容量格式（如 128）',
  },
  types_hash_max_size: {
    pattern: SIZE_PATTERN,
    tip: '取值应为容量格式（如 2048）',
  },
  proxy_connect_timeout: {
    pattern: TIME_PATTERN,
    tip: '取值应为时间格式（如 60s）',
  },
  default_type: {
    pattern: /^[a-z][a-z0-9/+.-]*$/,
    tip: '取值应为 MIME 类型（如 application/octet-stream）',
  },
};

/** log_format 变量快查（http 上下文常用） */
const LOG_VARIABLES = [
  '$remote_addr',
  '$remote_user',
  '$time_local',
  '$request',
  '$status',
  '$body_bytes_sent',
  '$http_referer',
  '$http_user_agent',
  '$http_x_forwarded_for',
  '$request_time',
  '$upstream_response_time',
  '$upstream_addr',
];

function addDirective() {
  editDirectives.value.push({ name: '', value: '' });
}

function removeDirective(index: number) {
  editDirectives.value.splice(index, 1);
}

function filterOption(input: string, option: { label: string }) {
  return option.label.toLowerCase().includes(input.toLowerCase());
}

/** 校验指令列表，返回错误信息（null 表示通过） */
function validateDirectives(): null | string {
  const seen = new Set<string>();
  for (const [index, item] of editDirectives.value.entries()) {
    const name = item.name.trim();
    const value = item.value.trim();
    if (!name || !value) {
      return `第 ${index + 1} 条指令的指令名与取值均为必填项`;
    }
    if (!NAME_PATTERN.test(name)) {
      return `第 ${index + 1} 条指令名「${name}」不合法：仅支持小写字母、数字与下划线`;
    }
    if (seen.has(name)) {
      return `指令 ${name} 重复定义，请合并或删除重复项`;
    }
    seen.add(name);
    if (name === 'log_format') {
      return 'log_format 请在下方「日志格式」区块单独配置';
    }
    const rule = HTTP_VALUE_RULES[name];
    if (rule && !rule.pattern.test(value)) {
      return `指令 ${name} ${rule.tip}`;
    }
  }
  return null;
}

async function saveLogFormat() {
  const name = logFormatName.value.trim();
  const format = logFormatValue.value.trim();
  if (!name || !format) {
    message.warning('日志格式名称与格式串均为必填项');
    return;
  }
  if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(name)) {
    message.warning('格式名称不合法：仅支持字母、数字与下划线，且以字母开头');
    return;
  }
  if (!format.includes('$')) {
    message.warning('格式串需至少包含一个 $ 变量（如 $remote_addr）');
    return;
  }
  savingLogFormat.value = true;
  try {
    await updateHttpLogFormatApi({ format, name });
    message.success('HTTP 块日志格式已保存');
  } catch {
    // 错误提示由请求拦截器统一处理
  } finally {
    savingLogFormat.value = false;
  }
}

async function loadAll() {
  try {
    const res = await getHttpDirectivesApi();
    editDirectives.value = (res ?? [])
      .filter((item: OrpDirective) => item.name !== 'log_format')
      .map((item: OrpDirective) => ({ ...item }));
  } catch {
    editDirectives.value = [];
  }
  try {
    const res = await getHttpLogFormatApi();
    logFormatName.value = res?.name ?? '';
    logFormatValue.value = res?.format ?? '';
  } catch {
    logFormatName.value = '';
    logFormatValue.value = '';
  }
  try {
    errorPages.value = await getHttpErrorPagesApi() ?? {};
    const first = Object.keys(errorPages.value).sort()[0];
    if (first) selectErrorPage(first);
    else {
      activeErrorKey.value = '';
      errorStatus.value = '404';
      errorContentType.value = '';
      errorContent.value = '';
    }
  } catch {
    errorPages.value = {};
  }
}

onMounted(loadAll);

watch(editDirectives, () => markDirty(), { deep: true });
watch([logFormatName, logFormatValue], () => markDirty());
watch(errorPages, () => markDirty(), { deep: true });
watch([errorStatus, errorContentType, errorContent], () => markDirty());

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onBeforeClose,
  async onConfirm() {
    const directiveError = validateDirectives();
    if (directiveError) {
      message.error(directiveError);
      return;
    }
    if (hasUnsavedErrorPage() && !saveErrorPage()) return;
    modalApi.lock();
    try {
      await updateHttpDirectivesApi({
        directives: editDirectives.value.map((item) => ({
          name: item.name.trim(),
          value: item.value.trim(),
        })),
      });
      await updateHttpErrorPagesApi({ errorPages: errorPages.value });
      message.success('HTTP 块全局指令已保存');
      resetDirty();
      modalApi.close();
      emit('success');
    } catch {
      // 错误提示由请求拦截器统一处理
    } finally {
      modalApi.unlock();
    }
  },
  async onOpenChange(isOpen: boolean) {
    if (!isOpen) {
      return;
    }
    await loadAll();
    modalApi.setState({ title: 'HTTP 块配置（http {} 上下文）' });
    resetDirty();
  },
});
</script>

<template>
  <Modal class="w-[680px]">
    <div
      class="mb-3 rounded border border-gray-200 bg-gray-50 p-3 text-xs leading-5 text-gray-500 dark:border-gray-700 dark:bg-gray-800"
    >
      维护 nginx.conf 中 http {} 上下文内、server 块外的全局指令（如 sendfile、gzip、keepalive_timeout）。指令随配置版本下发至各中心节点，半动态 Reload 生效。
    </div>
    <div class="mb-2 flex items-center justify-between">
      <span class="text-sm font-medium">指令列表（{{ editDirectives.length }} 项）</span>
      <Button size="small" type="dashed" @click="addDirective">+ 添加指令</Button>
    </div>
    <div class="max-h-96 space-y-2 overflow-y-auto pr-1">
      <div
        v-for="(directive, index) in editDirectives"
        :key="index"
        class="flex flex-wrap items-center gap-2 rounded border border-gray-200 bg-gray-50 p-2 dark:border-gray-700 dark:bg-gray-800"
      >
        <Tag class="shrink-0" color="blue">#{{ index + 1 }}</Tag>
        <AutoComplete
          v-model:value="directive.name"
          :filter-option="filterOption"
          :options="HTTP_DIRECTIVE_OPTIONS"
          class="!w-64"
          placeholder="指令名，如 keepalive_timeout"
          size="small"
        />
        <Input
          v-model:value="directive.value"
          class="!w-52"
          placeholder="取值，如 on / 65 / 10m"
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
        暂无指令，点击「+ 添加指令」开始配置
      </div>
    </div>

    <Divider class="!my-4">
      <span class="text-xs text-gray-400">log_format 日志格式</span>
    </Divider>

    <div class="rounded border border-gray-200 p-3 dark:border-gray-700">
      <div class="mb-2 flex items-center justify-between">
        <Tooltip
          title="http {} 上下文的 log_format 定义，供 server 块 access_log 引用。支持 http 变量：$remote_addr、$request、$status、$body_bytes_sent、$request_time 等。"
        >
          <span class="text-sm font-medium">访问日志格式（log_format）</span>
        </Tooltip>
        <Button
          :loading="savingLogFormat"
          size="small"
          type="primary"
          @click="saveLogFormat"
        >
          保存日志格式
        </Button>
      </div>
      <div class="flex flex-col gap-2">
        <Input
          v-model:value="logFormatName"
          class="w-64"
          placeholder="格式名称，如 main"
        />
        <Input.TextArea
          v-model:value="logFormatValue"
          :rows="3"
          placeholder='$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_user_agent" $request_time'
        />
        <div class="flex flex-wrap gap-1">
          <Tag
            v-for="v in LOG_VARIABLES"
            :key="v"
            class="cursor-pointer select-none"
            color="geekblue"
            @click="logFormatValue = `${logFormatValue} ${v}`.trim()"
          >
            {{ v }}
          </Tag>
        </div>
        <div class="text-xs text-gray-400">
          点击变量名可快速插入；配置后按「保存日志格式」即时生效。
        </div>
      </div>
    </div>

    <Divider class="!my-4">
      <span class="text-xs text-gray-400">HTTP 块默认错误页面</span>
    </Divider>
    <div class="rounded border border-gray-200 p-3 dark:border-gray-700">
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
          当前没有默认错误页；可在下方新增，未覆盖状态继续使用 OpenResty 默认行为。
        </span>
      </div>
      <div class="grid grid-cols-2 gap-2">
        <Input v-model:value="errorStatus" placeholder="状态码，如 404" />
        <Input v-model:value="errorContentType" placeholder="可选 Content-Type，如 application/json" />
      </div>
      <Input.TextArea
        v-model:value="errorContent"
        class="!mt-2"
        :rows="5"
        placeholder="错误页面内容（最多 1 MB）"
      />
      <div class="mt-2 flex justify-end gap-2">
        <Button v-if="activeErrorKey" danger size="small" @click="deleteErrorPage">删除当前页面</Button>
        <Button size="small" type="primary" @click="saveErrorPage">加入错误页配置</Button>
      </div>
    </div>
  </Modal>
</template>
