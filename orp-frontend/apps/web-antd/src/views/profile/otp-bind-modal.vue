<script lang="ts" setup>
/**
 * OTP 绑定 / 关闭弹窗（双模式）
 * - mode=bind：setupOtpApi 生成密钥 → 伪二维码（确定性渲染）+ Base32 密钥 → 六分格验证启用
 * - mode=disable：直接六分格验证关闭
 */
import type { ProfileApi } from '#/api';

import { computed, nextTick, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { disableOtpApi, enableOtpApi, setupOtpApi } from '#/api';

defineOptions({ name: 'OtpBindModal' });

const emit = defineEmits<{ success: [] }>();

type Mode = 'bind' | 'disable';

const mode = ref<Mode>('bind');
const setupInfo = ref<null | ProfileApi.OtpSetupResult>(null);
const loading = ref(false);
const cells = ref<string[]>(Array.from({ length: 6 }, () => ''));
const cellRefs = ref<HTMLInputElement[]>([]);
const errorMessage = ref('');

const code = computed(() => cells.value.join(''));
const filled = computed(() => code.value.length === 6);

const [Modal, modalApi] = useVbenModal({
  class: 'w-[460px]',
  closable: true,
  fullscreenButton: false,
  footer: false,
  showCancelButton: false,
  showConfirmButton: false,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      cells.value = Array.from({ length: 6 }, () => '');
      errorMessage.value = '';
      if (mode.value === 'bind' && !setupInfo.value) {
        apiLoading();
      }
      await nextTick();
      cellRefs.value[0]?.focus();
    }
  },
});

async function apiLoading() {
  loading.value = true;
  try {
    setupInfo.value = await setupOtpApi();
  } catch {
    // 错误由拦截器统一提示，直接关闭弹窗
    modalApi.close();
  } finally {
    loading.value = false;
  }
}

/** 打开弹窗（bind 模式首次生成密钥；disable 模式直接验证） */
function open(targetMode: Mode) {
  mode.value = targetMode;
  if (targetMode === 'bind') {
    setupInfo.value = null;
  }
  modalApi.open();
}

defineExpose({ open });

watch(code, async (v) => {
  if (v.length === 6 && !loading.value) {
    await nextTick();
    submit();
  }
});

/** 确定性伪二维码：基于密钥哈希渲染 21×21 模块（含三个定位图案） */
const qrCells = computed(() => {
  const secret = (setupInfo.value?.otpSecret ?? '').replaceAll(/\s/g, '');
  if (!secret) return [];
  const size = 21;
  const grid: boolean[][] = [];
  let hash = 7;
  for (let i = 0; i < secret.length; i++) {
    hash = (hash * 31 + (secret.codePointAt(i) ?? 0)) % 100_003;
  }
  const rnd = () => {
    hash = (hash * 1_103_515_245 + 12_345) % 2_147_483_648;
    return hash / 2_147_483_648;
  };
  for (let y = 0; y < size; y++) {
    const row: boolean[] = [];
    for (let x = 0; x < size; x++) {
      row.push(rnd() > 0.5);
    }
    grid.push(row);
  }
  // 三个定位图案（7×7：外框+中心块）
  const drawFinder = (fx: number, fy: number) => {
    for (let y = 0; y < 7; y++) {
      for (let x = 0; x < 7; x++) {
        const edge = x === 0 || x === 6 || y === 0 || y === 6;
        const core = x >= 2 && x <= 4 && y >= 2 && y <= 4;
        const row = grid[fy + y];
        if (row) row[fx + x] = edge || core;
      }
    }
  };
  drawFinder(0, 0);
  drawFinder(size - 7, 0);
  drawFinder(0, size - 7);
  return grid;
});

