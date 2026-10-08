<template>
  <div class="history-view">
    <!-- UI Mockup Notice Banner -->
    <div class="mock-notice">
      <div class="notice-badge">UI Mockup</div>
      <span class="notice-text">
        [UI Mockup] ส่วนของเพื่อน (Part E): หน้าประวัติการสำรอง & ตรวจสอบความถูกต้องของไฟล์ในโฟลเดอร์ปลายทาง
      </span>
    </div>

    <!-- Header -->
    <header class="view-header">
      <div class="header-text">
        <h1 class="view-title">ประวัติ & ตรวจสอบ</h1>
        <p class="view-subtitle">ประวัติการสำรองจากฐานข้อมูลของปลายทางปัจจุบัน</p>
      </div>

      <div class="header-actions">
        <button class="btn-change-dest" @click="handleChangeDest">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
          </svg>
          เปลี่ยนปลายทาง
        </button>

        <button class="btn-check-integrity" :class="{ 'is-checking': isChecking }" @click="handleCheckIntegrity">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="{ 'spin-anim': isChecking }">
            <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
          </svg>
          <span>{{ isChecking ? 'กำลังตรวจสอบ...' : 'ตรวจสอบความถูกต้อง' }}</span>
        </button>
      </div>
    </header>

    <!-- Destination Path Banner -->
    <div class="dest-banner">
      <div class="dest-info">
        <svg class="folder-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
        </svg>
        <span class="dest-label">ปลายทาง:</span>
        <span class="dest-path">{{ currentDest }}</span>
      </div>

      <div class="check-time-pill">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/>
          <polyline points="12 6 12 12 16 14"/>
        </svg>
        <span>ตรวจล่าสุดใน {{ lastCheckDuration }}</span>
      </div>
    </div>

    <!-- Stats Summary Cards (4 Cards) -->
    <div class="stats-grid">
      <!-- Card 1: ทั้งหมดในฐานข้อมูล -->
      <div class="stat-card card-total">
        <span class="stat-label">ทั้งหมดในฐานข้อมูล</span>
        <span class="stat-number">{{ stats.total }}</span>
      </div>

      <!-- Card 2: ตรวจสอบผ่าน -->
      <div class="stat-card card-passed">
        <span class="stat-label text-emerald">ตรวจสอบผ่าน</span>
        <span class="stat-number text-emerald">{{ stats.passed }}</span>
      </div>

      <!-- Card 3: ไฟล์สูญหาย -->
      <div class="stat-card card-missing">
        <span class="stat-label text-rose">ไฟล์สูญหาย</span>
        <span class="stat-number text-rose">{{ stats.missing }}</span>
      </div>

      <!-- Card 4: ลบแล้ว -->
      <div class="stat-card card-deleted">
        <span class="stat-label">ลบแล้ว</span>
        <span class="stat-number text-gray">{{ stats.deleted }}</span>
      </div>
    </div>

    <!-- Table Card -->
    <div class="table-container">
      <table class="history-table">
        <thead>
          <tr>
            <th class="col-filename">ชื่อไฟล์</th>
            <th class="col-size">ขนาด</th>
            <th class="col-date">วันที่สำรอง</th>
            <th class="col-ai">AI</th>
            <th class="col-status">สถานะ</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="row in rows" 
            :key="row.id" 
            :class="{ 'row-missing': row.status === 'missing' }"
          >
            <!-- Filename with colorful preview icon -->
            <td class="col-filename">
              <div class="file-item">
                <div class="thumb-box" :style="{ background: row.gradient }">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                    <circle cx="8.5" cy="8.5" r="1.5"/>
                    <polyline points="21 15 16 10 5 21"/>
                  </svg>
                </div>
                <span class="filename-text">{{ row.name }}</span>
              </div>
            </td>

            <!-- File Size -->
            <td class="col-size">{{ row.size }}</td>

            <!-- Backup Date -->
            <td class="col-date">{{ row.date }}</td>

            <!-- AI Status -->
            <td class="col-ai">
              <span class="ai-pill">
                <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z"/>
                </svg>
                {{ row.ai }}
              </span>
            </td>

            <!-- Integrity Status -->
            <td class="col-status">
              <!-- ผ่าน -->
              <span v-if="row.status === 'passed'" class="status-pill status-passed">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
                ตรวจสอบผ่าน
              </span>

              <!-- สูญหาย -->
              <span v-else-if="row.status === 'missing'" class="status-pill status-missing">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                  <line x1="12" y1="9" x2="12" y2="13"/>
                  <line x1="12" y1="17" x2="12.01" y2="17"/>
                </svg>
                ไฟล์สูญหาย
              </span>

              <!-- ลบแล้ว -->
              <span v-else-if="row.status === 'deleted'" class="status-pill status-deleted">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="3 6 5 6 21 6"/>
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                </svg>
                ลบแล้ว
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const currentDest = ref('~/Pictures/PhotoBackup')
const lastCheckDuration = ref('12.4 ms')
const isChecking = ref(false)

