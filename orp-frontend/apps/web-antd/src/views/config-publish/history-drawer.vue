<script lang="ts" setup>
import type { TableColumnsType, TablePaginationConfig } from 'ant-design-vue';

import type {
  Center,
  NodeItem,
  PublishHistoryItem,
  PublishHistoryStats,
} from '#/api';

import { computed, reactive, ref, watch } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { Button, message, Popconfirm, Select, Spin, Table, Tag } from 'ant-design-vue';

import {
  getPublishHistoryStatsApi,
  listCentersApi,
  listNodesApi,
  listPublishHistoryApi,
  rollbackPublishApi,
} from '#/api';

defineOptions({ name: 'PublishHistoryDrawer' });

const emit = defineEmits<{ rollback: [batchId: number] }>();

const MODULE_TEXT: Record<string, string> = {
  center: '中心管理',
  dns: 'DNS 解析器',
  http: 'HTTP 流量',
  node: '节点实例',
  rollback: '版本回滚',
  stream: 'Stream 流量',
  tls: 'TLS 证书',
  upstream: '上游服务器组',
};
const ACTION_TEXT: Record<string, string> = {
  create: '新增',
  delete: '删除',
  rollback: '回滚',
  update: '修改',
};
const ACTION_COLOR: Record<string, string> = {
  create: 'green',
  delete: 'red',
  rollback: 'orange',
  update: 'blue',
};

const centers = ref<Center[]>([]);
const nodes = ref<NodeItem[]>([]);
const rows = ref<PublishHistoryItem[]>([]);
const loading = ref(false);
const total = ref(0);
const filter = reactive<{
  centerId?: number;
  nodeId?: number;
}>({});
const pager = reactive<{ current: number; pageSize: number }>({
  current: 1,
  pageSize: 10,
});

const [Drawer, drawerApi] = useVbenDrawer<{ nodeId?: number }>({
  class: 'w-[860px]',
  confirmText: '关闭',
  showCancelButton: false,
  onConfirm() {
    drawerApi.close();
  },
  onOpenChange(isOpen) {
    if (isOpen) {
      filter.nodeId = drawerApi.getData()?.nodeId;
      filter.centerId = undefined;
      pager.current = 1;
      if (!nodes.value.length) {
        loadOptions();
      }
      query();
      loadStats();
    }
  },
  title: '配置下发历史',
});

const nodeOptions = computed(() => {
  const list = filter.centerId
    ? nodes.value.filter((n) => n.centerId === filter.centerId)
    : nodes.value;
  return list.map((n) => ({
    label: `${n.name}（${n.host}）`,
    value: n.id,
  }));
});

/** 顶部统计卡片数据（发布成功率 + 高频变更中心） */
const stats = ref<null | PublishHistoryStats>(null);

/** 成功率阈值配色：≥95% 绿 / 80-95% 橙 / <80% 红 */
const rateClass = computed(() => {
  const r = stats.value?.successRate ?? 100;
  if (r >= 95) return 'text-green-600 dark:text-green-400';
  if (r >= 80) return 'text-amber-600 dark:text-amber-400';
  return 'text-rose-600 dark:text-rose-400';
});
const rateBarClass = computed(() => {
  const r = stats.value?.successRate ?? 100;
  if (r >= 95) return 'bg-green-500';
  if (r >= 80) return 'bg-amber-500';
  return 'bg-rose-500';
});

async function loadStats() {
  try {
    stats.value = await getPublishHistoryStatsApi();
  } catch {
    // 统计加载失败不阻塞历史列表
  }
}

async function loadOptions() {
  try {
    const [c, n] = await Promise.all([
      listCentersApi({ page: 1, pageSize: 100 }),
      listNodesApi({ page: 1, pageSize: 500 }),
    ]);
    centers.value = c.items;
    nodes.value = n.items;
  } catch {
    // 选项加载失败不阻塞历史查询
  }
}

async function query() {
  loading.value = true;
  try {
    const data = await listPublishHistoryApi({
      centerId: filter.centerId,
      nodeId: filter.nodeId,
      page: pager.current,
      pageSize: pager.pageSize,
    });
    rows.value = data.items;
    total.value = data.total;
  } finally {
    loading.value = false;
  }
}

watch(
  () => filter.centerId,
  () => {
    filter.nodeId = undefined;
    pager.current = 1;
    query();
  },
);

watch(
  () => filter.nodeId,
  () => {
    if (filter.nodeId != null) {
      pager.current = 1;
      query();
    }
  },
);

function onTableChange(pag: TablePaginationConfig) {
  pager.current = pag.current ?? 1;
  pager.pageSize = pag.pageSize ?? 10;
  query();
}

const rollingId = ref<null | number>(null);

/** 回滚会冻结历史快照，在目标节点重新预检并启动真实下发。 */
async function onRollback(row: PublishHistoryItem) {
  rollingId.value = row.id;
  try {
    const res = await rollbackPublishApi(row.id);
    message.success(`实例 ${res.nodeName} 已通过回滚预检，正在发布历史配置`);
    drawerApi.close();
    emit('rollback', res.batchId);
    await query();
  } catch {
    // requestClient 已统一弹出错误提示
  } finally {
    rollingId.value = null;
  }
}

const columns: TableColumnsType = [
  { dataIndex: 'createdAt', title: '下发时间', width: 155 },
  { dataIndex: 'centerName', title: '中心', width: 120 },
  { dataIndex: 'nodeName', title: '实例', width: 135 },
  { key: 'modules', title: '变更模块', width: 125 },
  { dataIndex: 'operator', title: '操作人', width: 85 },
  { key: 'result', title: '结果', width: 75 },
  { key: 'lines', title: '配置行变更', width: 105 },
  { key: 'duration', title: '耗时', width: 75 },
  { key: 'action', title: '操作', fixed: 'right', width: 70 },
];
</script>

