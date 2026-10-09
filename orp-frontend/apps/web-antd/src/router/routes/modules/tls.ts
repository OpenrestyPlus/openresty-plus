import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:shield-check',
      activeIcon: 'ph:shield-check-fill',
      order: 5,
      title: 'TLS 证书',
    },
    name: 'TlsCertificate',
    path: '/tls',
    component: () => import('#/views/tls/index.vue'),
  },
];

export default routes;
