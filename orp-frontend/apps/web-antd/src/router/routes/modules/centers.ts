import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:building-2',
      activeIcon: 'ph:buildings-fill',
      order: 1,
      title: '中心管理',
    },
    name: 'Center',
    path: '/centers',
    component: () => import('#/views/center/index.vue'),
  },
];

export default routes;
