<template>
  <div class="settings-view">
    <!-- Header -->
    <header class="view-header">
      <div class="header-text">
        <h1 class="view-title">ตั้งค่า</h1>
        <p class="view-subtitle">
          จัดการ API Keys และตั้งค่าการทำงานของระบบ
        </p>
      </div>
      <div class="header-actions">
        <button class="btn-outline" @click="handleResetCooldowns" :disabled="isSaving">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/>
            <path d="M3 3v5h5"/>
            <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/>
            <path d="M16 21h5v-5"/>
          </svg>
          รีเซ็ตโควตา AI
        </button>
      </div>
    </header>

    <!-- Status Toast -->
    <div v-if="statusMessage" :class="['status-banner', statusType]">
      <span>{{ statusMessage }}</span>
      <button class="banner-close" @click="statusMessage = ''">✕</button>
    </div>

    <div class="settings-container">
      <!-- Summary Card -->
      <div class="settings-card allocator-card">
        <div class="card-header-flex">
          <h3 class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="3"/>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
            สถานะการเชื่อมต่อ AI
          </h3>
          <span class="tier-pill">{{ allocationTierLabel }}</span>
        </div>

        <div class="metrics-grid">
          <div class="metric-box">
            <span class="metric-label">โปรเจกต์ที่เปิดใช้</span>
            <span class="metric-value">{{ settings.active_key_count }} <small>/ {{ settings.total_project_count }}</small></span>
          </div>
          <div class="metric-box">
            <span class="metric-label">เส้นทางที่ใช้งาน</span>
            <span class="metric-value">{{ settings.primary_routes }} <small>เส้นทาง</small></span>
          </div>
          <div class="metric-box">
            <span class="metric-label">จำนวน Workers</span>
            <span class="metric-value">{{ settings.allocated_workers }} <small>Workers ({{ settings.batch_size }} รูป/รอบ)</small></span>
          </div>
          <div class="metric-box">
            <span class="metric-label">โควตารวม</span>
            <span class="metric-value">{{ settings.estimated_rpm }} <small>RPM | {{ settings.estimated_rpd.toLocaleString() }} RPD</small></span>
          </div>
        </div>

        <p class="allocator-desc">
          {{ allocationDescription }}
        </p>
      </div>

      <!-- Section: Add / Update Project & API Key -->
      <div class="settings-card">
        <div class="card-header-flex">
          <h3 class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
            </svg>
            จัดการโปรเจกต์และ API Keys
          </h3>
        </div>

        <!-- Mode Selector Tabs -->
        <div class="mode-tabs">
          <button
            :class="['mode-tab', { active: formMode === 'existing' }]"
            @click="switchMode('existing')"
          >
            แก้ไขโปรเจกต์เดิม
          </button>
          <button
            :class="['mode-tab', { active: formMode === 'new' }]"
            :disabled="settings.projects.length >= 10"
            :title="settings.projects.length >= 10 ? 'จำกัดสูงสุด 10 โปรเจกต์ (ครบแล้ว)' : 'เพิ่มโปรเจกต์ใหม่'"
            @click="switchMode('new')"
          >
            เพิ่มโปรเจกต์ใหม่ {{ settings.projects.length >= 10 ? '(ครบ 10 แล้ว)' : '' }}
          </button>
        </div>

        <!-- Form Body -->
        <div class="api-form-box">
          <div v-if="formMode === 'existing'" class="form-grid">
            <div class="form-group">
              <label class="form-label">โปรเจกต์</label>
              <select v-model="selectedProjectId" class="input-control" @change="onSelectExistingProject">
                <option v-for="proj in settings.projects" :key="proj.id" :value="proj.id">
                  {{ proj.project_name }} {{ proj.project_number ? `(${proj.project_number})` : '' }}
                  — {{ proj.api_key ? 'มีคีย์แล้ว' : 'ยังไม่มีคีย์' }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">Project Number</label>
              <input
                v-model="formProjectNumber"
                type="text"
                class="input-control"
                placeholder="เช่น 1061737622178"
              />
            </div>
          </div>

          <div v-else class="form-grid">
            <div class="form-group">
              <label class="form-label">ชื่อโปรเจกต์ (ไม่บังคับ)</label>
              <input
                v-model="formProjectName"
                type="text"
                class="input-control"
                placeholder="เช่น Project 11"
              />
            </div>
            <div class="form-group">
              <label class="form-label">Project Number (ไม่บังคับ)</label>
              <input
                v-model="formProjectNumber"
                type="text"
                class="input-control"
                placeholder="เช่น 1061737622178"
              />
            </div>
          </div>

          <div class="form-group full-width">
            <label class="form-label">
              API Key
              <span class="label-hint">(วางได้หลายคีย์พร้อมกัน คั่นด้วยลูกน้ำหรือขึ้นบรรทัดใหม่)</span>
            </label>
            <div class="key-input-row">
              <textarea
                v-model="formApiKey"
                rows="2"
                class="input-control textarea-key"
                placeholder="AQ.Ab8RN6... หรือ AIzaSy..."
              ></textarea>
              <button class="btn-primary btn-save-key" @click="handleSaveProjectKey" :disabled="isSaving">
                {{ formMode === 'existing' ? 'บันทึกคีย์' : 'เพิ่มโปรเจกต์' }}
              </button>
            </div>
          </div>
        </div>

        <!-- Project Pool List -->
        <div class="project-list-header">
          <span>โปรเจกต์ทั้งหมด ({{ settings.projects.length }})</span>
          <button class="btn-text" @click="showAllKeys = !showAllKeys">
            {{ showAllKeys ? 'ซ่อนคีย์' : 'แสดงคีย์ทั้งหมด' }}
          </button>
        </div>

        <div class="project-table-wrap">
          <table class="project-table">
            <thead>
              <tr>
                <th style="width: 46px;">#</th>
                <th>โปรเจกต์</th>
                <th>API Key</th>
                <th style="width: 150px;">สถานะ</th>
                <th style="width: 190px; text-align: right;">จัดการ</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(proj, idx) in settings.projects" :key="proj.id" :class="{ 'row-disabled': !proj.enabled }">
                <td class="col-idx">{{ idx + 1 }}</td>
                <td>
                  <div class="proj-name">{{ proj.project_name }}</div>
                  <div class="proj-num" v-if="proj.project_number">{{ proj.project_number }}</div>
                </td>
                <td class="col-key">
                  <div v-if="inlineEditId === proj.id" class="inline-edit-row">
                    <input
                      v-model="inlineApiKey"
                      type="text"
                      class="input-control input-sm"
                      placeholder="วาง API Key ใหม่..."
                      @keyup.enter="saveInlineKey(proj)"
                    />
                    <button class="btn-xs btn-primary" @click="saveInlineKey(proj)">บันทึก</button>
                    <button class="btn-xs btn-ghost" @click="inlineEditId = ''">ยกเลิก</button>
                  </div>
                  <div v-else class="key-display">
                    <code v-if="proj.api_key" class="key-code">
                      {{ showAllKeys ? proj.api_key : maskApiKey(proj.api_key) }}
                    </code>
                    <span v-else class="missing-key-text">ยังไม่มีคีย์</span>
                  </div>
                </td>
                <td>
                  <span :class="['status-pill', statusClass(proj.status)]">
                    {{ statusLabel(proj) }}
                  </span>
                </td>
                <td class="col-actions">
                  <button
                    class="btn-xs btn-outline"
                    @click="startInlineEdit(proj)"
                    title="แก้ไข API Key"
                  >
                    {{ proj.api_key ? 'เปลี่ยนคีย์' : 'ใส่คีย์' }}
                  </button>
                  <label class="toggle-switch toggle-sm" :title="proj.enabled ? 'ปิดใช้งาน' : 'เปิดใช้งาน'">
                    <input
                      type="checkbox"
                      :checked="proj.enabled"
                      :disabled="!proj.api_key"
                      @change="handleToggleProject(proj)"
                    />
                    <span class="slider"></span>
                  </label>
                  <button
                    class="btn-xs btn-danger-ghost"
                    @click="handleDeleteProject(proj)"
                    title="ลบโปรเจกต์"
                  >
                    ลบ
                  </button>
                </td>
              </tr>
              <tr v-if="settings.projects.length === 0">
                <td colspan="5" class="empty-row">
                  ยังไม่มีโปรเจกต์ในระบบ
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Section: AI & Backup Engine Configuration -->
      <div class="settings-card">
        <h3 class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
            <line x1="8" y1="21" x2="16" y2="21"/>
            <line x1="12" y1="17" x2="12" y2="21"/>
          </svg>
          ตั้งค่าทั่วไป
        </h3>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">วิเคราะห์ภาพอัตโนมัติ</span>
            <span class="setting-desc">สร้างคำอธิบายและแท็กทันทีหลังย้ายรูปเสร็จ</span>
          </div>
          <div class="setting-control">
            <label class="toggle-switch">
              <input
                type="checkbox"
                v-model="settings.auto_analyze"
                @change="handleSaveSystemSettings"
              />
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">โมเดล AI</span>
            <span class="setting-desc">เลือกโมเดลหลักสำหรับวิเคราะห์รูปภาพ</span>
          </div>
          <div class="setting-control">
            <select
              v-model="settings.primary_model"
              class="select-box"
              @change="handleSaveSystemSettings"
            >
              <option value="auto">อัตโนมัติ (แนะนำ)</option>
              <option value="gemini-3.5-flash-lite">Gemini 3.5 Flash-Lite</option>
              <option value="gemini-3.1-flash-lite">Gemini 3.1 Flash-Lite</option>
              <option value="gemini-3.8-flash">Gemini 3.8 Flash</option>
            </select>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">ป้องกันไฟล์ซ้ำ</span>
            <span class="setting-desc">ข้ามไฟล์ที่มีชื่อหรือค่า SHA256 ซ้ำกับที่มีอยู่แล้ว</span>
          </div>
          <div class="setting-control">
            <label class="toggle-switch">
              <input
                type="checkbox"
                v-model="settings.duplicate_guard"
                @change="handleSaveSystemSettings"
              />
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">จำนวน Worker คัดลอกไฟล์</span>
            <span class="setting-desc">จำนวนงานคัดลอกไฟล์พร้อมกัน</span>
          </div>
          <div class="setting-control">
            <select
              v-model.number="settings.copy_workers"
              class="select-box"
              @change="handleSaveSystemSettings"
            >
              <option :value="4">4 Workers</option>
              <option :value="8">8 Workers</option>
              <option :value="16">16 Workers (แนะนำ)</option>
              <option :value="24">24 Workers</option>
              <option :value="32">32 Workers</option>
            </select>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">ฐานข้อมูล SQLite</span>
            <span class="setting-desc">บันทึกข้อมูลแบบ WAL Mode</span>
          </div>
          <div class="setting-control">
            <span class="status-pill pill-active">ปกติ</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface ProjectApiKeyEntry {
  id: string
  project_name: string
  project_number: string
  api_key: string
  enabled: boolean
  status: string
  cooldown_count: number
}

interface AppSettingsDTO {
  auto_analyze: boolean
  primary_model: string
  duplicate_guard: boolean
  copy_workers: number
  active_key_count: number
  total_project_count: number
  primary_routes: number
  total_routes: number
  allocated_workers: number
  batch_size: number
  estimated_rpm: number
  estimated_rpd: number
  projects: ProjectApiKeyEntry[]
}

const settings = ref<AppSettingsDTO>({
  auto_analyze: true,
  primary_model: 'auto',
  duplicate_guard: true,
  copy_workers: 16,
  active_key_count: 0,
  total_project_count: 0,
  primary_routes: 0,
  total_routes: 0,
  allocated_workers: 6,
  batch_size: 25,
  estimated_rpm: 0,
  estimated_rpd: 0,
  projects: []
})

const formMode = ref<'existing' | 'new'>('existing')
const selectedProjectId = ref<string>('')
const formProjectName = ref<string>('')
const formProjectNumber = ref<string>('')
const formApiKey = ref<string>('')

const inlineEditId = ref<string>('')
const inlineApiKey = ref<string>('')
const showAllKeys = ref<boolean>(false)
const isSaving = ref<boolean>(false)
const statusMessage = ref<string>('')
const statusType = ref<'success' | 'error'>('success')

function getBackendBridge() {
  const w = window as any
  return w?.go?.main?.App || w?.go?.service?.BackupService || null
}

function showStatus(msg: string, type: 'success' | 'error' = 'success') {
  statusMessage.value = msg
  statusType.value = type
  setTimeout(() => {
    if (statusMessage.value === msg) {
      statusMessage.value = ''
    }
  }, 4500)
}

const allocationTierLabel = computed(() => {
  const k = settings.value.active_key_count
  if (k <= 0) return 'ยังไม่เปิดใช้งาน'
  return `เปิดใช้ ${k} โปรเจกต์`
})

const allocationDescription = computed(() => {
  const k = settings.value.active_key_count
  const w = settings.value.allocated_workers
  const r = settings.value.primary_routes
  if (k <= 0) {
    return 'กรุณาเพิ่ม API Key อย่างน้อย 1 โปรเจกต์เพื่อเริ่มใช้งาน'
  }
  return `เปิดใช้งาน ${k} คีย์ (${r} เส้นทาง, ${w} Workers)`
})

function applySettingsResponse(res: AppSettingsDTO | null | undefined) {
  if (!res) return
  settings.value = res
  if (!selectedProjectId.value && res.projects && res.projects.length > 0) {
    // Default to the first project that is missing a key, or the first project
    const missing = res.projects.find(p => !p.api_key)
    selectedProjectId.value = missing ? missing.id : (res.projects[0]?.id ?? '')
    onSelectExistingProject()
  }
}

async function loadSettings() {
  const bridge = getBackendBridge()
  if (!bridge || typeof bridge.GetAppSettings !== 'function') return
  try {
    const res = await bridge.GetAppSettings()
    applySettingsResponse(res)
  } catch (err: any) {
    console.error('Failed to load settings:', err)
  }
}

function switchMode(mode: 'existing' | 'new') {
  formMode.value = mode
  if (mode === 'existing') {
    onSelectExistingProject()
  } else {
    formProjectName.value = ''
    formProjectNumber.value = ''
    formApiKey.value = ''
  }
}

function onSelectExistingProject() {
  const found = settings.value.projects.find(p => p.id === selectedProjectId.value)
  if (found) {
    formProjectName.value = found.project_name
    formProjectNumber.value = found.project_number || ''
  }
}

async function handleSaveProjectKey() {
  const bridge = getBackendBridge()
  if (!bridge || typeof bridge.SaveProjectApiKey !== 'function') return

  const trimmedKey = formApiKey.value.trim()
  if (!trimmedKey) {
    showStatus('กรุณาวาง Gemini API Key ก่อนกดบันทึก', 'error')
    return
  }

  isSaving.value = true
  try {
    const targetId = formMode.value === 'existing' ? selectedProjectId.value : 'new'
    const res = await bridge.SaveProjectApiKey(
      targetId,
      formProjectName.value.trim(),
      formProjectNumber.value.trim(),
      trimmedKey
    )
    applySettingsResponse(res)
    formApiKey.value = ''

    // Move selection to next empty project if in existing mode
    if (formMode.value === 'existing' && res?.projects) {
      const nextEmpty = res.projects.find((p: ProjectApiKeyEntry) => !p.api_key)
      if (nextEmpty) {
        selectedProjectId.value = nextEmpty.id
        onSelectExistingProject()
      }
    } else {
      formProjectName.value = ''
      formProjectNumber.value = ''
    }

    showStatus('บันทึก API Key เรียบร้อยแล้ว')
  } catch (err: any) {
    showStatus(err?.message || String(err), 'error')
  } finally {
    isSaving.value = false
  }
}

function startInlineEdit(proj: ProjectApiKeyEntry) {
  inlineEditId.value = proj.id
  inlineApiKey.value = proj.api_key || ''
}

async function saveInlineKey(proj: ProjectApiKeyEntry) {
  const bridge = getBackendBridge()
  if (!bridge || typeof bridge.SaveProjectApiKey !== 'function') return

  const keyVal = inlineApiKey.value.trim()
  if (!keyVal) {
    showStatus('กรุณาระบุ API Key', 'error')
    return
  }

  isSaving.value = true
  try {
    const res = await bridge.SaveProjectApiKey(proj.id, proj.project_name, proj.project_number, keyVal)
    applySettingsResponse(res)
    inlineEditId.value = ''
    inlineApiKey.value = ''
    showStatus(`อัปเดตคีย์ ${proj.project_name} เรียบร้อยแล้ว`)
  } catch (err: any) {
    showStatus(err?.message || String(err), 'error')
  } finally {
    isSaving.value = false
  }
}

async function handleToggleProject(proj: ProjectApiKeyEntry) {
  const bridge = getBackendBridge()
  if (!bridge || typeof bridge.ToggleProjectApiKey !== 'function') return

  try {
    const res = await bridge.ToggleProjectApiKey(proj.id, !proj.enabled)
    applySettingsResponse(res)
    showStatus(`อัปเดตสถานะ ${proj.project_name} เรียบร้อยแล้ว`)
  } catch (err: any) {
    showStatus(err?.message || String(err), 'error')
  }
}

async function handleDeleteProject(proj: ProjectApiKeyEntry) {
  if (!confirm(`ต้องการลบ "${proj.project_name}" ใช่หรือไม่?`)) return

  const bridge = getBackendBridge()
  if (!bridge || typeof bridge.DeleteProjectApiKey !== 'function') return

  try {
    const res = await bridge.DeleteProjectApiKey(proj.id)
    applySettingsResponse(res)
    showStatus(`ลบ ${proj.project_name} เรียบร้อยแล้ว`)
  } catch (err: any) {
    showStatus(err?.message || String(err), 'error')
  }
}

async function handleSaveSystemSettings() {
  const bridge = getBackendBridge()
  if (!bridge || typeof bridge.UpdateSystemSettings !== 'function') return

  try {
    const res = await bridge.UpdateSystemSettings(
      settings.value.auto_analyze,
      settings.value.primary_model,
      settings.value.duplicate_guard,
      Number(settings.value.copy_workers) || 16
    )
    applySettingsResponse(res)
    showStatus('บันทึกการตั้งค่าแล้ว')
  } catch (err: any) {
    showStatus(err?.message || String(err), 'error')
  }
}

async function handleResetCooldowns() {
  const bridge = getBackendBridge()
  if (!bridge || typeof bridge.ResetAiQuotaCooldowns !== 'function') return

  isSaving.value = true
  try {
    const res = await bridge.ResetAiQuotaCooldowns()
    applySettingsResponse(res)
    showStatus('รีเซ็ตโควตาเรียบร้อยแล้ว')
  } catch (err: any) {
    showStatus(err?.message || String(err), 'error')
  } finally {
    isSaving.value = false
  }
}

function maskApiKey(key: string): string {
  if (!key) return ''
  if (key.length <= 14) return key
  return `${key.slice(0, 10)}••••••••••••${key.slice(-6)}`
}

function statusLabel(proj: ProjectApiKeyEntry): string {
  switch (proj.status) {
    case 'active':
      return 'พร้อมใช้งาน'
    case 'cooldown':
      return `พักชั่วคราว (${proj.cooldown_count} โมเดล)`
    case 'daily_limit':
      return 'โควตารายวันเต็ม'
    case 'missing_key':
      return 'รอใส่ API Key'
    default:
      return 'ปิดใช้งาน'
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'active':
      return 'pill-active'
    case 'cooldown':
      return 'pill-warn'
    case 'daily_limit':
      return 'pill-danger'
    case 'missing_key':
      return 'pill-missing'
    default:
      return 'pill-muted'
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.settings-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 22px 28px 28px 28px;
  box-sizing: border-box;
  background-color: #F4F5FA;
  overflow-y: auto;
  font-family: 'Prompt', 'Inter', sans-serif;
}

.view-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
  max-width: 980px;
}

.view-title {
  font-size: 22px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px 0;
}

.view-subtitle {
  font-size: 13px;
  color: #6B7280;
  margin: 0;
}

.status-banner {
  max-width: 980px;
  padding: 10px 14px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.status-banner.success {
  background: #ECFDF5;
  border: 1px solid #A7F3D0;
  color: #065F46;
}

.status-banner.error {
  background: #FEF2F2;
  border: 1px solid #FECACA;
  color: #991B1B;
}

.banner-close {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  font-size: 14px;
}

.settings-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 980px;
}

.settings-card {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 14px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}

.allocator-card {
  background: linear-gradient(180deg, #FFFFFF 0%, #F9FAFB 100%);
  border: 1px solid #C7D2FE;
}

.card-header-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.card-title {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.tier-pill {
  background: #EEF2FF;
  color: #4338CA;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 999px;
  border: 1px solid #C7D2FE;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}

.metric-box {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 10px;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-label {
  font-size: 11.5px;
  color: #6B7280;
  font-weight: 500;
}

.metric-value {
  font-size: 18px;
  font-weight: 700;
  color: #111827;
}

.metric-value small {
  font-size: 11.5px;
  font-weight: 500;
  color: #6B7280;
}

.allocator-desc {
  margin: 0;
  font-size: 12.5px;
  color: #374151;
  background: #EEF2FF;
  padding: 9px 12px;
  border-radius: 8px;
  line-height: 1.5;
}

/* Mode Tabs */
.mode-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
}

.mode-tab {
  flex: 1;
  padding: 9px 14px;
  border-radius: 8px;
  border: 1px solid #D1D5DB;
  background: #F9FAFB;
  color: #374151;
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.15s;
}

.mode-tab.active {
  background: #4F46E5;
  color: #FFFFFF;
  border-color: #4F46E5;
}

.api-form-box {
  background: #F9FAFB;
  border: 1px solid #E5E7EB;
  border-radius: 10px;
  padding: 14px;
  margin-bottom: 16px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.form-label {
  font-size: 12.5px;
  font-weight: 600;
  color: #1F2937;
}

.label-hint {
  font-weight: 400;
  color: #6B7280;
  font-size: 11.5px;
  margin-left: 4px;
}

.input-control {
  background: #FFFFFF;
  border: 1px solid #D1D5DB;
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 13px;
  font-family: inherit;
  color: #111827;
  outline: none;
}

.input-control:focus {
  border-color: #4F46E5;
  box-shadow: 0 0 0 2px rgba(79, 70, 229, 0.12);
}

.textarea-key {
  flex: 1;
  resize: vertical;
  font-family: monospace;
  font-size: 12.5px;
}

.key-input-row {
  display: flex;
  gap: 10px;
  align-items: stretch;
}

.btn-primary {
  background: #4F46E5;
  color: #FFFFFF;
  border: none;
  border-radius: 8px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-primary:hover:not(:disabled) {
  background: #4338CA;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-save-key {
  white-space: nowrap;
  padding: 0 20px;
}

.btn-outline {
  background: #FFFFFF;
  color: #374151;
  border: 1px solid #D1D5DB;
  border-radius: 8px;
  padding: 7px 12px;
  font-size: 12.5px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-outline:hover:not(:disabled) {
  background: #F3F4F6;
}

.btn-text {
  background: none;
  border: none;
  color: #4F46E5;
  font-size: 12.5px;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
}

.project-list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
}

.project-table-wrap {
  border: 1px solid #E5E7EB;
  border-radius: 10px;
  overflow: hidden;
}

.project-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.project-table th {
  background: #F9FAFB;
  color: #4B5563;
  font-weight: 600;
  font-size: 12px;
  text-align: left;
  padding: 10px 12px;
  border-bottom: 1px solid #E5E7EB;
}

.project-table td {
  padding: 10px 12px;
  border-bottom: 1px solid #F3F4F6;
  vertical-align: middle;
}

.project-table tr:last-child td {
  border-bottom: none;
}

.row-disabled {
  background: #FAFAFA;
  opacity: 0.78;
}

.col-idx {
  color: #6B7280;
  font-weight: 600;
}

.proj-name {
  font-weight: 600;
  color: #111827;
}

.proj-num {
  font-size: 11.5px;
  color: #6B7280;
  font-family: monospace;
}

.key-code {
  background: #F3F4F6;
  padding: 3px 8px;
  border-radius: 5px;
  font-size: 12px;
  color: #1F2937;
  word-break: break-all;
}

.missing-key-text {
  font-size: 12px;
  color: #9CA3AF;
  font-style: italic;
}

.inline-edit-row {
  display: flex;
  gap: 6px;
  align-items: center;
}

.input-sm {
  padding: 5px 8px;
  font-size: 12px;
  flex: 1;
  font-family: monospace;
}

.col-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.btn-xs {
  padding: 4px 9px;
  font-size: 11.5px;
  border-radius: 6px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
}

.btn-ghost {
  background: transparent;
  border: 1px solid #D1D5DB;
  color: #4B5563;
}

.btn-danger-ghost {
  background: transparent;
  border: 1px solid #FECACA;
  color: #DC2626;
}

.btn-danger-ghost:hover {
  background: #FEF2F2;
}

.status-pill {
  display: inline-block;
  font-size: 11.5px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 999px;
}

.pill-active {
  background: #ECFDF5;
  color: #059669;
}

.pill-warn {
  background: #FFFBEB;
  color: #D97706;
}

.pill-danger {
  background: #FEF2F2;
  color: #DC2626;
}

.pill-missing {
  background: #EFF6FF;
  color: #2563EB;
}

.pill-muted {
  background: #F3F4F6;
  color: #6B7280;
}

.empty-row {
  text-align: center;
  color: #6B7280;
  padding: 20px;
}

/* Setting Rows */
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 13px 0;
  border-bottom: 1px solid #F3F4F6;
}

.setting-row:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.setting-name {
  font-size: 13.5px;
  font-weight: 600;
  color: #1F2937;
}

.setting-desc {
  font-size: 12px;
  color: #6B7280;
}

.select-box {
  background: #F9FAFB;
  border: 1px solid #D1D5DB;
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 12.5px;
  font-family: inherit;
  color: #374151;
}

/* Switch */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
}

.toggle-switch.toggle-sm {
  width: 36px;
  height: 20px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #E5E7EB;
  transition: .25s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .25s;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0,0,0,0.15);
}

.toggle-sm .slider:before {
  height: 14px;
  width: 14px;
  left: 3px;
  bottom: 3px;
}

input:checked + .slider {
  background-color: #4F46E5;
}

input:checked + .slider:before {
  transform: translateX(20px);
}

.toggle-sm input:checked + .slider:before {
  transform: translateX(16px);
}
</style>
