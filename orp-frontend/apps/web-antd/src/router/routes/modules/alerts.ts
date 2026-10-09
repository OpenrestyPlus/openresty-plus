import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:bell-ring',
      activeIcon: 'ph:bell-ringing-fill',
      order: 5.5,
      title: '告警通知',
    },
    name: 'AlertNotifications',
    path: '/alerts',
    component: () => import('#/views/alerts/index.vue'),
  },
];

export default routes;
