import { ref } from 'vue';

import { defineStore } from 'pinia';

/**
 * OTP 动态口令弹窗状态
 * 登录被 OTP_REQUIRED 拦截时，记录登录参数与重放回调，
 * 由挂载在布局中的 OtpVerifyModal 消费并驱动六分格输入。
 */
export const useOtpStore = defineStore('otp', () => {
  /** 弹窗可见态 */
  const visible = ref(false);
  /** 登录参数（用户名/密码），校验通过后与 otpCode 合并重放 */
  const loginParams = ref<Record<string, any>>({});
  /** 校验通过回调（携带 6 位口令重放登录） */
  let replayLogin: null | ((otpCode: string) => Promise<void>) = null;

  function open(
    params: Record<string, any>,
    onVerified: (otpCode: string) => Promise<void>,
  ) {
    loginParams.value = params;
    replayLogin = onVerified;
    visible.value = true;
  }

  function close() {
    visible.value = false;
    loginParams.value = {};
    replayLogin = null;
  }

  async function submit(otpCode: string) {
    await replayLogin?.(otpCode);
    close();
  }

  function $reset() {
    close();
  }

  return {
    $reset,
    close,
    loginParams,
    open,
    submit,
    visible,
  };
});
