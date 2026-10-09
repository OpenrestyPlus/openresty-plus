<script lang="ts" setup>
/**
 * 个人中心（profile-2「深空账户册」设计稿实现）
 * 四区：账户概览 / 个人信息 / 安全设置 / 操作记录
 * 视觉签名：FIG 编号贯穿、等宽数据层（sig-mono）、直角工艺、网格背景、OTP 状态灯
 */
import type { ProfileApi } from '#/api';

import { computed, nextTick, onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Form,
  FormItem,
  Input,
  InputPassword,
  message,
  Pagination,
  RadioButton,
  RadioGroup,
  Table,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import {
  getProfileApi,
  getProfileAuditApi,
  listCentersApi,
  listNodesApi,
  getPublishSummaryApi,
  updatePasswordApi,
  updateProfileApi,
} from '#/api';
import { useUserStore } from '@vben/stores';

import OtpBindModal from './otp-bind-modal.vue';

defineOptions({ name: 'Profile' });

const userStore = useUserStore();
const otpBindModal = ref<InstanceType<typeof OtpBindModal>>();

/** 当前个人信息 */
const profile = ref<ProfileApi.ProfileInfo>({});
/** 页面级加载态 */
const pageLoading = ref(true);
/** 信息表单保存中 */
const savingProfile = ref(false);
/** 密码表单保存中 */
const savingPassword = ref(false);

/** 统计卡 */
const stats = reactive({
  centers: 0,
  changes7d: 0,
  nodes: 0,
  pending: 0,
});

/** 角色徽章文案 */
const roleLabel = computed(() => {
  const role = profile.value.roles?.[0];
  if (role === 'super') return 'SUPER ADMIN';
  if (role === 'admin') return 'ADMIN';
  return 'USER';
});

const avatarText = computed(
  () => profile.value.avatarText || profile.value.username?.slice(0, 2).toUpperCase() || 'U',
);

/** 个人中心内容页签 */
const tabs = [
  { id: 'overview', label: '账户概览', no: '01' },
  { id: 'info', label: '个人信息', no: '02' },
  { id: 'security', label: '安全设置', no: '03' },
  { id: 'activity', label: '操作记录', no: '04' },
];
const activeTab = ref('overview');

function handleTabKeydown(event: KeyboardEvent, index: number) {
  let nextIndex = index;
  if (event.key === 'ArrowRight') nextIndex = (index + 1) % tabs.length;
  else if (event.key === 'ArrowLeft') {
    nextIndex = (index - 1 + tabs.length) % tabs.length;
  } else if (event.key === 'Home') nextIndex = 0;
  else if (event.key === 'End') nextIndex = tabs.length - 1;
  else return;

  event.preventDefault();
  const nextTab = tabs[nextIndex];
  if (!nextTab) return;
  activeTab.value = nextTab.id;
  nextTick(() => document.getElementById(`pf-tab-${nextTab.id}`)?.focus());
}

/** 个人信息表单（手机/邮箱留空 = 保持不变，placeholder 展示掩码） */
const infoFormRef = ref();
const infoForm = reactive({
  email: '',
  gender: '男',
  nickname: '',
  phone: '',
});

const infoRules = {
  email: [
    { message: '邮箱格式不正确', pattern: /^[\w.+-]+@[\w-]+\.[\w.]+$/ },
  ],
  nickname: [
    { message: '昵称为 2-20 位字符', min: 2, required: true },
    { max: 20, message: '昵称为 2-20 位字符' },
  ],
  phone: [
    { message: '手机号格式不正确', pattern: /^1[3-9]\d{9}$/ },
  ],
};

/** 密码表单 */
const pwdFormRef = ref();
const pwdForm = reactive({
  confirmPassword: '',
  newPassword: '',
  oldPassword: '',
});

function validateConfirm(_rule: unknown, value: string) {
  if (value && value !== pwdForm.newPassword) {
    return Promise.reject(new Error('两次输入的新密码不一致'));
  }
  return Promise.resolve();
}

const pwdRules = {
  confirmPassword: [
    { required: true, message: '请再次输入新密码' },
    { validator: validateConfirm },
  ],
  newPassword: [
    { min: 8, message: '新密码长度必须是 8-32 位', required: true },
    { max: 32, message: '新密码长度必须是 8-32 位' },
    {
      message: '新密码必须同时包含大写字母、小写字母与数字',
      pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,32}$/,
      required: true,
    },
  ],
  oldPassword: [{ message: '请输入原密码', required: true }],
};

