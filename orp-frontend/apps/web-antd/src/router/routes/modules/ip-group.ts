import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:shield',
      activeIcon: 'ph:shield-fill',
      order: 7,
      title: 'IP 组管理',
    },
    name: 'IpGroup',
    path: '/ip-group',
    component: () => import('#/views/ip-group/index.vue'),
  },
];

export default routes;
