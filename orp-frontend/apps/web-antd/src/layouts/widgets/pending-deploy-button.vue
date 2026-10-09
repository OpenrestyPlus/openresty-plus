<script lang="ts" setup>
import { computed } from 'vue';
import { useRouter } from 'vue-router';

import { createIconifyIcon } from '@vben/icons';

import { Badge, Tooltip } from 'ant-design-vue';

import { usePublishStore } from '#/store';

const GitCompareIcon = createIconifyIcon('lucide:git-compare-arrows');

const router = useRouter();
const publishStore = usePublishStore();

const tooltipTitle = computed(() =>
  publishStore.pendingCount > 0
    ? `${publishStore.pendingCount} 个实例配置变更待下发，点击进入配置比对`
    : '配置比对与下发',
);

function goPublish() {
  router.push('/config-publish');
}
</script>

<template>
  <Tooltip :title="tooltipTitle" placement="bottom">
    <Badge :count="publishStore.pendingCount" :offset="[-4, 4]" :overflow-count="99">
      <button
        aria-label="配置比对与下发"
        class="hover:bg-accent hover:text-foreground flex size-8 cursor-pointer items-center justify-center rounded-md bg-transparent text-sm"
        type="button"
        @click="goPublish"
      >
        <GitCompareIcon class="size-[18px]" />
      </button>
    </Badge>
  </Tooltip>
</template>
