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
  getStreamDirectivesApi,
  getStreamLogFormatApi,
  updateStreamDirectivesApi,
  updateStreamLogFormatApi,
} from '#/api';
import { useDirtyGuard } from '#/composables/use-dirty-guard';

defineOptions({ name: 'StreamDirectivesModal' });

const emit = defineEmits<{ success: [] }>();

interface EditableDirective {
  name: string;
  value: string;
}

const { markDirty, onBeforeClose, resetDirty } = useDirtyGuard();

const editDirectives = ref<EditableDirective[]>([]);

/** stream {} 上下文 log_format 配置 */
const logFormatName = ref('');
const logFormatValue = ref('');
const savingLogFormat = ref(false);

/** 常用 stream {} 上下文全局指令预设（名称 · 说明） */
const STREAM_DIRECTIVE_OPTIONS = [
  { label: 'tcp_nodelay · 禁用 Nagle 算法（on / off）', value: 'tcp_nodelay' },
  {
    label: 'proxy_buffer_size · 代理缓冲区（如 16k）',
    value: 'proxy_buffer_size',
  },
  {
    label: 'resolver · DNS 解析器（如 10.60.0.11 valid=30s）',
    value: 'resolver',
  },
  {
    label: 'resolver_timeout · DNS 解析超时（如 5s）',
    value: 'resolver_timeout',
  },
  {
    label: 'proxy_connect_timeout · 上游连接超时（如 5s）',
    value: 'proxy_connect_timeout',
  },
  {
    label: 'proxy_timeout · 代理会话超时（如 600s）',
    value: 'proxy_timeout',
  },
  {
    label: 'error_log · 错误日志（如 logs/stream.err warn）',
    value: 'error_log',
  },
  {
    label: 'log_format · 四层日志格式（在下方区块配置）',
    value: 'log_format',
  },
];

/** 指令名格式（nginx 命名规范） */
const NAME_PATTERN = /^[a-z_][a-z0-9_]*$/;
const TIME_PATTERN = /^\d{1,4}(ms|s|m|h)?$/;
const SIZE_PATTERN = /^\d{1,7}(k|m|g)?$/;

const STREAM_VALUE_RULES: Record<string, { pattern: RegExp; tip: string }> = {
  tcp_nodelay: { pattern: /^(on|off)$/, tip: '取值应为 on 或 off' },
  proxy_buffer_size: {
    pattern: SIZE_PATTERN,
    tip: '取值应为容量格式（如 16k）',
  },
  resolver_timeout: {
    pattern: TIME_PATTERN,
    tip: '取值应为时间格式（如 5s）',
  },
  proxy_connect_timeout: {
    pattern: TIME_PATTERN,
    tip: '取值应为时间格式（如 5s）',
  },
  proxy_timeout: { pattern: TIME_PATTERN, tip: '取值应为时间格式（如 600s）' },
};

/** log_format 变量快查（stream 上下文常用） */
const LOG_VARIABLES = [
  '$remote_addr',
  '$remote_port',
  '$local_addr',
  '$local_port',
  '$protocol',
  '$status',
  '$bytes_sent',
  '$bytes_received',
  '$session_time',
  '$time_local',
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
    if (name === 'log_format') {
      return 'log_format 请在下方「日志格式」区块单独配置';
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
    await updateStreamLogFormatApi({ format, name });
    message.success('Stream 块日志格式已保存');
  } catch {
    // 错误提示由请求拦截器统一处理
  } finally {
    savingLogFormat.value = false;
  }
}

async function loadAll() {
  try {
    const res = await getStreamDirectivesApi();
    editDirectives.value = (res ?? [])
      .filter((item: OrpDirective) => item.name !== 'log_format')
      .map((item: OrpDirective) => ({ ...item }));
  } catch {
    editDirectives.value = [];
  }
  try {
    const res = await getStreamLogFormatApi();
    logFormatName.value = res?.name ?? '';
    logFormatValue.value = res?.format ?? '';
  } catch {
    logFormatName.value = '';
    logFormatValue.value = '';
  }
}

onMounted(loadAll);

watch(editDirectives, () => markDirty(), { deep: true });
watch([logFormatName, logFormatValue], () => markDirty());

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onBeforeClose,
  async onConfirm() {
    const directiveError = validateDirectives();
    if (directiveError) {
      message.error(directiveError);
      return;
    }
    modalApi.lock();
    try {
      await updateStreamDirectivesApi({
        directives: editDirectives.value.map((item) => ({
          name: item.name.trim(),
          value: item.value.trim(),
        })),
      });
      message.success('Stream 块全局指令已保存');
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
    modalApi.setState({ title: 'Stream 块配置（stream {} 上下文）' });
    resetDirty();
  },
});
</script>

<template>
  <Modal class="w-[680px]">
    <div
      class="mb-3 rounded border border-gray-200 bg-gray-50 p-3 text-xs leading-5 text-gray-500 dark:border-gray-700 dark:bg-gray-800"
    >
      维护 nginx.conf 中 stream {} 上下文内、server 块外的全局指令（如 tcp_nodelay、resolver）与 log_format 日志格式。指令随配置版本下发至各中心节点，半动态 Reload 生效。
    </div>
    <div class="mb-2 flex items-center justify-between">
      <span class="text-sm font-medium">指令列表（{{ editDirectives.length }} 项）</span>
      <Button size="small" type="dashed" @click="addDirective">+ 添加指令</Button>
    </div>
    <div class="max-h-72 space-y-2 overflow-y-auto pr-1">
      <div
        v-for="(directive, index) in editDirectives"
        :key="index"
        class="flex flex-wrap items-center gap-2 rounded border border-gray-200 bg-gray-50 p-2 dark:border-gray-700 dark:bg-gray-800"
      >
        <Tag class="shrink-0" color="blue">#{{ index + 1 }}</Tag>
        <AutoComplete
          v-model:value="directive.name"
          :filter-option="filterOption"
          :options="STREAM_DIRECTIVE_OPTIONS"
          class="!w-60"
          placeholder="指令名，如 tcp_nodelay"
          size="small"
        />
        <Input
          v-model:value="directive.value"
          class="!w-56"
          placeholder="取值，如 on / 16k / 5s"
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
          title="stream {} 上下文的 log_format 定义，供 server 块 access_log 引用。支持 stream 变量：$remote_addr、$protocol、$status、$bytes_sent、$bytes_received、$session_time 等。"
        >
          <span class="text-sm font-medium">四层访问日志格式（log_format）</span>
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
          placeholder="格式名称，如 tcp_main"
        />
        <Input.TextArea
          v-model:value="logFormatValue"
          :rows="3"
          placeholder='$remote_addr [$time_local] $protocol $status $bytes_sent $session_time'
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
  </Modal>
</template>
