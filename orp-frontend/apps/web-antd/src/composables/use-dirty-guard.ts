import { ref } from 'vue';

import { Modal as AntModal } from 'ant-design-vue';

/**
 * 表单脏数据防误关守卫
 *
 * 用法：
 * 1. useVbenForm 选项挂 handleValuesChange: () => markDirty()（用户输入即标记脏）
 * 2. useVbenModal / useVbenDrawer 选项挂 onBeforeClose（关闭前拦截确认）
 * 3. onOpenChange(true) 填充数据完成后调用 resetDirty()
 * 4. 保存成功路径在 modalApi.close() 前调用 resetDirty()
 */
export function useDirtyGuard() {
  const dirty = ref(false);
  let suppress = false;

  function markDirty() {
    if (!suppress) {
      dirty.value = true;
    }
  }

  /**
   * 清除脏标记：立即清一次（供保存路径 close 直接过闸），
   * 并吞掉紧随其后的 resetForm/setValues 同步触发的迟到 handleValuesChange
   */
  function resetDirty() {
    dirty.value = false;
    // 短窗口：回显填充触发的 values-change 全部吞掉，窗口结束后任何输入都正常标脏
    suppress = true;
    setTimeout(() => {
      dirty.value = false;
      suppress = false;
    }, 0);
  }

  /** 挂到 onBeforeClose：干净时直接放行；脏时弹确认框，确认放弃才放行 */
  function onBeforeClose(): Promise<boolean> {
    if (!dirty.value) {
      return Promise.resolve(true);
    }
    return new Promise<boolean>((resolve) => {
      AntModal.confirm({
        cancelText: '继续编辑',
        content: '关闭后本次修改内容将丢失，是否确认放弃？',
        okButtonProps: { danger: true },
        okText: '放弃修改',
        onCancel: () => resolve(false),
        onOk: () => {
          dirty.value = false;
          resolve(true);
        },
        title: '存在未保存的修改',
      });
    });
  }

  return { markDirty, onBeforeClose, resetDirty };
}
