<script lang="ts" setup>
import type { DrawerProps, ExtendedDrawerApi } from './drawer';

import {
  computed,
  onDeactivated,
  provide,
  ref,
  unref,
  useId,
  watch,
} from 'vue';

import {
  useIsMobile,
  usePriorityValues,
  useSimpleLocale,
} from '@vben-core/composables';
import { X } from '@vben-core/icons';
import {
  Separator,
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  VbenButton,
  VbenHelpTooltip,
  VbenIconButton,
  VbenLoading,
  VisuallyHidden,
} from '@vben-core/shadcn-ui';
import { ELEMENT_ID_MAIN_CONTENT } from '@vben-core/shared/constants';
import { globalShareState } from '@vben-core/shared/global-state';
import { cn } from '@vben-core/shared/utils';

interface Props extends DrawerProps {
  drawerApi?: ExtendedDrawerApi;
}

const props = withDefaults(defineProps<Props>(), {
  appendToMain: false,
  closeIconPlacement: 'right',
  destroyOnClose: false,
  drawerApi: undefined,
  submitting: false,
  zIndex: 1000,
});

const components = globalShareState.getComponents();

const id = useId();
provide('DISMISSABLE_DRAWER_ID', id);

// @ts-expect-error unused
const wrapperRef = ref<HTMLElement>();
const { $t } = useSimpleLocale();
const { isMobile } = useIsMobile();

const state = props.drawerApi?.useStore?.();

const {
  appendToMain,
  cancelText,
  class: drawerClass,
  closable,
  closeIconPlacement,
  closeOnClickModal,
  closeOnPressEscape,
  confirmLoading,
  confirmText,
  contentClass,
  description,
  destroyOnClose,
  footer: showFooter,
  footerClass,
  header: showHeader,
  headerClass,
  loading: showLoading,
  modal,
  openAutoFocus,
  overlayBlur,
  placement,
  showCancelButton,
  showConfirmButton,
  submitting,
  title,
  titleTooltip,
  zIndex,
} = usePriorityValues(props, state);

// watch(
//   () => showLoading.value,
//   (v) => {
//     if (v && wrapperRef.value) {
//       wrapperRef.value.scrollTo({
//         // behavior: 'smooth',
//         top: 0,
//       });
//     }
//   },
// );

/**
 * 在开启keepAlive情况下 直接通过浏览器按钮/手势等返回 不会关闭弹窗
 */
onDeactivated(() => {
  // 如果弹窗没有被挂载到内容区域，则关闭弹窗
  if (!appendToMain.value) {
    props.drawerApi?.close();
  }
});

function interactOutside(e: Event) {
  if (!closeOnClickModal.value || submitting.value) {
    e.preventDefault();
  }
}
function escapeKeyDown(e: KeyboardEvent) {
  if (!closeOnPressEscape.value || submitting.value) {
    e.preventDefault();
  }
}
// pointer-down-outside
function pointerDownOutside(e: Event) {
  const target = e.target as HTMLElement;
  const dismissableDrawer = target?.dataset.dismissableDrawer;
  if (
    submitting.value ||
    !closeOnClickModal.value ||
    dismissableDrawer !== id
  ) {
    e.preventDefault();
  }
}

function handerOpenAutoFocus(e: Event) {
  if (!openAutoFocus.value) {
    e?.preventDefault();
  }
}

function handleFocusOutside(e: Event) {
  e.preventDefault();
  e.stopPropagation();
}

const getAppendTo = computed(() => {
  return appendToMain.value ? `#${ELEMENT_ID_MAIN_CONTENT}` : undefined;
});

/**
 * destroyOnClose功能完善
 */
// 是否打开过
const hasOpened = ref(false);
const isClosed = ref(true);
watch(
  () => state?.value?.isOpen,
  (value) => {
    isClosed.value = false;
    if (value && !unref(hasOpened)) {
      hasOpened.value = true;
    }
  },
  { immediate: true },
);
function handleClosed() {
  isClosed.value = true;
  props.drawerApi?.onClosed();
}
const getForceMount = computed(() => {
  return !unref(destroyOnClose) && unref(hasOpened);
});

/**
 * 抽屉边缘拖拽调整宽度（左侧/右侧抽屉生效，移动端与上下抽屉禁用）
 */
const canResize = computed(() => {
  const side = placement.value ?? 'right';
  return !isMobile.value && (side === 'left' || side === 'right');
});
const drawerWidth = ref<number | undefined>();
let resizeState = { pointerId: -1, startX: 0, startWidth: 0 };