const stats = ref({
  total: 342,
  passed: 338,
  missing: 2,
  deleted: 2
})

interface PhotoHistoryItem {
  id: number
  name: string
  size: string
  date: string
  ai: string
  status: 'passed' | 'missing' | 'deleted'
  gradient: string
}

const rows = ref<PhotoHistoryItem[]>([
  {
    id: 1,
    name: 'beach_trip.png',
    size: '3.4 MB',
    date: '3 ต.ค. 07:21',
    ai: 'วิเคราะห์แล้ว',
    status: 'passed',
    gradient: 'linear-gradient(135deg, #38BDF8, #0284C7)'
  },
  {
    id: 2,
    name: 'IMG_2041.jpg',
    size: '2.8 MB',
    date: '3 ต.ค. 07:21',
    ai: 'วิเคราะห์แล้ว',
    status: 'passed',
    gradient: 'linear-gradient(135deg, #60A5FA, #2563EB)'
  },
  {
    id: 3,
    name: 'IMG_2043.heic',
    size: '1.9 MB',
    date: '3 ต.ค. 07:21',
    ai: 'วิเคราะห์แล้ว',
    status: 'passed',
    gradient: 'linear-gradient(135deg, #34D399, #059669)'
  },
  {
    id: 4,
    name: 'IMG_0412.jpg',
    size: '2.2 MB',
    date: '1 ต.ค. 18:05',
    ai: 'วิเคราะห์แล้ว',
    status: 'missing',
    gradient: 'linear-gradient(135deg, #FB923C, #EA580C)'
  },
  {
    id: 5,
    name: 'project_cover.png',
    size: '4.1 MB',
    date: '1 ต.ค. 18:05',
    ai: 'วิเคราะห์แล้ว',
    status: 'passed',
    gradient: 'linear-gradient(135deg, #FBBF24, #D97706)'
  },
  {
    id: 6,
    name: 'IMG_0988.heic',
    size: '2.0 MB',
    date: '30 ก.ย. 21:40',
    ai: 'วิเคราะห์แล้ว',
    status: 'missing',
    gradient: 'linear-gradient(135deg, #4ADE80, #16A34A)'
  },
  {
    id: 7,
    name: 'old_family.jpg',
    size: '1.2 MB',
    date: '28 ก.ย. 10:12',
    ai: 'วิเคราะห์แล้ว',
    status: 'deleted',
    gradient: 'linear-gradient(135deg, #A855F7, #7C3AED)'
  }
])

function handleCheckIntegrity() {
  if (isChecking.value) return
  isChecking.value = true
  setTimeout(() => {
    isChecking.value = false
    lastCheckDuration.value = '9.8 ms'
  }, 600)
}

function handleChangeDest() {
  // Mockup destination change
  if (currentDest.value === '~/Pictures/PhotoBackup') {
    currentDest.value = '~/Documents/Vault2026'
  } else {
    currentDest.value = '~/Pictures/PhotoBackup'
  }
}
</script>

<style scoped>
.history-view {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  padding: 24px 32px 32px 32px;
  box-sizing: border-box;
  background-color: #F4F5FA;
  overflow-y: auto;
  user-select: none;
  font-family: 'Prompt', 'Inter', sans-serif;
}

