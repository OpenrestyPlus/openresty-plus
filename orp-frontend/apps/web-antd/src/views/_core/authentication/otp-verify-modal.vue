<script lang="ts" setup>
/**
 * OTP 动态口令验证弹窗（登录二次验证）
 * - 460px 六分格输入（设计稿 sig-otp：直角工艺、等宽数字、居中光标）
 * - 粘贴/输入自动前移、退格回跳、填满 6 位自动提交
 * - 校验通过 → otpStore.submit(otpCode) → store/auth.ts 重放登录
 */
import { computed, nextTick, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useOtpStore } from '#/store';

defineOptions({ name: 'OtpVerifyModal' });

const otpStore = useOtpStore();

/** 六分格内容（每格一位数字） */
const cells = ref<string[]>(Array.from({ length: 6 }, () => ''));
/** 每格原生 input 引用，用于聚焦控制 */
const cellRefs = ref<HTMLInputElement[]>([]);
const submitting = ref(false);
const errorMessage = ref('');

const [Modal, modalApi] = useVbenModal({
  class: 'w-[460px]',
  closable: true,
  fullscreenButton: false,
  footer: false,
  showCancelButton: false,
  showConfirmButton: false,
  title: '两步验证',
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      cells.value = Array.from({ length: 6 }, () => '');
      errorMessage.value = '';
      await nextTick();
      cellRefs.value[0]?.focus();
    }
  },
});

const code = computed(() => cells.value.join(''));
const filled = computed(() => code.value.length === 6);

watch(
  () => otpStore.visible,
  (v) => {
    if (v) {
      modalApi.open();
    } else {
      modalApi.close();
    }
  },
);

/** 单格输入：仅数字，写入后自动前移 */
function onInput(idx: number, event: Event) {
  const target = event.target as HTMLInputElement;
  const raw = target.value;
  // 兼容输入法/粘贴多字符：取全部数字，逐格填充
  const digits = raw.replaceAll(/\D/g, '');
  if (digits.length > 1) {
    fillFrom(idx, digits);
    return;
  }
  cells.value[idx] = digits;
  if (digits && idx < 5) {
    cellRefs.value[idx + 1]?.focus();
  }
  errorMessage.value = '';
}

/** 键盘事件：退格回跳、方向键移动、回车提交 */
function onKeydown(idx: number, event: KeyboardEvent) {
  if (event.key === 'Backspace' && !cells.value[idx] && idx > 0) {
    event.preventDefault();
    cells.value[idx - 1] = '';
    cellRefs.value[idx - 1]?.focus();
  } else if (event.key === 'ArrowLeft' && idx > 0) {
    event.preventDefault();
    cellRefs.value[idx - 1]?.focus();
  } else if (event.key === 'ArrowRight' && idx < 5) {
    event.preventDefault();
    cellRefs.value[idx + 1]?.focus();
  } else if (event.key === 'Enter' && filled.value) {
    event.preventDefault();
    submit();
  }
}

/** 从 idx 起逐格填充（粘贴 6 位口令场景） */
function fillFrom(idx: number, digits: string) {
  for (let i = 0; i < 6; i++) {
    cells.value[i] = i >= idx && i - idx < digits.length ? (digits[i - idx] ?? '') : (cells.value[i] ?? '');
  }
  const next = Math.min(idx + digits.length, 5);
  cellRefs.value[next]?.focus();
}

/** 整段粘贴（任意一格触发） */
function onPaste(event: ClipboardEvent) {
  const text = event.clipboardData?.getData('text') ?? '';
  const digits = text.replaceAll(/\D/g, '').slice(0, 6);
  if (!digits) return;
  event.preventDefault();
  fillFrom(0, digits);
}

/** 填满自动提交 */
watch(code, async (v) => {
  if (v.length === 6 && !submitting.value) {
    await nextTick();
    submit();
  }
});

async function submit() {
  if (!filled.value || submitting.value) return;
  submitting.value = true;
  errorMessage.value = '';
  try {
    await otpStore.submit(code.value);
  } catch (error: any) {
    // 服务器拒绝：清空重输，展示错误文案（全局拦截器已 toast，这里内联展示便于重试）
    const msg =
      error?.response?.data?.message ?? '动态口令校验失败，请重试';
    errorMessage.value = msg;
    cells.value = Array.from({ length: 6 }, () => '');
    await nextTick();
    cellRefs.value[0]?.focus();
  } finally {
    submitting.value = false;
  }
}

