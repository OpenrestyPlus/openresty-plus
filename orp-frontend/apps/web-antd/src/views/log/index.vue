<script lang="ts" setup>
import type {
  HttpListener,
  LogLine,
  NodeItem,
  StreamService,
} from '#/api';

import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { useVbenDrawer } from '@vben/common-ui';

import {
  Button,
  Card,
  Input,
  Select,
  Space,
  Switch,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import {
  getHttpLogFormatApi,
  consumeLogEvents,
  getStreamLogFormatApi,
  listHttpListenersApi,
  listNodesApi,
  listStreamServicesApi,
} from '#/api';

import AdvancedSearchDrawer from './advanced-search-drawer.vue';

defineOptions({ name: 'RealtimeLog' });

const router = useRouter();

/** 目标类型 */
const targetType = ref<'http' | 'node' | 'stream'>('node');
const targetId = ref<number>();
const keyword = ref('');
const running = ref(false);
const autoScroll = ref(true);

/** 高级搜索条件（抽屉中编辑，实时生效） */
interface AdvancedQuery {
  codeMax: null | number;
  codeMin: null | number;
  codes: number[];
  exclude: string[];
  include: string[];
}
const adv = ref<AdvancedQuery>({
  codeMax: null,
  codeMin: null,
  codes: [],
  exclude: [],
  excludeInput: '',
  include: [],
  includeInput: '',
} as any);

/** 抽屉组件句柄 */
const [AdvDrawer, advApi] = useVbenDrawer({
  connectedComponent: AdvancedSearchDrawer,
});

/** 高级搜索生效中（任一条件非空） */
const advActive = computed(
  () =>
    adv.value.codes.length > 0 ||
    adv.value.codeMin != null ||
    adv.value.codeMax != null ||
    adv.value.include.length > 0 ||
    adv.value.exclude.length > 0,
);

/** 从日志行提取状态码（access 日志 "GET /path HTTP/1.1" 200 xx 模式） */
function extractStatusCode(line: string): number | null {
  const m = line.match(/"(?:GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)[^"]*"\s+(\d{3})\b/);
  if (m) return Number(m[1]);
  const m2 = line.match(/\b(2\d\d|3\d\d|4\d\d|5\d\d)\b/);
  return m2 ? Number(m2[1]) : null;
}

/** 关键词过滤 + 高级搜索过滤后的行 */
const filteredLines = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  const a = adv.value;
  let out = lines.value;
  if (kw) {
    out = out.filter((l) => l.line.toLowerCase().includes(kw));
  }
  if (advActive.value) {
    out = out.filter((l) => {
      // 状态码多选
      if (a.codes.length > 0) {
        const code = extractStatusCode(l.line);
        if (code == null || !a.codes.includes(code)) return false;
      }
      // 状态码区间
      if (a.codeMin != null || a.codeMax != null) {
        const code = extractStatusCode(l.line);
        if (code == null) return false;
        if (a.codeMin != null && code < a.codeMin) return false;
        if (a.codeMax != null && code > a.codeMax) return false;
      }
      // 包含关键词（任一命中即保留）
      if (a.include.length > 0) {
        const low = l.line.toLowerCase();
        if (!a.include.some((k) => low.includes(k.toLowerCase()))) return false;
      }
      // 排除关键词（任一命中即剔除）
      if (a.exclude.length > 0) {
        const low = l.line.toLowerCase();
        if (a.exclude.some((k) => low.includes(k.toLowerCase()))) return false;
      }
      return true;
    });
  }
  return out;
});

/** 需要高亮的关键词集合（高级搜索包含词 + 顶栏关键词） */
const highlightWords = computed(() => {
  const ws = [...adv.value.include, keyword.value.trim()].filter(Boolean);
  return [...new Set(ws.map((w) => w.toLowerCase()))];
});

/** 拆分行文本为高亮片段 */
function splitHighlight(text: string): { hit: boolean; text: string }[] {
  const words = highlightWords.value;
  if (words.length === 0) return [{ hit: false, text }];
  // 按所有关键词首次出现位置切分
  const lower = text.toLowerCase();
  const marks: [number, number][] = [];
  for (const w of words) {
    let from = 0;
    while (from < lower.length) {
      const idx = lower.indexOf(w, from);
      if (idx === -1) break;
      marks.push([idx, idx + w.length]);
      from = idx + w.length;
    }
  }
  if (marks.length === 0) return [{ hit: false, text }];
  marks.sort((a, b) => a[0] - b[0]);
  // 合并重叠区间
  const merged: [number, number][] = [];
  for (const m of marks) {
    const last = merged[merged.length - 1];
    if (last && m[0] <= last[1]) {
      last[1] = Math.max(last[1], m[1]);
    } else {
      merged.push([...m]);
    }
  }
  const out: { hit: boolean; text: string }[] = [];
  let pos = 0;
  for (const [s2, e2] of merged) {
    if (s2 > pos) out.push({ hit: false, text: text.slice(pos, s2) });
    out.push({ hit: true, text: text.slice(s2, e2) });
    pos = e2;
  }
  if (pos < text.length) out.push({ hit: false, text: text.slice(pos) });
  return out;
}

