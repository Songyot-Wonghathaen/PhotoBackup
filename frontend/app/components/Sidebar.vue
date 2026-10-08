<template>
  <aside class="sidebar">
    <!-- Brand Header -->
    <div class="brand">
      <div class="brand-icon">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
          <circle cx="8.5" cy="8.5" r="1.5"/>
          <polyline points="21 15 16 10 5 21"/>
        </svg>
      </div>
      <span class="brand-title">PhotoVault</span>
    </div>

    <!-- Navigation Menu via NuxtLink Routing -->
    <nav class="nav-menu">
      <NuxtLink 
        to="/" 
        class="nav-item" 
        active-class="active"
        exact-active-class="active"
        title="สำรองรูปภาพ (ระบบหลัก Part B)"
      >
        <svg class="nav-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="17 8 12 3 7 8"/>
          <line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
        <span>สำรองรูปภาพ</span>
      </NuxtLink>

      <!-- Mock navigation items for teammates via NuxtLink Routing -->
      <NuxtLink 
        to="/gallery" 
        class="nav-item mock-item" 
        active-class="active"
        title="ฟีเจอร์ส่วนของเพื่อน (UI Mockup)"
      >
        <svg class="nav-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="3" y="3" width="7" height="7"/>
          <rect x="14" y="3" width="7" height="7"/>
          <rect x="14" y="14" width="7" height="7"/>
          <rect x="3" y="14" width="7" height="7"/>
        </svg>
        <span>แกลเลอรี & ค้นหา</span>
      </NuxtLink>

      <NuxtLink 
        to="/history" 
        class="nav-item mock-item" 
        active-class="active"
        title="ฟีเจอร์ส่วนของเพื่อน (UI Mockup)"
      >
        <svg class="nav-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/>
          <polyline points="12 6 12 12 16 14"/>
        </svg>
        <span>ประวัติ & ตรวจสอบ</span>
      </NuxtLink>

      <NuxtLink 
        to="/settings" 
        class="nav-item mock-item" 
        active-class="active"
        title="การตั้งค่าระบบ"
      >
        <svg class="nav-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="3"/>
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
        </svg>
        <span>ตั้งค่า</span>
      </NuxtLink>
    </nav>

    <!-- Bottom Database Info Card -->
    <div class="db-card">
      <div class="db-header">
        <svg class="db-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#A5B4FC" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <ellipse cx="12" cy="5" rx="9" ry="3"/>
          <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
          <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
        </svg>
        <span class="db-name">SQLite · GORM</span>
      </div>
      <div class="db-path" :title="destPath || 'ยังไม่ได้กำหนดปลายทาง'">
        dest: {{ displayDest }}
      </div>
      <div class="db-badge">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#6EE7B7" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="20 6 9 17 4 12"/>
        </svg>
        <span>เชื่อมต่อแล้ว</span>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  destPath?: string
}>()

const displayDest = computed(() => {
  if (!props.destPath) return '~/Pictures/PhotoBackup'
  if (props.destPath.length > 22) {
    return '...' + props.destPath.slice(-19)
  }
  return props.destPath
})
</script>

<style scoped>
.sidebar {
  width: 220px;
  background-color: #0F172A;
  color: #FFFFFF;
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 24px 14px 16px 14px;
  box-sizing: border-box;
  flex-shrink: 0;
  user-select: none;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 6px 28px 6px;
}

.brand-icon {
  width: 36px;
  height: 36px;
  background: #4F46E5;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #FFFFFF;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.35);
}

.brand-title {
  font-family: 'Prompt', 'Inter', sans-serif;
  font-size: 18px;
  font-weight: 700;
  color: #FFFFFF;
  letter-spacing: -0.2px;
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 11px 14px;
  border-radius: 10px;
  border: none;
  background: transparent;
  color: #94A3B8;
  font-family: 'Prompt', 'Inter', sans-serif;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  text-align: left;
  text-decoration: none;
  transition: all 0.15s ease-in-out;
  box-sizing: border-box;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #E2E8F0;
}

.nav-item.active {
  background: #4F46E5;
  color: #FFFFFF;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

.nav-icon {
  flex-shrink: 0;
}

.mock-item {
  position: relative;
}

.db-card {
  background: #1E293B;
  border-radius: 12px;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.db-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.db-name {
  font-size: 12px;
  font-weight: 600;
  color: #F8FAFC;
}

.db-path {
  font-size: 11px;
  color: #94A3B8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: monospace;
}

.db-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #064E3B;
  color: #6EE7B7;
  font-size: 11px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: 9999px;
  width: fit-content;
  margin-top: 2px;
}
</style>
