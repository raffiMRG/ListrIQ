<script setup>
import { computed, ref, watch } from 'vue'
import AppChart from '../AppChart.vue'
import { fmt } from '../live.js'

const today = new Date(Date.now() - new Date().getTimezoneOffset() * 60e3).toISOString()
const kind = ref('daily')
const date = ref(today.slice(0, 10))
const month = ref(today.slice(0, 7))
const rows = ref([])

async function load() {
  const q = kind.value === 'daily' ? `daily?date=${date.value}` : `monthly?month=${month.value}`
  const r = await fetch(`/api/reports/${q}`)
  rows.value = r.ok ? await r.json() : []
}
watch([kind, date, month], load, { immediate: true })

const label = (ts) => kind.value === 'daily'
  ? new Date(ts).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
  : new Date(ts).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
const total = computed(() => rows.value.reduce((s, r) => s + (r.energy || 0), 0))
const peak = computed(() => Math.max(0, ...rows.value.map((r) => r.max_power || 0)))
const labels = computed(() => rows.value.map((r) => label(r.ts)))
const datasets = computed(() => [{ label: 'Energi (kWh)', data: rows.value.map((r) => r.energy), backgroundColor: '#2563eb' }])
</script>

<template>
  <div class="card">
    <div class="row">
      <button :class="{ on: kind === 'daily' }" @click="kind = 'daily'">Harian</button>
      <button :class="{ on: kind === 'monthly' }" @click="kind = 'monthly'">Bulanan</button>
      <input v-if="kind === 'daily'" type="date" v-model="date" />
      <input v-else type="month" v-model="month" />
    </div>
    <div class="row">
      <strong>Total: {{ fmt(total, 3) }} kWh</strong>
      <span class="muted">· Daya puncak: {{ fmt(peak, 0) }} W</span>
    </div>
    <p v-if="!rows.length" class="muted">Belum ada data.</p>
    <template v-else>
      <AppChart type="bar" :labels="labels" :datasets="datasets" />
      <table>
        <thead><tr><th>{{ kind === 'daily' ? 'Jam' : 'Tanggal' }}</th><th>Energi (kWh)</th><th>Daya rata-rata (W)</th><th>Daya maks (W)</th></tr></thead>
        <tbody>
          <tr v-for="r in rows" :key="r.ts">
            <td>{{ label(r.ts) }}</td><td>{{ fmt(r.energy, 3) }}</td><td>{{ fmt(r.avg_power, 1) }}</td><td>{{ fmt(r.max_power, 1) }}</td>
          </tr>
        </tbody>
      </table>
    </template>
  </div>
</template>
