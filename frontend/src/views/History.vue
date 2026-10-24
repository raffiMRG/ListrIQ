<script setup>
import { computed, ref } from 'vue'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'
import AppChart from '../AppChart.vue'
import { METRICS } from '../live.js'

const dateRange = ref([new Date(Date.now() - 86400e3), new Date()])
const resolution = ref('auto')
const metric = ref('power')
const rows = ref([])
const error = ref('')

async function load() {
  error.value = ''
  const [from, to] = dateRange.value || []
  if (!(from instanceof Date) || !(to instanceof Date) || from >= to) {
    error.value = 'Pilih rentang waktu yang lengkap dan valid.'
    return
  }
  const q = new URLSearchParams({ from: from.toISOString(), to: to.toISOString(), resolution: resolution.value })
  const r = await fetch(`/api/readings/history?${q}`)
  if (!r.ok) return (error.value = await r.text())
  rows.value = await r.json()
}
load()

const m = computed(() => METRICS.find((x) => x[0] === metric.value))
const labels = computed(() => rows.value.map((p) => new Date(p.ts).toLocaleString('id-ID', { dateStyle: 'short', timeStyle: 'short' })))
const datasets = computed(() => {
  const ds = [{ label: `${m.value[1]} (${m.value[2]})`, data: rows.value.map((p) => p[metric.value]), borderColor: '#2563eb' }]
  if (metric.value === 'power' && rows.value[0]?.max_power !== undefined)
    ds.push({ label: 'Daya maks (W)', data: rows.value.map((p) => p.max_power), borderColor: '#f59e0b' })
  return ds
})
</script>

<template>
  <div class="card">
    <div class="row">
      <label class="range-picker">Rentang waktu
        <VueDatePicker v-model="dateRange" range :enable-time-picker="true" :locale="'id-ID'" format="dd MMM yyyy, HH:mm" :teleport="true" :clearable="false" placeholder="Pilih rentang waktu" />
      </label>
      <select v-model="resolution">
        <option value="auto">Otomatis</option>
        <option value="raw">Raw</option>
        <option value="hourly">Per jam</option>
        <option value="daily">Per hari</option>
      </select>
      <select v-model="metric">
        <option v-for="[key, label] in METRICS.filter((x) => x[0] !== 'energy')" :key="key" :value="key">{{ label }}</option>
      </select>
      <button class="on" @click="load">Tampilkan</button>
    </div>
    <p v-if="error" class="muted">Gagal: {{ error }}</p>
    <p v-else-if="!rows.length" class="muted">Tidak ada data pada rentang ini.</p>
    <AppChart v-else :labels="labels" :datasets="datasets" />
  </div>
</template>

<style scoped>
.range-picker { display: grid; gap: 4px; font-size: 13px; color: var(--muted); }
.range-picker :deep(.dp__input) { min-width: 260px; border-color: var(--border); color: var(--text); background: var(--card); }
</style>
