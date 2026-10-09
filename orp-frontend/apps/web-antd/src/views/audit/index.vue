<script lang="ts" setup>
import type { AuditLog } from '#/api';
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenDrawer } from '@vben/common-ui';

import { Button, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { listAuditLogsApi } from '#/api';

import AuditDetailDrawer from './detail-drawer.vue';

defineOptions({ name: 'AuditLogManagement' });

const MODULE_TEXT: Record<string, string> = {
  center: '中心管理',
  user: '用户管理',
  dns: 'DNS 解析器',
  http: 'HTTP 流量',
  'ip-group': 'IP 组管理',
  node: '节点实例',
  publish: '配置下发',
  role: '角色管理',
  setting: '系统配置',
  stream: 'Stream 流量',
  tls: 'TLS 证书',
  upstream: '上游服务器组',
};

const ACTION_COLOR: Record<string, string> = {
  abort: 'orange',
  'reset-password': 'gold',
  advance: 'cyan',
  assign: 'purple',
  create: 'green',
  delete: 'red',
  offline: 'orange',
  update: 'blue',
};

const ACTION_TEXT: Record<string, string> = {
  abort: '中止',
  advance: '推进',
  assign: '分配',
  create: '新增',
  delete: '删除',
  offline: '下线',
  'reset-password': '重置密码',
  update: '修改',
};

const formOptions: VbenFormProps = {
  collapsed: false,
  fieldMappingTime: [['timeRange', ['startTime', 'endTime']]],
  schema: [
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        showSearch: true,
        optionFilterProp: 'label',
        options: [
          { label: '全部模块', value: 'all' },
          ...Object.entries(MODULE_TEXT).map(([value, label]) => ({ label, value })),
        ],
        placeholder: '按模块筛选',
      },
      fieldName: 'module',
      label: '变更模块',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        showSearch: true,
        optionFilterProp: 'label',
        options: [
          { label: '全部类型', value: 'all' },
          ...Object.entries(ACTION_TEXT).map(([value, label]) => ({ label, value })),
        ],
        placeholder: '按类型筛选',
      },
      fieldName: 'action',
      label: '操作类型',
    },
    {
      component: 'Input',
      componentProps: { allowClear: true, placeholder: '操作人关键字' },
      fieldName: 'operator',
      label: '操作人',
    },
    {
      component: 'RangePicker',
      componentProps: {
        valueFormat: 'YYYY-MM-DD',
      },
      fieldName: 'timeRange',
      label: '时间范围',
    },
  ],
  submitOnChange: true,
};

const gridOptions: VxeGridProps<AuditLog> = {
  columns: [
    { title: '序号', type: 'seq', width: 60 },
    { field: 'createdAt', title: '操作时间', width: 170 },
    { field: 'operator', title: '操作人', width: 110 },
    {
      field: 'module',
      slots: { default: 'module' },
      title: '变更模块',
      width: 120,
    },
    {
      field: 'action',
      slots: { default: 'action' },
      title: '操作类型',
      width: 90,
    },
    { field: 'target', minWidth: 180, showOverflow: true, title: '目标资源' },
    { field: 'detail', minWidth: 260, showOverflow: true, title: '变更详情' },
    {
      field: 'actionBtn',
      fixed: 'right',
      slots: { default: 'actionBtn' },
      title: '操作',
      width: 100,
    },
  ],
  height: 'auto',
  pagerConfig: {},
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        return await listAuditLogsApi({
          action: formValues?.action,
          endTime: formValues?.endTime,
          module: formValues?.module,
          operator: formValues?.operator,
          page: page.currentPage,
          pageSize: page.pageSize,
          startTime: formValues?.startTime,
        });
      },
    },
  },
  toolbarConfig: { custom: true, refresh: true, zoom: true },
};

const [Grid] = useVbenVxeGrid({ formOptions, gridOptions });

const [AuditDrawer, auditDrawerApi] = useVbenDrawer({
  connectedComponent: AuditDetailDrawer,
});

function onDetail(row: AuditLog) {
  auditDrawerApi.setData({ record: row }).open();
}
</script>

<template>
  <Page
    auto-content-height
    description="平台内所有资源变更操作的审计流水，不可篡改。"
    title="审计日志"
  >
    <Grid table-title="审计记录">
      <template #toolbar-tools>
        <span class="text-xs text-gray-400">
          今日新增审计记录自动滚动展示，支持按模块 / 操作人 / 时间范围检索
        </span>
      </template>
      <template #module="{ row }">
        <Tag color="geekblue">{{ MODULE_TEXT[row.module] ?? row.module }}</Tag>
      </template>
      <template #action="{ row }">
        <Tag :color="ACTION_COLOR[row.action]">
          {{ ACTION_TEXT[row.action] ?? row.action }}
        </Tag>
      </template>
      <template #actionBtn="{ row }">
        <Button size="small" type="link" @click="onDetail(row)">详情</Button>
      </template>
    </Grid>
    <AuditDrawer />
  </Page>
</template>
