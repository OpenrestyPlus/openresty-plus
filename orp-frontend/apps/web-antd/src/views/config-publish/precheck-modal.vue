<script lang="ts" setup>
import type { PublishPrecheckResult } from '#/api';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Button, Spin, Tag } from 'ant-design-vue';

import { precheckPublishApi } from '#/api';

defineOptions({ name: 'PublishPrecheckModal' });

const emit = defineEmits<{ confirmed: [nodeIds: number[]] }>();

const loading = ref(false);
const checking = ref(false);
const result = ref<PublishPrecheckResult | null>(null);
const nodeIds = ref<number[]>([]);

const failedItems = () => result.value?.items.filter((x) => !x.passed) ?? [];

async function runCheck(ids: number[]) {
  checking.value = true;
  result.value = null;
  try {
    result.value = await precheckPublishApi(ids);
  } catch {
    // 错误提示由请求拦截器统一处理，关闭后可重试
  } finally {
    checking.value = false;
  }
}

const [PrecheckModal, modalApi] = useVbenModal<{ nodeIds?: number[] }>({
  class: 'w-[640px]',
  footer: false,
  onOpenChange(isOpen) {
    if (isOpen) {
      nodeIds.value = modalApi.getData()?.nodeIds ?? [];
      result.value = null;
      runCheck(nodeIds.value);
    } else {
      loading.value = false;
    }
  },
  title: 'Nginx 配置语法预检',
});

/** 重新预检 */
function recheck() {
  runCheck(nodeIds.value);
}

/** 全部通过后通知父组件继续下发（由父组件在下发完成后关闭本弹窗） */
function onConfirm() {
  if (checking.value || !result.value?.passed) {
    return;
  }
  loading.value = true;
  emit('confirmed', nodeIds.value);
}
</script>

<template>
  <PrecheckModal>
    <div class="flex flex-col gap-3 py-1">
      <Alert
        v-if="result && result.passed"
        message="全部实例语法校验通过，可继续下发"
        show-icon
        type="success"
      />
      <Alert
        v-else-if="result"
        show-icon
        type="error"
      >
        <template #message>
          <span
            >{{ failedItems().length }} 个实例语法校验未通过，已阻断下发：{{ failedItems().map((x) => x.nodeName).join('、') }}</span
          >
        </template>
        <template #description>
          请先在对应模块修正配置错误（缺少分号、括号不匹配、指令值非法等），修正后重新发起预检。
        </template>
      </Alert>

      <div
        v-for="item in result?.items ?? []"
        :key="item.nodeId"
        class="rounded-md border border-gray-200 p-3 dark:border-gray-700"
      >
        <div class="flex flex-wrap items-center gap-2">
          <span class="font-medium">{{ item.nodeName }}</span>
          <Tag class="!m-0" color="geekblue">{{ item.centerName }}</Tag>
          <Tag class="!m-0" :color="item.passed ? 'green' : 'red'">
            {{ item.passed ? 'PASS · 语法正常' : 'FAILED · 语法错误' }}
          </Tag>
        </div>
        <div class="mt-1 font-mono text-xs text-gray-400 dark:text-gray-500">
          $ {{ item.command }}
        </div>
		<div v-if="item.candidateDigest" class="mt-2 text-xs text-gray-500 dark:text-gray-400">
		  候选摘要：<span class="font-mono">{{ item.candidateDigest }}</span>
		  · 文本差异 <span class="text-emerald-600">+{{ item.addCount ?? 0 }}</span>
		  / <span class="text-rose-600">-{{ item.delCount ?? 0 }}</span>
		</div>
		<div v-if="item.semanticDiff?.length" class="mt-2 max-h-24 overflow-y-auto text-xs text-gray-500 dark:text-gray-400">
		  <div v-for="(change, index) in item.semanticDiff" :key="index">
		    {{ change.action }} · {{ change.kind }} #{{ change.id }}
		    <span v-if="change.fields?.length">（{{ change.fields.join('、') }}）</span>
		  </div>
		</div>
        <pre
          class="mt-2 max-h-32 overflow-y-auto rounded bg-gray-50 p-2 font-mono text-xs whitespace-pre-wrap dark:bg-gray-900"
          :class="
            item.passed
              ? 'text-gray-500 dark:text-gray-400'
              : 'text-rose-600 dark:text-rose-400'
          "
          >{{ item.output }}</pre
        >
      </div>

      <div
        v-if="checking"
        class="flex items-center justify-center gap-2 py-8 text-sm text-gray-400"
      >
        <Spin size="small" />
        正在目标节点的候选目录执行 nginx -t…
      </div>

      <div class="flex justify-end gap-2 border-t border-gray-100 pt-3 dark:border-gray-800">
        <Button :disabled="checking" @click="modalApi.close()">取消</Button>
        <Button v-if="result && !result.passed" :loading="checking" @click="recheck">
          重新预检
        </Button>
        <Button
          :disabled="!result?.passed"
          :loading="loading || checking"
          danger
          type="primary"
          @click="onConfirm"
        >
          确认无误，继续下发
        </Button>
      </div>
    </div>
  </PrecheckModal>
</template>
