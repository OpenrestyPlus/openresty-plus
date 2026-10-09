<script lang="ts" setup>
import type { EChartsOption } from 'echarts';

import type { LatencyHistory, NodeItem } from '#/api';

import {
  Button,
  Descriptions,
  DescriptionsItem,
  Empty,
  Segmented,
  Select,
  Spin,
  Tag,
} from 'ant-design-vue';
import { computed, ref, shallowRef, watch } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { getNodeLatencyHistoryApi, listNodesApi } from '#/api';

import EChart from '../dashboard/echart.vue';

defineOptions({ name: 'NodeLatencyDrawer' });

/** 最多同时对比的节点数 */
const MAX_COMPARE = 3;

const loading = ref(false);
const histories = shallowRef<LatencyHistory[]>([]);
const selectedIds = ref<number[]>([]);
const limit = ref(50);
const nodeOptions = ref<{ label: string; value: number }[]>([]);

/** 健康状态标签配色 */
/** 兜底标签（在线），规避 Record<string> 索引 possibly undefined */
const HEALTH_TAG_ONLINE = { color: 'success', text: '在线' };
const HEALTH_TAG: Record<string, { color: string; text: string }> = {
  degraded: { color: 'warning', text: '不稳定' },
  offline: { color: 'error', text: '离线' },
  online: HEALTH_TAG_ONLINE,
};

const LIMIT_OPTIONS = [
  { label: '近 20 次', value: 20 },
  { label: '近 50 次', value: 50 },
  { label: '近 100 次', value: 100 },
];

/** 当前抽屉主体节点（首个选中项） */
const primary = computed(() => histories.value[0]);

async function loadOptions(centerId: number) {
  try {
    const res = await listNodesApi({ centerId, page: 1, pageSize: 100 });
    nodeOptions.value = res.items.map((item) => ({
      label: item.name,
      value: item.id,
    }));
  } catch {
    nodeOptions.value = [];
  }
}

async function loadHistories() {
  if (selectedIds.value.length === 0) {
    histories.value = [];
    return;
  }
  loading.value = true;
  try {
    const list = await Promise.all(
      selectedIds.value.map((id) => getNodeLatencyHistoryApi(id, limit.value)),
    );
    histories.value = list;
  } catch (err) {
    console.error('[latency] 采样数据加载失败', err);
    histories.value = [];
  } finally {
    loading.value = false;
  }
}

/** 延迟趋势折线图：横轴采样时间、纵轴 RTT（ms）；超时样本断线呈现 */
const chartOption = computed<EChartsOption>(() => ({
  dataZoom: [
    { type: 'inside' },
    { bottom: 6, height: 16, type: 'slider' },
  ],
  grid: { bottom: 52, containLabel: true, left: 12, right: 16, top: 40 },
  legend: { top: 8, type: 'scroll' },
  series: histories.value.map((h) => ({
    areaStyle: histories.value.length === 1 ? { opacity: 0.08 } : undefined,
    connectNulls: false,
    data: h.samples.map((s) => [s.ts, s.rtt]),
    lineStyle: { width: 1.6 },
    name: h.nodeName,
    showSymbol: false,
    smooth: true,
    type: 'line',
  })),
  tooltip: {
    trigger: 'axis',
    valueFormatter: (value) => `${value ?? '-'} ms`,
  },
  xAxis: {
    axisLabel: { formatter: '{HH}:{mm}:{ss}' },
    type: 'time',
  },
  yAxis: { name: 'ms', scale: true, type: 'value' },
}));

watch(selectedIds, (value) => {
  if (value.length > MAX_COMPARE) {
    selectedIds.value = value.slice(-3);
  } else {
    loadHistories();
  }
});
watch(limit, () => loadHistories());

const [Drawer, drawerApi] = useVbenDrawer<{ record?: NodeItem }>({
  class: 'w-[680px]',
  footer: false,
  onOpenChange(isOpen) {
    if (isOpen) {
      const record = drawerApi.getData()?.record;
      if (record) {
        selectedIds.value = [record.id];
        loadOptions(record.centerId);
        loadHistories();
      }
    }
  },
  title: '延迟响应趋势',
});
</script>

