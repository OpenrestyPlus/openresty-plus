<script lang="ts" setup>
import type { Center } from '#/api';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import { createCenterApi, updateCenterApi } from '#/api';
import { useDirtyGuard } from '#/composables/use-dirty-guard';

defineOptions({ name: 'CenterFormModal' });

const emit = defineEmits<{ success: [] }>();

const isEdit = ref(false);
const editId = ref(0);

const { markDirty, onBeforeClose, resetDirty } = useDirtyGuard();

const [Form, formApi] = useVbenForm({
  handleValuesChange: () => markDirty(),
  handleSubmit: onSubmit,
  schema: [
    {
      component: 'Input',
      componentProps: { maxLength: 64, placeholder: '例如：华东生产中心' },
      fieldName: 'name',
      label: '中心名称',
      rules: z.string().min(1, { message: '请输入中心名称' }),
    },
    {
      component: 'Input',
      componentProps: {
        maxLength: 32,
        placeholder: '例如：cn-east-1',
      },
      fieldName: 'code',
      help: '唯一标识码，仅允许小写字母、数字与中划线，创建后不可修改',
      label: '标识码',
      rules: z
        .string()
        .min(2, { message: '标识码长度 2-32' })
        .regex(/^[a-z0-9-]{2,32}$/, {
          message: '仅允许小写字母、数字与中划线',
        }),
    },
    {
      component: 'Textarea',
      componentProps: {
        maxLength: 200,
        placeholder: '中心的用途、归属等备注描述（可选）',
        rows: 3,
        showCount: true,
      },
      fieldName: 'description',
      label: '备注描述',
    },
    {
      component: 'InputNumber',
      componentProps: { min: -90, max: 90, precision: 6, placeholder: '例如：31.2304', style: { width: '100%' } },
      fieldName: 'latitude',
      help: '与经度同时填写后，大屏可标出该中心位置',
      label: '纬度（可选）',
    },
    {
      component: 'InputNumber',
      componentProps: { min: -180, max: 180, precision: 6, placeholder: '例如：121.4737', style: { width: '100%' } },
      fieldName: 'longitude',
      label: '经度（可选）',
    },
  ],
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal<{ record?: Center }>({
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
      await formApi.updateSchema([
        { componentProps: { disabled: isEdit.value }, fieldName: 'code' },
      ]);
      modalApi.setState({ title: record ? '编辑中心' : '新建中心' });
      if (record) {
        formApi.setValues({
          code: record.code,
          description: record.description,
          latitude: record.latitude,
          longitude: record.longitude,
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
    const latitude = values.latitude === '' || values.latitude == null ? null : Number(values.latitude);
    const longitude = values.longitude === '' || values.longitude == null ? null : Number(values.longitude);
    if ((latitude == null) !== (longitude == null)) {
      message.error('纬度和经度需要同时填写');
      return;
    }
    const payload = {
      code: String(values.code),
      description: String(values.description ?? ''),
      latitude,
      longitude,
      name: String(values.name),
    };
    if (isEdit.value) {
      await updateCenterApi(editId.value, payload);
      message.success('中心已更新');
    } else {
      await createCenterApi(payload);
      message.success('中心创建成功');
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
