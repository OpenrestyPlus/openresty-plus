<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import { computed, markRaw, onMounted } from 'vue';

import {
  AuthenticationLogin,
  SliderCaptcha,
  z,
} from '@vben/common-ui';
import { $t } from '@vben/locales';

import { useAuthStore } from '#/store';

defineOptions({ name: 'Login' });

const authStore = useAuthStore();
const localDevPassword = import.meta.env.VITE_LOCAL_DEV_PASSWORD || '';
const defaultCredentials = import.meta.env.DEV || localDevPassword
  ? { username: 'vben', password: localDevPassword }
  : { username: '', password: '' };

onMounted(() => {
  if (!import.meta.env.DEV) return;
  try {
    const savedUsername = window.localStorage.getItem('orp.local-dev-username');
    if (savedUsername === 'vben') defaultCredentials.username = savedUsername;
  } catch {
    // 本地凭据读取失败时使用开发环境默认账号。
  }
});

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: $t('authentication.usernameTip'),
      },
      defaultValue: defaultCredentials.username,
      fieldName: 'username',
      label: $t('authentication.username'),
      rules: z.string().min(1, { message: $t('authentication.usernameTip') }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: $t('authentication.password'),
      },
      defaultValue: defaultCredentials.password,
      fieldName: 'password',
      label: $t('authentication.password'),
      rules: z.string().min(1, { message: $t('authentication.passwordTip') }),
    },
    {
      component: markRaw(SliderCaptcha),
      fieldName: 'captcha',
      rules: z.boolean().refine((value) => value, {
        message: $t('authentication.verifyRequiredTip'),
      }),
    },
  ];
});
</script>

<template>
  <AuthenticationLogin
    :form-schema="formSchema"
    :loading="authStore.loginLoading"
    :show-code-login="false"
    :show-forget-password="false"
    :show-qrcode-login="false"
    :show-register="false"
    :show-remember-me="false"
    @submit="authStore.authLogin"
  />
</template>
