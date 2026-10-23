<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import Chart from 'chart.js/auto'

// Parent wajib mengganti array (bukan mutasi) agar watch terpicu.
const props = defineProps({
  labels: Array,
  datasets: Array,
  type: { type: String, default: 'line' },
})
const canvas = ref()
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
    },
  })
})
watch(() => [props.labels, props.datasets], () => {
  chart.data = data()
  chart.update()
})
onBeforeUnmount(() => chart.destroy())
</script>

<template>
  <div class="chart"><canvas ref="canvas" /></div>
</template>

<style scoped>
.chart { position: relative; height: 300px; }
</style>
