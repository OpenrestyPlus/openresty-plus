<script lang="ts" setup>
import type { RbacPermissionNode } from '#/api';

import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Card, Input, Tag } from 'ant-design-vue';

import { listRbacPermissionsApi } from '#/api';

defineOptions({ name: 'SystemPermManagement' });

const loading = ref(false);
const tree = ref<RbacPermissionNode[]>([]);
const keyword = ref('');

onMounted(async () => {
  loading.value = true;
  try {
    tree.value = await listRbacPermissionsApi();
  } finally {
    loading.value = false;
  }
});

/** 关键字过滤 */
const filteredTree = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  if (!kw) return tree.value;
  const filter = (nodes: RbacPermissionNode[]): RbacPermissionNode[] => {
    const out: RbacPermissionNode[] = [];
    nodes.forEach((n) => {
      const kids = n.children ? filter(n.children) : undefined;
      if (
        n.title.toLowerCase().includes(kw) ||
        n.key.toLowerCase().includes(kw) ||
        (kids && kids.length > 0)
      ) {
        out.push({ ...n, children: kids });
      }
    });
    return out;
  };
  return filter(tree.value);
});

const totalKeys = computed(() => {
  let c = 0;
  const walk = (ns: RbacPermissionNode[]) => ns.forEach((n) => { c += 1; if (n.children) walk(n.children); });
  walk(tree.value);
  return c;
});

function countType(nodes: RbacPermissionNode[], type: RbacPermissionNode['type']): number {
  return nodes.reduce(
    (count, node) => count + Number(node.type === type) + countType(node.children ?? [], type),
    0,
  );
}

const menuCount = computed(() => countType(tree.value, 'menu'));
const actionCount = computed(() => countType(tree.value, 'action'));
</script>

<template>
  <Page
    title="权限定义"
    description="平台全部菜单与操作点权限的唯一定义源；角色分配权限时从这里选择。"
  >
    <div class="flex flex-col gap-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="flex items-center gap-2 text-sm text-gray-500">
          <Tag color="blue">菜单节点 {{ menuCount }} 个</Tag>
          <Tag color="cyan">操作点 {{ actionCount }} 个</Tag>
          <Tag color="purple">权限键合计 {{ totalKeys }} 个</Tag>
        </div>
        <Input
          v-model:value="keyword"
          allow-clear
          class="!w-64"
          placeholder="按名称或权限标识搜索"
        />
      </div>

      <Card :loading="loading" class="flex-1" size="small">
        <div class="grid grid-cols-1 gap-3 p-2 md:grid-cols-2 xl:grid-cols-3">
          <div
            v-for="node in filteredTree"
            :key="node.key"
            class="rounded-lg border border-gray-200 dark:border-gray-700"
          >
            <div class="flex items-center justify-between border-b border-gray-100 px-3 py-2 dark:border-gray-800">
              <span class="flex items-center gap-2">
                <Tag class="!m-0" :color="node.type === 'menu' ? 'blue' : 'cyan'">
                  {{ node.type === 'menu' ? '菜单' : '操作点' }}
                </Tag>
                <span class="text-sm font-medium">{{ node.title }}</span>
              </span>
              <span class="font-mono text-xs text-gray-400">{{ node.key }}</span>
            </div>
            <div class="space-y-1.5 px-3 py-2">
              <!-- 系统管理三级结构 -->
              <template v-if="node.key === 'system'">
                <div
                  v-for="sub in node.children ?? []"
                  :key="sub.key"
                  class="rounded bg-gray-50 p-2 dark:bg-gray-800/60"
                >
                  <div class="flex items-center justify-between">
                    <span class="text-xs font-medium">{{ sub.title }}</span>
                    <span class="font-mono text-xs text-gray-400">{{ sub.key }}</span>
                  </div>
                  <div class="mt-1 flex flex-wrap gap-x-3 gap-y-1">
                    <span
                      v-for="leaf in sub.children ?? []"
                      :key="leaf.key"
                      class="inline-flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300"
                    >
                      <span class="h-1 w-1 rounded-full bg-emerald-500" />
                      {{ leaf.title }}
                    </span>
                  </div>
                </div>
              </template>
              <!-- 常规二级结构 -->
              <template v-else>
                <div
                  v-for="child in node.children ?? []"
                  :key="child.key"
                  class="flex items-center justify-between"
                >
                  <span class="inline-flex items-center gap-1 text-xs text-gray-600 dark:text-gray-300">
                    <span class="h-1 w-1 rounded-full bg-emerald-500" />
                    {{ child.title }}
                  </span>
                  <span class="font-mono text-xs text-gray-400">{{ child.key }}</span>
                </div>
              </template>
            </div>
          </div>
        </div>
        <div v-if="!loading && filteredTree.length === 0" class="py-10 text-center text-sm text-gray-400">
          未命中任何权限节点
        </div>
      </Card>
    </div>
  </Page>
</template>
