import type { NodeStatus } from '#/api';

/** 节点运行状态视觉元数据：色点 + 描边徽章（含暗色主题变体），列表标签与状态修改弹窗共用 */
export const STATUS_META: Record<
  NodeStatus,
  { dot: string; label: string; pill: string }
> = {
  running: {
    dot: 'bg-emerald-500',
    label: '运行中',
    pill: 'bg-emerald-50 border-emerald-200 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300',
  },
  paused: {
    dot: 'bg-sky-500',
    label: '已暂停',
    pill: 'bg-sky-50 border-sky-200 text-sky-700 dark:border-sky-800 dark:bg-sky-950 dark:text-sky-300',
  },
  stopped: {
    dot: 'bg-rose-500',
    label: '已停止',
    pill: 'bg-rose-50 border-rose-200 text-rose-700 dark:border-rose-800 dark:bg-rose-950 dark:text-rose-300',
  },
  maintenance: {
    dot: 'bg-amber-500',
    label: '维护中',
    pill: 'bg-amber-50 border-amber-200 text-amber-700 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-300',
  },
};

/** 运行状态枚举顺序（列表筛选与弹窗选项共用） */
export const STATUS_VALUES = [
  'running',
  'paused',
  'stopped',
  'maintenance',
] as const;

/** 列表筛选用下拉选项（label 纯文本） */
export const STATUS_OPTIONS = STATUS_VALUES.map((value) => ({
  label: STATUS_META[value].label,
  value,
}));