/** 操作审计 */
const auditItems = ref<ProfileApi.ProfileAuditItem[]>([]);
const auditTotal = ref(0);
const auditPage = ref(1);
const auditPageSize = 10;
const auditLoading = ref(false);

const auditColumns = [
  { dataIndex: 'createdAt', title: '时间', width: 170 },
  { dataIndex: 'action', title: '动作', width: 90 },
  { dataIndex: 'module', title: '模块', width: 100 },
  { dataIndex: 'detail', ellipsis: true, title: '操作内容' },
  { dataIndex: 'ip', title: '来源 IP', width: 130 },
];

const ACTION_LABEL: Record<string, { color: string; text: string }> = {
  create: { color: 'green', text: '创建' },
  delete: { color: 'red', text: '删除' },
  update: { color: 'blue', text: '更新' },
};

async function loadProfile() {
  profile.value = await getProfileApi();
  userStore.setUserInfo({
    ...userStore.userInfo,
    avatar: userStore.userInfo?.avatar,
    realName: profile.value.realName,
  } as any);
}

async function loadStats() {
  const [centers, nodes, summary, audit] = await Promise.all([
    listCentersApi({ page: 1, pageSize: 1 }),
    listNodesApi({ page: 1, pageSize: 1 }),
    getPublishSummaryApi(),
    getProfileAuditApi({ page: 1, pageSize: 200 }),
  ]);
  stats.centers = centers.total;
  stats.nodes = nodes.total;
  stats.pending = summary.pendingCount;
  const weekAgo = Date.now() - 7 * 24 * 3600 * 1000;
  stats.changes7d = audit.items.filter((item) => {
    const ts = Date.parse(item.createdAt.replaceAll('-', '/'));
    return Number.isFinite(ts) && ts >= weekAgo;
  }).length;
}

async function loadAudit() {
  auditLoading.value = true;
  try {
    const result = await getProfileAuditApi({
      page: auditPage.value,
      pageSize: auditPageSize,
    });
    auditItems.value = result.items;
    auditTotal.value = result.total;
  } finally {
    auditLoading.value = false;
  }
}

function fillInfoForm() {
  infoForm.nickname = profile.value.nickname ?? profile.value.realName ?? '';
  infoForm.gender = profile.value.gender || '男';
  infoForm.phone = '';
  infoForm.email = '';
}

async function handleSaveProfile() {
  await infoFormRef.value?.validate();
  savingProfile.value = true;
  try {
    // 留空字段不提交（服务端语义：字段缺省 = 保持原值）
    const payload: Record<string, string> = {
      gender: infoForm.gender,
      nickname: infoForm.nickname.trim(),
    };
    if (infoForm.phone.trim()) payload.phone = infoForm.phone.trim();
    if (infoForm.email.trim()) payload.email = infoForm.email.trim();
    await updateProfileApi(payload);
    message.success('个人信息已更新');
    await loadProfile();
    fillInfoForm();
  } catch (error: any) {
    // 表单校验错误静默；其余错误由拦截器提示
    if (!error?.errorFields) {
      // eslint-disable-next-line no-console
    }
  } finally {
    savingProfile.value = false;
  }
}

async function handleChangePassword() {
  await pwdFormRef.value?.validate();
  savingPassword.value = true;
  try {
    await updatePasswordApi({
      newPassword: pwdForm.newPassword,
      oldPassword: pwdForm.oldPassword,
    });
    message.success('密码已修改，下次登录请使用新密码');
    pwdFormRef.value?.resetFields();
  } catch (error: any) {
    // 表单校验错误静默；其余错误由拦截器提示
    if (!error?.errorFields) {
      // eslint-disable-next-line no-console
    }
  } finally {
    savingPassword.value = false;
  }
}

function handleOtpToggle() {
  const target: 'bind' | 'disable' = profile.value.otpEnabled
    ? 'disable'
    : 'bind';
  otpBindModal.value?.open(target);
}

async function handleOtpSuccess() {
  await loadProfile();
  await loadAudit();
}

function handleAuditPageChange(page: number) {
  auditPage.value = page;
  loadAudit();
}

