<script lang="ts" setup>
import type { NodeItem, NodeMetrics } from '#/api';

import { ref, watch } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { Descriptions, DescriptionsItem, Progress, Spin, Tag } from 'ant-design-vue';

import { getNodeMetricsApi } from '#/api';

defineOptions({ name: 'NodeMetricsDrawer' });

const loading = ref(false);
const node = ref<NodeItem>();
const metrics = ref<NodeMetrics>();

const STATUS_COLOR: Record<string, string> = {
  maintenance: 'processing',
  offline: 'default',
  online: 'success',
};

const STATUS_TEXT: Record<string, string> = {
  maintenance: '维护中',
  offline: '离线',
  online: '在线',
};

/** 加载指标数据 */
async function loadMetrics(id: number) {
  loading.value = true;
  try {
    const res = await getNodeMetricsApi(id);
    node.value = res.node;
    metrics.value = res.metrics;
  } catch {
    // 错误提示由请求拦截器统一处理
  } finally {
    loading.value = false;
  }
}

const [Drawer, drawerApi] = useVbenDrawer<{ record?: NodeItem }>({
  class: 'w-[560px]',
  onOpenChange(isOpen) {
    if (isOpen) {
      const record = drawerApi.getData()?.record;
      if (record) {
        node.value = record;
        metrics.value = undefined;
        loadMetrics(record.id);
      }
    }
  },
  title: '节点指标',
});

watch(
  () => node.value?.id,
  () => {
    drawerApi.setState({
      title: node.value ? `节点指标 - ${node.value.name}` : '节点指标',
    });
  },
);
</script>

<template>
  <Drawer>
    <Spin :spinning="loading">
      <div class="flex flex-col gap-4">
        <Descriptions
          :column="2"
          bordered
          size="small"
          title="基础信息"
        >
          <DescriptionsItem label="节点名称" :span="1">
            {{ node?.name }}
          </DescriptionsItem>
          <DescriptionsItem label="运行状态" :span="1">
            <Tag v-if="node" :color="STATUS_COLOR[node.status]">
              {{ STATUS_TEXT[node.status] }}
            </Tag>
          </DescriptionsItem>
          <DescriptionsItem label="主机地址" :span="1">
            {{ node?.host }}
          </DescriptionsItem>
          <DescriptionsItem label="所属中心" :span="1">
            {{ node?.centerName }}
          </DescriptionsItem>
          <DescriptionsItem label="活动版本" :span="1">
            {{ node?.activeVersion }}
          </DescriptionsItem>
          <DescriptionsItem label="通信端点" :span="1">
            {{ node?.controlEndpoint }}
          </DescriptionsItem>
        </Descriptions>

        <Descriptions :column="2" bordered size="small" title="运行指标">
          <DescriptionsItem label="CPU 使用率">
            <div class="flex items-center gap-2">
              <Progress
                :percent="metrics?.cpuPercent ?? 0"
                :status="(metrics?.cpuPercent ?? 0) > 80 ? 'exception' : 'normal'"
                :stroke-width="10"
                size="small"
                style="max-width: 140px; min-width: 100px;"
              />
            </div>
          </DescriptionsItem>
          <DescriptionsItem label="内存使用率">
            <div class="flex items-center gap-2">
              <Progress
                :percent="metrics?.memPercent ?? 0"
                :status="(metrics?.memPercent ?? 0) > 80 ? 'exception' : 'normal'"
                :stroke-width="10"
                size="small"
                style="max-width: 140px; min-width: 100px;"
              />
            </div>
          </DescriptionsItem>
          <DescriptionsItem label="QPS">
            {{ metrics?.qps.toLocaleString() ?? '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="活动连接数">
            {{ metrics?.connections.toLocaleString() ?? '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="入带宽 (Mbps)">
            {{ metrics?.bandwidthInMbps ?? '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="出带宽 (Mbps)">
            {{ metrics?.bandwidthOutMbps ?? '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="24h 请求量">
            {{ metrics?.requestTotal24h.toLocaleString() ?? '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="错误率">
            {{ metrics ? `${metrics.errorRatePercent}%` : '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="持续运行" :span="2">
            {{ metrics ? `${metrics.uptimeDays} 天` : '-' }}
          </DescriptionsItem>
        </Descriptions>
      </div>
    </Spin>
    <span v-if="!loading && metrics" class="mt-2 block text-xs text-gray-400">
      指标为 Mock 演示数据，对接真实控制面后展示实时采集值。
    </span>
  </Drawer>
</template>
