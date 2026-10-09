import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:layers',
      activeIcon: 'ph:stack-fill',
      order: 7,
      title: '上游服务器组',
    },
    name: 'UpstreamGroup',
    path: '/upstream',
    component: () => import('#/views/upstream/index.vue'),
  },
];

export default routes;
