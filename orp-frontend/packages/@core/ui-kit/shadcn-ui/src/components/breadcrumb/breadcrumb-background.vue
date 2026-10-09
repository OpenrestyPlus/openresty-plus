<script lang="ts" setup>
import type { BreadcrumbProps } from './types';

import { VbenIcon } from '../icon';

interface Props extends BreadcrumbProps {}

defineOptions({ name: 'Breadcrumb' });
const { breadcrumbs, showIcon } = defineProps<Props>();

const emit = defineEmits<{ select: [string] }>();

function handleClick(index: number, path?: string) {
  if (!path || index === breadcrumbs.length - 1) {
    return;
  }
  emit('select', path);
}
</script>
<template>
  <ul class="flex items-center select-none">
    <TransitionGroup name="breadcrumb-transition">
      <template
        v-for="(item, index) in breadcrumbs"
        :key="`${item.path}-${item.title}-${index}`"
      >
        <li class="breadcrumb-item h-7">
          <a
            href="javascript:void 0"
            class="breadcrumb-link"
            :class="{
              'is-active cursor-default': index === breadcrumbs.length - 1,
              'cursor-pointer': index !== breadcrumbs.length - 1,
            }"
            @click.stop="handleClick(index, item.path)"
          >
            <span class="flex-center relative z-10 h-full">
              <VbenIcon
                v-if="showIcon"
                :icon="item.icon"
                class="mr-1.5 size-4 shrink-0 text-primary"
              />
              <span
                :class="{
                  'text-foreground font-semibold':
                    index === breadcrumbs.length - 1,
                  'text-muted-foreground font-normal hover:text-foreground':
                    index !== breadcrumbs.length - 1,
                }"
                class="text-[13px] tracking-wide"
                >{{ item.title }}
              </span>
            </span>
          </a>
        </li>
      </template>
    </TransitionGroup>
  </ul>
</template>
<style scoped>
@reference "@vben/tailwind-config/theme";

li.breadcrumb-item {
  height: 28px;
}

li.breadcrumb-item a.breadcrumb-link {
  position: relative;
  display: flex;
  height: 28px;
  align-items: center;
  background-color: hsl(var(--accent));
  border-top: 1px solid hsl(var(--border));
  border-bottom: 1px solid hsl(var(--border));
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  margin-right: 22px;
  padding: 0 14px 0 24px;
  font-size: 13px;
  line-height: 28px;
  transition: all 0.2s ease;
}

li.breadcrumb-item:first-child a.breadcrumb-link {
  border-left: 1px solid hsl(var(--border));
  border-top-left-radius: 3px;
  border-bottom-left-radius: 3px;
  padding-left: 14px;
}

li.breadcrumb-item:last-child a.breadcrumb-link {
  border-right: 1px solid hsl(var(--border));
  border-top-right-radius: 3px;
  border-bottom-right-radius: 3px;
  margin-right: 0;
  padding-right: 14px;
}

/* 两个箭头的缝隙嵌合：使用纯 CSS border 三角形 */
li.breadcrumb-item a.breadcrumb-link::before,
li.breadcrumb-item a.breadcrumb-link::after {
  content: '';
  position: absolute;
  top: -1px;
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 14px 0 14px 14px;
  transition: border-color 0.2s ease;
}

/* 左侧内凹三角（非第一项）：左边界内凹，颜色与页面顶栏背景一致 */
li.breadcrumb-item:not(:first-child) a.breadcrumb-link::before {
  left: 0;
  z-index: 1;
  border-color: transparent transparent transparent hsl(var(--header, var(--background)));
}

/* 右侧外凸三角（非最后一项）：向右凸出，与当前项背景色一致 */
li.breadcrumb-item:not(:last-child) a.breadcrumb-link::after {
  left: 100%;
  z-index: 2;
  border-color: transparent transparent transparent hsl(var(--accent));
}

/* 带有上下边框描边的右侧三角线效果（避免断边） */
li.breadcrumb-item:not(:last-child) a.breadcrumb-link:hover {
  background-color: hsl(var(--accent-hover));
}

li.breadcrumb-item:not(:last-child) a.breadcrumb-link:hover::after {
  border-color: transparent transparent transparent hsl(var(--accent-hover));
}
</style>
