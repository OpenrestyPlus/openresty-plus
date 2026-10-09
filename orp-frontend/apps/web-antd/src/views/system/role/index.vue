<script lang="ts" setup>
import type { RbacRoleItem } from '#/api';
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page, useVbenDrawer, useVbenModal } from '@vben/common-ui';

import { Button, message, Popconfirm, Tag, Tooltip } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { deleteRbacRoleApi, listRbacRolesApi } from '#/api';

import RoleFormModal from './role-form-modal.vue';
import RolePermDrawer from './role-perm-drawer.vue';

defineOptions({ name: 'SystemRoleManagement' });

const formOptions: VbenFormProps = {
  collapsed: false,
  schema: [
    {
      component: 'Input',
      componentProps: { allowClear: true, placeholder: '角色名称或编码' },
      fieldName: 'keyword',
      label: '关键字',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: '启用', value: 'enabled' },
          { label: '停用', value: 'disabled' },
        ],
        placeholder: '按状态筛选',
      },
      fieldName: 'status',
      label: '状态',
    },
  ],
  submitOnChange: true,
};

const gridOptions: VxeGridProps<RbacRoleItem> = {
  columns: [
    { title: '序号', type: 'seq', width: 60 },
    { field: 'name', minWidth: 130, title: '角色名称' },
    { field: 'code', minWidth: 130, title: '角色编码', },
    {
      field: 'builtin',
      slots: { default: 'builtin' },
      title: '内置',
      width: 80,
    },
    {
      field: 'status',
      slots: { default: 'status' },
      title: '状态',
      width: 90,
    },
    { field: 'userCount', slots: { default: 'userCount' }, title: '绑定用户', width: 100 },
    { field: 'permissionKeys', slots: { default: 'perms' }, title: '权限数', width: 90 },
    { field: 'description', minWidth: 220, showOverflow: true, title: '角色描述' },
    { field: 'createdAt', title: '创建时间', width: 170 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 250,
    },
  ],
  height: 'auto',
  pagerConfig: {},
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await listRbacRolesApi({
          ...formValues,
          page: page.currentPage,
          pageSize: page.pageSize,
        });
        return { items: res.items, total: res.total };
      },
    },
  },
};

const [BasicGrid, gridApi] = useVbenVxeGrid({ formOptions, gridOptions });

const [RoleFormView, roleFormApi] = useVbenModal({
  connectedComponent: RoleFormModal,
});

const [RolePermView, rolePermApi] = useVbenDrawer({
  connectedComponent: RolePermDrawer,
});

function onAdd() {
  roleFormApi.setData({ record: null }).open();
}

function onEdit(record: RbacRoleItem) {
  roleFormApi.setData({ record }).open();
}

function onAssign(record: RbacRoleItem) {
  rolePermApi.setData({ record }).open();
}

async function onDelete(record: RbacRoleItem) {
  await deleteRbacRoleApi(record.id);
  message.success(`角色 ${record.name} 已删除`);
  gridApi.query();
}

function onFormSuccess() {
  gridApi.query();
}
</script>

<template>
  <Page
    title="角色管理"
    description="以角色为载体聚合菜单与操作权限；用户通过绑定角色获得权限，内置角色不可删除。"
  >
    <BasicGrid table-title="角色列表">
      <template #toolbar-actions>
        <Button type="primary" @click="onAdd">新增角色</Button>
      </template>
      <template #builtin="{ row }">
        <Tag v-if="row.builtin" class="!m-0" color="purple">内置</Tag>
        <span v-else class="text-gray-400">—</span>
      </template>
      <template #status="{ row }">
        <span :class="row.status === 'enabled' ? 'inline-flex items-center gap-1.5 rounded border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-xs text-emerald-700' : 'inline-flex items-center gap-1.5 rounded border border-rose-200 bg-rose-50 px-2 py-0.5 text-xs text-rose-700'">
          <span :class="row.status === 'enabled' ? 'h-1.5 w-1.5 rounded-full bg-emerald-500' : 'h-1.5 w-1.5 rounded-full bg-rose-500'" />
          {{ row.status === 'enabled' ? '启用' : '停用' }}
        </span>
      </template>
      <template #userCount="{ row }">
        <Tag class="!m-0" :color="row.userCount ? 'blue' : 'default'">{{ row.userCount ?? 0 }} 人</Tag>
      </template>
      <template #perms="{ row }">
        <Tag class="!m-0" color="cyan">{{ row.permissionKeys.length }} 项</Tag>
      </template>
      <template #action="{ row }">
        <Button size="small" type="link" @click="onAssign(row)">分配权限</Button>
        <Button size="small" type="link" @click="onEdit(row)">编辑</Button>
        <Popconfirm v-if="!row.builtin" :title="`确认删除角色 ${row.name}？`" @confirm="onDelete(row)">
          <Button danger size="small" type="link">删除</Button>
        </Popconfirm>
        <Tooltip v-else title="内置角色不可删除">
          <Button disabled size="small" type="link">删除</Button>
        </Tooltip>
      </template>
    </BasicGrid>
    <RoleFormView @success="onFormSuccess" />
    <RolePermView @success="onFormSuccess" />
  </Page>
</template>