onMounted(async () => {
  try {
    await Promise.all([loadProfile(), loadStats()]);
    fillInfoForm();
    await loadAudit();
  } finally {
    pageLoading.value = false;
  }
});
</script>

<template>
  <Page>
    <div :class="['pf-page', { loading: pageLoading }]" data-profile-page>
      <nav class="pf-tabs" aria-label="个人中心内容" role="tablist" data-pf-tabs>
        <button
          v-for="(item, index) in tabs"
          :id="`pf-tab-${item.id}`"
          :key="item.id"
          :aria-controls="`pf-panel-${item.id}`"
          :aria-selected="activeTab === item.id"
          :class="['pf-tab', { active: activeTab === item.id }]"
          :tabindex="activeTab === item.id ? 0 : -1"
          role="tab"
          type="button"
          @click="activeTab = item.id"
          @keydown="handleTabKeydown($event, index)"
        >
          <span class="pf-mono pf-tab-no">{{ item.no }}</span>
          <span>{{ item.label }}</span>
        </button>
      </nav>

      <!-- ============ 01 账户概览 ============ -->
      <section
        id="pf-panel-overview"
        v-show="activeTab === 'overview'"
        aria-labelledby="pf-tab-overview"
        class="pf-grid-bg"
        data-pf-overview
        role="tabpanel"
        tabindex="0"
      >
        <div class="pf-mono pf-fig">FIG 1.1 · ACCOUNT SCALE 1:1</div>
        <div class="pf-overview-head">
          <div class="pf-overview-id">
            <div class="pf-avatar" data-pf-avatar>{{ avatarText }}</div>
            <div class="pf-overview-meta">
              <div class="pf-name-row">
                <h1 class="pf-name">{{ profile.realName || profile.username }}</h1>
                <span class="pf-badge pf-badge-role" data-pf-role>{{ roleLabel }}</span>
                <span
                  :class="[
                    'pf-badge',
                    profile.otpEnabled ? 'pf-badge-otp-on' : 'pf-badge-otp-off',
                  ]"
                  data-pf-otp-badge
                >
                  OTP.{{ profile.otpEnabled ? 'ENABLED' : 'DISABLED' }}
                </span>
              </div>
              <p class="pf-mono pf-sub">
                ACC {{ profile.username }} · {{ profile.phone || 'PHONE 未设置' }} ·
                {{ profile.email || 'EMAIL 未设置' }}
              </p>
              <p class="pf-mono pf-sub">
                LAST LOGIN {{ profile.lastLoginAt || '-' }} · CREATED
                {{ profile.createdAt || '-' }}
              </p>
            </div>
          </div>
          <p class="pf-overview-desc">
            个人信息、安全凭据与业务数据在同一坐标系下管理；敏感操作要求二次确认，账户变更自动进入审计留档。
          </p>
        </div>

        <dl class="pf-stats" data-pf-stats>
          <div class="pf-stat">
            <dt class="pf-mono pf-stat-label">MY CENTERS / 负责中心</dt>
            <dd class="pf-mono pf-stat-value" data-pf-stat-centers>
              {{ String(stats.centers).padStart(2, '0') }}
            </dd>
          </div>
          <div class="pf-stat">
            <dt class="pf-mono pf-stat-label">MY NODES / 纳管节点</dt>
            <dd class="pf-mono pf-stat-value" data-pf-stat-nodes>
              {{ String(stats.nodes).padStart(2, '0') }}
            </dd>
          </div>
          <div class="pf-stat">
            <dt class="pf-mono pf-stat-label">CHANGES / 近 7 日变更</dt>
            <dd class="pf-mono pf-stat-value" data-pf-stat-changes>
              {{ String(stats.changes7d).padStart(2, '0') }}
            </dd>
          </div>
          <div class="pf-stat">
            <dt class="pf-mono pf-stat-label">PENDING / 待下发变更</dt>
            <dd class="pf-mono pf-stat-value" data-pf-stat-pending>
              {{ String(stats.pending).padStart(2, '0') }}
            </dd>
          </div>
        </dl>
      </section>

      <!-- ============ 02 个人信息 ============ -->
      <section
        id="pf-panel-info"
        v-show="activeTab === 'info'"
        aria-labelledby="pf-tab-info"
        class="pf-section"
        data-pf-info
        role="tabpanel"
        tabindex="0"
      >
        <div class="pf-section-head">
          <div class="pf-mono pf-fig">FIG 2.1 · PERSONAL INFO</div>
          <h2 class="pf-section-title">个人信息</h2>
        </div>
        <Form
          ref="infoFormRef"
          :model="infoForm"
          :rules="infoRules"
          class="pf-form"
          layout="vertical"
        >
          <div class="pf-form-grid">
            <FormItem label="昵称" name="nickname">
              <Input
                v-model:value="infoForm.nickname"
                :maxlength="20"
                data-pf-input-nickname
                placeholder="2-20 位字符"
              />
            </FormItem>
            <FormItem label="性别" name="gender">
              <RadioGroup v-model:value="infoForm.gender" data-pf-gender>
                <RadioButton value="男">男</RadioButton>
                <RadioButton value="女">女</RadioButton>
                <RadioButton value="保密">保密</RadioButton>
              </RadioGroup>
            </FormItem>
            <FormItem label="手机号" name="phone">
              <Input
                v-model:value="infoForm.phone"
                :placeholder="`当前：${profile.phone || '未设置'}，输入新号码以更换`"
                data-pf-input-phone
              />
            </FormItem>
            <FormItem label="邮箱" name="email">
              <Input
                v-model:value="infoForm.email"
                :placeholder="`当前：${profile.email || '未设置'}，输入新邮箱以更换`"
                data-pf-input-email
              />
            </FormItem>
          </div>
          <div class="pf-form-actions">
            <Button
              :loading="savingProfile"
              data-pf-save-info
              type="primary"
              @click="handleSaveProfile"
            >
              保存修改
            </Button>
            <span class="pf-mono pf-form-hint">留空项保持原值不变</span>
          </div>
        </Form>
      </section>

      <!-- ============ 03 安全设置 ============ -->
      <section
        id="pf-panel-security"
        v-show="activeTab === 'security'"
        aria-labelledby="pf-tab-security"
        class="pf-section"
        data-pf-security
        role="tabpanel"
        tabindex="0"
      >
        <div class="pf-section-head">
          <div class="pf-mono pf-fig">FIG 3.1 · SECURITY</div>
          <h2 class="pf-section-title">安全设置</h2>
        </div>

        <div class="pf-security-grid">
          <!-- 密码修改 -->
          <div class="pf-card" data-pf-password-card>
            <div class="pf-card-title">登录密码</div>
            <p class="pf-card-desc">
              密码要求 8-32 位，且同时包含大写字母、小写字母与数字；修改成功后下次登录生效
            </p>
            <Form
              ref="pwdFormRef"
              :model="pwdForm"
              :rules="pwdRules"
              class="pf-form"
              layout="vertical"
            >
              <FormItem label="原密码" name="oldPassword">
                <InputPassword
                  v-model:value="pwdForm.oldPassword"
                  autocomplete="current-password"
                  data-pf-input-oldpwd
                />
              </FormItem>
              <FormItem label="新密码" name="newPassword">
                <InputPassword
                  v-model:value="pwdForm.newPassword"
                  autocomplete="new-password"
                  data-pf-input-newpwd
                  placeholder="8-32 位，含大小写字母与数字"
                />
              </FormItem>
              <FormItem label="确认新密码" name="confirmPassword">
                <InputPassword
                  v-model:value="pwdForm.confirmPassword"
                  autocomplete="new-password"
                  data-pf-input-confirmpwd
                />
              </FormItem>
              <Button
                :loading="savingPassword"
                data-pf-save-password
                type="primary"
                @click="handleChangePassword"
              >
                修改密码
              </Button>
            </Form>
          </div>

          <!-- OTP 两步验证 -->
          <div class="pf-card" data-pf-otp-card>
            <div class="pf-card-title">
              OTP 两步验证
              <span class="pf-otp-state">
                <span
                  :class="['pf-live-dot', { off: !profile.otpEnabled }]"
                  data-pf-otp-dot
                ></span>
                <span class="pf-mono" data-pf-otp-state-text>
                  OTP.{{ profile.otpEnabled ? 'ENABLED · TOTP' : 'DISABLED' }}
                </span>
              </span>
            </div>
            <p class="pf-card-desc">
              开启后，登录时需输入验证器应用（TOTP）生成的 6 位动态口令，为账户增加第二道防线
            </p>
            <ul class="pf-otp-list">
              <li>口令每 30 秒轮换，单次有效</li>
              <li>绑定过程需扫描二维码或手动输入 Base32 密钥</li>
              <li>关闭两步验证需通过动态口令二次确认</li>
            </ul>
            <Button
              :danger="profile.otpEnabled"
              :type="profile.otpEnabled ? 'default' : 'primary'"
              data-pf-otp-toggle
              @click="handleOtpToggle"
            >
              {{ profile.otpEnabled ? '关闭两步验证' : '开启两步验证' }}
            </Button>
          </div>
        </div>
      </section>

      <!-- ============ 04 操作记录 ============ -->
      <section
        id="pf-panel-activity"
        v-show="activeTab === 'activity'"
        aria-labelledby="pf-tab-activity"
        class="pf-section"
        data-pf-activity
        role="tabpanel"
        tabindex="0"
      >
        <div class="pf-section-head">
          <div class="pf-mono pf-fig">FIG 4.1 · ACTIVITY LOG</div>
          <h2 class="pf-section-title">操作记录</h2>
          <span class="pf-mono pf-total">TOTAL {{ auditTotal }}</span>
        </div>
        <div class="pf-table-wrap">
          <Table
            :columns="auditColumns"
            :data-source="auditItems"
            :loading="auditLoading"
            :pagination="false"
            :row-key="(record: any) => record.id"
            data-pf-audit-table
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'action'">
                <Tag :color="ACTION_LABEL[record.action]?.color ?? 'default'">
                  {{ ACTION_LABEL[record.action]?.text ?? record.action }}
                </Tag>
              </template>
              <template v-else-if="column.dataIndex === 'createdAt'">
                <span class="pf-mono">{{ record.createdAt }}</span>
              </template>
              <template v-else-if="column.dataIndex === 'detail'">
                <Tooltip :title="record.detail">
                  <span class="pf-detail">{{ record.detail }}</span>
                </Tooltip>
              </template>
            </template>
          </Table>
        </div>
        <div class="pf-pagination">
          <Pagination
            :current="auditPage"
            :page-size="auditPageSize"
            :show-size-changer="false"
            :total="auditTotal"
            data-pf-audit-pagination
            @change="handleAuditPageChange"
          />
        </div>
      </section>

      <!-- OTP 绑定 / 关闭弹窗 -->
      <OtpBindModal ref="otpBindModal" @success="handleOtpSuccess" />
    </div>
  </Page>
