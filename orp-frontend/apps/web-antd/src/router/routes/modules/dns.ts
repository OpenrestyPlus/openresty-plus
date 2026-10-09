import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:network',
      activeIcon: 'ph:graph-fill',
      order: 6,
      title: 'DNS 解析器',
    },
    name: 'DnsResolver',
    path: '/dns',
    component: () => import('#/views/dns/index.vue'),
  },
];

export default routes;
