<script setup>
import { computed, onBeforeUnmount, ref } from 'vue'
import { Icon } from '@iconify/vue'
import AppChart from '../AppChart.vue'
import GaugeChart from '../GaugeChart.vue'
import { live, METRICS, fmt } from '../live.js'

const metric = ref('power')
const gauges = [
  { key: 'power', label: 'Daya', icon: 'solar:bolt-bold-duotone', unit: 'W', decimals: 1, max: 3200, color: '#2563eb' },
  { key: 'current', label: 'Arus', icon: 'solar:flashlight-on-bold-duotone', unit: 'A', decimals: 2, max: 16, color: '#16a34a' },
  { key: 'voltage', label: 'Tegangan', icon: 'solar:plug-circle-bold-duotone', unit: 'V', decimals: 1, max: 250, color: '#d97706' },
]
const now = ref(Date.now())
const tick = setInterval(() => (now.value = Date.now()), 1000)
onBeforeUnmount(() => clearInterval(tick))

const ago = computed(() => {
  if (!live.lastSeen) return 'belum pernah'
  const s = Math.max(0, Math.round((now.value - new Date(live.lastSeen)) / 1000))
  if (s < 60) return `${s} detik lalu`
  if (s < 3600) return `${Math.floor(s / 60)} menit lalu`
  if (s < 86400) return `${Math.floor(s / 3600)} jam lalu`
  return new Date(live.lastSeen).toLocaleString('id-ID')
})

const m = computed(() => METRICS.find((x) => x[0] === metric.value))
const labels = computed(() => live.points.map((p) => new Date(p.ts).toLocaleTimeString('id-ID')))
const datasets = computed(() => [
  { label: `${m.value[1]} (${m.value[2]})`, data: live.points.map((p) => p[metric.value]), borderColor: '#2563eb', backgroundColor: '#2563eb33', fill: true },
])
</script>

<template>
  <div class="card status" aria-live="polite">
    <span class="badge" :class="live.status"><Icon icon="solar:link-round-angle-bold" />{{ live.status === 'online' ? 'Online' : 'Offline' }}</span>
    <span v-if="live.reading" class="badge" :class="live.reading.sensor_ok ? 'online' : 'offline'"><Icon icon="solar:shield-check-bold" />
      Sensor {{ live.reading.sensor_ok ? 'OK' : 'Error' }}
    </span>
    <span class="last-seen"><Icon icon="solar:clock-circle-bold" />Data terakhir: {{ ago }}</span>
    <span v-if="!live.connected" class="last-seen"><Icon icon="solar:refresh-circle-bold" />Menyambung ulang…</span>
    <span v-if="live.reading?.alarm" class="alarm"><Icon icon="solar:danger-triangle-bold" />Alarm daya aktif</span>
  </div>

  <section class="gauge-grid" aria-label="Pengukuran listrik realtime">
    <div v-for="gauge in gauges" :key="gauge.key" class="card metric gauge-card">
      <div class="metric-label"><Icon :icon="gauge.icon" aria-hidden="true" />{{ gauge.label }}</div>
      <GaugeChart v-bind="gauge" :value="live.reading?.[gauge.key]" />
    </div>
  </section>
  <section class="summary-grid" aria-label="Pengukuran tambahan">
    <div v-for="[key, label, unit, dec] in METRICS.slice(3)" :key="key" class="card metric">
      <div class="muted">{{ label }}</div>
      <div class="value">{{ fmt(live.reading?.[key], dec) }} <small>{{ unit }}</small></div>
    </div>
  </section>

  <div class="card">
    <div class="row">
      <button v-for="[key, label] in METRICS.slice(0, 3)" :key="key" :class="{ on: metric === key }" @click="metric = key">{{ label }}</button>
      <span class="muted">15 menit terakhir</span>
    </div>
    <AppChart :labels="labels" :datasets="datasets" />
  </div>
</template>

<style scoped>
.status { display: flex; flex-wrap: wrap; gap: 10px 16px; align-items: center; }
.badge, .last-seen, .alarm { display: inline-flex; align-items: center; gap: 6px; }
.badge { padding: 4px 10px; border-radius: 999px; color: #fff; font-size: 13px; font-weight: 600; background: var(--bad); }
.badge svg, .last-seen svg, .alarm svg { font-size: 17px; }
.badge.online { background: var(--ok); }
.last-seen { color: var(--muted); font-size: 13px; }
.alarm { color: var(--bad); font-size: 13px; font-weight: 600; }
.gauge-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 16px; margin-bottom: 16px; }
.summary-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 16px; margin-bottom: 16px; }
.gauge-grid .card, .summary-grid .card { margin: 0; }
.gauge-card { min-width: 0; padding-bottom: 12px; overflow: hidden; }
.metric-label { display: flex; align-items: center; gap: 7px; font-weight: 650; }
.metric-label svg { font-size: 21px; color: var(--accent); }
.gauge-card:nth-child(2) .metric-label svg { color: var(--ok); }
.gauge-card:nth-child(3) .metric-label svg { color: #d97706; }
.value { font-size: 26px; font-weight: 600; font-variant-numeric: tabular-nums; }
.value small { font-size: 14px; color: var(--muted); }
@media (max-width: 760px) {
  .gauge-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
  .gauge-card:first-child { grid-column: 1 / -1; }
  .summary-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px; }
  .summary-grid .card { padding: 12px; }
  .value { font-size: 21px; }
}
@media (max-width: 420px) {
  .status { align-items: flex-start; gap: 8px 12px; }
  .gauge-grid { grid-template-columns: 1fr; }
  .gauge-card:first-child { grid-column: auto; }
  .summary-grid { grid-template-columns: 1fr; }
}
</style>