<template>
  <Drawer title="延迟响应趋势">
    <Spin :spinning="loading">
      <div class="mb-3 flex flex-wrap items-center gap-2">
        <Select
          v-model:value="selectedIds"
          :max-tag-count="2"
          :options="nodeOptions"
          class="min-w-64"
          mode="multiple"
          option-filter-prop="label"
          placeholder="选择同中心节点对比（最多 3 个）"
          show-search
        />
        <Segmented v-model:value="limit" :options="LIMIT_OPTIONS" />
        <Button size="small" @click="loadHistories()">刷新</Button>
      </div>

      <template v-if="histories.length > 0">
        <!-- 单节点：完整统计摘要 -->
        <Descriptions
          v-if="histories.length === 1 && primary"
          :column="3"
          bordered
          size="small"
          title="探活统计摘要"
        >
          <DescriptionsItem label="健康状态">
            <Tag
              :color="(HEALTH_TAG[primary.healthStatus] ?? HEALTH_TAG_ONLINE).color"
            >
              {{ (HEALTH_TAG[primary.healthStatus] ?? HEALTH_TAG_ONLINE).text }}
            </Tag>
          </DescriptionsItem>
          <DescriptionsItem label="平均延迟">
            {{
              primary.summary.avgMs != null ? `${primary.summary.avgMs} ms` : '-'
            }}
          </DescriptionsItem>
          <DescriptionsItem label="P95 延迟">
            {{
              primary.summary.p95Ms != null ? `${primary.summary.p95Ms} ms` : '-'
            }}
          </DescriptionsItem>
          <DescriptionsItem label="最大延迟">
            {{
              primary.summary.maxMs != null ? `${primary.summary.maxMs} ms` : '-'
            }}
          </DescriptionsItem>
          <DescriptionsItem label="最小延迟">
            {{
              primary.summary.minMs != null ? `${primary.summary.minMs} ms` : '-'
            }}
          </DescriptionsItem>
          <DescriptionsItem label="丢包率">
            {{ primary.summary.lossRate }}%
          </DescriptionsItem>
          <DescriptionsItem label="采样次数">
            {{ primary.summary.sampleCount }} 次
          </DescriptionsItem>
          <DescriptionsItem label="采样间隔">
            {{ primary.intervalSec }} 秒
          </DescriptionsItem>
          <DescriptionsItem label="超时阈值">
            {{ primary.summary.timeoutMs }} ms
          </DescriptionsItem>
        </Descriptions>

        <!-- 多节点对比：逐节点摘要行 -->
        <div v-else class="mb-2 flex flex-col gap-2">
          <div
            v-for="h in histories"
            :key="h.nodeId"
            class="flex flex-wrap items-center gap-x-4 gap-y-1 rounded border border-gray-200 px-3 py-2 text-xs dark:border-gray-700"
          >
            <span class="font-medium">{{ h.nodeName }}</span>
            <Tag
              :color="(HEALTH_TAG[h.healthStatus] ?? HEALTH_TAG_ONLINE).color"
              class="m-0"
            >
              {{ (HEALTH_TAG[h.healthStatus] ?? HEALTH_TAG_ONLINE).text }}
            </Tag>
            <span>
              平均 {{ h.summary.avgMs != null ? `${h.summary.avgMs} ms` : '-' }}
            </span>
            <span>
              P95 {{ h.summary.p95Ms != null ? `${h.summary.p95Ms} ms` : '-' }}
            </span>
            <span>丢包 {{ h.summary.lossRate }}%</span>
            <span>{{ h.summary.sampleCount }} 次采样</span>
          </div>
        </div>

        <EChart v-if="!loading" :option="chartOption" height="320px" />
      </template>

      <Empty
        v-else-if="!loading"
        description="暂无探活采样数据，请稍后刷新"
        style="margin-top: 80px"
      />
    </Spin>
  </Drawer>
</template>
