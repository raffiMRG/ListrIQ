import { reactive } from 'vue'

const DEVICE = 'main'

// [key, label, satuan, desimal]
export const METRICS = [
  ['power', 'Daya', 'W', 1],
  ['current', 'Arus', 'A', 2],
  ['voltage', 'Tegangan', 'V', 1],
  ['energy', 'Energi', 'kWh', 3],
  ['frequency', 'Frekuensi', 'Hz', 1],
  ['power_factor', 'Power Factor', '', 2],
]

export const fmt = (v, dec) => (v == null ? '–' : Number(v).toLocaleString('id-ID', { minimumFractionDigits: dec, maximumFractionDigits: dec }))
export const TIMEFRAMES = [
  { minutes: 1, label: '1 menit terakhir' },
  { minutes: 10, label: '10 menit terakhir' },
  { minutes: 30, label: '30 menit terakhir' },
  { minutes: 60, label: '1 jam terakhir' },
]

// State realtime bersama: diisi WebSocket, dibaca Dashboard.
export const live = reactive({ reading: null, status: 'offline', lastSeen: null, points: [], connected: false, loading: false })
let timeframeMinutes = 10
let historyController

function trimPoints() {
  const cutoff = Date.now() - timeframeMinutes * 60e3
  live.points = live.points.filter((point) => new Date(point.ts).getTime() >= cutoff)
}

function push(r) {
  live.reading = r
  live.points.push(r)
  trimPoints()
}

export async function loadTimeframe(minutes) {
  timeframeMinutes = minutes
  historyController?.abort()
  const controller = new AbortController()
  historyController = controller
  live.loading = true
  const to = new Date()
  const from = new Date(to.getTime() - minutes * 60e3)
  try {
    const response = await fetch(`/api/readings/history?resolution=raw&from=${from.toISOString()}&to=${to.toISOString()}`, {
      signal: controller.signal,
    })
    if (!response.ok) throw new Error(`Riwayat: ${response.status}`)
    live.points = await response.json()
    trimPoints()
  } catch (error) {
    if (error.name !== 'AbortError') console.error(error)
  } finally {
    if (historyController === controller) live.loading = false
  }
}

function connect() {
  const ws = new WebSocket(`${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws`)
  ws.onopen = () => (live.connected = true)
  ws.onclose = () => {
    live.connected = false
    setTimeout(connect, 3000)
  }
  ws.onmessage = (e) => {
    const { type, data } = JSON.parse(e.data)
    if (data.device_id !== DEVICE) return
    if (type === 'reading') {
      push(data)
      live.status = 'online'
      live.lastSeen = data.ts
    } else if (type === 'status') {
      live.status = data.status
      live.lastSeen = data.last_seen
    }
  }
}

export async function start() {
  const status = await fetch('/api/status').then((r) => r.json())
  const s = status.find((d) => d.device_id === DEVICE)
  if (s) Object.assign(live, { status: s.status, lastSeen: s.last_seen })
  await loadTimeframe(timeframeMinutes)
  connect()
}
