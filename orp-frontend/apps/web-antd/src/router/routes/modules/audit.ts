import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:scroll-text',
      activeIcon: 'ph:scroll-fill',
      order: 7,
      title: '审计日志',
    },
    name: 'AuditLog',
    path: '/audit',
    component: () => import('#/views/audit/index.vue'),
  },
];

export default routes;