<template>
  <Drawer>
    <div
      v-if="stats"
      class="mb-4 grid grid-cols-1 gap-3 md:grid-cols-3"
    >
      <div
        class="rounded-lg border border-gray-200 p-4 dark:border-gray-700 dark:bg-gray-800/40"
      >
        <div class="text-xs text-gray-400">发布成功率</div>
        <div class="mt-1 flex items-baseline gap-1">
          <span class="text-3xl font-semibold" :class="rateClass">
            {{ stats.successRate.toFixed(1) }}
          </span>
          <span class="text-sm text-gray-400">%</span>
        </div>
        <div class="mt-2 h-1.5 w-full rounded-full bg-gray-200 dark:bg-gray-700">
          <div
            class="h-full rounded-full transition-all"
            :class="rateBarClass"
            :style="{ width: `${stats.successRate}%` }"
          ></div>
        </div>
        <div class="mt-2 text-xs text-gray-400">
          成功 {{ stats.successCount }} · 失败 {{ stats.failCount }} ·
          {{ stats.totalBatches }} 个批次
        </div>
      </div>
      <div
        class="rounded-lg border border-gray-200 p-4 md:col-span-2 dark:border-gray-700 dark:bg-gray-800/40"
      >
        <div class="mb-1 text-xs text-gray-400">高频变更中心 Top5</div>
        <div
          v-for="c in stats.topCenters"
          :key="c.name"
          class="mt-2.5"
        >
          <div class="flex items-center justify-between text-xs">
            <span class="font-medium">{{ c.name }}</span>
            <span class="text-gray-400">{{ c.count }} 次 · {{ c.percent }}%</span>
          </div>
          <div class="mt-1 h-1.5 w-full rounded-full bg-gray-200 dark:bg-gray-700">
            <div
              class="h-full rounded-full bg-blue-500 transition-all dark:bg-blue-400"
              :style="{ width: `${c.percent}%` }"
            ></div>
          </div>
        </div>
      </div>
    </div>
    <div class="mb-3 flex flex-wrap items-center gap-2">
      <Select
        v-model:value="filter.centerId"
        :options="[
          ...centers.map((c) => ({ label: c.name, value: c.id })),
        ]"
        allow-clear
        class="w-44"
        placeholder="按中心筛选"
        show-search
        option-filter-prop="label"
      />
      <Select
        v-model:value="filter.nodeId"
        :options="nodeOptions"
        allow-clear
        class="w-52"
        placeholder="按实例筛选"
        show-search
        option-filter-prop="label"
      />
      <span class="ml-auto text-xs text-gray-400">共 {{ total }} 条记录</span>
    </div>

    <Spin :spinning="loading">
      <Table
        :columns="columns"
        :data-source="rows"
        :pagination="{
          current: pager.current,
          pageSize: pager.pageSize,
          total,
          showSizeChanger: false,
          size: 'small',
        }"
        :row-key="(r: PublishHistoryItem) => r.id"
        :scroll="{ x: 1050 }"
        size="small"
        @change="onTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'modules'">
            <Tag
              v-for="m in record.modules"
              :key="m"
              class="!m-0 mr-1"
              color="geekblue"
            >
              {{ MODULE_TEXT[m] ?? m }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'result'">
            <Tag :color="record.result === 'success' ? 'success' : 'error'">
              {{ record.result === 'success' ? '成功' : '失败' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'lines'">
            <span class="text-emerald-600">+{{ record.addCount }}</span>
            <span class="mx-1 text-gray-400">/</span>
            <span class="text-rose-600">-{{ record.delCount }}</span>
          </template>
          <template v-else-if="column.key === 'duration'">
            {{ (record.durationMs / 1000).toFixed(1) }}s
          </template>
          <template v-else-if="column.key === 'action'">
            <Popconfirm
              v-if="record.result === 'success'"
              ok-text="确认回滚"
              cancel-text="取消"
              :disabled="rollingId != null"
              :title="`回滚至 ${record.createdAt} 版本？`"
              @confirm="onRollback(record as PublishHistoryItem)"
            >
              <template #description>
                <div class="max-w-72 text-xs leading-5">
                  将历史配置冻结为新候选，在实例 {{ record.nodeName }} 上重新执行 nginx -t，
                  通过后立即激活并 reload。当前数据库草稿仍保留，可再次预检发布。
                </div>
              </template>
              <Button
                danger
                :loading="rollingId === record.id"
                size="small"
                type="link"
              >
                回滚
              </Button>
            </Popconfirm>
            <span v-else class="text-xs text-gray-400">—</span>
          </template>
        </template>
        <template #expandedRowRender="{ record }">
          <div class="flex flex-col gap-1.5 py-1">
            <div
              v-for="(ev, ix) in record.events"
              :key="ix"
              class="flex flex-wrap items-center gap-2 text-xs"
            >
              <span class="text-gray-400">{{ ev.ts }}</span>
              <Tag class="!m-0" :color="ACTION_COLOR[ev.action]">
                {{ ACTION_TEXT[ev.action] ?? ev.action }}
              </Tag>
              <span class="text-gray-500">
                {{ MODULE_TEXT[ev.module] ?? ev.module }}
              </span>
              <span class="font-mono">{{ ev.target }}</span>
            </div>
            <div v-if="!record.events.length" class="text-xs text-gray-400">
              暂无明细事件
            </div>
          </div>
        </template>
      </Table>
    </Spin>
  </Drawer>
</template>
