<script lang="ts" setup>
import { computed } from 'vue';

import { AuthPageLayout } from '@vben/layouts';
import { preferences } from '@vben/preferences';

import { $t } from '#/locales';
import OtpVerifyModal from '#/views/_core/authentication/otp-verify-modal.vue';

const appName = computed(() => preferences.app.name);
const logo = computed(() => preferences.logo.source);
const logoDark = computed(() => preferences.logo.sourceDark);
</script>

<template>
  <AuthPageLayout
    :app-name="appName"
    :logo="logo"
    :logo-dark="logoDark"
    :page-description="$t('authentication.pageDesc')"
    :page-title="$t('authentication.pageTitle')"
  >
    <!-- 自定义工具栏 -->
    <!-- <template #toolbar></template> -->
  </AuthPageLayout>
  <!-- OTP 两步验证弹窗：登录页触发 OTP_REQUIRED 时在此拉起（auth.ts 拦截 → otp store）。
       注意：AuthPageLayout 无默认插槽，必须平级挂载（antd Modal teleport 至 body） -->
  <OtpVerifyModal />
</template>