function onInput(idx: number, event: Event) {
  const target = event.target as HTMLInputElement;
  const digits = target.value.replaceAll(/\D/g, '');
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

function fillFrom(idx: number, digits: string) {
  for (let i = 0; i < 6; i++) {
    const keep = cells.value[i] ?? '';
    cells.value[i] =
      i >= idx && i - idx < digits.length ? (digits[i - idx] ?? '') : keep;
  }
  const next = Math.min(idx + digits.length, 5);
  cellRefs.value[next]?.focus();
}

function onPaste(event: ClipboardEvent) {
  const text = event.clipboardData?.getData('text') ?? '';
  const digits = text.replaceAll(/\D/g, '').slice(0, 6);
  if (!digits) return;
  event.preventDefault();
  fillFrom(0, digits);
}

async function submit() {
  if (!filled.value || loading.value) return;
  loading.value = true;
  errorMessage.value = '';
  try {
    if (mode.value === 'bind') {
      await enableOtpApi(code.value);
      message.success('两步验证已开启');
    } else {
      await disableOtpApi(code.value);
      message.success('两步验证已关闭');
    }
    modalApi.close();
    emit('success');
  } catch (error: any) {
    errorMessage.value =
      error?.response?.data?.message ?? '动态口令校验失败，请重试';
    cells.value = Array.from({ length: 6 }, () => '');
    await nextTick();
    cellRefs.value[0]?.focus();
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <Modal>
    <div class="otp-bind-body">
      <!-- 绑定模式：密钥 + 伪二维码 -->
      <template v-if="mode === 'bind'">
        <p class="otp-bind-desc">
          使用验证器应用（如 Google Authenticator）扫描二维码，或手动输入
          Base32 密钥完成绑定，然后输入 6 位动态口令启用两步验证
        </p>
        <div v-if="loading && !setupInfo" class="otp-bind-loading">
          正在生成绑定密钥…
        </div>
        <template v-else-if="setupInfo">
          <div class="otp-bind-qr-wrap">
            <svg
              class="otp-bind-qr"
              data-otp-qr
              height="168"
              viewBox="0 0 21 21"
              width="168"
            >
              <rect fill="#fff" height="21" width="21" x="0" y="0" />
              <template v-for="(row, y) in qrCells" :key="y">
                <rect
                  v-for="(cell, x) in row"
                  :key="x"
                  :fill="cell ? '#101223' : '#fff'"
                  :height="1"
                  :width="1"
                  :x="x"
                  :y="y"
                />
              </template>
            </svg>
          </div>
          <div class="otp-bind-secret" data-otp-secret>
            <span class="otp-bind-secret-label">BASE32 密钥</span>
            <code class="otp-bind-secret-value">{{ setupInfo.otpSecret }}</code>
          </div>
        </template>
      </template>

      <!-- 关闭模式 -->
      <p v-else class="otp-bind-desc">
        关闭两步验证将降低账户安全等级，请输入当前验证器中的 6 位动态口令确认操作
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
        <button class="otp-verify-btn ghost" type="button" @click="modalApi.close()">
          取消
        </button>
        <button
          :class="['otp-verify-btn', 'primary', { disabled: !filled || loading }]"
          :disabled="!filled || loading"
          data-otp-submit
          type="button"
          @click="submit"
        >
          {{ loading ? '校验中…' : mode === 'bind' ? '验证并启用' : '验证并关闭' }}
        </button>
      </div>
      <p class="otp-verify-hint">
        口令每 30 秒轮换一次；若持续失败，请校准设备时间后重试
      </p>
    </div>
  </Modal>
</template>

<style scoped>
.otp-bind-body {
  padding: 4px 2px 8px;
}

.otp-bind-desc {
  margin: 0 0 16px;
  font-size: 13px;
  line-height: 1.6;
  color: hsl(var(--muted-foreground));
  text-align: left;
}

.otp-bind-loading {
  padding: 24px 0;
  font-size: 13px;
  color: hsl(var(--muted-foreground));
  text-align: center;
}

/* 伪二维码：白底黑块，四周留白装裱 */
.otp-bind-qr-wrap {
  display: flex;
  justify-content: center;
  padding: 12px;
  margin-bottom: 14px;
  background: hsl(var(--popover));
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
}

.otp-bind-qr {
  display: block;
  border: 4px solid #fff;
}

.otp-bind-secret {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  margin-bottom: 18px;
  background: hsl(var(--secondary));
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
}

.otp-bind-secret-label {
  font-size: 11px;
  letter-spacing: 0.14em;
  color: hsl(var(--muted-foreground));
}

.otp-bind-secret-value {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 14px;
  font-weight: 600;
  word-break: break-all;
  color: hsl(var(--foreground));
}

/* 六分格（与登录验证弹窗同源视觉） */
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
