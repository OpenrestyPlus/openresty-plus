<script lang="ts" setup>
import type { OrpSettingGroup } from '#/api';

import { computed, reactive, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { HolderOutlined } from '@ant-design/icons-vue';
import {
  Alert,
  Button,
  Checkbox,
  Input,
  InputNumber,
  Modal as AntModal,
  Radio,
  RadioGroup,
  Select,
  Tag,
  Textarea,
  Tooltip,
  message,
} from 'ant-design-vue';

import {
  createSettingGroupApi,
  deleteSettingGroupApi,
  listSettingGroupEntitiesApi,
  reorderSettingGroupsApi,
  updateSettingGroupApi,
} from '#/api';
import { useDirtyGuard } from '#/composables/use-dirty-guard';

defineOptions({ name: 'SettingGroupModal' });

const emit = defineEmits<{ change: [] }>();

/** 视图：list 分组列表 / form 新增编辑 / danger 删除处理 */
type ViewMode = 'danger' | 'form' | 'list';
const view = ref<ViewMode>('list');

const groups = ref<OrpSettingGroup[]>([]);
const saving = ref(false);

/** 拖拽起点 / 悬停行号（视觉提示） */
const dragIndex = ref(-1);
const dragOverIndex = ref(-1);

/** 表单编辑目标（undefined = 新增） */
const formTarget = ref<OrpSettingGroup | undefined>();
const form = reactive({
  code: '',
  description: '',
  name: '',
  sortOrder: 1,
});
const errors = reactive<{ code?: string; name?: string }>({});
let formSnapshot = { ...form };

/** 删除处理目标（组内含配置项时进入 danger 视图） */
const dangerTarget = ref<OrpSettingGroup>();
const dangerMode = ref<'cascade' | 'move'>('move');
const dangerTargetGroup = ref<undefined | string>(undefined);
const riskConfirmed = ref(false);

const { markDirty, onBeforeClose, resetDirty } = useDirtyGuard();

const isEditForm = computed(() => Boolean(formTarget.value));
const isSystemForm = computed(() => formTarget.value?.isSystem === true);

/** 转移目标分组选项（排除待删除分组自身，防止自循环） */
const moveOptions = computed(() =>
  groups.value
    .filter((g) => g.id !== dangerTarget.value?.id)
    .map((g) => ({ label: g.name, value: g.name })),
);

/** 确认删除按钮可用条件：move 需选目标分组；cascade 需勾选风险确认 */
const dangerOkDisabled = computed(() =>
  dangerMode.value === 'move'
    ? !dangerTargetGroup.value
    : !riskConfirmed.value,
);

async function loadGroups() {
  groups.value = await listSettingGroupEntitiesApi();
}

function backToList() {
  view.value = 'list';
  resetDirty();
}

function snapshotForm() {
  formSnapshot = { ...form };
}

function formDirty(): boolean {
  return JSON.stringify({ ...form }) !== JSON.stringify(formSnapshot);
}

/** 表单返回列表：有未保存修改时二次确认 */
function onFormBack() {
  if (formDirty()) {
    AntModal.confirm({
      content: '当前分组编辑内容尚未保存，确认放弃并返回列表？',
      okButtonProps: { danger: true },
      okText: '放弃修改',
      onOk: () => backToList(),
      title: '未保存的修改',
    });
    return;
  }
  backToList();
}

function onAdd() {
  formTarget.value = undefined;
  Object.assign(form, {
    code: '',
    description: '',
    name: '',
    sortOrder: groups.value.length + 1,
  });
  errors.code = undefined;
  errors.name = undefined;
  snapshotForm();
  view.value = 'form';
  resetDirty();
}

function onEdit(g: OrpSettingGroup) {
  formTarget.value = g;
  Object.assign(form, {
    code: g.code,
    description: g.description,
    name: g.name,
    sortOrder: g.sortOrder,
  });
  errors.code = undefined;
  errors.name = undefined;
  snapshotForm();
  view.value = 'form';
  resetDirty();
}

function validateForm(): boolean {
  errors.code = undefined;
  errors.name = undefined;
  const name = form.name.trim();
  if (!name) {
    errors.name = '分组名称为必填项';
  } else if (name.length < 2 || name.length > 32) {
    errors.name = '分组名称长度必须是 2-32 位';
  }
  if (!isEditForm.value) {
    const code = form.code.trim();
    if (!/^[a-z][a-z0-9_]{1,30}$/.test(code)) {
      errors.code =
        '分组编码需以小写字母开头，2-31 位小写字母/数字/下划线（如 custom_alert）';
    }
  }
  return !errors.code && !errors.name;
}

async function onFormSave() {
  if (!validateForm()) return;
  saving.value = true;
  try {
    const payload = {
      description: form.description.trim(),
      name: form.name.trim(),
      sortOrder: Math.max(1, Math.round(form.sortOrder) || 1),
      ...(isEditForm.value ? {} : { code: form.code.trim() }),
    };
    if (isEditForm.value && formTarget.value) {
      await updateSettingGroupApi(formTarget.value.id, payload);
      message.success('分组已更新，配置页导航展示已同步刷新');
    } else {
      await createSettingGroupApi(payload);
      message.success(`自定义分组「${payload.name}」创建成功`);
    }
    resetDirty();
    await loadGroups();
    emit('change');
    backToList();
  } catch {
    // 错误提示由请求拦截器统一处理
  } finally {
    saving.value = false;
  }
}

/** 列表内直接编辑排序序号（即时生效） */
async function onSortChange(g: OrpSettingGroup, value: number | string | null) {
  const next = Math.round(Number(value));
  if (!Number.isFinite(next) || next < 1 || next === g.sortOrder) return;
  try {
    await updateSettingGroupApi(g.id, {
      description: g.description,
      name: g.name,
      sortOrder: next,
    });
    message.success(`分组「${g.name}」排序序号已调整为 ${next}`);
    await loadGroups();
    emit('change');
  } catch {
    await loadGroups();
  }
}

/** 拖拽释放：本地乐观重排 + 按新顺序重编序号提交 */
async function onDrop(index: number) {
  const from = dragIndex.value;
  dragIndex.value = -1;
  dragOverIndex.value = -1;
  if (from < 0 || from === index) return;
  const list = [...groups.value];
  const [moved] = list.splice(from, 1);
  if (!moved) return;
  list.splice(index, 0, moved);
  groups.value = list;
  try {
    await reorderSettingGroupsApi(list.map((g) => g.id));
    message.success(`分组排序已保存：「${moved?.name}」移至第 ${index + 1} 位`);
    await loadGroups();
    emit('change');
  } catch {
    await loadGroups();
  }
}

/** 删除：空分组直接确认；含配置项进入处理视图（转移 / 级联） */
function onDelete(g: OrpSettingGroup) {
  if (g.isSystem) return;
  if (g.settingCount === 0) {
    AntModal.confirm({
      content: '该分组下暂无配置项，删除后将从配置页导航中移除。',
      okButtonProps: { danger: true },
      okText: '确认删除',
      onOk: async () => {
        try {
          await deleteSettingGroupApi(g.id);
          message.success(`分组「${g.name}」已删除`);
          await loadGroups();
          emit('change');
        } catch {
          // 错误提示由请求拦截器统一处理
        }
      },
      title: `确认删除分组「${g.name}」？`,
    });
    return;
  }
  dangerTarget.value = g;
  dangerMode.value = 'move';
  dangerTargetGroup.value = undefined;
  riskConfirmed.value = false;
  view.value = 'danger';
  resetDirty();
}

async function onDangerOk() {
  const g = dangerTarget.value;
  if (!g) return;
  saving.value = true;
  try {
    if (dangerMode.value === 'move') {
      await deleteSettingGroupApi(g.id, {
        mode: 'move',
        targetGroup: dangerTargetGroup.value,
      });
      message.success(
        `组内 ${g.settingCount} 个配置项已转移至「${dangerTargetGroup.value}」，分组已删除`,
      );
    } else {
      await deleteSettingGroupApi(g.id, {
        mode: 'cascade',
        riskConfirmed: true,
      });
      message.success(
        `分组「${g.name}」及组内 ${g.settingCount} 个配置项已级联删除`,
      );
    }
    backToList();
    await loadGroups();
    emit('change');
  } catch {
    // 错误提示由请求拦截器统一处理
  } finally {
    saving.value = false;
  }
}

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onBeforeClose,
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      modalApi.setState({ title: '分组管理' });
      view.value = 'list';
      try {
        await loadGroups();
      } catch {
        // 错误提示由请求拦截器统一处理
      }
      resetDirty();
    }
  },
});
</script>

