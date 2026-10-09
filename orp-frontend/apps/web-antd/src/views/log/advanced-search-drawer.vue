<script lang="ts" setup>
/** 实时日志高级搜索抽屉：状态码多选/区间 + 关键词包含/排除 + 命中黄色高亮 */
import { reactive } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { Alert, Button, Card, Input, Select, Tag } from 'ant-design-vue';

defineOptions({ name: 'LogAdvancedSearchDrawer' });

const emit = defineEmits<{ apply: [query: AdvQuery] }>();

interface AdvQuery {
  codeMax: null | number;
  codeMin: null | number;
  codes: number[];
  exclude: string[];
  include: string[];
}

/** 常用 HTTP 状态码 */
const CODE_OPTIONS = [200, 201, 204, 301, 302, 304, 400, 401, 403, 404, 408, 429, 500, 502, 503, 504].map(
  (c) => ({ label: String(c), value: c }),
);

const draft = reactive<AdvQuery & { excludeInput: string; includeInput: string }>({
  codeMax: null,
  codeMin: null,
  codes: [],
  exclude: [],
  excludeInput: '',
  include: [],
  includeInput: '',
});

function normalize(q: Partial<AdvQuery>): AdvQuery {
  return {
    codeMax: q.codeMax ?? null,
    codeMin: q.codeMin ?? null,
    codes: (q.codes ?? []).map(Number).filter((x) => Number.isInteger(x) && x >= 100 && x <= 599),
    exclude: (q.exclude ?? []).map(String).filter(Boolean),
    include: (q.include ?? []).map(String).filter(Boolean),
  };
}

/** 添加包含词（回车或点加号） */
function addInclude() {
  const w = draft.includeInput.trim();
  if (w && !draft.include.includes(w)) {
    draft.include.push(w);
  }
  draft.includeInput = '';
}
function addExclude() {
  const w = draft.excludeInput.trim();
  if (w && !draft.exclude.includes(w)) {
    draft.exclude.push(w);
  }
  draft.excludeInput = '';
}

function resetAll() {
  draft.codes = [];
  draft.codeMin = null;
  draft.codeMax = null;
  draft.include = [];
  draft.exclude = [];
  draft.includeInput = '';
  draft.excludeInput = '';
  emit('apply', normalize(draft));
}

const [Drawer, drawerApi] = useVbenDrawer<{ query?: AdvQuery }>({
  class: 'w-[480px]',
  footer: true,
  onOpenChange(isOpen) {
    if (isOpen) {
      const q = drawerApi.getData()?.query;
      const n = normalize(q ?? {});
      draft.codes = [...n.codes];
      draft.codeMin = n.codeMin;
      draft.codeMax = n.codeMax;
      draft.include = [...n.include];
      draft.exclude = [...n.exclude];
      draft.includeInput = '';
      draft.excludeInput = '';
    }
  },
  title: '高级搜索',
});

function onApply() {
  emit('apply', normalize(draft));
  drawerApi.close();
}
</script>

<template>
  <Drawer>
    <div class="flex flex-col gap-4">
      <Alert
        message="条件实时作用于日志缓冲区，命中关键词将黄色高亮；清除后恢复全量显示。"
        show-icon
        type="info"
      />

      <!-- 状态码 -->
      <Card size="small" title="状态码过滤">
        <div class="flex flex-col gap-3">
          <div>
            <div class="mb-1.5 text-xs text-gray-500">多选（命中任一状态码）</div>
            <Select
              v-model:value="draft.codes"
              :options="CODE_OPTIONS"
              allow-clear
              class="w-full"
              mode="multiple"
              placeholder="例如：200、404、502"
            />
          </div>
          <div>
            <div class="mb-1.5 text-xs text-gray-500">区间（最小/最大，可单边）</div>
            <div class="flex items-center gap-2">
              <InputNumber
                v-model:value="draft.codeMin"
                :max="599"
                :min="100"
                class="!w-full"
                placeholder="最小"
              />
              <span class="text-gray-400">—</span>
              <InputNumber
                v-model:value="draft.codeMax"
                :max="599"
                :min="100"
                class="!w-full"
                placeholder="最大"
              />
            </div>
          </div>
        </div>
      </Card>

      <!-- 包含关键词 -->
      <Card size="small" title="关键词 · 包含（任一命中即保留）">
        <div class="flex gap-2">
          <Input
            v-model:value="draft.includeInput"
            placeholder="输入关键词后回车添加"
            @keydown.enter.prevent="addInclude"
          />
          <Button @click="addInclude">添加</Button>
        </div>
        <div class="mt-2 flex flex-wrap gap-1.5">
          <Tag
            v-for="w in draft.include"
            :key="w"
            closable
            color="green"
            @close="draft.include = draft.include.filter((x) => x !== w)"
          >
            {{ w }}
          </Tag>
          <span v-if="draft.include.length === 0" class="text-xs text-gray-400">未设置，默认不过滤</span>
        </div>
      </Card>

      <!-- 排除关键词 -->
      <Card size="small" title="关键词 · 排除（任一命中即剔除）">
        <div class="flex gap-2">
          <Input
            v-model:value="draft.excludeInput"
            placeholder="输入关键词后回车添加"
            @keydown.enter.prevent="addExclude"
          />
          <Button @click="addExclude">添加</Button>
        </div>
        <div class="mt-2 flex flex-wrap gap-1.5">
          <Tag
            v-for="w in draft.exclude"
            :key="w"
            closable
            color="red"
            @close="draft.exclude = draft.exclude.filter((x) => x !== w)"
          >
            {{ w }}
          </Tag>
          <span v-if="draft.exclude.length === 0" class="text-xs text-gray-400">未设置，默认不过滤</span>
        </div>
      </Card>
    </div>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <Button @click="resetAll">清除全部</Button>
        <Button type="primary" @click="onApply">应用搜索</Button>
      </div>
    </template>
  </Drawer>
</template>
