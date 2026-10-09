import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:globe',
      activeIcon: 'ph:globe-fill',
      order: 3,
      title: 'HTTP 流量',
    },
    name: 'HttpTraffic',
    path: '/http',
    component: () => import('#/views/http/index.vue'),
  },
];

export default routes;