/** 级别着色 */
function levelColor(level: string) {
  if (level === 'ERROR') return 'red';
  if (level === 'WARN') return 'orange';
  if (level === 'DEBUG') return 'default';
  return 'green';
}

/** 行文本色 */
function lineClass(level: string) {
  if (level === 'ERROR') return 'text-red-500';
  if (level === 'WARN') return 'text-orange-500';
  if (level === 'DEBUG') return 'text-gray-400';
  return 'text-emerald-700 dark:text-emerald-400';
}

/** 可选目标列表 */
const nodeOptions = ref<{ label: string; value: number }[]>([]);
const httpOptions = ref<{ label: string; value: number }[]>([]);
const streamOptions = ref<{ label: string; value: number }[]>([]);

/** 日志缓冲（上限 2000 行） */
const MAX_LINES = 2000;
const lines = ref<LogLine[]>([]);
const linesTotal = ref(0);
const seenLogIds = new Set<number>();

/** log_format 提示 */
const httpLogFormat = ref('');
const streamLogFormat = ref('');

/** SSE 实时连接控制 */
let streamController: AbortController | null = null;
let streamGeneration = 0;

const currentTarget = computed(() => {
  if (!targetId.value) return '';
  return `${targetType.value}:${targetId.value}`;
});

const targetTypeOptions = [
  { label: '节点（nginx error/access）', value: 'node' },
  { label: 'HTTP Server（访问日志）', value: 'http' },
  { label: 'Stream Server（四层会话日志）', value: 'stream' },
];

const targetOptions = computed(() => {
  if (targetType.value === 'node') return nodeOptions.value;
  if (targetType.value === 'http') return httpOptions.value;
  return streamOptions.value;
});

const containerRef = ref<HTMLElement>();
async function loadOptions() {
  try {
    const res = await listNodesApi({ page: 1, pageSize: 200 });
    nodeOptions.value = res.items.map((item: NodeItem) => ({
      label: `${item.name}（${item.host}）`,
      value: item.id,
    }));
  } catch {
    nodeOptions.value = [];
  }
  try {
    const res = await listHttpListenersApi({ page: 1, pageSize: 200 });
    httpOptions.value = res.items.map((item: HttpListener) => ({
      label: `${item.domain}:${item.port}`,
      value: item.id,
    }));
  } catch {
    httpOptions.value = [];
  }
  try {
    const res = await listStreamServicesApi({ page: 1, pageSize: 200 });
    streamOptions.value = res.items.map((item: StreamService) => ({
      label: `${item.description || '未命名'} · ${item.listenAddress}:${item.listenPort}/${item.protocol.toUpperCase()}`,
      value: item.id,
    }));
  } catch {
    streamOptions.value = [];
  }
}

async function loadLogFormats() {
  try {
    const res = await getHttpLogFormatApi();
    httpLogFormat.value = `${res?.name ?? ''}：${res?.format ?? ''}`;
  } catch {
    httpLogFormat.value = '';
  }
  try {
    const res = await getStreamLogFormatApi();
    streamLogFormat.value = `${res?.name ?? ''}：${res?.format ?? ''}`;
  } catch {
    streamLogFormat.value = '';
  }
}

function appendLines(incoming: LogLine[]) {
  const batch = incoming.filter((line) => {
    if (seenLogIds.has(line.id)) return false;
    seenLogIds.add(line.id);
    return true;
  });
  if (batch.length > 0) {
    lines.value.push(...batch);
    linesTotal.value += batch.length;
    if (lines.value.length > MAX_LINES) {
      lines.value = lines.value.slice(-MAX_LINES);
      seenLogIds.clear();
      for (const line of lines.value) seenLogIds.add(line.id);
    }
    if (autoScroll.value) {
      requestAnimationFrame(() => {
        const el = containerRef.value;
        if (el) el.scrollTop = el.scrollHeight;
      });
    }
  }
}

function start() {
  if (!currentTarget.value) return;
  pause();
  running.value = true;
  const target = currentTarget.value;
  const generation = ++streamGeneration;
  const controller = new AbortController();
  streamController = controller;
  void (async () => {
    let retryDelay = 1000;
    while (running.value && generation === streamGeneration) {
      try {
        await consumeLogEvents(
          target,
          controller.signal,
          appendLines,
          (message) => console.warn(message),
        );
      } catch (error) {
        if (controller.signal.aborted || generation !== streamGeneration) break;
        console.warn('日志实时连接中断，准备重连', error);
      }
      if (!running.value || generation !== streamGeneration) break;
      await new Promise((resolve) => setTimeout(resolve, retryDelay));
      retryDelay = Math.min(retryDelay * 2, 15_000);
    }
  })();
}

function pause() {
  running.value = false;
  streamGeneration++;
  streamController?.abort();
  streamController = null;
}

function clearScreen() {
  lines.value = [];
  linesTotal.value = 0;
  seenLogIds.clear();
}

watch(targetType, () => {
  if (suppressTypeWatch) {
    suppressTypeWatch = false;
    return;
  }
  targetId.value = undefined;
  pause();
  clearScreen();
});

