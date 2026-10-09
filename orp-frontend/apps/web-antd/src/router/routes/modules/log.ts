import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:terminal',
      activeIcon: 'ph:terminal-window-fill',
      order: 9,
      title: '实时日志',
    },
    name: 'RealtimeLog',
    path: '/log',
    component: () => import('#/views/log/index.vue'),
  },
];

export default routes;
