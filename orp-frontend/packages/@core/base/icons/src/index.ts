// 侧边栏激活态 Phosphor 填充图标离线集合（副作用注册，先于任何 VbenIcon 渲染执行）
import './phosphor';
// 侧边栏默认态 lucide 线性图标与框架内置图标离线集合（同机制副作用注册）
import './builtin';

export * from './create-icon';

export * from './lucide';

export type { IconifyIcon as IconifyIconStructure } from '@iconify/vue';
export {
  addCollection,
  addIcon,
  Icon as IconifyIcon,
  listIcons,
} from '@iconify/vue';
