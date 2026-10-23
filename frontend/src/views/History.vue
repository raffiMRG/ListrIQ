<script setup>
import { computed, ref } from 'vue'
import AppChart from '../AppChart.vue'
import { METRICS } from '../live.js'

// datetime-local butuh "YYYY-MM-DDTHH:mm" waktu lokal
const local = (d) => new Date(d - d.getTimezoneOffset() * 60e3).toISOString().slice(0, 16)
const from = ref(local(new Date(Date.now() - 86400e3)))
const to = ref(local(new Date()))
const resolution = ref('auto')
const metric = ref('power')
const rows = ref([])
const error = ref('')

async function load() {
  error.value = ''
  const q = new URLSearchParams({ from: new Date(from.value).toISOString(), to: new Date(to.value).toISOString(), resolution: resolution.value })
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
      <label>Dari <input type="datetime-local" v-model="from" /></label>
      <label>Sampai <input type="datetime-local" v-model="to" /></label>
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
