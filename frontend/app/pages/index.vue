<template>
  <div class="backup-view">
    <!-- Header -->
    <header class="view-header">
      <div class="header-text">
        <h1 class="view-title">สำรองรูปภาพ</h1>
        <p class="view-subtitle">ย้ายรูปจาก Flash Drive / โทรศัพท์ ไปยังเครื่อง พร้อมให้ AI วิเคราะห์คำอธิบายและ Tag</p>
      </div>

      <!-- Missing files badge in header if any -->
      <div v-if="integrityResult && integrityResult.missing_in_disk && integrityResult.missing_in_disk.length > 0" 
           class="header-alert-pill"
           @click="showMissingModal = true">
        <span class="pill-icon">⚠️</span>
        <span>ไฟล์สูญหาย {{ integrityResult.missing_in_disk.length }} ไฟล์</span>
      </div>
    </header>

    <!-- Source and Destination Selection Row -->
    <div class="folder-selector-row">
      <!-- Source Box -->
      <div class="folder-card source-card">
        <div class="card-top">
          <span class="card-label">ต้นทาง (Source)</span>
          <div class="source-type-pills">
            <button 
              class="type-pill" 
              :class="{ active: sourceType === 'usb' }"
              @click="sourceType = 'usb'"
            >
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="7" y="2" width="10" height="20" rx="2"/>
                <line x1="10" y1="6" x2="14" y2="6"/>
                <line x1="10" y1="10" x2="14" y2="10"/>
              </svg>
              Flash Drive
            </button>
            <button 
              class="type-pill" 
              :class="{ active: sourceType === 'phone' }"
              @click="sourceType = 'phone'"
            >
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="5" y="2" width="14" height="20" rx="2"/>
                <line x1="12" y1="18" x2="12.01" y2="18"/>
              </svg>
              โทรศัพท์
            </button>
          </div>
        </div>

        <div class="path-select-box">
          <div class="path-icon-container">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
            </svg>
          </div>
          <span class="folder-path" :title="sourcePath || 'ยังไม่ได้เลือกโฟลเดอร์ต้นทาง'">
            {{ sourcePath || 'กรุณาเลือกโฟลเดอร์ Flash Drive หรือ รูปภาพ' }}
          </span>
          <button class="btn-select" @click="handleSelectSource" :disabled="isProcessing">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
            </svg>
            เลือกโฟลเดอร์
          </button>
        </div>

        <div class="card-bottom-tags">
          <span class="tag-pill tag-blue">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="18" height="18" rx="2"/>
              <circle cx="8.5" cy="8.5" r="1.5"/>
              <polyline points="21 15 16 10 5 21"/>
            </svg>
            พบรูปภาพ {{ sourceFiles.length }} ไฟล์
          </span>
          <span class="tag-pill tag-gray">
            {{ formatTotalSize(sourceTotalBytes) }}
          </span>
        </div>
      </div>

      <!-- Arrow Connector -->
      <div class="arrow-container">
        <div class="arrow-circle">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#FFFFFF" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="5" y1="12" x2="19" y2="12"/>
            <polyline points="12 5 19 12 12 19"/>
          </svg>
        </div>
      </div>

      <!-- Destination Box -->
      <div class="folder-card dest-card">
        <div class="card-top">
          <span class="card-label">ปลายทาง (Destination)</span>
        </div>

        <div class="path-select-box">
          <div class="path-icon-container">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
            </svg>
          </div>
          <span class="folder-path" :title="destPath">
            {{ destPath || 'กรุณาเลือกโฟลเดอร์ปลายทาง' }}
          </span>
          <button class="btn-select" @click="handleSelectDest" :disabled="isProcessing">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
            </svg>
            เลือกโฟลเดอร์
          </button>
        </div>

        <div class="card-bottom-tags">
          <button class="tag-pill tag-green tag-btn" @click="handleOpenDestList" title="คลิกเพื่อดูรายการไฟล์จริงในโฟลเดอร์ปลายทาง (/list dest)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
            สำรองแล้ว {{ destCount }} รูป (ดูไฟล์จริง)
          </button>
          <span class="tag-pill tag-purple">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <ellipse cx="12" cy="5" rx="9" ry="3"/>
              <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
              <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
            </svg>
            SQLite พร้อมใช้งาน
          </span>
        </div>
      </div>
    </div>

    <!-- Integrity Missing Files Warning Banner (Matching Figma) -->
    <div 
      v-if="integrityResult && integrityResult.missing_in_disk && integrityResult.missing_in_disk.length > 0"
      class="warning-banner"
    >
      <div class="warning-content">
        <svg class="warning-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#D97706" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
          <line x1="12" y1="9" x2="12" y2="13"/>
          <line x1="12" y1="17" x2="12.01" y2="17"/>
        </svg>
        <span class="warning-text">
          ตรวจพบไฟล์สูญหาย: <strong>{{ displayMissingPreview }}</strong> — มีประวัติในฐานข้อมูลแต่ไม่พบไฟล์ในโฟลเดอร์ปลายทาง
        </span>
      </div>
      <button class="btn-warning-detail" @click="showMissingModal = true">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        ดูรายละเอียด
      </button>
    </div>

    <!-- Photos Grid Section -->
    <section class="photos-section">
      <div class="section-toolbar">
        <div class="toolbar-left">
          <h2 class="section-heading">รูปในต้นทาง</h2>
          <span class="selected-counter" v-if="selectedFiles.length > 0">
            เลือกแล้ว {{ selectedFiles.length }}
          </span>
        </div>

        <div class="toolbar-actions">
          <button 
            class="btn-tool btn-outline" 
            @click="toggleSelectAll"
            :disabled="sourceFiles.length === 0"
          >
            {{ allSelected ? 'ยกเลิกการเลือก' : 'เลือกทั้งหมด' }}
          </button>

          <button 
            class="btn-tool btn-primary-light" 
            @click="handleMoveSelected"
            :disabled="selectedFiles.length === 0 || isProcessing"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="17 8 12 3 7 8"/>
              <line x1="12" y1="3" x2="12" y2="15"/>
            </svg>
            ย้ายที่เลือก ({{ selectedFiles.length }})
          </button>

          <button 
            class="btn-tool btn-primary-solid" 
            @click="handleMoveAll"
            :disabled="sourceFiles.length === 0 || isProcessing"
          >
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="17 8 12 3 7 8"/>
              <line x1="12" y1="3" x2="12" y2="15"/>
            </svg>
            สำรองทั้งหมด
          </button>
        </div>
      </div>

      <!-- Photo Grid Cards -->
      <div class="photo-grid" v-if="sourceFiles.length > 0">
        <div 
          v-for="(photo, index) in sourceFiles" 
          :key="photo.filename"
          class="photo-card"
          :class="{ selected: isSelected(photo.filename) }"
          @click="toggleSelect(photo.filename)"
        >
          <!-- Thumbnail preview container with Figma artistic gradient waves -->
          <div class="thumb-wrapper" :style="getArtisticGradient(index)">
            <!-- Top Checkbox -->
            <div class="checkbox-container" @click.stop="toggleSelect(photo.filename)">
              <div class="custom-checkbox" :class="{ checked: isSelected(photo.filename) }">
                <svg v-if="isSelected(photo.filename)" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#FFFFFF" stroke-width="3">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
              </div>
            </div>

            <!-- Stylized sun / moon & wave shape matching Figma SVG -->
            <div class="thumb-art">
              <div class="art-sun" :style="getSunStyle(index)"></div>
              <svg class="art-wave" viewBox="0 0 200 80" preserveAspectRatio="none">
                <path d="M0,45 C50,20 120,65 200,35 L200,80 L0,80 Z" :fill="getWaveColor(index)"/>
              </svg>
            </div>
          </div>

          <!-- Card Info -->
          <div class="card-info">
            <span class="filename" :title="photo.filename">{{ photo.filename }}</span>
            <div class="badge-row">
              <span v-if="photo.is_duplicate" class="badge badge-duplicate">
                ⚠️ ซ้ำ · ข้าม
              </span>
              <span v-else class="badge badge-new">
                ใหม่
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#94A3B8" stroke-width="1.5">
          <rect x="3" y="3" width="18" height="18" rx="2"/>
          <circle cx="8.5" cy="8.5" r="1.5"/>
          <polyline points="21 15 16 10 5 21"/>
        </svg>
        <p class="empty-title">ไม่พบรูปภาพในโฟลเดอร์ต้นทาง</p>
        <p class="empty-subtitle">กรุณาเลือกโฟลเดอร์ต้นทาง (Flash Drive / โทรศัพท์) เพื่อเริ่มสำรองข้อมูล</p>
        <button class="btn-select empty-btn" @click="handleSelectSource">
          เลือกโฟลเดอร์ต้นทาง
        </button>
      </div>
    </section>

    <!-- Bottom Status Bar (ผลการสำรองล่าสุด) -->
    <footer class="bottom-status-bar" v-if="lastSummary.has_run">
      <div class="status-left">
        <span class="status-title">ผลการสำรองล่าสุด</span>
        <span class="status-pill pill-success">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          ย้ายสำเร็จ {{ lastSummary.moved_files }}
        </span>
        <span class="status-pill pill-warning">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
          </svg>
          ข้ามไฟล์ซ้ำ {{ lastSummary.skipped_files }}
        </span>
        <span class="status-pill pill-ai">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
          </svg>
          AI วิเคราะห์ {{ lastSummary.moved_files }}/{{ lastSummary.moved_files }}
        </span>
      </div>

      <div class="status-right">
        <div class="timer-box">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
          <span class="timer-label">เวลาที่ใช้ทั้งหมด</span>
          <span class="timer-value">{{ lastSummary.duration_formatted }}</span>
        </div>
      </div>
    </footer>
    <footer class="bottom-status-bar" v-else>
      <div class="status-left">
        <span class="status-title">สถานะระบบ</span>
        <span class="status-pill pill-neutral">
          พร้อมสำหรับการสำรองรูปภาพ
        </span>
      </div>
      <div class="status-right">
        <div class="timer-box">
          <span class="timer-label">ความเร็วระบบ</span>
          <span class="timer-value">16 Goroutines · SQLite WAL</span>
        </div>
      </div>
    </footer>

    <!-- Missing Files Detail Modal -->
    <div v-if="showMissingModal" class="modal-backdrop" @click.self="showMissingModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-title-row">
            <span class="modal-icon">⚠️</span>
            <h3>รายการไฟล์สูญหาย (Missing Files)</h3>
          </div>
          <button class="btn-close" @click="showMissingModal = false">✕</button>
        </div>
        <div class="modal-body">
          <p class="modal-desc">
            ไฟล์เหล่านี้มีประวัติบันทึกในฐานข้อมูล (สถานะ <code>missing</code>) แต่ไม่พบไฟล์จริงในโฟลเดอร์ปลายทาง
          </p>
          <ul class="missing-list">
            <li v-for="file in (integrityResult?.missing_in_disk || [])" :key="file">
              <span class="file-icon">📄</span>
              <span class="missing-filename">{{ file }}</span>
            </li>
          </ul>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showMissingModal = false">ปิด</button>
        </div>
      </div>
    </div>

    <!-- Destination Actual Files Modal (Feature 10: /list dest) -->
    <div v-if="showDestFilesModal" class="modal-backdrop" @click.self="showDestFilesModal = false">
      <div class="modal-card modal-large">
        <div class="modal-header">
          <div class="modal-title-row">
            <span class="modal-icon">📁</span>
            <div>
              <h3>รายการไฟล์จริงในโฟลเดอร์ปลายทาง (/list dest)</h3>
              <p class="modal-subtitle">{{ destPath }} (พบ {{ destFilesList.length }} ไฟล์)</p>
            </div>
          </div>
          <button class="btn-close" @click="showDestFilesModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="destFilesList.length === 0" class="empty-modal-text">
            ไม่พบไฟล์ในโฟลเดอร์ปลายทาง
          </div>
          <div v-else class="modal-table-wrap">
            <table class="modal-table">
              <thead>
                <tr>
                  <th>ชื่อไฟล์</th>
                  <th>ขนาด</th>
                  <th>เวลาแก้ไขล่าสุด</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in destFilesList" :key="item.filename">
                  <td class="font-semibold">{{ item.filename }}</td>
                  <td>{{ formatTotalSize(item.size_bytes) }}</td>
                  <td class="text-muted">{{ item.mod_time || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showDestFilesModal = false">ปิด</button>
        </div>
      </div>
    </div>

    <!-- Hidden folder input for browser environment fallback -->
    <input 
      type="file" 
      ref="folderInputRef" 
      webkitdirectory 
      directory 
      multiple 
      style="display: none" 
      @change="handleBrowserFolderUpload" 
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface PhotoItem {
  filename: string
  size_bytes: number
  is_duplicate: boolean
}

// Props & Emits
const props = defineProps<{
  sourcePathProp?: string
  destPathProp?: string
}>()

const emit = defineEmits<{
  (e: 'update:dest', path: string): void
  (e: 'update:source', path: string): void
}>()

// State - 100% Real Data
const folderInputRef = ref<HTMLInputElement | null>(null)
const sourceType = ref<'usb' | 'phone'>('usb')
const sourcePath = ref(props.sourcePathProp || '')
const destPath = ref(props.destPathProp || '')
const destCount = ref(0)
const isProcessing = ref(false)
const showMissingModal = ref(false)
const showDestFilesModal = ref(false)
const destFilesList = ref<any[]>([])
const selectedFiles = ref<string[]>([])

// Integrity Result (starts null - loaded live by backend CheckIntegrity)
const integrityResult = ref<any>(null)

// Last Move Summary (starts has_run: false - populated when backup completes)
const lastSummary = ref<{
  total_files: number
  moved_files: number
  skipped_files: number
  failed_files: number
  duration_ms: number
  duration_formatted: string
  has_run: boolean
}>({
  total_files: 0,
  moved_files: 0,
  skipped_files: 0,
  failed_files: 0,
  duration_ms: 0,
  duration_formatted: '-',
  has_run: false
})

// Photo items in source (starts empty [] - populated with real files from device)
const sourceFiles = ref<PhotoItem[]>([])

const sourceTotalBytes = computed(() => {
  return sourceFiles.value.reduce((acc, f) => acc + f.size_bytes, 0)
})

const displayMissingPreview = computed(() => {
  if (!integrityResult.value || !integrityResult.value.missing_in_disk) return ''
  return integrityResult.value.missing_in_disk.join(', ')
})

const allSelected = computed(() => {
  return sourceFiles.value.length > 0 && selectedFiles.value.length === sourceFiles.value.length
})

// Selection helpers
function isSelected(filename: string): boolean {
  return selectedFiles.value.includes(filename)
}

function toggleSelect(filename: string) {
  const index = selectedFiles.value.indexOf(filename)
  if (index > -1) {
    selectedFiles.value.splice(index, 1)
  } else {
    selectedFiles.value.push(filename)
  }
}

function toggleSelectAll() {
  if (allSelected.value) {
    selectedFiles.value = []
  } else {
    selectedFiles.value = sourceFiles.value.map(f => f.filename)
  }
}

function formatTotalSize(bytes: number): string {
  if (bytes > 1024 * 1024 * 1024) {
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
  }
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

// Art gradients and curves matching Figma's design
const artThemes = [
  { bg: 'linear-gradient(180deg, #93C5FD 0%, #DBEAFE 100%)', sun: '#FFFFFF', wave: '#64748B' },
  { bg: 'linear-gradient(180deg, #818CF8 0%, #C7D2FE 100%)', sun: '#FDE047', wave: '#4338CA' },
  { bg: 'linear-gradient(180deg, #FB923C 0%, #FED7AA 100%)', sun: '#FEF08A', wave: '#9A3412' },
  { bg: 'linear-gradient(180deg, #818CF8 0%, #E0E7FF 100%)', sun: '#FEF08A', wave: '#3730A3' },
  { bg: 'linear-gradient(180deg, #A78BFA 0%, #EDE9FE 100%)', sun: '#FEF08A', wave: '#5B21B6' },
  { bg: 'linear-gradient(180deg, #38BDF8 0%, #BAE6FD 100%)', sun: '#FFFFFF', wave: '#0284C7' },
  { bg: 'linear-gradient(180deg, #FB923C 0%, #FFEDD5 100%)', sun: '#FEF08A', wave: '#C2410C' },
  { bg: 'linear-gradient(180deg, #38BDF8 0%, #E0F2FE 100%)', sun: '#FFFFFF', wave: '#0284C7' },
  { bg: 'linear-gradient(180deg, #FBBF24 0%, #FEF3C7 100%)', sun: '#FFFFFF', wave: '#D97706' },
  { bg: 'linear-gradient(180deg, #334155 0%, #475569 100%)', sun: '#E2E8F0', wave: '#0F172A' },
  { bg: 'linear-gradient(180deg, #38BDF8 0%, #E0F2FE 100%)', sun: '#FFFFFF', wave: '#0369A1' },
  { bg: 'linear-gradient(180deg, #818CF8 0%, #DDD6FE 100%)', sun: '#FEF08A', wave: '#4C1D95' }
]

function getArtisticGradient(index: number) {
  const theme = artThemes[index % artThemes.length]
  return { background: theme.bg }
}

function getSunStyle(index: number) {
  const theme = artThemes[index % artThemes.length]
  return { background: theme.sun }
}

function getWaveColor(index: number) {
  const theme = artThemes[index % artThemes.length]
  return theme.wave
}

// Wails Backend Interactions & Real Folder Picker
async function handleSelectSource() {
  try {
    if (typeof (window as any)?.go?.main?.App?.SelectSourceFolder === 'function') {
      const selected = await (window as any).go.main.App.SelectSourceFolder()
      if (selected) {
        sourcePath.value = selected
        emit('update:source', selected)
        await refreshSourceFiles()
      }
    } else {
      // In browser preview, trigger native folder chooser
      folderInputRef.value?.click()
    }
  } catch (err) {
    console.error('Failed to select source folder:', err)
  }
}

function handleBrowserFolderUpload(event: Event) {
  const target = event.target as HTMLInputElement
  if (!target.files || target.files.length === 0) return
  const files = Array.from(target.files)
  
  const firstPath = files[0].webkitRelativePath || files[0].name
  const folderName = firstPath.includes('/') ? firstPath.split('/')[0] : 'โฟลเดอร์ที่เลือก'
  sourcePath.value = folderName
  emit('update:source', folderName)

  const validExts = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp', '.tif', '.heic', '.raw', '.svg']
  const images = files.filter(f => {
    const name = f.name.toLowerCase()
    return validExts.some(ext => name.endsWith(ext))
  })

  const destLower = new Set(destFilesList.value.map(f => f.filename.toLowerCase()))

  sourceFiles.value = images.map(f => ({
    filename: f.name,
    size_bytes: f.size,
    is_duplicate: destLower.has(f.name.toLowerCase())
  }))
  selectedFiles.value = sourceFiles.value.filter(f => !f.is_duplicate).map(f => f.filename)
}

async function handleSelectDest() {
  try {
    if (typeof (window as any)?.go?.main?.App?.SelectDestFolder === 'function') {
      const res = await (window as any).go.main.App.SelectDestFolder()
      if (res) {
        const dest = await (window as any).go.main.App.GetDestPath()
        destPath.value = dest
        emit('update:dest', dest)
        integrityResult.value = res
        destCount.value = res.total_disk || 0
      }
    } else {
      const chosen = prompt('กำหนดโฟลเดอร์ปลายทางสำหรับจัดเก็บรูปภาพ (เช่น D:/PhotoBackup):', destPath.value || 'D:/PhotoBackup')
      if (chosen) {
        destPath.value = chosen
        emit('update:dest', chosen)
        destCount.value = destFilesList.value.length
      }
    }
  } catch (err) {
    console.error('Failed to select dest folder:', err)
  }
}

async function handleOpenDestList() {
  if (!destPath.value) {
    alert('กรุณาเลือกโฟลเดอร์ปลายทางก่อน!')
    return
  }
  try {
    if (typeof (window as any)?.go?.main?.App?.ListDestFileItems === 'function') {
      const items = await (window as any).go.main.App.ListDestFileItems(destPath.value)
      destFilesList.value = items || []
    }
  } catch (err) {
    console.error('Failed to list dest files:', err)
  }
  showDestFilesModal.value = true
}

async function refreshSourceFiles() {
  try {
    if (typeof (window as any)?.go?.main?.App?.ListSourceFileItems === 'function') {
      const items = await (window as any).go.main.App.ListSourceFileItems()
      if (items && Array.isArray(items)) {
        // Filter image files only
        const imageItems = items.filter((item: any) => item.is_image)

        // Feature 7: Detect duplicates with destination disk files
        let destFiles: string[] = []
        if (destPath.value && typeof (window as any)?.go?.main?.App?.ListDestPhotos === 'function') {
          try {
            destFiles = await (window as any).go.main.App.ListDestPhotos(destPath.value) || []
          } catch (_) {}
        }
        const destLower = new Set((destFiles || []).map((f: string) => f.toLowerCase()))

        sourceFiles.value = imageItems.map((item: any) => ({
          filename: item.filename,
          size_bytes: item.size_bytes,
          is_duplicate: destLower.has(item.filename.toLowerCase())
        }))
        selectedFiles.value = sourceFiles.value.filter(f => !f.is_duplicate).map(f => f.filename)
      } else {
        sourceFiles.value = []
        selectedFiles.value = []
      }
    }
  } catch (err) {
    console.error('Failed to list source items:', err)
    sourceFiles.value = []
    selectedFiles.value = []
  }
}

async function handleMoveSelected() {
  // Feature 11: State Guard (Validate source & dest before processing)
  if (!sourcePath.value) {
    alert('กรุณาเลือกโฟลเดอร์ต้นทาง (Flash Drive / โทรศัพท์) ก่อนทำการย้ายไฟล์!')
    return
  }
  if (!destPath.value) {
    alert('กรุณาเลือกโฟลเดอร์ปลายทางสำหรับเก็บภาพก่อนทำการย้ายไฟล์!')
    return
  }
  if (selectedFiles.value.length === 0) {
    alert('กรุณาเลือกรูปภาพอย่างน้อย 1 ไฟล์ที่ต้องการย้าย!')
    return
  }
  isProcessing.value = true
  try {
    if (typeof (window as any)?.go?.main?.App?.MovePhotos === 'function') {
      const summary = await (window as any).go.main.App.MovePhotos(selectedFiles.value, false)
      lastSummary.value = {
        total_files: summary.total_files || selectedFiles.value.length,
        moved_files: summary.moved_files || 0,
        skipped_files: summary.skipped_files || 0,
        failed_files: summary.failed_files || 0,
        duration_ms: summary.duration_ms || 0,
        duration_formatted: summary.duration_formatted || `${summary.duration_ms} ms`,
        has_run: true
      }
      await refreshSourceFiles()
      if (typeof (window as any)?.go?.main?.App?.CheckIntegrity === 'function') {
        integrityResult.value = await (window as any).go.main.App.CheckIntegrity(destPath.value)
        destCount.value = integrityResult.value?.total_disk || 0
      }
      alert(`สำรองรูปภาพที่เลือกสำเร็จ ${summary.moved_files} ไฟล์ (ข้ามไฟล์ซ้ำ ${summary.skipped_files} ไฟล์) ใช้เวลา ${summary.duration_formatted}`)
    } else {
      // Browser preview mode with real selected files
      const count = selectedFiles.value.length
      destCount.value += count
      lastSummary.value = {
        total_files: count,
        moved_files: count,
        skipped_files: 0,
        failed_files: 0,
        duration_ms: 18.5,
        duration_formatted: '18.50 ms',
        has_run: true
      }
      sourceFiles.value = sourceFiles.value.filter(f => !selectedFiles.value.includes(f.filename))
      selectedFiles.value = []
      alert(`จำลองการย้ายไฟล์สำเร็จ ${count} รูป (เวลา ${lastSummary.value.duration_formatted})`)
    }
  } catch (err: any) {
    alert('เกิดข้อผิดพลาดในการย้าย: ' + (err?.message || err))
  } finally {
    isProcessing.value = false
  }
}

async function handleMoveAll() {
  // Feature 11: State Guard (Validate source & dest before processing)
  if (!sourcePath.value) {
    alert('กรุณาเลือกโฟลเดอร์ต้นทาง (Flash Drive / โทรศัพท์) ก่อนทำการสำรองข้อมูล!')
    return
  }
  if (!destPath.value) {
    alert('กรุณาเลือกโฟลเดอร์ปลายทางสำหรับเก็บภาพก่อนทำการสำรองข้อมูล!')
    return
  }
  if (sourceFiles.value.length === 0) {
    alert('ไม่พบรูปภาพในโฟลเดอร์ต้นทางที่จะสำรอง!')
    return
  }
  isProcessing.value = true
  try {
    if (typeof (window as any)?.go?.main?.App?.MovePhotos === 'function') {
      const summary = await (window as any).go.main.App.MovePhotos([], true)
      lastSummary.value = {
        total_files: summary.total_files || sourceFiles.value.length,
        moved_files: summary.moved_files || 0,
        skipped_files: summary.skipped_files || 0,
        failed_files: summary.failed_files || 0,
        duration_ms: summary.duration_ms || 0,
        duration_formatted: summary.duration_formatted || `${summary.duration_ms} ms`,
        has_run: true
      }
      await refreshSourceFiles()
      if (typeof (window as any)?.go?.main?.App?.CheckIntegrity === 'function') {
        integrityResult.value = await (window as any).go.main.App.CheckIntegrity(destPath.value)
        destCount.value = integrityResult.value?.total_disk || 0
      }
      alert(`สำรองรูปภาพทั้งหมดสำเร็จ ${summary.moved_files} ไฟล์ (ข้ามไฟล์ซ้ำ ${summary.skipped_files} ไฟล์) ใช้เวลา ${summary.duration_formatted}`)
    } else {
      // Browser preview mode with real files
      const count = sourceFiles.value.filter(f => !f.is_duplicate).length
      const skipped = sourceFiles.value.filter(f => f.is_duplicate).length
      destCount.value += count
      lastSummary.value = {
        total_files: sourceFiles.value.length,
        moved_files: count,
        skipped_files: skipped,
        failed_files: 0,
        duration_ms: 32.4,
        duration_formatted: '32.40 ms',
        has_run: true
      }
      sourceFiles.value = sourceFiles.value.filter(f => f.is_duplicate)
      selectedFiles.value = []
      alert(`จำลองการสำรองทั้งหมดสำเร็จ ${count} รูป (ข้ามไฟล์ซ้ำ ${skipped} รูป) เวลา ${lastSummary.value.duration_formatted}`)
    }
  } catch (err: any) {
    alert('เกิดข้อผิดพลาดในการย้าย: ' + (err?.message || err))
  } finally {
    isProcessing.value = false
  }
}

onMounted(async () => {
  // If running inside Wails, read live paths and real data from Go backend
  if (typeof (window as any)?.go?.main?.App?.GetSourcePath === 'function') {
    const src = await (window as any).go.main.App.GetSourcePath()
    if (src) {
      sourcePath.value = src
      await refreshSourceFiles()
    }
  }
  if (typeof (window as any)?.go?.main?.App?.GetDestPath === 'function') {
    const dst = await (window as any).go.main.App.GetDestPath()
    if (dst) {
      destPath.value = dst
      integrityResult.value = await (window as any).go.main.App.CheckIntegrity(dst)
      destCount.value = integrityResult.value?.total_disk || 0
    }
  }
})
</script>

<style scoped>
.backup-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 24px 28px 16px 28px;
  box-sizing: border-box;
  background-color: #F4F5FA;
  overflow-y: auto;
  user-select: none;
  font-family: 'Prompt', 'Inter', sans-serif;
}

/* Header */
.view-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}

.view-title {
  font-size: 24px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px 0;
  letter-spacing: -0.3px;
}

.view-subtitle {
  font-size: 13px;
  color: #6B7280;
  margin: 0;
}

.header-alert-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #FEF3C7;
  color: #B45309;
  border: 1px solid #FCD34D;
  font-size: 12px;
  font-weight: 600;
  padding: 6px 14px;
  border-radius: 9999px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.header-alert-pill:hover {
  background: #FDE68A;
}

/* Folder Selection Row */
.folder-selector-row {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 14px;
}

.folder-card {
  flex: 1;
  background: #FFFFFF;
  border-radius: 14px;
  padding: 16px 18px;
  border: 1px solid #E5E7EB;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-label {
  font-size: 12px;
  font-weight: 600;
  color: #4B5563;
}

.source-type-pills {
  display: flex;
  background: #F3F4F6;
  border-radius: 8px;
  padding: 2px;
  gap: 2px;
}

.type-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: #4B5563;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
}

