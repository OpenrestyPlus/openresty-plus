<script lang="ts" setup>
import type { AuditLog } from '#/api';

import { ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';

defineOptions({ name: 'AuditDetailDrawer' });

const record = ref<AuditLog>();

const MODULE_TEXT: Record<string, string> = {
  center: '中心管理',
  dns: 'DNS 解析器',
  http: 'HTTP 流量',
  node: '节点实例',
  stream: 'Stream 流量',
  tls: 'TLS 证书',
};

const ACTION_COLOR: Record<string, string> = {
  create: 'green',
  delete: 'red',
  offline: 'orange',
  update: 'blue',
};

const ACTION_TEXT: Record<string, string> = {
  create: '新增',
  delete: '删除',
  offline: '下线',
  update: '修改',
};

const [Drawer, drawerApi] = useVbenDrawer<{ record?: AuditLog }>({
  class: 'w-[560px]',
  onOpenChange(isOpen) {
    if (isOpen) {
      const log = drawerApi.getData()?.record;
      record.value = log;
      if (log) {
        drawerApi.setState({
          title: `审计详情 #${log.id}`,
        });
      }
    }
  },
  title: '审计详情',
});
</script>

<template>
  <Drawer>
    <Descriptions v-if="record" :column="1" bordered size="small" title="操作记录">
      <DescriptionsItem label="记录 ID">{{ record.id }}</DescriptionsItem>
      <DescriptionsItem label="操作时间">{{ record.createdAt }}</DescriptionsItem>
      <DescriptionsItem label="操作人">
        {{ record.operator }}（{{ record.ip }}）
      </DescriptionsItem>
      <DescriptionsItem label="变更模块">
        <Tag color="geekblue">{{ MODULE_TEXT[record.module] ?? record.module }}</Tag>
      </DescriptionsItem>
      <DescriptionsItem label="操作类型">
        <Tag :color="ACTION_COLOR[record.action]">
          {{ ACTION_TEXT[record.action] ?? record.action }}
        </Tag>
      </DescriptionsItem>
      <DescriptionsItem label="目标资源">
        <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-gray-800">
          {{ record.target }}
        </code>
      </DescriptionsItem>
      <DescriptionsItem label="变更详情">
        <p class="mb-0 leading-6 text-gray-700 dark:text-gray-300">
          {{ record.detail }}
        </p>
      </DescriptionsItem>
    </Descriptions>
  </Drawer>
</template>
