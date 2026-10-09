import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      hideInMenu: true,
      icon: 'lucide:id-card',
      activeIcon: 'ph:identification-card-fill',
      order: 999,
      title: '个人中心',
    },
    name: 'Profile',
    path: '/profile',
    component: () => import('#/views/profile/index.vue'),
  },
];

export default routes;