/** 换用其他方式 / 放弃 */
function handleCancel() {
  otpStore.close();
  message.info('已取消两步验证，登录未完成');
}
</script>

<template>
  <Modal>
    <div class="otp-verify-body" data-otp-modal-body>
      <p class="otp-verify-desc">
        账户已开启两步验证（TOTP），请输入验证器应用中的 6 位动态口令完成登录
      </p>
      <div class="otp-verify-cells" data-otp-cells>
        <input
          v-for="(cell, idx) in cells"
          :key="idx"
          :ref="(el) => (cellRefs[idx] = el as HTMLInputElement)"
          :class="['otp-verify-cell', { filled: !!cell }]"
          :value="cell"
          autocomplete="one-time-code"
          data-otp-cell
          inputmode="numeric"
          maxlength="2"
          type="text"
          @input="onInput(idx, $event)"
          @keydown="onKeydown(idx, $event)"
          @paste="onPaste"
        />
      </div>
      <div v-if="errorMessage" class="otp-verify-error" data-otp-error>
        {{ errorMessage }}
      </div>
      <div class="otp-verify-actions">
        <button class="otp-verify-btn ghost" type="button" @click="handleCancel">
          取消登录
        </button>
        <button
          :class="['otp-verify-btn', 'primary', { disabled: !filled || submitting }]"
          :disabled="!filled || submitting"
          data-otp-submit
          type="button"
          @click="submit"
        >
          {{ submitting ? '校验中…' : '验证并登录' }}
        </button>
      </div>
      <p class="otp-verify-hint">
        口令每 30 秒轮换一次；若持续失败，请校准设备时间后重试
      </p>
    </div>
  </Modal>
</template>

<style scoped>
.otp-verify-body {
  padding: 4px 2px 8px;
}

.otp-verify-desc {
  margin: 0 0 18px;
  font-size: 13px;
  line-height: 1.6;
  color: hsl(var(--muted-foreground));
  text-align: left;
}

/* 六分格：直角工艺（--radius 极小值体系），等宽数字 */
.otp-verify-cells {
  display: flex;
  gap: 8px;
}

.otp-verify-cell {
  width: 100%;
  height: 52px;
  flex: 1;
  min-width: 0;
  padding: 0;
  font-size: 22px;
  font-weight: 700;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  text-align: center;
  color: hsl(var(--foreground));
  background: hsl(var(--card));
  border: 1px solid hsl(var(--input));
  border-radius: var(--radius);
  outline: none;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease,
    background 0.15s ease;
}

.otp-verify-cell.filled {
  border-color: hsl(var(--primary) / 0.6);
  background: hsl(var(--secondary));
}

.otp-verify-cell:focus {
  border-color: hsl(var(--ring));
  box-shadow: 0 0 0 3px hsl(var(--primary) / 0.25);
}

.otp-verify-error {
  margin-top: 12px;
  font-size: 12px;
  color: hsl(var(--destructive));
  text-align: left;
}

.otp-verify-actions {
  display: flex;
  gap: 8px;
  margin-top: 20px;
}

.otp-verify-btn {
  flex: 1;
  height: 38px;
  padding: 0 12px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border-radius: var(--radius);
  transition:
    opacity 0.15s ease,
    background 0.15s ease;
}

.otp-verify-btn.ghost {
  color: hsl(var(--foreground) / 0.8);
  background: transparent;
  border: 1px solid hsl(var(--border));
}

.otp-verify-btn.ghost:hover {
  background: hsl(var(--accent) / 0.1);
}

.otp-verify-btn.primary {
  color: hsl(var(--primary-foreground));
  background: hsl(var(--primary));
  border: 1px solid hsl(var(--primary));
}

.otp-verify-btn.primary:hover {
  opacity: 0.88;
}

.otp-verify-btn.primary.disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.otp-verify-hint {
  margin: 14px 0 0;
  font-size: 12px;
  color: hsl(var(--muted-foreground) / 0.8);
  text-align: left;
}
</style>
