import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:arrow-left-right',
      activeIcon: 'ph:arrows-out-line-horizontal-fill',
      order: 4,
      title: 'Stream 流量',
    },
    name: 'StreamTraffic',
    path: '/stream',
    component: () => import('#/views/stream/index.vue'),
  },
];

export default routes;
