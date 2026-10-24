<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { Icon } from '@iconify/vue'

const authenticated = ref(false)
const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const storage = ref(null)
const selectedFile = ref(null)
const restoreJob = ref(null)
const backupJob = ref(null)
const confirmation = ref('')
let poller

const bytes = (value) => value == null ? '–' : new Intl.NumberFormat('id-ID', { maximumFractionDigits: 1 }).format(value / 1024 / 1024) + ' MB'

async function request(url, options) {
  const response = await fetch(url, options)
  if (!response.ok) throw new Error(await response.text())
  return response.status === 204 ? null : response.json()
}

async function check() {
  const result = await request('/api/auth/me')
  authenticated.value = result.authenticated
  if (result.authenticated) await loadStorage()
}

async function login() {
  error.value = ''; loading.value = true
  try {
    await request('/api/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ username: username.value, password: password.value }) })
    password.value = ''; authenticated.value = true; await loadStorage()
  } catch (e) { error.value = e.message } finally { loading.value = false }
}

async function logout() { await request('/api/auth/logout', { method: 'POST' }); authenticated.value = false; storage.value = null }
async function loadStorage() { try { storage.value = await request('/api/admin/storage') } catch (e) { error.value = e.message } }
async function createBackup() { error.value = ''; try { backupJob.value = await request('/api/admin/backup', { method: 'POST' }); poll(backupJob.value, 'backup') } catch (e) { error.value = e.message } }
function selectFile(event) { selectedFile.value = event.target.files[0] || null; restoreJob.value = null; confirmation.value = '' }
async function uploadRestore() {
  if (!selectedFile.value) return
  error.value = ''; loading.value = true
  try { const form = new FormData(); form.append('backup', selectedFile.value); restoreJob.value = await request('/api/admin/restore', { method: 'POST', body: form }) } catch (e) { error.value = e.message } finally { loading.value = false }
}
async function confirmRestore() {
  try { restoreJob.value = await request(`/api/admin/restore/${restoreJob.value.id}/confirm`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ confirmation: confirmation.value }) }); poll(restoreJob.value, 'restore') } catch (e) { error.value = e.message }
}
function poll(job, kind) {
  clearInterval(poller)
  poller = setInterval(async () => {
    try {
      const next = await request(`/api/admin/jobs/${job.id}`)
      if (kind === 'backup') backupJob.value = next; else restoreJob.value = next
      if (['ready', 'completed', 'failed'].includes(next.status)) { clearInterval(poller); await loadStorage() }
    } catch (e) { error.value = e.message; clearInterval(poller) }
  }, 1000)
}

onMounted(check)
onBeforeUnmount(() => clearInterval(poller))
</script>

<template>
  <section v-if="!authenticated" class="login card">
    <div class="title"><Icon icon="solar:lock-keyhole-bold-duotone" /><div><h2>Pengaturan Admin</h2><p class="muted">Masuk untuk mengelola backup dan pemulihan database.</p></div></div>
    <form @submit.prevent="login">
      <label>Username <input v-model="username" autocomplete="username" required /></label>
      <label>Password <input v-model="password" type="password" autocomplete="current-password" required /></label>
      <p v-if="error" class="error">{{ error }}</p>
      <button class="on" :disabled="loading">{{ loading ? 'Memproses…' : 'Masuk' }}</button>
    </form>
    <p class="warning">Akses admin memakai HTTP. Gunakan hanya dari jaringan LAN tepercaya sampai HTTPS dipasang.</p>
  </section>

  <template v-else>
    <div class="card heading"><div><h2>Pengaturan</h2><p class="muted">Backup, restore, dan penggunaan penyimpanan.</p></div><button @click="logout">Keluar</button></div>
    <div class="card"><div class="section-title"><h3>Penyimpanan</h3><button @click="loadStorage">Muat ulang</button></div>
      <div v-if="storage" class="storage-grid"><div><span>Database</span><strong>{{ bytes(storage.database_bytes) }}</strong></div><div><span>Raw readings + indeks</span><strong>{{ bytes(storage.readings_bytes) }}</strong></div><div><span>Backup pengaman</span><strong>{{ bytes(storage.backup_bytes) }}</strong></div></div>
      <p v-else class="muted">Memuat penggunaan disk…</p><p class="muted">Log aplikasi dikelola Docker host dan tidak dihitung di halaman ini.</p>
      <ul v-if="storage?.backup_files?.length" class="backups"><li v-for="file in storage.backup_files" :key="file.name"><span>{{ file.name }}</span><span>{{ bytes(file.size) }}</span></li></ul>
    </div>
    <div class="card"><h3>Backup Database</h3><p class="muted">Buat backup PostgreSQL format .dump untuk diunduh dan dipulihkan kembali.</p><button class="on" :disabled="backupJob?.status === 'running'" @click="createBackup">Buat Backup</button>
      <p v-if="backupJob" class="job">Status backup: <strong>{{ backupJob.status }}</strong><a v-if="backupJob.status === 'ready'" :href="`/api/admin/backup/${backupJob.id}/download`">Unduh backup</a><span v-if="backupJob.error" class="error">{{ backupJob.error }}</span></p>
    </div>
    <div class="card"><h3>Restore Database</h3><p class="warning">Restore mengganti seluruh data monitoring. Data perangkat yang masuk saat proses restore tidak disimpan. Backup pengaman otomatis dibuat dan tidak dihapus otomatis.</p>
      <input type="file" accept=".dump,application/octet-stream" @change="selectFile" /><button :disabled="!selectedFile || loading" @click="uploadRestore">Validasi Backup</button>
      <div v-if="restoreJob?.status === 'awaiting_confirmation'" class="confirm"><p>Backup valid: {{ bytes(restoreJob.size) }}. Ketik <code>RESTORE</code> untuk melanjutkan.</p><input v-model="confirmation" placeholder="RESTORE" /><button class="danger" :disabled="confirmation !== 'RESTORE'" @click="confirmRestore">Pulihkan Database</button></div>
      <p v-if="restoreJob && restoreJob.status !== 'awaiting_confirmation'" class="job">Status restore: <strong>{{ restoreJob.status }}</strong><span v-if="restoreJob.error" class="error">{{ restoreJob.error }}</span></p>
    </div>
    <p v-if="error" class="error">{{ error }}</p>
  </template>
</template>

<style scoped>
h2,h3 { margin: 0; } .title,.heading,.section-title { display:flex; align-items:center; justify-content:space-between; gap:12px; } .title { justify-content:flex-start; } .title svg { font-size:36px; color:var(--accent); } form { display:grid; gap:12px; max-width:360px; margin-top:18px; } label { display:grid; gap:5px; } .warning { color:#b45309; } .error { color:var(--bad); white-space:pre-wrap; } .storage-grid { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:12px; margin:14px 0; } .storage-grid div { padding:12px; border:1px solid var(--border); border-radius:8px; } .storage-grid span,.storage-grid strong { display:block; } .storage-grid span { color:var(--muted); font-size:13px; } .storage-grid strong { font-size:20px; } .backups { padding:0; list-style:none; } .backups li { display:flex; justify-content:space-between; gap:12px; padding:6px 0; border-top:1px solid var(--border); font-variant-numeric:tabular-nums; } .job { display:flex; flex-wrap:wrap; gap:8px; align-items:center; } .job a { color:var(--accent); font-weight:600; } .confirm { display:grid; gap:9px; max-width:420px; margin-top:14px; } .danger { color:#fff; background:var(--bad); border-color:var(--bad); } @media(max-width:560px){ .storage-grid { grid-template-columns:1fr; } }
</style>
