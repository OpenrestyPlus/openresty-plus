import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      hideInMenu: true,
      icon: 'lucide:file-code-2',
      activeIcon: 'ph:file-code-fill',
      order: 8,
      title: '接口文档',
    },
    name: 'ApiDocs',
    path: '/api-docs',
    component: () => import('#/views/api-docs/index.vue'),
  },
];

export default routes;
