<script setup>
import { ref } from 'vue'
import { Icon } from '@iconify/vue'
import Dashboard from './views/Dashboard.vue'
import History from './views/History.vue'
import Report from './views/Report.vue'
import { start } from './live.js'

const pages = {
  dashboard: ['Dashboard', 'solar:widget-5-bold', Dashboard],
  history: ['Histori', 'solar:chart-2-bold', History],
  report: ['Laporan', 'solar:document-text-bold', Report],
}
const page = ref(location.hash.slice(1) in pages ? location.hash.slice(1) : 'dashboard')
window.addEventListener('hashchange', () => (page.value = location.hash.slice(1) in pages ? location.hash.slice(1) : 'dashboard'))

start()
</script>

<template>
  <header>
    <h1><Icon icon="solar:bolt-bold-duotone" aria-hidden="true" /> ListrIQ</h1>
    <nav>
      <a v-for="(p, key) in pages" :key="key" :href="'#' + key" :class="{ active: page === key }">
        <Icon :icon="p[1]" aria-hidden="true" />
        <span>{{ p[0] }}</span>
      </a>
    </nav>
  </header>
  <main><component :is="pages[page][2]" /></main>
</template>

<style>
:root {
  --bg: #f5f6f8; --card: #fff; --text: #1d2330; --muted: #6b7280; --border: #e3e6ea;
  --accent: #2563eb; --ok: #16a34a; --bad: #dc2626;
  color-scheme: light dark;
}
@media (prefers-color-scheme: dark) {
  :root { --bg: #111418; --card: #1a1e24; --text: #e6e8eb; --muted: #9aa1ab; --border: #2a2f37; --accent: #60a5fa; }
}
* { box-sizing: border-box; }
  body { margin: 0; background: var(--bg); color: var(--text); font: 15px/1.5 system-ui, sans-serif; }
  header { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px; padding: 12px max(16px, calc((100vw - 1168px) / 2)); border-bottom: 1px solid var(--border); background: var(--card); }
  h1 { display: flex; align-items: center; gap: 8px; font-size: 18px; margin: 0; }
  h1 svg { color: #eab308; font-size: 22px; }
  nav { display: flex; gap: 4px; }
  nav a { display: inline-flex; align-items: center; gap: 6px; padding: 7px 9px; border-radius: 7px; color: var(--muted); text-decoration: none; }
  nav a svg { font-size: 18px; }
  nav a.active { color: var(--accent); background: color-mix(in srgb, var(--accent) 10%, transparent); font-weight: 600; }
  main { max-width: 1200px; margin: 0 auto; padding: 20px 16px; }
  .card { background: var(--card); border: 1px solid var(--border); border-radius: 14px; padding: 16px; margin-bottom: 16px; }
.row { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; margin-bottom: 12px; }
input, select, button { font: inherit; padding: 6px 10px; border: 1px solid var(--border); border-radius: 6px; background: var(--card); color: var(--text); }
button { cursor: pointer; }
button.on { background: var(--accent); color: #fff; border-color: var(--accent); }
table { width: 100%; border-collapse: collapse; }
th, td { text-align: right; padding: 6px 8px; border-bottom: 1px solid var(--border); font-variant-numeric: tabular-nums; }
th:first-child, td:first-child { text-align: left; }
  .muted { color: var(--muted); }
  @media (max-width: 560px) {
    header { padding: 10px 14px; align-items: flex-start; }
    h1 { font-size: 16px; }
    nav { width: 100%; justify-content: space-between; }
    nav a { flex: 1; justify-content: center; padding: 7px 4px; font-size: 13px; }
    main { padding: 14px 12px; }
    .card { padding: 14px; border-radius: 12px; }
  }
</style>
