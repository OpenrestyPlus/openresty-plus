<script lang="ts" setup>
import type { DnsResolver } from '#/api';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import { createDnsResolverApi, listCentersApi, updateDnsResolverApi } from '#/api';
import { useDirtyGuard } from '#/composables/use-dirty-guard';

defineOptions({ name: 'DnsFormModal' });

const emit = defineEmits<{ success: [] }>();

const isEdit = ref(false);
const editId = ref(0);

const { markDirty, onBeforeClose, resetDirty } = useDirtyGuard();

const [Form, formApi] = useVbenForm({
  handleValuesChange: () => markDirty(),
  handleSubmit: onSubmit,
  schema: [
    {
      component: 'ApiSelect',
      componentProps: {
        allowClear: false,
        showSearch: true,

        optionFilterProp: 'label',

        api: async () => {
          const res = await listCentersApi({ page: 1, pageSize: 200 });
          return res.items.map((item) => ({ label: item.name, value: item.id }));
        },
        placeholder: '请选择所属中心',
      },
      fieldName: 'centerId',
      label: '所属中心',
      rules: z.number({ message: '请选择所属中心' }).min(1, { message: '请选择所属中心' }),
    },
    {
      component: 'Input',
      componentProps: { placeholder: '例如：10.60.0.11' },
      fieldName: 'address',
      label: '解析服务器地址',
      rules: z
        .string()
        .min(1, { message: '请输入解析服务器地址' })
        .regex(/^(\d{1,3}\.){3}\d{1,3}$/, { message: '必须是 IPv4 格式' }),
    },
    {
      component: 'InputNumber',
      componentProps: { max: 65_535, min: 1, placeholder: '默认 53', precision: 0 },
      defaultValue: 53,
      fieldName: 'port',
      label: '端口',
      rules: z
        .number({ message: '请输入端口' })
        .int()
        .min(1, { message: '端口范围 1-65535' })
        .max(65_535, { message: '端口范围 1-65535' }),
    },
    {
      component: 'InputNumber',
      componentProps: { max: 60, min: 1, placeholder: '单位：秒', precision: 0 },
      defaultValue: 2,
      fieldName: 'timeoutSec',
      label: '超时时间',
      help: '1-60 秒',
      rules: z
        .number({ message: '请输入超时时间' })
        .int()
        .min(1, { message: '范围 1-60 秒' })
        .max(60, { message: '范围 1-60 秒' }),
    },
    {
      component: 'InputNumber',
      componentProps: { max: 3600, min: 1, placeholder: '单位：秒', precision: 0 },
      defaultValue: 30,
      fieldName: 'cacheTtlSec',
      label: '有效缓存时间',
      help: '1-3600 秒',
      rules: z
        .number({ message: '请输入有效缓存时间' })
        .int()
        .min(1, { message: '范围 1-3600 秒' })
        .max(3600, { message: '范围 1-3600 秒' }),
    },
  ],
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal<{ record?: DnsResolver }>({
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
      modalApi.setState({ title: record ? '编辑解析器' : '新增解析器' });
      if (record) {
        formApi.setValues({
          address: record.address,
          cacheTtlSec: record.cacheTtlSec,
          centerId: record.centerId,
          port: record.port,
          timeoutSec: record.timeoutSec,
        });
      } else {
        formApi.setValues({ centerId: undefined });
      }
      resetDirty();
    }
  },
});

async function onSubmit(values: Record<string, any>) {
  modalApi.lock();
  try {
    const payload = {
      address: String(values.address),
      cacheTtlSec: Number(values.cacheTtlSec),
      centerId: Number(values.centerId),
      port: Number(values.port),
      timeoutSec: Number(values.timeoutSec),
    };
    if (isEdit.value) {
      await updateDnsResolverApi(editId.value, payload);
      message.success('解析器已更新');
    } else {
      await createDnsResolverApi(payload);
      message.success('解析器创建成功');
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