</template>

<style scoped>
/* ===== 设计签名：等宽数据层 ===== */
.pf-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-variant-numeric: tabular-nums;
}

.pf-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.pf-page.loading {
  opacity: 0.6;
}

/* ===== 01 概览：网格背景 + 身份行 ===== */
.pf-grid-bg {
  position: relative;
  padding: 28px 28px 24px;
  overflow: hidden;
  background-color: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
  background-image:
    linear-gradient(hsl(var(--primary) / 0.05) 1px, transparent 1px),
    linear-gradient(90deg, hsl(var(--primary) / 0.05) 1px, transparent 1px);
  background-size: 24px 24px;
}

.pf-fig {
  font-size: 11px;
  letter-spacing: 0.18em;
  color: hsl(var(--primary));
  text-transform: uppercase;
}

.pf-overview-head {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-end;
  justify-content: space-between;
  margin-top: 14px;
}

.pf-overview-id {
  display: flex;
  gap: 20px;
  align-items: center;
}

.pf-avatar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 24px;
  font-weight: 700;
  color: hsl(var(--primary));
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
}

.pf-name-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}

.pf-name {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  line-height: 1.2;
  color: hsl(var(--foreground));
}

.pf-badge {
  padding: 2px 8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 11px;
  border-radius: var(--radius);
}

