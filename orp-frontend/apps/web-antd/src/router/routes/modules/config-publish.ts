import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:git-compare-arrows',
      activeIcon: 'ph:arrows-left-right-fill',
      order: 7.5,
      title: '配置比对与下发',
    },
    name: 'ConfigPublish',
    path: '/config-publish',
    component: () => import('#/views/config-publish/index.vue'),
  },
];

export default routes;
