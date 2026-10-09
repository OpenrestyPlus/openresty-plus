<script lang="ts" setup>
/** 下发策略配置弹窗：全量 / 权重分批（金丝雀优先）+ 观察等待期 */
import { computed, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Button, InputNumber, Radio, RadioGroup, Switch, Tag } from 'ant-design-vue';

import { listNodesApi } from '#/api';

defineOptions({ name: 'PublishStrategyModal' });

const emit = defineEmits<{
  confirmed: [payload: { nodeIds: number[]; strategy: PublishStrategyInput }];
}>();

interface PublishStrategyInput {
  batchCount?: number;
  canary?: { enabled: boolean; waitSeconds: number };
  mode: 'all' | 'weighted';
}

/** 节点权重信息（用于分批预览） */
interface NodeWeightInfo {
  id: number;
  name: string;
  weight: number;
}

const nodeInfos = ref<NodeWeightInfo[]>([]);
const mode = ref<'all' | 'weighted'>('all');
const batchCount = ref(2);
const canaryEnabled = ref(true);
const canaryWait = ref(30);
const nodeIds = ref<number[]>([]);
const confirming = ref(false);

const WAIT_OPTIONS = [10, 30, 60, 300];

/** 权重升序分批预览（低权重金丝雀优先） */
const batchesPreview = computed(() => {
  if (mode.value !== 'weighted' || nodeInfos.value.length === 0) return [];
  const k = Math.max(1, Math.min(batchCount.value ?? 1, 5, nodeInfos.value.length));
  const sorted = [...nodeInfos.value].sort((a, b) => a.weight - b.weight || a.id - b.id);
  const sizes: number[] = Array.from({ length: k }, () => Math.floor(sorted.length / k));
  for (let r = 0; r < sorted.length % k; r++) {
    sizes[k - 1 - r] = (sizes[k - 1 - r] ?? 0) + 1;
  }
  const out: NodeWeightInfo[][] = [];
  let ix = 0;
  for (let b = 0; b < k; b++) {
    const size = sizes[b] ?? 0;
    out.push(sorted.slice(ix, ix + size));
    ix += size;
  }
  return out;
});

watch(batchCount, (v) => {
  if (v == null || !Number.isInteger(v) || v < 1) batchCount.value = 1;
  if (v != null && v > 5) batchCount.value = 5;
});

const [Modal, modalApi] = useVbenModal<{ nodeIds?: number[] }>({
  class: 'w-[640px]',
  footer: false,
  async onOpenChange(isOpen) {
    if (isOpen) {
      nodeIds.value = modalApi.getData()?.nodeIds ?? [];
      confirming.value = false;
      if (nodeInfos.value.length === 0) {
        try {
          const res = await listNodesApi({ page: 1, pageSize: 200 });
          nodeInfos.value = res.items.map((n) => ({ id: n.id, name: n.name, weight: n.weight ?? 10 }));
        } catch {
          nodeInfos.value = [];
        }
      }
    }
  },
  title: '下发策略配置',
});

function onConfirm() {
  if (confirming.value) return;
  confirming.value = true;
  emit('confirmed', {
    nodeIds: nodeIds.value,
    strategy: {
      batchCount: mode.value === 'weighted' ? (batchCount.value ?? 2) : undefined,
      canary:
        mode.value === 'weighted'
          ? { enabled: canaryEnabled.value, waitSeconds: canaryWait.value ?? 30 }
          : undefined,
      mode: mode.value,
    },
  });
}
</script>

<template>
  <Modal>
    <div class="flex flex-col gap-4 py-1">
      <Alert
        message="语法预检已全部通过，请选择下发策略：权重分批将按节点权重升序分批执行，低权重实例优先金丝雀灰度验证。"
        show-icon
        type="success"
      />

      <div class="rounded-md border border-gray-200 p-3 dark:border-gray-700">
        <RadioGroup v-model:value="mode" class="flex flex-col gap-3">
          <Radio value="all">
            <div class="inline-flex items-center gap-2">
              <span class="text-sm font-medium">全量下发</span>
              <Tag class="!m-0" color="blue">一次到位</Tag>
            </div>
            <div class="mt-0.5 pl-6 text-xs text-gray-500">
              逐实例执行已选节点，下发失败会在进度页标明具体原因
            </div>
          </Radio>
          <Radio value="weighted">
            <div class="inline-flex items-center gap-2">
              <span class="text-sm font-medium">按节点权重分批</span>
              <Tag class="!m-0" color="orange">金丝雀优先</Tag>
            </div>
            <div class="mt-0.5 pl-6 text-xs text-gray-500">
              低权重实例最先下发验证，观察无异常后再推后续批次，降低故障爆炸半径
            </div>
          </Radio>
        </RadioGroup>
      </div>

      <template v-if="mode === 'weighted'">
        <div class="rounded-md border border-gray-200 p-3 dark:border-gray-700">
          <div class="mb-2 flex items-center justify-between">
            <span class="text-sm font-medium">批次划分</span>
            <div class="flex items-center gap-2">
              <span class="text-xs text-gray-500">分批数（1-5）</span>
              <InputNumber v-model:value="batchCount" :max="5" :min="1" size="small" />
            </div>
          </div>
          <div class="space-y-1.5">
            <div
              v-for="(batch, ix) in batchesPreview"
              :key="ix"
              class="flex flex-wrap items-center gap-2 rounded bg-gray-50 px-2.5 py-1.5 dark:bg-gray-800/60"
            >
              <Tag class="!m-0" :color="ix === 0 ? 'orange' : 'geekblue'">
                {{ ix === 0 ? '金丝雀批' : `第 ${ix + 1} 批` }}
              </Tag>
              <span
                v-for="n in batch"
                :key="n.id"
                class="rounded border border-gray-200 bg-white px-1.5 py-0.5 font-mono text-xs dark:border-gray-600 dark:bg-gray-900"
              >
                {{ n.name }} <span class="text-amber-600">W{{ n.weight }}</span>
              </span>
              <span class="ml-auto text-xs text-gray-400">Σ权重 {{ batch.reduce((s, n) => s + n.weight, 0) }}</span>
            </div>
          </div>
        </div>

        <div class="rounded-md border border-gray-200 p-3 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-sm font-medium">金丝雀观察等待期</div>
              <div class="mt-0.5 text-xs text-gray-500">
                金丝雀批下发完成后暂停，观察确认无异常再自动推进（期间可中止或立即推进）
              </div>
            </div>
            <Switch v-model:checked="canaryEnabled" />
          </div>
          <div v-if="canaryEnabled" class="mt-3 flex items-center gap-2">
            <span class="text-xs text-gray-500">等待时长</span>
            <RadioGroup v-model:value="canaryWait" size="small">
              <Radio v-for="w in WAIT_OPTIONS" :key="w" :value="w">
                {{ w >= 60 ? `${w / 60} 分钟` : `${w} 秒` }}
              </Radio>
            </RadioGroup>
          </div>
        </div>
      </template>

      <div class="flex justify-end gap-2 border-t border-gray-100 pt-3 dark:border-gray-800">
        <Button @click="modalApi.close()">取消</Button>
        <Button :loading="confirming" danger type="primary" @click="onConfirm">
          按此策略下发（{{ nodeIds.length }} 实例）
        </Button>
      </div>
    </div>
  </Modal>
</template>
