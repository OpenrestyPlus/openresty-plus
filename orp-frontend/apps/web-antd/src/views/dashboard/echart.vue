<script lang="ts" setup>
/**
 * ECharts 按需引入的轻量封装：传入 option，自动渲染 + ResizeObserver 自适应。
 */
import type { EChartsOption } from 'echarts';

import { BarChart, LineChart, PieChart } from 'echarts/charts';
import {
  DataZoomInsideComponent,
  DataZoomSliderComponent,
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent,
} from 'echarts/components';
import * as echarts from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';

echarts.use([
  BarChart,
  LineChart,
  PieChart,
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent,
  DataZoomInsideComponent,
  DataZoomSliderComponent,
  CanvasRenderer,
]);

const props = defineProps<{
  height?: string;
  option: EChartsOption;
}>();

const el = ref<HTMLDivElement>();
let chart: echarts.ECharts | null = null;
let observer: null | ResizeObserver = null;

function render() {
  if (!el.value) return;
  if (!chart) {
    chart = echarts.init(el.value);
  }
  chart.setOption(props.option, true);
}

onMounted(() => {
  render();
  observer = new ResizeObserver(() => chart?.resize());
  if (el.value) observer.observe(el.value);
});

watch(
  () => props.option,
  () => render(),
  { deep: true },
);

onBeforeUnmount(() => {
  observer?.disconnect();
  chart?.dispose();
  chart = null;
});
</script>

<template>
  <div ref="el" :style="{ height: height ?? '260px', width: '100%' }"></div>
</template>
