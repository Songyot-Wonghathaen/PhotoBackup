<template>
  <div class="history-view">
    <!-- Notice Banner that this is a Mockup of a friend's part -->
    <div class="mock-notice">
      <div class="notice-badge">UI Mockup</div>
      <span class="notice-text">
        หน้านี้เป็นภาพตัวอย่าง (ส่วนของเพื่อน): ระบบประวัติการสำรอง และรายงานการตรวจสอบความสมบูรณ์ของไฟล์ (Integrity Log)
      </span>
    </div>

    <!-- Header -->
    <header class="view-header">
      <div class="header-text">
        <h1 class="view-title">ประวัติ & ตรวจสอบ</h1>
        <p class="view-subtitle">ดูบันทึกรอบการสำรองข้อมูลย้อนหลัง และตรวจสอบความสมบูรณ์ของไฟล์ในดิสก์</p>
      </div>

      <button class="btn-refresh-history">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10"/>
          <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
        </svg>
        สแกนตรวจสอบไฟล์
      </button>
    </header>

    <!-- Stats Summary Row -->
    <div class="stats-row">
      <div class="stat-card">
        <span class="stat-label">รอบการสำรองทั้งหมด</span>
        <span class="stat-value">28 รอบ</span>
        <span class="stat-sub">ตั้งแต่ 1 ต.ค. 2026</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">รูปภาพในฐานข้อมูล</span>
        <span class="stat-value text-indigo">342 รูป</span>
        <span class="stat-sub">ความจุรวม 1.84 GB</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">ไฟล์ที่ตรวจพบสูญหาย</span>
        <span class="stat-value text-amber">2 รูป</span>
        <span class="stat-sub text-amber">ต้องทำการตรวจสอบ</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">ความเร็วเฉลี่ย</span>
        <span class="stat-value text-green">42.5 ms</span>
        <span class="stat-sub">16 Concurrency Workers</span>
      </div>
    </div>

    <!-- History Table Card -->
    <div class="history-card">
      <div class="card-title-row">
        <h3>บันทึกรอบการสำรองล่าสุด</h3>
        <span class="badge-count">แสดง 5 รายการล่าสุด</span>
      </div>

      <table class="history-table">
        <thead>
          <tr>
            <th>วันที่และเวลา</th>
            <th>ต้นทาง</th>
            <th>ปลายทาง</th>
            <th>จำนวนไฟล์</th>
            <th>เวลาที่ใช้</th>
            <th>สถานะ</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, idx) in historyRows" :key="idx">
            <td class="date-col">
              <span class="primary-text">{{ row.date }}</span>
              <span class="secondary-text">{{ row.time }}</span>
            </td>
            <td class="mono-col">{{ row.source }}</td>
            <td class="mono-col">{{ row.dest }}</td>
            <td>
              <span class="count-pill">{{ row.count }} รูป</span>
            </td>
            <td class="time-col">{{ row.duration }}</td>
            <td>
              <span class="status-badge" :class="row.statusClass">{{ row.statusText }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const historyRows = ref([
  {
    date: '8 ต.ค. 2026',
    time: '15:20:10',
    source: 'E:\\DCIM\\Camera',
    dest: '~/Pictures/PhotoBackup',
    count: 12,
    duration: '45.20 ms',
    statusClass: 'status-success',
    statusText: 'สำเร็จ 100%'
  },
  {
    date: '7 ต.ค. 2026',
    time: '18:45:00',
    source: 'E:\\DCIM\\100APPLE',
    dest: '~/Pictures/PhotoBackup',
    count: 45,
    duration: '124.30 ms',
    statusClass: 'status-success',
    statusText: 'สำเร็จ 100%'
  },
  {
    date: '5 ต.ค. 2026',
    time: '11:12:40',
    source: 'D:\\PhoneBackup',
    dest: '~/Pictures/PhotoBackup',
    count: 120,
    duration: '310.15 ms',
    statusClass: 'status-warning',
    statusText: 'มีไฟล์ซ้ำ 4 รูป'
  },
  {
    date: '2 ต.ค. 2026',
    time: '09:05:12',
    source: 'E:\\DCIM\\Camera',
    dest: '~/Pictures/PhotoBackup',
    count: 85,
    duration: '215.80 ms',
    statusClass: 'status-success',
    statusText: 'สำเร็จ 100%'
  },
  {
    date: '1 ต.ค. 2026',
    time: '14:30:00',
    source: 'E:\\Travel2026',
    dest: '~/Pictures/PhotoBackup',
    count: 80,
    duration: '198.40 ms',
    statusClass: 'status-success',
    statusText: 'สำเร็จ 100%'
  }
])
</script>

<style scoped>
.history-view {
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

.mock-notice {
  background: #EEF2FF;
  border: 1px dashed #6366F1;
  border-radius: 10px;
  padding: 8px 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
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
}

.view-subtitle {
  font-size: 13px;
  color: #6B7280;
  margin: 0;
}

.btn-refresh-history {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #FFFFFF;
  border: 1px solid #D1D5DB;
  color: #374151;
  font-size: 12.5px;
  font-weight: 600;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-refresh-history:hover {
  background: #F9FAFB;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin-bottom: 20px;
}

.stat-card {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-size: 12px;
  color: #6B7280;
  font-weight: 500;
}

.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: #111827;
}

.text-indigo { color: #4F46E5; }
.text-amber { color: #D97706; }
.text-green { color: #059669; }

.stat-sub {
  font-size: 11px;
  color: #9CA3AF;
}

.history-card {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 14px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}

.card-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.card-title-row h3 {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
  margin: 0;
}

.badge-count {
  font-size: 11.5px;
  color: #6B7280;
  background: #F3F4F6;
  padding: 3px 10px;
  border-radius: 9999px;
}

.history-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.history-table th {
  padding: 10px 14px;
  background: #F9FAFB;
  font-size: 12px;
  font-weight: 600;
  color: #4B5563;
  border-bottom: 1px solid #E5E7EB;
}

.history-table td {
  padding: 12px 14px;
  font-size: 12.5px;
  border-bottom: 1px solid #F3F4F6;
  color: #374151;
}

.date-col {
  display: flex;
  flex-direction: column;
}

.primary-text {
  font-weight: 600;
  color: #111827;
}

.secondary-text {
  font-size: 11px;
  color: #9CA3AF;
}

.mono-col {
  font-family: monospace;
  font-size: 12px;
  color: #4B5563;
}

.count-pill {
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 11.5px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 9999px;
}

.time-col {
  font-family: monospace;
  font-weight: 600;
  color: #4F46E5;
}

.status-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 6px;
}

.status-success {
  background: #ECFDF5;
  color: #059669;
}

.status-warning {
  background: #FEF3C7;
  color: #B45309;
}
</style>