.pf-badge-role {
  color: hsl(var(--primary));
  border: 1px solid hsl(var(--primary) / 0.4);
}

.pf-badge-otp-on {
  color: hsl(var(--primary));
  border: 1px solid hsl(var(--primary) / 0.4);
}

.pf-badge-otp-off {
  color: hsl(var(--muted-foreground));
  border: 1px solid hsl(var(--border));
}

.pf-sub {
  margin: 6px 0 0;
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}

.pf-overview-desc {
  max-width: 360px;
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: hsl(var(--muted-foreground));
  text-align: left;
}

/* 统计卡：1px 缝隙网格 */
.pf-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1px;
  margin: 22px 0 0;
  background: hsl(var(--border));
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
}

@media (min-width: 768px) {
  .pf-stats {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.pf-stat {
  padding: 18px;
  background: hsl(var(--card));
}

.pf-stat-label {
  font-size: 11px;
  letter-spacing: 0.14em;
  color: hsl(var(--muted-foreground));
}

.pf-stat-value {
  margin: 10px 0 0;
  font-size: 36px;
  font-weight: 700;
  line-height: 1;
  color: hsl(var(--foreground));
}

/* ===== 内容页签 ===== */
.pf-tabs {
  display: flex;
  gap: 4px;
  overflow-x: auto;
  padding: 6px;
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
}

.pf-tab {
  display: flex;
  flex: 1 0 auto;
  gap: 8px;
  align-items: center;
  justify-content: center;
  min-height: 40px;
  padding: 0 14px;
  font-size: 13px;
  color: hsl(var(--muted-foreground));
  cursor: pointer;
  user-select: none;
  background: transparent;
  border: 0;
  border-bottom: 2px solid transparent;
  border-radius: var(--radius);
  transition:
    color 0.15s ease,
    background 0.15s ease,
    border-color 0.15s ease;
}

.pf-tab:hover {
  color: hsl(var(--foreground));
  background: hsl(var(--accent) / 0.08);
}

.pf-tab.active {
  color: hsl(var(--foreground));
  background: hsl(var(--secondary));
  border-bottom-color: hsl(var(--primary));
}

.pf-tab:focus-visible {
  outline: 2px solid hsl(var(--primary));
  outline-offset: 2px;
}

.pf-tab-no {
  font-size: 11px;
  color: hsl(var(--primary));
}

.pf-tab:not(.active) .pf-tab-no {
  color: hsl(var(--muted-foreground) / 0.7);
}

/* ===== 区块通用 ===== */
.pf-section {
  padding: 24px 28px;
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
}

.pf-section-head {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: baseline;
  margin-bottom: 18px;
}

.pf-section-title {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: hsl(var(--foreground));
}

.pf-total {
  font-size: 11px;
  color: hsl(var(--muted-foreground));
}

/* ===== 表单 ===== */
.pf-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 24px;
}

@media (max-width: 767px) {
  .pf-form-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

.pf-form-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.pf-form-hint {
  font-size: 11px;
  color: hsl(var(--muted-foreground) / 0.8);
}

/* ===== 安全设置 ===== */
.pf-security-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

@media (max-width: 767px) {
  .pf-security-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

.pf-card {
  padding: 18px;
  background: hsl(var(--background));
  border: 1px solid hsl(var(--border));
  border-radius: var(--radius);
}

.pf-card-title {
  display: flex;
  gap: 10px;
  align-items: center;
  font-size: 14px;
  font-weight: 600;
  color: hsl(var(--foreground));
}

.pf-card-desc {
  margin: 8px 0 16px;
  font-size: 12px;
  line-height: 1.7;
  color: hsl(var(--muted-foreground));
  text-align: left;
}

/* 状态灯 */
.pf-otp-state {
  display: flex;
  gap: 6px;
  align-items: center;
  margin-left: auto;
}

.pf-live-dot {
  position: relative;
  width: 8px;
  height: 8px;
  overflow: hidden;
  background: hsl(var(--primary));
  border-radius: 9999px;
}

.pf-live-dot::after {
  position: absolute;
  inset: 0;
  content: '';
  background: hsl(var(--primary));
  border-radius: inherit;
  animation: pf-pulse 2s ease-in-out infinite;
}

.pf-live-dot.off {
  background: hsl(var(--muted-foreground) / 0.4);
}

.pf-live-dot.off::after {
  display: none;
}

@keyframes pf-pulse {
  0%,
  100% {
    opacity: 0.4;
  }

  50% {
    opacity: 1;
  }
}

.pf-otp-list {
  margin: 0 0 16px;
  padding-left: 18px;
  font-size: 12px;
  line-height: 2;
  color: hsl(var(--muted-foreground));
  text-align: left;
}

/* ===== 操作记录 ===== */
.pf-table-wrap {
  min-width: 0;
  overflow-x: auto;
  background: hsl(var(--card));
}

.pf-detail {
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}

.pf-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 14px;
}
</style>
