<script lang="ts" setup>
import type { NodeItem } from '#/api';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';
import { ref } from 'vue';

import { updateNodeStatusApi } from '#/api';

import type { NodeStatus } from '#/api';

import { STATUS_META, STATUS_VALUES } from './status-meta';

defineOptions({ name: 'NodeStatusModal' });

const emit = defineEmits<{ success: [] }>();

const nodeName = ref('');
const nodeHost = ref('');
const currentStatus = ref<NodeStatus>('running');

/** 模板取状态元数据（索引收窄，杜绝 possibly undefined） */
const metaOf = (value: string | NodeStatus) =>
  (STATUS_META[value as NodeStatus] ?? STATUS_META.running)!;
const target = ref<string>('running');

const [Modal, modalApi] = useVbenModal<{ record?: NodeItem }>({
  fullscreenButton: false,
  async onConfirm() {
    const record = modalApi.getData()?.record;
    if (!record) return;
    // 目标状态与当前一致：直接关闭，不产生冗余变更与审计记录
    if (target.value === record.status) {
      modalApi.close();
      return;
    }
    modalApi.lock();
    try {
      await updateNodeStatusApi(record.id, target.value as never);
      message.success(
        `节点「${record.name}」运行状态已更新为「${(STATUS_META[target.value as NodeStatus] ?? { label: target.value }).label}」`,
      );
      modalApi.close();
      emit('success');
    } catch {
      // 错误提示由请求拦截器统一处理
    } finally {
      modalApi.unlock();
    }
  },
  onOpenChange(isOpen: boolean) {
    if (isOpen) {
      const record = modalApi.getData()?.record;
      nodeName.value = record?.name ?? '';
      nodeHost.value = record?.host ?? '';
      currentStatus.value = record?.status ?? 'running';
      target.value = record?.status ?? 'running';
      modalApi.setState({ title: '修改运行状态' });
    }
  },
});
</script>

<template>
  <Modal class="w-[420px]">
    <div class="text-sm">
      <!-- 当前节点与状态概览 -->
      <div
        class="flex items-center justify-between gap-3 rounded border border-gray-200 bg-gray-50 px-3 py-2 dark:border-gray-700 dark:bg-gray-800"
      >
        <div class="flex min-w-0 items-center gap-2">
          <span class="shrink-0 text-gray-500 dark:text-gray-400">节点</span>
          <span class="truncate font-medium">{{ nodeName }}</span>
          <span class="shrink-0 text-xs text-gray-400 dark:text-gray-500">
            {{ nodeHost }}
          </span>
        </div>
        <span
          :class="`inline-flex shrink-0 items-center gap-1 rounded border px-2 py-0.5 text-xs ${metaOf(currentStatus).pill}`"
        >
          <span
            :class="`h-1.5 w-1.5 rounded-full ${metaOf(currentStatus).dot}`"
          ></span>
          {{ metaOf(currentStatus).label }}
        </span>
      </div>

      <!-- 目标状态选择 -->
      <div class="mt-4 font-medium">目标状态</div>
      <div class="mt-2 grid grid-cols-2 gap-2">
        <button
          v-for="value in STATUS_VALUES"
          :key="value"
          data-status-option
          :data-value="value"
          type="button"
          :class="[
            'inline-flex items-center justify-center gap-1.5 rounded border px-3 py-2 text-sm transition-colors',
            target === value
              ? 'border-emerald-500 bg-emerald-50 font-medium text-emerald-700 ring-1 ring-emerald-500/30 dark:border-emerald-600 dark:bg-emerald-950 dark:text-emerald-300'
              : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:border-gray-600',
          ]"
          @click="target = value"
        >
          <span :class="`h-1.5 w-1.5 rounded-full ${metaOf(value).dot}`"></span>
          {{ metaOf(value).label }}
        </button>
      </div>

      <div class="mt-3 text-xs text-gray-400 dark:text-gray-500">
        确认后立即生效并记录审计日志；选择与当前一致的状态不会产生变更。
      </div>
    </div>
  </Modal>
</template>
