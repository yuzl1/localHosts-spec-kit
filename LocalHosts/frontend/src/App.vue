<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { applyHosts, canWriteHosts, ensureWriteAccess, getHosts, saveHosts } from './services/hosts'

const loading = ref(true)
const syncing = ref(false)
const authorizing = ref(false)
const authorized = ref(false)
const error = ref('')
const status = ref('')

const workspace = reactive({
  composeEnabled: true,
  theme: 'auto',
  groups: [],
  entries: [],
  readonlySystemLines: [],
  systemHostsLines: []
})

const selectedGroupId = ref('system')
const groupText = ref('')
const groupTextArea = ref(null)

const groupModalOpen = ref(false)
const groupModalMode = ref('add')
const groupModalGroupId = ref('')
const groupModalName = ref('')

let saveTimer = 0
let lastAppliedKey = ''

const selectedGroup = computed(() => {
  if (selectedGroupId.value === 'system') return null
  return workspace.groups.find((g) => g.id === selectedGroupId.value) || null
})

function id() {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID()
  return String(Date.now()) + String(Math.random()).slice(2)
}

function toPlainWorkspace() {
  return {
    composeEnabled: workspace.composeEnabled,
    theme: workspace.theme,
    groups: workspace.groups.map((g) => ({ ...g })),
    entries: workspace.entries.map((e) => ({ ...e })),
    readonlySystemLines: [...workspace.readonlySystemLines],
    systemHostsLines: [...workspace.systemHostsLines]
  }
}

