<script lang="ts" setup>
import type { OrpIpGroup } from '#/api';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import { createIpGroupApi, updateIpGroupApi } from '#/api';
import { useDirtyGuard } from '#/composables/use-dirty-guard';

defineOptions({ name: 'IpGroupFormModal' });

const emit = defineEmits<{ success: [] }>();

const isEdit = ref(false);
const editId = ref(0);

const { markDirty, onBeforeClose, resetDirty } = useDirtyGuard();

/** IP / CIDR 格式（IPv4 段 0-255、掩码 /0-/32；兼容 IPv6 及 /0-/128） */
const IPV4_CIDR_PATTERN =
  /^(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.){3}(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\/(?:3[0-2]|[12]?\d))?$/;
const IPV6_CIDR_PATTERN =
  /^(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}(?:\/(?:12[0-8]|1[01]\d|[0-9]?\d))?$/;

const [Form, formApi] = useVbenForm({
  handleValuesChange: () => markDirty(),
  handleSubmit: onSubmit,
  schema: [
    {
      component: 'Input',
      componentProps: { placeholder: '例如：Office_Internal_Net' },
      fieldName: 'name',
      label: '组名称',
      rules: z
        .string()
        .min(2, { message: '组名称长度 2-64 位' })
        .max(64, { message: '组名称长度 2-64 位' }),
    },
    {
      component: 'Textarea',
      componentProps: {
        placeholder: '用途说明，例如：总部与分支办公内网网段',
        rows: 2,
      },
      fieldName: 'description',
      label: '描述',
    },
    {
      component: 'Select',
      componentProps: {
        mode: 'tags',
        open: false,
        placeholder:
          '输入 IP 或 CIDR 网段后回车添加，支持逗号/空格分隔批量录入',
        tokenSeparators: [',', ' '],
      },
      defaultValue: [],
      fieldName: 'members',
      help: '支持 IPv4 / IPv6 与 CIDR 网段（如 192.168.1.10、10.0.0.0/8），回车确认、标签可删',
      label: '成员 IP / CIDR',
      rules: z
        .array(z.string())
        .min(1, { message: '至少添加一个 IP 或 CIDR 网段' })
        .refine(
          (arr: string[]) =>
            arr.every(
              (item: string) =>
                IPV4_CIDR_PATTERN.test(item) || IPV6_CIDR_PATTERN.test(item),
            ),
          {
            message:
              '存在不合法的 IP 地址或 CIDR 网段（示例：192.168.1.10 / 10.0.0.0/8）',
          },
        ),
    },
  ],
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal<{ record?: OrpIpGroup }>({
  fullscreenButton: false,
  onBeforeClose,
  async onConfirm() {
    await formApi.validateAndSubmitForm();
  },
  async onOpenChange(isOpen: boolean) {
    if (isOpen) {
      await formApi.resetForm();
      const record = modalApi.getData()?.record;
      isEdit.value = Boolean(record);
      editId.value = record?.id ?? 0;
      modalApi.setState({ title: record ? '编辑 IP 组' : '新增 IP 组' });
      if (record) {
        formApi.setValues({
          description: record.description,
          members: [...record.members],
          name: record.name,
        });
      }
      resetDirty();
    }
  },
});

async function onSubmit(values: Record<string, any>) {
  modalApi.lock();
  try {
    const payload = {
      description: String(values.description ?? '').trim(),
      members: values.members
        .map((item: unknown) => String(item).trim())
        .filter(Boolean),
      name: String(values.name).trim(),
    };
    if (isEdit.value) {
      await updateIpGroupApi(editId.value, payload);
      message.success('IP 组已更新');
    } else {
      await createIpGroupApi(payload);
      message.success('IP 组创建成功');
    }
    resetDirty();
    modalApi.close();
    emit('success');
  } catch {
    // 错误提示由请求拦截器统一处理
  } finally {
    modalApi.unlock();
  }
}
</script>

<template>
  <Modal class="w-[680px]">
    <Form />
  </Modal>
</template>

<style scoped>
/* 成员 IP/CIDR 录入区：舒展的大宽幅标签输入框，多标签换行排版不拥挤 */
:deep(.ant-select-selector) {
  min-height: 96px;
  padding: 6px 10px;
}

:deep(.ant-select-selection-item) {
  background-color: hsl(var(--primary) / 8%);
  border-color: hsl(var(--primary) / 25%);
  border-radius: 6px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  line-height: 22px;
}
</style>