<template>
  <Modal class="w-[520px]">
    <!-- 分组列表视图 -->
    <div v-if="view === 'list'">
      <div class="mb-3 flex items-center justify-between gap-2">
        <span class="text-muted-foreground text-xs">
          共 {{ groups.length }} 个分组（系统预置
          {{ groups.filter((g) => g.isSystem).length }} / 自定义
          {{ groups.filter((g) => !g.isSystem).length }}）；拖拽把手或编辑序号调整展示顺序
        </span>
        <Button size="small" type="primary" @click="onAdd">新增分组</Button>
      </div>
      <div class="max-h-[420px] space-y-2 overflow-y-auto pr-1">
        <div
          v-for="(g, i) in groups"
          :key="g.id"
          class="flex items-center gap-2 rounded-md border border-border px-3 py-2"
          :class="
            dragIndex >= 0 && dragOverIndex === i && dragIndex !== i
              ? 'border-primary bg-accent'
              : ''
          "
          @dragover="
            (e) => {
              e.preventDefault();
              dragOverIndex = i;
            }
          "
          @drop="onDrop(i)"
        >
          <span
            class="text-muted-foreground cursor-move select-none"
            draggable="true"
            title="拖拽调整排序"
            @dragend="
              dragIndex = -1;
              dragOverIndex = -1;
            "
            @dragstart="dragIndex = i"
          >
            <HolderOutlined />
          </span>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="truncate text-sm font-medium">{{ g.name }}</span>
              <Tag :color="g.isSystem ? 'blue' : 'purple'" class="text-xs">
                {{ g.isSystem ? '系统内置' : '自定义' }}
              </Tag>
            </div>
            <div class="text-muted-foreground truncate font-mono text-xs">
              {{ g.code }} · {{ g.settingCount }} 个配置项
            </div>
          </div>
          <InputNumber
            :min="1"
            :value="g.sortOrder"
            size="small"
            style="width: 72px"
            title="调整排序序号（越小越靠前，即时生效）"
            @change="(v) => onSortChange(g, v)"
          />
          <Button size="small" type="link" @click="onEdit(g)">编辑</Button>
          <Tooltip v-if="g.isSystem" title="系统预置分组禁止删除">
            <Button disabled size="small" type="link">删除</Button>
          </Tooltip>
          <Button
            v-else
            danger
            size="small"
            type="link"
            @click="onDelete(g)"
          >
            删除
          </Button>
        </div>
      </div>
    </div>

    <!-- 新增 / 编辑表单视图 -->
    <div v-else-if="view === 'form'" class="space-y-4">
      <Alert
        v-if="isSystemForm"
        message="系统预置分组：名称与编码锁定不可修改，仅支持调整排序序号与描述"
        show-icon
        type="info"
      />
      <div>
        <div class="mb-1 text-sm">
          <span class="text-red-500">*</span> 分组名称
        </div>
        <Input
          v-model:value="form.name"
          :disabled="isSystemForm"
          placeholder="例如：告警集成配置"
          @input="markDirty"
        />
        <div v-if="errors.name" class="mt-1 text-xs text-red-500">
          {{ errors.name }}
        </div>
      </div>
      <div>
        <div class="mb-1 text-sm">
          <span v-if="!isEditForm" class="text-red-500">*</span> 分组编码
        </div>
        <Input
          v-model:value="form.code"
          :disabled="isEditForm"
          placeholder="例如：custom_alert"
          @input="markDirty"
        />
        <div v-if="errors.code" class="mt-1 text-xs text-red-500">
          {{ errors.code }}
        </div>
        <div v-else class="text-muted-foreground mt-1 text-xs">
          小写字母开头，2-31 位小写字母/数字/下划线，全局唯一，创建后不可修改
        </div>
      </div>
      <div>
        <div class="mb-1 text-sm">排序序号</div>
        <InputNumber
          v-model:value="form.sortOrder"
          :min="1"
          class="w-full"
          @change="markDirty"
        />
        <div class="text-muted-foreground mt-1 text-xs">
          数值越小在配置页导航中越靠前，保存后即时生效
        </div>
      </div>
      <div>
        <div class="mb-1 text-sm">描述说明</div>
        <Textarea
          v-model:value="form.description"
          :rows="2"
          placeholder="分组用途说明，例如：第三方告警平台对接参数"
          @input="markDirty"
        />
      </div>
    </div>

    <!-- 删除处理视图（组内含配置项） -->
    <div v-else class="space-y-4">
      <Alert
        :message="`分组「${dangerTarget?.name}」下包含 ${dangerTarget?.settingCount} 个配置项，删除前请选择处理方式`"
        show-icon
        type="warning"
      />
      <RadioGroup v-model:value="dangerMode" class="w-full">
        <div class="space-y-3">
          <Radio value="move">移动到其他分组并删除</Radio>
          <div v-if="dangerMode === 'move'" class="pl-6">
            <Select
              v-model:value="dangerTargetGroup"
              :options="moveOptions"
              placeholder="选择目标分组（组内全部配置项将转移至此）"
              show-search
              @change="markDirty"
            />
          </div>
          <Radio value="cascade">级联删除分组及组内全部配置项</Radio>
        </div>
      </RadioGroup>
      <template v-if="dangerMode === 'cascade'">
        <Alert
          :message="`警告：将同步永久删除该分组内的全部 ${dangerTarget?.settingCount} 个配置项，此操作不可逆`"
          show-icon
          type="error"
        />
        <Checkbox v-model:checked="riskConfirmed" @change="markDirty">
          我已知晓风险，确认级联删除
        </Checkbox>
      </template>
    </div>

    <template #footer>
      <Button v-if="view === 'list'" @click="modalApi.close()">关闭</Button>
      <template v-else-if="view === 'form'">
        <Button :disabled="saving" @click="onFormBack">返回列表</Button>
        <Button :loading="saving" type="primary" @click="onFormSave">
          保存
        </Button>
      </template>
      <template v-else>
        <Button :disabled="saving" @click="backToList">取消</Button>
        <Button
          :danger="dangerMode === 'cascade'"
          :disabled="dangerOkDisabled"
          :loading="saving"
          type="primary"
          @click="onDangerOk"
        >
          确认删除
        </Button>
      </template>
    </template>
  </Modal>
</template>