/* UI Mockup Notice Banner */
.mock-notice {
  background: #EEF2FF;
  border: 1px dashed #6366F1;
  border-radius: 10px;
  padding: 8px 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 18px;
}

.notice-badge {
  background: #4F46E5;
  color: #FFFFFF;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 6px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.notice-text {
  font-size: 12.5px;
  color: #3730A3;
  font-weight: 500;
}

/* Header */
.view-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
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

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.btn-change-dest {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  color: #111827;
  font-size: 13px;
  font-weight: 600;
  padding: 9px 16px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s ease;
  box-shadow: 0 1px 2px rgba(0,0,0,0.03);
}

.btn-change-dest:hover {
  background: #F9FAFB;
  border-color: #D1D5DB;
}

.btn-check-integrity {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #4F46E5;
  border: 1px solid #4F46E5;
  color: #FFFFFF;
  font-size: 13px;
  font-weight: 600;
  padding: 9px 18px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s ease;
  box-shadow: 0 1px 3px rgba(79, 70, 229, 0.25);
}

.btn-check-integrity:hover {
  background: #4338CA;
  border-color: #4338CA;
}

.btn-check-integrity.is-checking {
  opacity: 0.85;
  cursor: wait;
}

.spin-anim {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Destination Banner */
.dest-banner {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 12px;
  padding: 12px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.02);
}

.dest-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13.5px;
}

.folder-icon {
  flex-shrink: 0;
}

.dest-label {
  color: #6B7280;
}

.dest-path {
  color: #111827;
  font-weight: 700;
}

.check-time-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 9999px;
}

/* 4 Summary Cards Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin-bottom: 20px;
}

.stat-card {
  border-radius: 12px;
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.02);
}

.stat-label {
  font-size: 12.5px;
  font-weight: 500;
  color: #6B7280;
}

.stat-number {
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  line-height: 1.1;
}

.card-total {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
}

.card-passed {
  background: #ECFDF5;
  border: 1px solid #A7F3D0;
}

.card-missing {
  background: #FEE2E2;
  border: 1px solid #FECACA;
}

.card-deleted {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
}

.text-emerald { color: #059669 !important; }
.text-rose { color: #DC2626 !important; }
.text-gray { color: #374151 !important; }

/* Table Container */
.table-container {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}

.history-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.history-table th {
  padding: 14px 20px;
  font-size: 12.5px;
  font-weight: 600;
  color: #6B7280;
  border-bottom: 1px solid #E5E7EB;
  background: #FAFAFA;
}

.history-table td {
  padding: 12px 20px;
  font-size: 13px;
  color: #374151;
  border-bottom: 1px solid #F3F4F6;
  vertical-align: middle;
}

.history-table tbody tr {
  transition: background-color 0.1s ease;
}

.history-table tbody tr:hover {
  background-color: #F9FAFB;
}

/* Missing file row special highlight */
.history-table tbody tr.row-missing {
  background-color: #FFF5F5;
  border-bottom-color: #FEE2E2;
}

.history-table tbody tr.row-missing:hover {
  background-color: #FEECEC;
}

/* Column Widths & Styles */
.col-filename {
  width: 32%;
}

.col-size {
  width: 15%;
  color: #4B5563;
  font-weight: 500;
}

.col-date {
  width: 18%;
  color: #4B5563;
  font-weight: 500;
}

.col-ai {
  width: 18%;
}

.col-status {
  width: 17%;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.thumb-box {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}

.filename-text {
  font-weight: 600;
  color: #111827;
  font-size: 13.5px;
}

/* AI Pill */
.ai-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 11.5px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 9999px;
}

/* Status Pills */
.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11.5px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 9999px;
  white-space: nowrap;
}

.status-passed {
  background: #ECFDF5;
  color: #059669;
}

.status-missing {
  background: #FEE2E2;
  color: #DC2626;
}

.status-deleted {
  background: #F3F4F6;
  color: #6B7280;
}
</style>

