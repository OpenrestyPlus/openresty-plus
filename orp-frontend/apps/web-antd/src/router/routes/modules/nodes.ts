import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:server',
      activeIcon: 'ph:hard-drives-fill',
      order: 2,
      title: '节点实例',
    },
    name: 'NodeInstance',
    path: '/nodes',
    component: () => import('#/views/node/index.vue'),
  },
];

export default routes;