/** 程序化设置目标时跳过 targetType 联动清空 */
let suppressTypeWatch = false;

/** 从其他页面跳转：/log?target=http:1 */
function initFromQuery() {
  const t = String(router.currentRoute.value.query.target ?? '');
  const [type, id] = t.split(':');
  if (
    (type === 'http' || type === 'node' || type === 'stream') &&
    Number(id) > 0
  ) {
    if (targetType.value !== type) {
      suppressTypeWatch = true;
      targetType.value = type;
    }
    targetId.value = Number(id);
  }
}

/** 已在日志页时再次跳转（query 变化）也响应 */
watch(
  () => router.currentRoute.value.query.target,
  (val) => {
    if (val) {
      pause();
      clearScreen();
      initFromQuery();
      if (currentTarget.value) {
        start();
      }
    }
  },
);

/** 打开高级搜索抽屉 */
function openAdvanced() {
  advApi.setData({ query: adv.value }).open();
}

onMounted(async () => {
  initFromQuery();
  await loadOptions();
  await loadLogFormats();
  if (currentTarget.value) {
    start();
  }
});

onBeforeUnmount(pause);
</script>

<template>
  <div class="p-4 md:p-6">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
      <div>
        <h2 class="text-lg font-semibold">实时日志</h2>
        <p class="mt-1 text-xs text-gray-500">
          按节点 / HTTP Server / Stream Server 维度实时查看运行日志，每 2 秒自动刷新
        </p>
      </div>
      <Space>
        <Tag v-if="running" color="green">● 实时推送中</Tag>
        <Tag v-else>已暂停</Tag>
        <Tag>累计 {{ linesTotal }} 行</Tag>
        <Tag>缓冲 {{ lines.length }} 行</Tag>
        <Tooltip v-if="advActive" title="高级搜索生效中，点击抽屉可调整或清除条件">
          <Tag color="gold" class="cursor-pointer" @click="openAdvanced">
            高级搜索 · {{ adv.codes.length + (adv.codeMin != null ? 1 : 0) + (adv.codeMax != null ? 1 : 0) + adv.include.length + adv.exclude.length }} 条件
          </Tag>
        </Tooltip>
      </Space>
  </div>

    <Card class="mb-4" size="small">
      <div class="flex flex-wrap items-center gap-3">
        <Select
          v-model:value="targetType"
          :options="targetTypeOptions"
          class="w-56"
          placeholder="日志目标类型"
        />
        <Select
          v-model:value="targetId"
          :options="targetOptions"
          :disabled="targetOptions.length === 0"
          class="min-w-64 flex-1"
          placeholder="选择具体目标"
          show-search
          option-filter-prop="label"
        />
        <Input
          v-model:value="keyword"
          allow-clear
          class="w-48"
          placeholder="关键词过滤"
        />
        <Button
          :type="advActive ? 'primary' : 'default'"
          ghost
          @click="openAdvanced"
        >
          高级搜索{{ advActive ? '（生效中）' : '' }}
        </Button>
        <Button
          v-if="!running"
          :disabled="!currentTarget"
          type="primary"
          @click="start"
        >
          开始 / 继续
        </Button>
        <Button v-else @click="pause">暂停</Button>
        <Button danger @click="clearScreen">清屏</Button>
        <Tooltip title="自动滚动到最新日志行">
          <div class="flex items-center gap-1">
            <span class="text-xs text-gray-500">自动滚动</span>
            <Switch v-model:checked="autoScroll" size="small" />
          </div>
        </Tooltip>
      </div>
      <div
        v-if="targetType === 'http' && httpLogFormat"
        class="mt-2 truncate text-xs text-gray-400"
      >
        当前 HTTP log_format（{{ httpLogFormat }}）
      </div>
      <div
        v-else-if="targetType === 'stream' && streamLogFormat"
        class="mt-2 truncate text-xs text-gray-400"
      >
        当前 Stream log_format（{{ streamLogFormat }}）
      </div>
    </Card>

    <div
      ref="containerRef"
      class="max-h-[62vh] min-h-96 overflow-y-auto rounded border border-gray-700 bg-gray-950 p-3 font-mono text-xs leading-5"
    >
      <div
        v-for="line in filteredLines"
        :key="line.id"
        class="flex gap-2 whitespace-pre-wrap break-all"
      >
        <span class="shrink-0 text-gray-500">{{ line.ts }}</span>
        <Tag class="!m-0 !h-4 !px-1 !text-[10px] !leading-4" :color="levelColor(line.level)">
          {{ line.level }}
        </Tag>
        <span :class="lineClass(line.level)"><template v-for="(seg, i) in splitHighlight(line.line)" :key="i"><mark v-if="seg.hit" class="rounded-sm bg-yellow-300/90 px-0.5 text-black">{{ seg.text }}</mark><template v-else>{{ seg.text }}</template></template></span>
      </div>
      <div v-if="filteredLines.length === 0" class="py-8 text-center text-gray-500">
        {{ running ? '等待日志输出…' : '选择目标后点击「开始 / 继续」查看实时日志' }}
      </div>
    </div>
    <AdvDrawer @apply="(q: any) => (adv = q)" />
  </div>
</template>