.type-pill.active {
  background: #FFFFFF;
  color: #4F46E5;
  font-weight: 600;
  box-shadow: 0 1px 2px rgba(0,0,0,0.06);
}

.path-select-box {
  display: flex;
  align-items: center;
  gap: 10px;
}

.path-icon-container {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #EEF2FF;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.folder-path {
  font-size: 13px;
  font-weight: 500;
  color: #1F2937;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  font-family: 'Inter', monospace;
}

.btn-select {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: 8px;
  background: #FFFFFF;
  border: 1px solid #D1D5DB;
  color: #374151;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}

.btn-select:hover {
  background: #F9FAFB;
  border-color: #9CA3AF;
}

.card-bottom-tags {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: 9999px;
}

.tag-blue {
  background: #EEF2FF;
  color: #4F46E5;
}

.tag-gray {
  background: #F3F4F6;
  color: #6B7280;
}

.tag-green {
  background: #ECFDF5;
  color: #059669;
}

.tag-purple {
  background: #F5F3FF;
  color: #7C3AED;
}

.arrow-container {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.arrow-circle {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: #4F46E5;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 10px rgba(79, 70, 229, 0.25);
}

/* Warning Banner */
.warning-banner {
  background: #FEF3C7;
  border: 1px solid #FCD34D;
  border-radius: 12px;
  padding: 12px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
}

.warning-content {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12.5px;
  color: #92400E;
}

.warning-icon {
  flex-shrink: 0;
}

.btn-warning-detail {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #FFFFFF;
  border: 1px solid #F59E0B;
  color: #92400E;
  font-size: 12px;
  font-weight: 600;
  padding: 6px 14px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-warning-detail:hover {
  background: #FFFBEB;
}

/* Photos Section */
.photos-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  margin-bottom: 16px;
}

.section-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.section-heading {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
  margin: 0;
}

.selected-counter {
  font-size: 12px;
  font-weight: 500;
  color: #4F46E5;
  background: #EEF2FF;
  padding: 2px 8px;
  border-radius: 9999px;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-tool {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  padding: 8px 14px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-outline {
  background: #FFFFFF;
  border: 1px solid #D1D5DB;
  color: #374151;
}

.btn-outline:hover:not(:disabled) {
  background: #F9FAFB;
}

.btn-primary-light {
  background: #EEF2FF;
  border: 1px solid #C7D2FE;
  color: #4F46E5;
}

.btn-primary-light:hover:not(:disabled) {
  background: #E0E7FF;
}

.btn-primary-solid {
  background: #4F46E5;
  border: 1px solid #4F46E5;
  color: #FFFFFF;
  box-shadow: 0 2px 6px rgba(79, 70, 229, 0.25);
}

.btn-primary-solid:hover:not(:disabled) {
  background: #4338CA;
}

.btn-tool:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Photo Grid */
.photo-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 12px;
}

.photo-card {
  background: #FFFFFF;
  border-radius: 12px;
  border: 1px solid #E5E7EB;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.15s ease;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}

.photo-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.06);
}

.photo-card.selected {
  border-color: #4F46E5;
  box-shadow: 0 0 0 2px rgba(79, 70, 229, 0.25);
}

.thumb-wrapper {
  position: relative;
  height: 90px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.checkbox-container {
  position: absolute;
  top: 8px;
  left: 8px;
  z-index: 2;
}

.custom-checkbox {
  width: 18px;
  height: 18px;
  border-radius: 5px;
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(0,0,0,0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.12s;
}

.custom-checkbox.checked {
  background: #4F46E5;
  border-color: #4F46E5;
}

.thumb-art {
  width: 100%;
  height: 100%;
  position: relative;
}

.art-sun {
  width: 22px;
  height: 14px;
  border-radius: 50%;
  position: absolute;
  top: 14px;
  right: 28px;
  opacity: 0.9;
}

.art-wave {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 48px;
}

.card-info {
  padding: 8px 10px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.filename {
  font-size: 11.5px;
  font-weight: 600;
  color: #1F2937;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.badge-row {
  display: flex;
  align-items: center;
}

.badge {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 7px;
  border-radius: 4px;
}

.badge-new {
  background: #ECFDF5;
  color: #059669;
}

.badge-duplicate {
  background: #FEF3C7;
  color: #B45309;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  background: #FFFFFF;
  border-radius: 14px;
  border: 2px dashed #E5E7EB;
  text-align: center;
  gap: 8px;
}

.empty-title {
  font-size: 15px;
  font-weight: 600;
  color: #374151;
  margin: 4px 0 0 0;
}

.empty-subtitle {
  font-size: 12.5px;
  color: #6B7280;
  margin: 0;
}

.empty-btn {
  margin-top: 10px;
  background: #4F46E5;
  color: #FFFFFF;
  border: none;
}

/* Bottom Status Bar */
.bottom-status-bar {
  background: #FFFFFF;
  border-radius: 14px;
  padding: 12px 18px;
  border: 1px solid #E5E7EB;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
  flex-shrink: 0;
}

.status-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-title {
  font-size: 13px;
  font-weight: 700;
  color: #1F2937;
  margin-right: 4px;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11.5px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 9999px;
}

.pill-success {
  background: #ECFDF5;
  color: #059669;
}

.pill-warning {
  background: #FEF3C7;
  color: #B45309;
}

.pill-ai {
  background: #EEF2FF;
  color: #4F46E5;
}

.pill-neutral {
  background: #F3F4F6;
  color: #4B5563;
}

.status-right {
  display: flex;
  align-items: center;
}

.timer-box {
  display: flex;
  align-items: center;
  gap: 6px;
}

.timer-label {
  font-size: 12px;
  color: #6B7280;
}

.timer-value {
  font-size: 13px;
  font-weight: 700;
  color: #4F46E5;
  font-family: 'Inter', monospace;
}

/* Modal */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-card {
  background: #FFFFFF;
  border-radius: 16px;
  width: 480px;
  max-width: 90vw;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.modal-header {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #E5E7EB;
}

.modal-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-title-row h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: #111827;
}

.btn-close {
  background: transparent;
  border: none;
  font-size: 16px;
  color: #6B7280;
  cursor: pointer;
}

.modal-body {
  padding: 18px 20px;
}

.modal-desc {
  font-size: 13px;
  color: #4B5563;
  margin: 0 0 14px 0;
}

.modal-desc code {
  background: #F3F4F6;
  padding: 2px 6px;
  border-radius: 4px;
  color: #DC2626;
}

.missing-list {
  list-style: none;
  padding: 0;
  margin: 0;
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
}

.missing-list li {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid #F3F4F6;
  font-size: 13px;
}

.missing-list li:last-child {
  border-bottom: none;
}

.missing-filename {
  font-weight: 500;
  color: #DC2626;
  font-family: monospace;
}

.modal-footer {
  padding: 12px 20px;
  background: #F9FAFB;
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid #E5E7EB;
}

.btn-secondary {
  padding: 6px 16px;
  border-radius: 8px;
  border: 1px solid #D1D5DB;
  background: #FFFFFF;
  color: #374151;
  font-size: 13px;
  cursor: pointer;
}

.tag-btn {
  border: none;
  cursor: pointer;
  transition: transform 0.1s, opacity 0.15s;
}

.tag-btn:hover {
  opacity: 0.88;
  transform: translateY(-1px);
}

.modal-large {
  max-width: 620px;
  width: 90%;
}

.modal-subtitle {
  font-size: 12px;
  color: #6B7280;
  margin: 2px 0 0 0;
  font-family: monospace;
}

.modal-table-wrap {
  max-height: 320px;
  overflow-y: auto;
  border: 1px solid #E5E7EB;
  border-radius: 8px;
}

.modal-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 13px;
}

.modal-table th {
  background: #F9FAFB;
  padding: 10px 14px;
  font-weight: 600;
  color: #4B5563;
  border-bottom: 1px solid #E5E7EB;
  position: sticky;
  top: 0;
}

.modal-table td {
  padding: 10px 14px;
  border-bottom: 1px solid #F3F4F6;
  color: #374151;
}

.empty-modal-text {
  text-align: center;
  padding: 24px;
  color: #9CA3AF;
  font-size: 13px;
}
</style>
