<template>
  <div class="history-view">
    <!-- Header -->
    <header class="view-header">
      <div class="header-text">
        <h1 class="view-title">ประวัติ & ตรวจสอบ</h1>
        <p class="view-subtitle">ประวัติการสำรองจากฐานข้อมูลของปลายทางปัจจุบัน (Part E: Features 20 - 25)</p>
      </div>

      <div class="header-actions">
        <!-- Change Destination (Feature 24) -->
        <button class="btn-change-dest" @click="changeDest" title="เปลี่ยนโฟลเดอร์ปลายทางเพื่อดูประวัติที่อื่น">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
          </svg>
          เปลี่ยนปลายทาง
        </button>

        <!-- Delete All Files in Dest (Feature 25) -->
        <button 
          class="btn-delete-all" 
          :disabled="isDeleting || stats.total === 0 || (stats.passed === 0 && stats.missing === 0)"
          @click="showDeleteAllModal = true"
          title="ลบไฟล์ทั้งหมดในโฟลเดอร์ปลายทางนี้ (/delete all)"
        >
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="3 6 5 6 21 6"/>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
            <line x1="10" y1="11" x2="10" y2="17"/>
            <line x1="14" y1="11" x2="14" y2="17"/>
          </svg>
          ลบทั้งหมด
        </button>

        <!-- Check Integrity (Feature 21) -->
        <button class="btn-check-integrity" :class="{ 'is-checking': isChecking }" :disabled="isChecking" @click="checkIntegrity()">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="{ 'spin-anim': isChecking }">
            <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
          </svg>
          <span>{{ isChecking ? 'กำลังตรวจสอบ...' : 'ตรวจสอบความถูกต้อง' }}</span>
        </button>
      </div>
    </header>

    <!-- Notification Alert Banner (Integrity report & actions) -->
    <transition name="fade">
      <div v-if="alertNotice" class="alert-banner" :class="[`alert-${alertNotice.type}`]">
        <div class="alert-content">
          <svg v-if="alertNotice.type === 'success'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          <svg v-else-if="alertNotice.type === 'warning'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
            <line x1="12" y1="9" x2="12" y2="13"/>
            <line x1="12" y1="17" x2="12.01" y2="17"/>
          </svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="16" x2="12" y2="12"/>
            <line x1="12" y1="8" x2="12.01" y2="8"/>
          </svg>
          <span class="alert-text">{{ alertNotice.message }}</span>
        </div>
        <button class="alert-close" @click="alertNotice = null">✕</button>
      </div>
    </transition>

    <!-- Destination Path Banner -->
    <div class="dest-banner">
      <div class="dest-info" @click="changeDest" title="คลิกเพื่อเปลี่ยนโฟลเดอร์ปลายทาง">
        <svg class="folder-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
        </svg>
        <span class="dest-label">ปลายทาง:</span>
        <span class="dest-path">{{ currentDest || 'ยังไม่ได้กำหนดโฟลเดอร์ปลายทาง' }}</span>
        <span class="badge-change">คลิกเปลี่ยน</span>
      </div>

      <div class="check-time-pill" title="เวลาที่ใช้ในการสแกนตรวจสอบความถูกต้องล่าสุด">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/>
          <polyline points="12 6 12 12 16 14"/>
        </svg>
        <span>ตรวจล่าสุดใน {{ lastCheckDuration }}</span>
      </div>
    </div>

    <!-- Stats Summary Cards (4 Cards - Feature 22) -->
    <div class="stats-grid">
      <!-- Card 1: ทั้งหมดในฐานข้อมูล -->
      <div 
        class="stat-card card-total" 
        :class="{ 'card-active': activeFilter === 'all' }"
        @click="activeFilter = 'all'"
        title="คลิกเพื่อแสดงไฟล์ทั้งหมด"
      >
        <div class="stat-card-top">
          <span class="stat-label">ทั้งหมดในฐานข้อมูล</span>
          <span class="filter-dot" v-if="activeFilter === 'all'">• ทั้งหมด</span>
        </div>
        <span class="stat-number">{{ stats.total }}</span>
      </div>

      <!-- Card 2: ตรวจสอบผ่าน -->
      <div 
        class="stat-card card-passed" 
        :class="{ 'card-active': activeFilter === 'passed' }"
        @click="activeFilter = activeFilter === 'passed' ? 'all' : 'passed'"
        title="คลิกเพื่อกรองเฉพาะไฟล์ที่ตรวจสอบผ่าน"
      >
        <div class="stat-card-top">
          <span class="stat-label text-emerald">ตรวจสอบผ่าน</span>
          <span class="filter-dot" v-if="activeFilter === 'passed'">• กำลังกรอง</span>
        </div>
        <span class="stat-number text-emerald">{{ stats.passed }}</span>
      </div>

      <!-- Card 3: ไฟล์สูญหาย -->
      <div 
        class="stat-card card-missing" 
        :class="{ 'card-active': activeFilter === 'missing' }"
        @click="activeFilter = activeFilter === 'missing' ? 'all' : 'missing'"
        title="คลิกเพื่อกรองเฉพาะไฟล์ที่สูญหายจากโฟลเดอร์"
      >
        <div class="stat-card-top">
          <span class="stat-label text-rose">ไฟล์สูญหาย</span>
          <span class="filter-dot" v-if="activeFilter === 'missing'">• กำลังกรอง</span>
        </div>
        <span class="stat-number text-rose">{{ stats.missing }}</span>
      </div>

      <!-- Card 4: ลบแล้ว -->
      <div 
        class="stat-card card-deleted" 
        :class="{ 'card-active': activeFilter === 'deleted' }"
        @click="activeFilter = activeFilter === 'deleted' ? 'all' : 'deleted'"
        title="คลิกเพื่อกรองเฉพาะไฟล์ที่ถูกลบแล้ว"
      >
        <div class="stat-card-top">
          <span class="stat-label">ลบแล้ว</span>
          <span class="filter-dot" v-if="activeFilter === 'deleted'">• กำลังกรอง</span>
        </div>
        <span class="stat-number text-gray">{{ stats.deleted }}</span>
      </div>
    </div>

    <!-- Filter & Search Toolbar -->
    <div class="table-toolbar">
      <div class="filter-tabs">
        <button 
          class="filter-tab-btn" 
          :class="{ active: activeFilter === 'all' }" 
          @click="activeFilter = 'all'"
        >
          ทั้งหมด ({{ stats.total }})
        </button>
        <button 
          class="filter-tab-btn tab-passed" 
          :class="{ active: activeFilter === 'passed' }" 
          @click="activeFilter = 'passed'"
        >
          ตรวจสอบผ่าน ({{ stats.passed }})
        </button>
        <button 
          class="filter-tab-btn tab-missing" 
          :class="{ active: activeFilter === 'missing' }" 
          @click="activeFilter = 'missing'"
        >
          ไฟล์สูญหาย ({{ stats.missing }})
        </button>
        <button 
          class="filter-tab-btn tab-deleted" 
          :class="{ active: activeFilter === 'deleted' }" 
          @click="activeFilter = 'deleted'"
        >
          ลบแล้ว ({{ stats.deleted }})
        </button>
      </div>

      <div class="search-box">
        <svg class="search-icon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#6B7280" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input 
          type="text" 
          v-model="searchQuery" 
          placeholder="ค้นหาชื่อไฟล์, AI description, หรือ Tag..." 
          class="search-input"
        />
        <button v-if="searchQuery" class="btn-clear-search" @click="searchQuery = ''">✕</button>
      </div>
    </div>

    <!-- Table Card (Feature 20) -->
    <div class="table-container">
      <div v-if="isLoading" class="loading-state">
        <svg class="spin-anim" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2.5">
          <circle cx="12" cy="12" r="10" stroke-opacity="0.25"/>
          <path d="M12 2a10 10 0 0 1 10 10"/>
        </svg>
        <span>กำลังโหลดประวัติจากฐานข้อมูล...</span>
      </div>

      <table v-else-if="filteredRows.length > 0" class="history-table">
        <thead>
          <tr>
            <th class="col-filename">ชื่อไฟล์</th>
            <th class="col-size">ขนาด</th>
            <th class="col-date">วันที่สำรอง</th>
            <th class="col-ai">AI</th>
            <th class="col-status">สถานะ</th>
            <th class="col-actions">จัดการ</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="row in filteredRows" 
            :key="row.id" 
            :class="{ 'row-missing': row.status === 'missing', 'row-deleted': row.status === 'deleted' }"
          >
            <!-- Filename with colorful preview icon / real thumbnail -->
            <td class="col-filename">
              <div class="file-item">
                <div class="thumb-box" :style="{ background: row.gradient }">
                  <img 
                    v-if="row.full_path && row.status !== 'missing' && row.status !== 'deleted'"
                    :src="`/api/thumbnail?path=${encodeURIComponent(row.full_path)}`"
                    :alt="row.name"
                    class="real-thumb-img"
                    @error="($event.target as HTMLElement).style.display = 'none'"
                  />
                  <svg class="fallback-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                    <circle cx="8.5" cy="8.5" r="1.5"/>
                    <polyline points="21 15 16 10 5 21"/>
                  </svg>
                </div>
                <div class="filename-group">
                  <span class="filename-text" :title="row.name">{{ row.name }}</span>
                  <span v-if="row.status === 'missing'" class="missing-hint">ไม่พบในโฟลเดอร์จริง</span>
                </div>
              </div>
            </td>

            <!-- File Size -->
            <td class="col-size">{{ row.size }}</td>

            <!-- Backup Date -->
            <td class="col-date">{{ row.date }}</td>

            <!-- AI Status (Clickable to view detail modal) -->
            <td class="col-ai">
              <button 
                v-if="row.ai === 'วิเคราะห์แล้ว'" 
                class="ai-pill ai-pill-clickable"
                @click="selectedPhotoForDetail = row"
                title="คลิกเพื่อดูรายละเอียดและ Tag ที่ AI วิเคราะห์"
              >
                <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z"/>
                </svg>
                {{ row.ai }}
              </button>
              <span v-else class="ai-pill ai-pill-empty">
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

            <!-- Action: Delete single file (Feature 23) -->
            <td class="col-actions">
              <button 
                v-if="row.status !== 'deleted'"
                class="btn-delete-row"
                @click="confirmDeleteSingle(row)"
                title="ลบไฟล์นี้จากโฟลเดอร์ปลายทางและบันทึกสถานะ 'ลบแล้ว' ในฐานข้อมูล"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="3 6 5 6 21 6"/>
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                </svg>
                <span>ลบ</span>
              </button>
              <span v-else class="text-deleted-label">ลบแล้ว</span>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Empty State -->
      <div v-else class="empty-state">
        <div class="empty-icon-box">
          <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="#9CA3AF" stroke-width="1.8">
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
            <circle cx="8.5" cy="8.5" r="1.5"/>
            <polyline points="21 15 16 10 5 21"/>
          </svg>
        </div>
        <h3 class="empty-title">ไม่พบประวัติการสำรอง</h3>
        <p class="empty-desc">
          {{ searchQuery ? 'ไม่พบไฟล์ที่ตรงกับคำค้นหา ลองเปลี่ยนคำค้นหา' : 'ยังไม่มีข้อมูลการสำรองรูปภาพในโฟลเดอร์ปลายทางนี้' }}
        </p>
        <button v-if="searchQuery" class="btn-reset-search" @click="searchQuery = ''">
          ล้างคำค้นหา
        </button>
        <button v-else class="btn-change-dest-empty" @click="changeDest">
          เลือกโฟลเดอร์ปลายทางอื่น
        </button>
      </div>
    </div>

    <!-- Modal 1: Confirm Delete Single File (Feature 23) -->
    <div v-if="photoToDelete" class="modal-backdrop" @click.self="photoToDelete = null">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon-warning">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#DC2626" stroke-width="2">
              <polyline points="3 6 5 6 21 6"/>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
            </svg>
          </div>
          <div>
            <h3 class="modal-title">ยืนยันการลบไฟล์</h3>
            <p class="modal-subtitle">ต้องการลบไฟล์นี้ออกจากโฟลเดอร์ปลายทางใช่หรือไม่?</p>
          </div>
        </div>

        <div class="modal-file-preview">
          <span class="preview-filename">{{ photoToDelete.name }}</span>
          <span class="preview-filesize">ขนาด: {{ photoToDelete.size }}</span>
        </div>

        <p class="modal-warning-text">
          ⚠️ ไฟล์จริงจะถูกลบออกจากดิสก์ และฐานข้อมูลจะบันทึกสถานะเป็น <strong>"ลบแล้ว"</strong> เพื่อเก็บประวัติ
        </p>

        <div class="modal-actions">
          <button class="btn-cancel" :disabled="isDeleting" @click="photoToDelete = null">ยกเลิก</button>
          <button class="btn-danger" :disabled="isDeleting" @click="executeDeleteSingle">
            {{ isDeleting ? 'กำลังลบ...' : 'ยืนยันลบไฟล์' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal 2: Confirm Delete All Files (Feature 25) -->
    <div v-if="showDeleteAllModal" class="modal-backdrop" @click.self="showDeleteAllModal = false">
      <div class="modal-card modal-card-danger">
        <div class="modal-header">
          <div class="modal-icon-danger">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#DC2626" stroke-width="2">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
              <line x1="12" y1="9" x2="12" y2="13"/>
              <line x1="12" y1="17" x2="12.01" y2="17"/>
            </svg>
          </div>
          <div>
            <h3 class="modal-title text-danger">ลบไฟล์ทั้งหมดในปลายทาง (/delete all)</h3>
            <p class="modal-subtitle">คำเตือน: การกระทำนี้ไม่สามารถย้อนกลับได้</p>
          </div>
        </div>

        <div class="modal-file-preview">
          <span class="preview-filename">โฟลเดอร์: {{ currentDest }}</span>
          <span class="preview-filesize">จำนวนไฟล์ในฐานข้อมูล: {{ stats.total }} ไฟล์</span>
        </div>

        <p class="modal-warning-text danger-box">
          ไฟล์จริงทั้งหมดในโฟลเดอร์ปลายทางจะถูกลบออกจากดิสก์ทันที และข้อมูลรูปภาพทั้งหมดจะถูกเปลี่ยนสถานะเป็น <strong>"ลบแล้ว"</strong> ในฐานข้อมูล
        </p>

        <div class="modal-actions">
          <button class="btn-cancel" :disabled="isDeleting" @click="showDeleteAllModal = false">ยกเลิก</button>
          <button class="btn-danger" :disabled="isDeleting" @click="executeDeleteAll">
            {{ isDeleting ? 'กำลังลบทั้งหมด...' : 'ยืนยันลบไฟล์ทั้งหมด' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal 3: AI Detail View Modal -->
    <div v-if="selectedPhotoForDetail" class="modal-backdrop" @click.self="selectedPhotoForDetail = null">
      <div class="modal-card modal-card-ai">
        <div class="modal-header">
          <div class="modal-icon-ai">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="#4F46E5">
              <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z"/>
            </svg>
          </div>
          <div>
            <h3 class="modal-title">ข้อมูลภาพและการวิเคราะห์ AI</h3>
            <p class="modal-subtitle">{{ selectedPhotoForDetail.name }}</p>
          </div>
          <button class="btn-close-modal" @click="selectedPhotoForDetail = null">✕</button>
        </div>

        <!-- Preview Image -->
        <div class="ai-preview-img-box" :style="{ background: selectedPhotoForDetail.gradient }">
          <img 
            v-if="selectedPhotoForDetail.full_path"
            :src="`/api/thumbnail?path=${encodeURIComponent(selectedPhotoForDetail.full_path)}`"
            :alt="selectedPhotoForDetail.name"
            class="ai-full-img"
          />
        </div>

        <div class="ai-meta-row">
          <span>ขนาด: <strong>{{ selectedPhotoForDetail.size }}</strong></span>
          <span>วันที่สำรอง: <strong>{{ selectedPhotoForDetail.date }}</strong></span>
          <span>สถานะ: <strong>{{ selectedPhotoForDetail.status }}</strong></span>
        </div>

        <!-- AI Description -->
        <div class="ai-section">
          <h4 class="ai-section-title">คำอธิบายจาก AI</h4>
          <p class="ai-desc-text">
            {{ selectedPhotoForDetail.description || 'ไม่มีคำอธิบายสำหรับภาพนี้' }}
          </p>
        </div>

        <!-- AI Tags -->
        <div class="ai-section">
          <h4 class="ai-section-title">Tags ({{ selectedPhotoForDetail.tags.length }})</h4>
          <div v-if="selectedPhotoForDetail.tags.length > 0" class="tags-cloud">
            <span v-for="tag in selectedPhotoForDetail.tags" :key="tag" class="tag-chip">
              #{{ tag }}
            </span>
          </div>
          <p v-else class="text-no-tags">ไม่มีแท็กสำหรับภาพนี้</p>
        </div>

        <div class="modal-actions">
          <button class="btn-primary" @click="selectedPhotoForDetail = null">ปิด</button>
        </div>
      </div>
    </div>

    <!-- Modal 4: Change Destination Manually (Feature 24 Fallback) -->
    <div v-if="showChangeDestModal" class="modal-backdrop" @click.self="showChangeDestModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon-folder">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
            </svg>
          </div>
          <div>
            <h3 class="modal-title">เปลี่ยนโฟลเดอร์ปลายทาง (/list db)</h3>
            <p class="modal-subtitle">ระบุพาธของโฟลเดอร์ปลายทางที่ต้องการดูประวัติ</p>
          </div>
        </div>

        <div class="input-group">
          <label class="input-label">พาธโฟลเดอร์ปลายทาง:</label>
          <input 
            type="text" 
            v-model="manualDestInput" 
            placeholder="เช่น C:/Users/.../Pictures/PhotoBackup" 
            class="dest-text-input"
            @keyup.enter="applyManualDest(manualDestInput)"
          />
        </div>

        <div class="modal-actions">
          <button class="btn-cancel" @click="showChangeDestModal = false">ยกเลิก</button>
          <button class="btn-primary" @click="applyManualDest(manualDestInput)">ตกลง</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useHistory } from '~/composables/useHistory'

const {
  currentDest,
  rows,
  filteredRows,
  stats,
  isChecking,
  isDeleting,
  isLoading,
  lastCheckDuration,
  activeFilter,
  searchQuery,
  alertNotice,

  selectedPhotoForDetail,
  photoToDelete,
  showDeleteAllModal,
  showChangeDestModal,
  manualDestInput,

  initHistory,
  checkIntegrity,
  changeDest,
  applyManualDest,
  confirmDeleteSingle,
  executeDeleteSingle,
  executeDeleteAll
} = useHistory()

onMounted(() => {
  initHistory()
})
</script>

<style scoped>
.history-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  max-height: 100%;
  padding: 24px 32px 32px 32px;
  box-sizing: border-box;
  background-color: #F4F5FA;
  overflow-y: auto;
  user-select: none;
  font-family: 'Prompt', 'Inter', sans-serif;
}

/* Header */
.view-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
  gap: 16px;
  flex-wrap: wrap;
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

.btn-delete-all {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  background: #FEF2F2;
  border: 1px solid #FECACA;
  color: #DC2626;
  font-size: 13px;
  font-weight: 600;
  padding: 9px 15px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-delete-all:hover:not(:disabled) {
  background: #FEE2E2;
  border-color: #FCA5A5;
}

.btn-delete-all:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.btn-check-integrity:hover:not(:disabled) {
  background: #4338CA;
  border-color: #4338CA;
}

.btn-check-integrity.is-checking,
.btn-check-integrity:disabled {
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

/* Alert Notification Banner */
.alert-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-radius: 10px;
  margin-bottom: 14px;
  font-size: 13px;
  font-weight: 500;
  box-shadow: 0 1px 2px rgba(0,0,0,0.03);
}

.alert-success {
  background: #ECFDF5;
  border: 1px solid #A7F3D0;
  color: #065F46;
}

.alert-warning {
  background: #FFF1F2;
  border: 1px solid #FECDD3;
  color: #9F1239;
}

.alert-info {
  background: #EEF2FF;
  border: 1px solid #C7D2FE;
  color: #3730A3;
}

.alert-content {
  display: flex;
  align-items: center;
  gap: 10px;
}

.alert-close {
  background: transparent;
  border: none;
  font-size: 14px;
  cursor: pointer;
  color: inherit;
  opacity: 0.7;
}

.alert-close:hover {
  opacity: 1;
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
  cursor: pointer;
}

.dest-info:hover .dest-path {
  color: #4F46E5;
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
  transition: color 0.15s ease;
}

.badge-change {
  font-size: 11px;
  background: #F3F4F6;
  color: #4F46E5;
  padding: 2px 7px;
  border-radius: 6px;
  font-weight: 600;
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
  margin-bottom: 18px;
}

.stat-card {
  border-radius: 12px;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.02);
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px -1px rgba(0,0,0,0.06);
}

.stat-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.filter-dot {
  font-size: 11px;
  font-weight: 600;
  color: #4F46E5;
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
  border: 1.5px solid #E5E7EB;
}

.card-total.card-active {
  border-color: #4F46E5;
  background: #F5F7FF;
}

.card-passed {
  background: #ECFDF5;
  border: 1.5px solid #A7F3D0;
}

.card-passed.card-active {
  border-color: #059669;
  background: #D1FAE5;
}

.card-missing {
  background: #FEE2E2;
  border: 1.5px solid #FECACA;
}

.card-missing.card-active {
  border-color: #DC2626;
  background: #FEE2E2;
  box-shadow: 0 0 0 2px rgba(220, 38, 38, 0.2);
}

.card-deleted {
  background: #FFFFFF;
  border: 1.5px solid #E5E7EB;
}

.card-deleted.card-active {
  border-color: #6B7280;
  background: #F9FAFB;
}

.text-emerald { color: #059669 !important; }
.text-rose { color: #DC2626 !important; }
.text-gray { color: #374151 !important; }

/* Filter & Search Toolbar */
.table-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  gap: 16px;
  flex-wrap: wrap;
}

.filter-tabs {
  display: flex;
  align-items: center;
  gap: 6px;
  background: #FFFFFF;
  padding: 4px;
  border-radius: 10px;
  border: 1px solid #E5E7EB;
}

.filter-tab-btn {
  background: transparent;
  border: none;
  font-size: 12.5px;
  font-weight: 600;
  color: #6B7280;
  padding: 6px 12px;
  border-radius: 7px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.filter-tab-btn:hover {
  color: #111827;
}

.filter-tab-btn.active {
  background: #4F46E5;
  color: #FFFFFF;
}

.filter-tab-btn.tab-passed.active {
  background: #059669;
}

.filter-tab-btn.tab-missing.active {
  background: #DC2626;
}

.filter-tab-btn.tab-deleted.active {
  background: #4B5563;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 10px;
  padding: 6px 14px;
  min-width: 320px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.02);
}

.search-icon {
  flex-shrink: 0;
}

.search-input {
  border: none;
  outline: none;
  font-size: 12.5px;
  color: #111827;
  width: 100%;
  font-family: inherit;
}

.search-input::placeholder {
  color: #9CA3AF;
}

.btn-clear-search {
  background: transparent;
  border: none;
  font-size: 12px;
  color: #9CA3AF;
  cursor: pointer;
}

/* Table Container */
.table-container {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 14px;
  overflow-x: auto;
  overflow-y: auto;
  max-height: 560px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}

.history-table {
  width: 100%;
  min-width: 820px;
  border-collapse: collapse;
  text-align: left;
}

.history-table th {
  position: sticky;
  top: 0;
  z-index: 5;
  padding: 14px 20px;
  font-size: 12.5px;
  font-weight: 600;
  color: #6B7280;
  border-bottom: 1px solid #E5E7EB;
  background: #FAFAFA;
  box-shadow: 0 1px 0 #E5E7EB;
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

.history-table tbody tr.row-deleted {
  opacity: 0.65;
}

/* Column Widths & Styles */
.col-filename {
  width: 32%;
}

.col-size {
  width: 12%;
  color: #4B5563;
  font-weight: 500;
}

.col-date {
  width: 16%;
  color: #4B5563;
  font-weight: 500;
}

.col-ai {
  width: 16%;
}

.col-status {
  width: 14%;
}

.col-actions {
  width: 10%;
  text-align: right;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.thumb-box {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
  position: relative;
  overflow: hidden;
}

.real-thumb-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  position: absolute;
  top: 0;
  left: 0;
}

.fallback-icon {
  z-index: 1;
}

.filename-group {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.filename-text {
  font-weight: 600;
  color: #111827;
  font-size: 13.5px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 260px;
}

.missing-hint {
  font-size: 11px;
  color: #DC2626;
  font-weight: 500;
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
  border: none;
  font-family: inherit;
}

.ai-pill-clickable {
  cursor: pointer;
  transition: all 0.15s ease;
}

.ai-pill-clickable:hover {
  background: #E0E7FF;
  transform: scale(1.02);
}

.ai-pill-empty {
  background: #F3F4F6;
  color: #9CA3AF;
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

/* Row Action Buttons */
.btn-delete-row {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: 1px solid #FECACA;
  color: #DC2626;
  font-size: 11.5px;
  font-weight: 600;
  padding: 4px 9px;
  border-radius: 7px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-delete-row:hover {
  background: #FEE2E2;
}

.text-deleted-label {
  font-size: 11.5px;
  color: #9CA3AF;
  font-weight: 500;
}

/* Loading & Empty States */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  gap: 12px;
  color: #6B7280;
  font-size: 13.5px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 56px 24px;
  text-align: center;
}

.empty-icon-box {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: #F3F4F6;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 14px;
}

.empty-title {
  font-size: 16px;
  font-weight: 700;
  color: #374151;
  margin: 0 0 6px 0;
}

.empty-desc {
  font-size: 13px;
  color: #6B7280;
  max-width: 340px;
  margin: 0 0 16px 0;
}

.btn-reset-search,
.btn-change-dest-empty {
  background: #4F46E5;
  color: #FFFFFF;
  border: none;
  font-size: 13px;
  font-weight: 600;
  padding: 8px 18px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.btn-reset-search:hover,
.btn-change-dest-empty:hover {
  background: #4338CA;
}

/* Modals */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(17, 24, 39, 0.45);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
  box-sizing: border-box;
}

.modal-card {
  background: #FFFFFF;
  border-radius: 16px;
  max-width: 480px;
  width: 100%;
  padding: 24px;
  box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1), 0 10px 10px -5px rgba(0,0,0,0.04);
  box-sizing: border-box;
  animation: modalPop 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-card-ai {
  max-width: 540px;
}

@keyframes modalPop {
  from {
    opacity: 0;
    transform: scale(0.96) translateY(10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modal-header {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 16px;
  position: relative;
}

.modal-icon-warning,
.modal-icon-danger,
.modal-icon-ai,
.modal-icon-folder {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.modal-icon-warning { background: #FEF2F2; }
.modal-icon-danger { background: #FEE2E2; }
.modal-icon-ai { background: #EEF2FF; }
.modal-icon-folder { background: #EEF2FF; }

.modal-title {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 3px 0;
}

.text-danger {
  color: #DC2626 !important;
}

.modal-subtitle {
  font-size: 12.5px;
  color: #6B7280;
  margin: 0;
}

.btn-close-modal {
  position: absolute;
  top: 0;
  right: 0;
  background: transparent;
  border: none;
  font-size: 16px;
  color: #9CA3AF;
  cursor: pointer;
}

.modal-file-preview {
  background: #F9FAFB;
  border: 1px solid #E5E7EB;
  border-radius: 10px;
  padding: 10px 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
  font-size: 13px;
}

.preview-filename {
  font-weight: 600;
  color: #111827;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 280px;
}

.preview-filesize {
  color: #6B7280;
  font-size: 12px;
}

.modal-warning-text {
  font-size: 12.5px;
  color: #4B5563;
  line-height: 1.5;
  margin: 0 0 20px 0;
}

.danger-box {
  background: #FEF2F2;
  border: 1px solid #FECACA;
  border-radius: 10px;
  padding: 12px 14px;
  color: #991B1B;
}

.modal-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

.btn-cancel {
  background: #FFFFFF;
  border: 1px solid #D1D5DB;
  color: #374151;
  font-size: 13px;
  font-weight: 600;
  padding: 8px 16px;
  border-radius: 9px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-cancel:hover:not(:disabled) {
  background: #F9FAFB;
}

.btn-danger {
  background: #DC2626;
  border: 1px solid #DC2626;
  color: #FFFFFF;
  font-size: 13px;
  font-weight: 600;
  padding: 8px 18px;
  border-radius: 9px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-danger:hover:not(:disabled) {
  background: #B91C1C;
}

.btn-primary {
  background: #4F46E5;
  border: 1px solid #4F46E5;
  color: #FFFFFF;
  font-size: 13px;
  font-weight: 600;
  padding: 8px 18px;
  border-radius: 9px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-primary:hover:not(:disabled) {
  background: #4338CA;
}

/* AI Detail Modal specific styles */
.ai-preview-img-box {
  width: 100%;
  height: 180px;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 14px;
}

.ai-full-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.ai-meta-row {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 12.5px;
  color: #6B7280;
  margin-bottom: 14px;
  padding-bottom: 12px;
  border-bottom: 1px solid #F3F4F6;
}

.ai-meta-row strong {
  color: #111827;
}

.ai-section {
  margin-bottom: 16px;
}

.ai-section-title {
  font-size: 12.5px;
  font-weight: 700;
  color: #374151;
  margin: 0 0 6px 0;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.ai-desc-text {
  font-size: 13.5px;
  color: #1F2937;
  background: #F9FAFB;
  border: 1px solid #E5E7EB;
  border-radius: 10px;
  padding: 10px 14px;
  line-height: 1.5;
  margin: 0;
}

.tags-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag-chip {
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 9999px;
}

.text-no-tags {
  font-size: 12px;
  color: #9CA3AF;
  font-style: italic;
  margin: 0;
}

/* Manual dest input */
.input-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 20px;
}

.input-label {
  font-size: 12.5px;
  font-weight: 600;
  color: #374151;
}

.dest-text-input {
  border: 1.5px solid #E5E7EB;
  border-radius: 10px;
  padding: 10px 14px;
  font-size: 13px;
  outline: none;
  font-family: inherit;
  transition: border-color 0.15s ease;
}

.dest-text-input:focus {
  border-color: #4F46E5;
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
