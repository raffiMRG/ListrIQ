<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import Chart from 'chart.js/auto'
import zoomPlugin from 'chartjs-plugin-zoom'

Chart.register(zoomPlugin)

// Parent wajib mengganti array (bukan mutasi) agar watch terpicu.
const props = defineProps({
  labels: Array,
  datasets: Array,
  type: { type: String, default: 'line' },
})
const canvas = ref()
const zoomed = ref(false)
let chart

// Salin ke array biasa: Chart.js tidak cocok dengan proxy reaktif Vue.
const data = () => ({
  labels: [...props.labels],
  datasets: props.datasets.map((d) => ({ ...d, data: [...d.data] })),
})

onMounted(() => {
  chart = new Chart(canvas.value, {
    type: props.type,
    data: data(),
    options: {
      animation: false,
      maintainAspectRatio: false,
      interaction: { mode: 'index', intersect: false },
      elements: { point: { radius: 0 }, line: { borderWidth: 2 } },
      plugins: {
        zoom: {
          pan: { enabled: true, mode: 'x', onPanComplete: () => (zoomed.value = true) },
          zoom: {
            drag: { enabled: true, borderColor: '#2563eb', backgroundColor: '#2563eb22' },
            wheel: { enabled: true },
            pinch: { enabled: true },
            mode: 'x',
            onZoomComplete: () => (zoomed.value = true),
          },
        },
      },
    },
  })
})
watch(() => [props.labels, props.datasets], () => {
  chart.data = data()
  chart.update()
})
onBeforeUnmount(() => chart.destroy())

function resetZoom() {
  chart.resetZoom()
  zoomed.value = false
}
</script>

<template>
  <div class="chart-wrap">
    <div class="chart"><canvas ref="canvas" /></div>
    <button v-if="zoomed" class="reset-zoom" type="button" @click="resetZoom">Reset zoom</button>
  </div>
</template>

<style scoped>
.chart { position: relative; height: 300px; }
.chart-wrap { position: relative; }
.reset-zoom { position: absolute; top: 8px; right: 8px; font-size: 12px; padding: 4px 8px; background: var(--card); }
</style>
