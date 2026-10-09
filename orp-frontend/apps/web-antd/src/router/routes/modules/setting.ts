import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:settings-2',
      activeIcon: 'ph:gear-fill',
      order: 8,
      title: '系统配置',
    },
    name: 'SystemSetting',
    path: '/system-setting',
    component: () => import('#/views/setting/index.vue'),
  },
];

export default routes;