function onResizeStart(e: PointerEvent) {
  const handle = e.currentTarget as HTMLElement;
  const panel = handle.parentElement;
  if (!panel) return;
  resizeState = {
    pointerId: e.pointerId,
    startWidth: panel.offsetWidth,
    startX: e.clientX,
  };
  handle.setPointerCapture(e.pointerId);
  document.body.style.cursor = 'ew-resize';
  document.body.style.userSelect = 'none';
}
function onResizeMove(e: PointerEvent) {
  if (e.pointerId !== resizeState.pointerId || !resizeState.startWidth) {
    return;
  }
  const side = placement.value ?? 'right';
  const delta =
    side === 'right' ? resizeState.startX - e.clientX : e.clientX - resizeState.startX;
  const min = 360;
  const max = Math.min(window.innerWidth * 0.8, 1280);
  drawerWidth.value = Math.round(
    Math.min(max, Math.max(min, resizeState.startWidth + delta)),
  );
}
function onResizeEnd(e: PointerEvent) {
  if (e.pointerId !== resizeState.pointerId) return;
  try {
    (e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId);
  } catch {
    // 指针可能已被浏览器释放
  }
  resizeState = { pointerId: -1, startX: 0, startWidth: 0 };
  document.body.style.cursor = '';
  document.body.style.userSelect = '';
}
watch(
  () => state?.value?.isOpen,
  (open) => {
    if (!open) drawerWidth.value = undefined;
  },
);
</script>
<template>
  <Sheet
    :modal="false"
    :open="state?.isOpen"
    @update:open="() => drawerApi?.close()"
  >
    <SheetContent
      :append-to="getAppendTo"
      :class="
        cn(
          'flex w-130 flex-col',
          {
            'w-full!':
              isMobile || placement === 'bottom' || placement === 'top',
            'max-h-screen': placement === 'bottom' || placement === 'top',
            hidden: isClosed,
          },
          drawerClass,
        )
      "
      :style="drawerWidth ? { width: `${drawerWidth}px` } : undefined"
      :modal="modal"
      :open="state?.isOpen"
      :side="placement"
      :z-index="zIndex"
      :force-mount="getForceMount"
      :overlay-blur="overlayBlur"
      @close-auto-focus="handleFocusOutside"
      @closed="handleClosed"
      @escape-key-down="escapeKeyDown"
      @focus-outside="handleFocusOutside"
      @interact-outside="interactOutside"
      @open-auto-focus="handerOpenAutoFocus"
      @opened="() => drawerApi?.onOpened()"
      @pointer-down-outside="pointerDownOutside"
    >
      <div
        v-if="canResize"
        :class="
          cn(
            'absolute inset-y-0 z-20 w-1.5 cursor-ew-resize touch-none select-none transition-colors hover:bg-primary/30 active:bg-primary/50',
            (placement ?? 'right') === 'left' ? 'right-0' : 'left-0',
          )
        "
        @pointerdown="onResizeStart"
        @pointermove="onResizeMove"
        @pointerup="onResizeEnd"
        @pointercancel="onResizeEnd"
      ></div>
      <SheetHeader
        v-if="showHeader"
        :class="
          cn(
            'flex! flex-row items-center justify-between border-b px-6 py-5',
            headerClass,
            {
              'px-4 py-3': closable,
              'pl-2': closable && closeIconPlacement === 'left',
            },
          )
        "
      >
        <div class="flex items-center">
          <SheetClose
            v-if="closable && closeIconPlacement === 'left'"
            as-child
            :disabled="submitting"
            class="ml-0.5 cursor-pointer rounded-full opacity-80 transition-opacity hover:opacity-100 focus:outline-hidden disabled:pointer-events-none data-[state=open]:bg-secondary"
          >
            <slot name="close-icon">
              <VbenIconButton>
                <X class="size-4" />
              </VbenIconButton>
            </slot>
          </SheetClose>
          <Separator
            v-if="closable && closeIconPlacement === 'left'"
            class="mr-2 ml-1 h-8"
            decorative
            orientation="vertical"
          />
          <SheetTitle v-if="title" class="text-left">
            <slot name="title">
              {{ title }}

              <VbenHelpTooltip v-if="titleTooltip" trigger-class="pb-1">
                {{ titleTooltip }}
              </VbenHelpTooltip>
            </slot>
          </SheetTitle>
          <SheetDescription v-if="description" class="mt-1 text-xs">
            <slot name="description">
              {{ description }}
            </slot>
          </SheetDescription>
        </div>

        <VisuallyHidden v-if="!title || !description">
          <SheetTitle v-if="!title" />
          <SheetDescription v-if="!description" />
        </VisuallyHidden>

        <div class="flex-center">
          <slot name="extra"></slot>
          <SheetClose
            v-if="closable && closeIconPlacement === 'right'"
            as-child
            :disabled="submitting"
            class="ml-0.5 cursor-pointer rounded-full opacity-80 transition-opacity hover:opacity-100 focus:outline-hidden disabled:pointer-events-none data-[state=open]:bg-secondary"
          >
            <slot name="close-icon">
              <VbenIconButton>
                <X class="size-4" />
              </VbenIconButton>
            </slot>
          </SheetClose>
        </div>
      </SheetHeader>
      <template v-else>
        <VisuallyHidden>
          <SheetTitle />
          <SheetDescription />
        </VisuallyHidden>
      </template>
      <div
        ref="wrapperRef"
        :class="
          cn('relative flex-1 overflow-y-auto p-3', contentClass, {
            'pointer-events-none': showLoading || submitting,
          })
        "
      >
        <slot></slot>
      </div>
      <VbenLoading v-if="showLoading || submitting" spinning />
      <SheetFooter
        v-if="showFooter"
        :class="
          cn(
            'w-full flex-row items-center justify-end border-t p-2 px-3',
            footerClass,
          )
        "
      >
        <slot name="prepend-footer"></slot>
        <slot name="footer">
          <component
            :is="components.DefaultButton || VbenButton"
            v-if="showCancelButton"
            variant="ghost"
            :disabled="submitting"
            @click="() => drawerApi?.onCancel()"
          >
            <slot name="cancelText">
              {{ cancelText || $t('cancel') }}
            </slot>
          </component>
          <slot name="center-footer"></slot>
          <component
            :is="components.PrimaryButton || VbenButton"
            v-if="showConfirmButton"
            :loading="confirmLoading || submitting"
            @click="() => drawerApi?.onConfirm()"
          >
            <slot name="confirmText">
              {{ confirmText || $t('confirm') }}
            </slot>
          </component>
        </slot>
        <slot name="append-footer"></slot>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