async function refresh() {
  loading.value = true
  error.value = ''
  status.value = ''
  try {
    const data = await getHosts()
    workspace.composeEnabled = !!data.composeEnabled
    workspace.theme = data.theme || 'auto'
    workspace.groups = data.groups || []
    workspace.entries = data.entries || []
    workspace.readonlySystemLines = data.readonlySystemLines || []
    workspace.systemHostsLines = data.systemHostsLines || []
    if (!selectedGroupId.value) selectedGroupId.value = 'system'
    syncGroupText()
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

watch(selectedGroupId, () => {
  syncGroupText()
})

function syncGroupText() {
  if (selectedGroupId.value === 'system') {
    groupText.value = workspace.systemHostsLines.join('\n')
    return
  }
  groupText.value = buildGroupText(selectedGroupId.value)
}

function scheduleSave() {
  globalThis.clearTimeout(saveTimer)
  saveTimer = globalThis.setTimeout(() => {
    persistAndApply()
  }, 2000)
}

async function persistAndApply() {
  if (!authorized.value) return
  const syncErr = syncEntriesFromText()
  if (syncErr) {
    error.value = syncErr
    return
  }

  const invalid = validateEntries()
  if (invalid) {
    error.value = invalid
    return
  }

  syncing.value = true
  error.value = ''
  status.value = ''
  try {
    const key = buildApplyKey()
    if (key === lastAppliedKey) {
      status.value = '已同步'
      return
    }
    const payload = toPlainWorkspace()
    await saveHosts(payload)
    await applyHosts(payload)
    lastAppliedKey = key
    try {
      const latest = await getHosts()
      workspace.systemHostsLines = latest.systemHostsLines || []
      workspace.readonlySystemLines = latest.readonlySystemLines || []
      if (selectedGroupId.value === 'system') {
        groupText.value = workspace.systemHostsLines.join('\n')
      }
    } catch (_) {
    }
    status.value = '已保存并写入 hosts'
  } catch (e) {
    error.value = formatApplyError(String(e))
  } finally {
    syncing.value = false
  }
}

async function authorize() {
  authorizing.value = true
  error.value = ''
  status.value = ''
  try {
    await ensureWriteAccess()
    authorized.value = true
    status.value = '授权成功，可编辑'
  } catch (e) {
    error.value = formatApplyError(String(e))
  } finally {
    authorizing.value = false
  }
}

function syncEntriesFromText() {
  if (selectedGroupId.value === 'system') return ''
  if (!selectedGroup.value) return '请选择一个分组'
  const parsed = parseGroupText(groupText.value, selectedGroupId.value)
  if (parsed.error) return parsed.error
  const existing = workspace.entries.filter((e) => e.groupId === selectedGroupId.value)
  const idByKey = new Map(existing.map((e) => [`${e.ip}\t${e.domain}\t${e.groupId}`, e.id]))
  for (const entry of parsed.entries) {
    const key = `${entry.ip}\t${entry.domain}\t${entry.groupId}`
    const existingID = idByKey.get(key)
    if (existingID) entry.id = existingID
  }
  const other = workspace.entries.filter((e) => e.groupId !== selectedGroupId.value)
  workspace.entries = [...other, ...parsed.entries]
  return ''
}

function buildApplyKey() {
  const groupPart = workspace.groups
    .slice()
    .sort((a, b) => (a.order || 0) - (b.order || 0))
    .map((g) => `${g.id}\t${g.name}\t${g.enabled ? 1 : 0}\t${g.order}`)
    .join('\n')
  const entryPart = workspace.entries.map((e) => `${e.groupId}\t${e.enabled ? 1 : 0}\t${e.ip}\t${e.domain}\t${e.comment || ''}`).join('\n')
  return `${workspace.composeEnabled ? 1 : 0}\n${groupPart}\n---\n${entryPart}`
}

function validateEntries() {
  for (const e of workspace.entries) {
    if (!e.ip || !e.domain) return '存在未填写的 IP 或域名'
    if (!isValidIP(e.ip)) return `IP 格式不合法: ${e.ip}`
    if (/\s/.test(e.domain) || e.domain.includes('#')) return `域名格式不合法: ${e.domain}`
  }
  return ''
}

function isValidIP(ip) {
  if (ip.includes(':')) return true
  const parts = ip.split('.')
  if (parts.length !== 4) return false
  for (const p of parts) {
    if (!/^\d+$/.test(p)) return false
    const n = Number(p)
    if (n < 0 || n > 255) return false
  }
  return true
}

function formatApplyError(raw) {
  if (raw.startsWith('CANCELLED:')) return '已取消授权，系统 hosts 未修改'
  if (raw.startsWith('PERMISSION_DENIED:')) return '权限不足，系统 hosts 未修改'
  if (raw.startsWith('WRITE_FAILED:')) return '写入失败，系统 hosts 未修改或已尝试回滚'
  return raw
}

function addGroup() {
  if (!authorized.value) return
  groupModalMode.value = 'add'
  groupModalGroupId.value = ''
  groupModalName.value = ''
  groupModalOpen.value = true
}

function openEditGroup(group) {
  if (!authorized.value) return
  groupModalMode.value = 'edit'
  groupModalGroupId.value = group.id
  groupModalName.value = group.name || ''
  groupModalOpen.value = true
}

function closeGroupModal() {
  groupModalOpen.value = false
}

function confirmGroupModal() {
  const name = (groupModalName.value || '').trim()
  if (!name) {
    error.value = '分组名不能为空'
    return
  }

  if (groupModalMode.value === 'edit') {
    const group = workspace.groups.find((g) => g.id === groupModalGroupId.value)
    if (group) group.name = name
  } else {
    const maxOrder = workspace.groups.reduce((m, g) => Math.max(m, Number(g.order) || 0), 0)
    const newId = 'group:' + name
    const exists = workspace.groups.some((g) => g.id === newId)
    workspace.groups.push({ id: exists ? id() : newId, name, enabled: true, order: maxOrder + 1 })
    selectedGroupId.value = exists ? workspace.groups[workspace.groups.length - 1].id : newId
  }

  closeGroupModal()
  scheduleSave()
}

function moveGroup(groupId, direction) {
  const groups = [...workspace.groups].sort((a, b) => (a.order || 0) - (b.order || 0))
  const index = groups.findIndex((g) => g.id === groupId)
  const targetIndex = index + direction
  if (index < 0 || targetIndex < 0 || targetIndex >= groups.length) return
  const a = groups[index]
  const b = groups[targetIndex]
  const tmp = a.order
  a.order = b.order
  b.order = tmp
  workspace.groups = groups
  scheduleSave()
}

onMounted(() => {
  refresh().then(async () => {
    try {
      const ok = await canWriteHosts()
      if (ok) {
        authorized.value = true
        return
      }
    } catch (_) {
      authorized.value = false
    }
    await authorize()
  })
})

function buildGroupText(groupId) {
  const entries = workspace.entries.filter((e) => e.groupId === groupId)
  return entries.map((e) => toEntryLine(e)).join('\n')
}

function toEntryLine(e) {
  const body = `${e.ip} ${e.domain}${e.comment ? ` # ${e.comment}` : ''}`
  return e.enabled ? body : `# ${body}`
}

function parseGroupText(text, groupId) {
  const out = []
  const lines = text.split('\n')
  for (const line of lines) {
    const parsed = parseEntryLine(line, groupId)
    if (parsed.error) return { entries: [], error: parsed.error }
    out.push(...parsed.entries)
  }
  return { entries: out, error: '' }
}

function parseEntryLine(line, groupId) {
  const trimmed = line.trim()
  if (!trimmed) return { entries: [], error: '' }

  let enabled = true
  let s = trimmed
  if (s.startsWith('#')) {
    enabled = false
    s = s.replace(/^#\s*/, '')
  }

  const idx = s.indexOf('#')
  const body = idx >= 0 ? s.slice(0, idx).trim() : s.trim()
  const comment = idx >= 0 ? s.slice(idx + 1).trim() : ''

  const parts = body.split(/\s+/).filter(Boolean)
  if (parts.length < 2) return { entries: [], error: `无法解析条目: ${line}` }

  const ip = parts[0]
  if (!isValidIP(ip)) return { entries: [], error: `IP 格式不合法: ${ip}` }

  const domains = parts.slice(1)
  for (const d of domains) {
    if (/\s/.test(d) || d.includes('#')) return { entries: [], error: `域名格式不合法: ${d}` }
  }

  const entries = domains.map((domain) => ({
    id: id(),
    ip,
    domain,
    comment,
    enabled,
    groupId,
    raw: line
  }))

  return { entries, error: '' }
}

function lineIndexAtCursor(text, cursor) {
  const before = text.slice(0, cursor)
  if (!before) return 0
  return before.split('\n').length - 1
}
</script>

<template>
  <div class="app" :data-theme="workspace.theme">
    <header class="header">
      <div class="title">LocalHosts</div>
      <div class="actions">
        <div class="theme-select">
          <select class="theme-native" v-model="workspace.theme" @change="scheduleSave">
            <option value="auto">跟随系统</option>
            <option value="light">浅色</option>
            <option value="dark">深色</option>
          </select>
          <div class="theme-arrow">▾</div>
        </div>
        <label class="toggle">
          <input type="checkbox" :disabled="!authorized" v-model="workspace.composeEnabled" @change="scheduleSave" />
          <span>按分组组装</span>
        </label>
        <button v-if="!authorized" class="btn primary" :disabled="loading || authorizing" @click="authorize">
          {{ authorizing ? '授权中…' : '授权写入' }}
        </button>
        <div v-else class="hint">{{ syncing ? '同步中…' : (status || '已授权') }}</div>
      </div>
    </header>

    <div v-if="error" class="banner error">{{ error }}</div>

    <div v-if="loading" class="loading">加载中…</div>

    <div v-else class="layout">
      <aside class="sidebar">
        <div class="sidebar-header">
          <div class="section-title">分组</div>
          <button class="btn small" :disabled="!authorized" @click="addGroup">新增分组</button>
        </div>

        <div class="group-item system" :class="{ active: selectedGroupId === 'system' }" @click="selectedGroupId = 'system'">
          <div class="group-enable-spacer"></div>
          <div class="group-name">系统hosts</div>
        </div>

        <div
          v-for="g in [...workspace.groups].sort((a, b) => (a.order || 0) - (b.order || 0))"
          :key="g.id"
          class="group-item"
          :class="{ active: selectedGroupId === g.id }"
          @click="selectedGroupId = g.id"
        >
          <label class="group-enable" @click.stop>
            <input type="checkbox" :disabled="!authorized" v-model="g.enabled" @change="scheduleSave" />
          </label>
          <div class="group-name">{{ g.name }}</div>
          <div class="group-move" @click.stop>
            <button class="btn icon" :disabled="!authorized" @click="moveGroup(g.id, -1)">↑</button>
            <button class="btn icon" :disabled="!authorized" @click="moveGroup(g.id, 1)">↓</button>
            <button class="btn icon" :disabled="!authorized" @click="openEditGroup(g)">编辑</button>
          </div>
        </div>
      </aside>

      <main class="main">
        <div class="main-header">
          <div class="main-title">
            <div class="section-title">条目</div>
            <div class="subhint">一域名一条</div>
          </div>
          <div class="entry-actions">
            <span class="subhint" style="margin-right: 8px;">作者: yuzl (yuzl6@foxmail.com)</span>
          </div>
        </div>

        <div class="editor">
          <div class="editor-hint" v-if="!authorized">未授权写入系统 hosts，当前仅可查看</div>
          <div class="editor-hint" v-else-if="selectedGroupId === 'system'">系统hosts：本地 hosts 全量内容（只读）</div>
          <textarea
            ref="groupTextArea"
            class="textarea"
            v-model="groupText"
            spellcheck="false"
            :readonly="!authorized || selectedGroupId === 'system'"
            @input="scheduleSave"
            placeholder="# 127.0.0.1 example.com # comment"
          />
        </div>
      </main>
    </div>

    <div v-if="groupModalOpen" class="modal-mask" @click.self="closeGroupModal">
      <div class="modal">
        <div class="modal-title">{{ groupModalMode === 'edit' ? '修改分组' : '新增分组' }}</div>
        <div class="modal-row">
          <div class="modal-field full">
            <div class="modal-field-label">分组名</div>
            <input class="input" v-model="groupModalName" placeholder="例如：Dev" />
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn" @click="closeGroupModal">取消</button>
          <button class="btn primary" @click="confirmGroupModal">确定</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Dark/Light mode theme variables */
.app {
  --bg-main: #f8fafc;
  --bg-surface: #ffffff;
  --bg-surface-active: #eff6ff;
  --bg-banner-ok: #dcfce7;
  --bg-banner-error: #fee2e2;
  --bg-modal-mask: rgba(15, 23, 42, 0.45);
  --bg-table-head: #f1f5f9;
  
  --text-primary: #0f172a;
  --text-secondary: #64748b;
  --text-ok: #166534;
  --text-error: #991b1b;
  --text-btn-primary: #ffffff;
  
  --border-color: #e2e8f0;
  --border-input: #cbd5e1;
  --border-focus: #93c5fd;
  
  --btn-bg: #ffffff;
  --btn-primary-bg: #2563eb;
  --btn-primary-border: #2563eb;
  --btn-danger-border: #ef4444;
  --btn-danger-text: #ef4444;

  height: 100vh;
  display: flex;
  flex-direction: column;
  color: var(--text-primary);
  background: var(--bg-main);
  transition: background-color 0.3s, color 0.3s;
}

.app[data-theme="dark"] {
  --bg-main: #0f172a;
  --bg-surface: #1e293b;
  --bg-surface-active: #334155;
  --bg-banner-ok: #064e3b;
  --bg-banner-error: #7f1d1d;
  --bg-modal-mask: rgba(0, 0, 0, 0.6);
  --bg-table-head: #1e293b;
  
  --text-primary: #f8fafc;
  --text-secondary: #94a3b8;
  --text-ok: #a7f3d0;
  --text-error: #fca5a5;
  --text-btn-primary: #ffffff;
  
  --border-color: #334155;
  --border-input: #475569;
  --border-focus: #3b82f6;
  
  --btn-bg: #1e293b;
  --btn-primary-bg: #3b82f6;
  --btn-primary-border: #3b82f6;
  --btn-danger-border: #f87171;
  --btn-danger-text: #f87171;
}

@media (prefers-color-scheme: dark) {
  .app[data-theme="auto"] {
    --bg-main: #0f172a;
    --bg-surface: #1e293b;
    --bg-surface-active: #334155;
    --bg-banner-ok: #064e3b;
    --bg-banner-error: #7f1d1d;
    --bg-modal-mask: rgba(0, 0, 0, 0.6);
    --bg-table-head: #1e293b;
    
    --text-primary: #f8fafc;
    --text-secondary: #94a3b8;
    --text-ok: #a7f3d0;
    --text-error: #fca5a5;
    --text-btn-primary: #ffffff;
    
    --border-color: #334155;
    --border-input: #475569;
    --border-focus: #3b82f6;
    
    --btn-bg: #1e293b;
    --btn-primary-bg: #3b82f6;
    --btn-primary-border: #3b82f6;
    --btn-danger-border: #f87171;
    --btn-danger-text: #f87171;
  }
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-surface);
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
}

.title {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.01em;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary);
}

.theme-select {
  position: relative;
  display: inline-flex;
  align-items: center;
  height: 32px;
  padding: 0 10px;
  border-radius: 8px;
  border: 1px solid var(--border-input);
  background: var(--bg-surface);
  color: var(--text-primary);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.theme-select:focus-within {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.25);
}

.theme-native {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  border: none;
  background: transparent;
  color: inherit;
  font-size: 13px;
  font-weight: 600;
  padding-right: 18px;
  cursor: pointer;
  outline: none;
}

.theme-arrow {
  position: absolute;
  right: 10px;
  font-size: 12px;
  color: var(--text-secondary);
  pointer-events: none;
  transform: translateY(-1px);
}

.btn {
  border: 1px solid var(--border-input);
  background: var(--btn-bg);
  color: var(--text-primary);
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:hover:not(:disabled) {
  border-color: var(--border-focus);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn.primary {
  border-color: var(--btn-primary-border);
  background: var(--btn-primary-bg);
  color: var(--text-btn-primary);
}

.btn.primary:hover:not(:disabled) {
  opacity: 0.9;
}

.btn.small {
  padding: 4px 8px;
  font-size: 12px;
}

.btn.danger {
  border-color: var(--btn-danger-border);
  color: var(--btn-danger-text);
}

.btn.icon {
  padding: 2px 6px;
  font-size: 12px;
}

.banner {
  padding: 10px 16px;
  font-size: 13px;
  font-weight: 500;
}

.banner.error {
  background: var(--bg-banner-error);
  color: var(--text-error);
}

.banner.ok {
  background: var(--bg-banner-ok);
  color: var(--text-ok);
}

.loading {
  padding: 18px 16px;
  font-size: 13px;
  color: var(--text-secondary);
}

.layout {
  flex: 1;
  display: grid;
  grid-template-columns: 240px 1fr;
  min-height: 0;
}

.sidebar {
  border-right: 1px solid var(--border-color);
  background: var(--bg-surface);
  padding: 16px 12px;
  overflow: auto;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.section-title {
  font-weight: 700;
  font-size: 13px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.group-item {
  display: grid;
  grid-template-columns: 20px 1fr auto;
  gap: 8px;
  align-items: center;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
  margin-bottom: 2px;
}

.group-item:hover {
  background: var(--bg-main);
}

.group-item.active {
  background: var(--bg-surface-active);
}

.group-name {
  font-size: 13px;
  font-weight: 500;
}

.group-enable input {
  cursor: pointer;
  accent-color: var(--btn-primary-bg);
}

.group-name-input {
  border: 1px solid transparent;
  border-radius: 4px;
  padding: 4px 6px;
  font-size: 13px;
  width: 100%;
  background: transparent;
  color: var(--text-primary);
}

.group-name-input:focus {
  outline: none;
  border-color: var(--border-focus);
  background: var(--bg-surface);
}

.group-move {
  display: flex;
  gap: 4px;
}

.main {
  padding: 16px 20px;
  overflow: auto;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: var(--bg-main);
}

.main-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.main-title {
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.subhint {
  font-size: 12px;
  color: var(--text-secondary);
}

.editor {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-surface);
  padding: 12px;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
}

.editor-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.textarea {
  width: 100%;
  flex: 1;
  box-sizing: border-box;
  min-height: 280px;
  resize: none;
  border: 1px solid var(--border-input);
  border-radius: 6px;
  padding: 12px;
  font-size: 13px;
  line-height: 1.6;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  background: var(--bg-surface);
  color: var(--text-primary);
  transition: border-color 0.2s;
}

.textarea:focus {
  outline: none;
  border-color: var(--border-focus);
}

.textarea:read-only {
  background: var(--bg-main);
  color: var(--text-secondary);
}

.input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid var(--border-input);
  border-radius: 6px;
  padding: 6px 10px;
  font-size: 13px;
  background: var(--bg-surface);
  color: var(--text-primary);
}

.input:focus {
  outline: none;
  border-color: var(--border-focus);
}

.modal-mask {
  position: fixed;
  inset: 0;
  background: var(--bg-modal-mask);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  backdrop-filter: blur(2px);
}

.modal {
  width: 480px;
  max-width: 100%;
  background: var(--bg-surface);
  border-radius: 10px;
  padding: 20px;
  border: 1px solid var(--border-color);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

.modal-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
}

.modal-row {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
}

.modal-label {
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-primary);
}

.modal-field {
  flex: 1;
  min-width: 0;
}

.modal-field.full {
  flex: 1;
}

.modal-field-label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}
</style>
