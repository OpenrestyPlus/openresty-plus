import { ref } from 'vue';

import { defineStore } from 'pinia';

import { getPublishSummaryApi } from '#/api';

/**
 * 配置下发全局状态：驱动顶栏待下发角标轮询
 */
export const usePublishStore = defineStore('publish', () => {
  /** 待下发实例总数（顶栏角标） */
  const pendingCount = ref(0);
  const loading = ref(false);

  /** 立即刷新待下发计数 */
  async function refresh() {
    if (loading.value) {
      return;
    }
    loading.value = true;
    try {
      const data = await getPublishSummaryApi();
      pendingCount.value = data.pendingCount;
    } catch {
      // 轮询静默失败，等待下一轮重试
    } finally {
      loading.value = false;
    }
  }

  return { pendingCount, refresh };
});
