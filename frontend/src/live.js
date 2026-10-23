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
const MAX_POINTS = 450 // ~15 menit pada interval 2 detik

// State realtime bersama: diisi WebSocket, dibaca Dashboard.
export const live = reactive({ reading: null, status: 'offline', lastSeen: null, points: [], connected: false })

function push(r) {
  live.reading = r
  live.points.push(r)
  if (live.points.length > MAX_POINTS) live.points.shift()
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
  const from = new Date(Date.now() - 15 * 60e3).toISOString()
  const [status, points] = await Promise.all([
    fetch('/api/status').then((r) => r.json()),
    fetch(`/api/readings/history?resolution=raw&from=${from}`).then((r) => r.json()),
  ])
  const s = status.find((d) => d.device_id === DEVICE)
  if (s) Object.assign(live, { status: s.status, lastSeen: s.last_seen })
  points.forEach(push)
  connect()
}
