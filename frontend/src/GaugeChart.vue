<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import Chart from 'chart.js/auto'
import { fmt } from './live.js'

const props = defineProps({
  value: Number,
  max: { type: Number, required: true },
  unit: { type: String, default: '' },
  decimals: { type: Number, default: 1 },
  color: { type: String, default: '#2563eb' },
})

const canvas = ref()
let chart

const numericValue = computed(() => Number.isFinite(props.value) ? props.value : 0)
const displayedValue = computed(() => props.value == null ? '–' : fmt(props.value, props.decimals))
const filledValue = computed(() => Math.min(props.max, Math.max(0, numericValue.value)))

function chartData() {
  return {
    datasets: [{
      data: [filledValue.value, props.max - filledValue.value],
      backgroundColor: [props.color, '#e5e7eb'],
      borderWidth: 0,
      borderRadius: 6,
    }],
  }
}

onMounted(() => {
  chart = new Chart(canvas.value, {
    type: 'doughnut',
    data: chartData(),
    options: {
      animation: false,
      responsive: true,
      maintainAspectRatio: false,
      rotation: -90,
      circumference: 180,
      cutout: '72%',
      plugins: { legend: { display: false }, tooltip: { enabled: false } },
    },
  })
})

watch([filledValue, () => props.max, () => props.color], () => {
  if (!chart) return
  chart.data = chartData()
  chart.update()
})

onBeforeUnmount(() => chart?.destroy())
</script>

<template>
  <div class="gauge" :aria-label="`${displayedValue} ${unit} dari maksimum ${max} ${unit}`">
    <canvas ref="canvas" />
    <div class="reading">
      <strong>{{ displayedValue }}</strong>
      <small>{{ unit }}</small>
    </div>
    <div class="range">maks. {{ fmt(max, decimals) }} {{ unit }}</div>
  </div>
</template>

<style scoped>
.gauge { position: relative; height: 180px; }
canvas { display: block; height: 100% !important; width: 100% !important; }
.reading { position: absolute; top: 54%; left: 0; width: 100%; text-align: center; font-variant-numeric: tabular-nums; }
.reading strong { display: block; font-size: clamp(24px, 3vw, 30px); line-height: 1.1; letter-spacing: -0.04em; }
.reading small, .range { color: var(--muted); }
.reading small { font-size: 14px; }
.range { position: absolute; bottom: 0; left: 0; width: 100%; text-align: center; font-size: 12px; }
@media (max-width: 560px) {
  .gauge { height: 155px; }
  .reading { top: 52%; }
}
</style>
