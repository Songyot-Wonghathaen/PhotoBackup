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
          <!-- Thumbnail preview container with real image preview -->
          <div class="thumb-wrapper">
            <!-- Top Checkbox -->
            <div class="checkbox-container" @click.stop="toggleSelect(photo.filename)">
              <div class="custom-checkbox" :class="{ checked: isSelected(photo.filename) }">
                <svg v-if="isSelected(photo.filename)" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#FFFFFF" stroke-width="3">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
              </div>
            </div>

            <!-- Real Image Thumbnail via Wails IPC Base64 / ObjectURL -->
            <img 
              v-if="thumbnailMap[photo.full_path || '']"
              :src="thumbnailMap[photo.full_path || '']"
              :alt="photo.filename"
              class="thumb-img"
              loading="lazy"
            />

            <!-- Stylized fallback if image loading fails or browser mode -->
            <div v-else class="thumb-art-fallback" :style="getArtisticGradient(index)">
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
        <button
          class="status-pill pill-ai clickable"
          @click="showAiResultsModal = true"
          title="คลิกเพื่อดูคำอธิบายและ Tag จาก AI ทั้งหมด"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
          </svg>
          AI วิเคราะห์ {{ lastSummary.moved_files }}/{{ lastSummary.moved_files }} (คลิกดูคำอธิบาย & Tag)
        </button>
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

    <!-- AI Analysis Status Tab (Bottom Right when analyzing) -->
    <div v-if="isAnalyzing" class="ai-status-tab">
      <div class="ai-tab-header">
        <div class="ai-tab-icon">
          <svg class="ai-spinner" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
          </svg>
        </div>
        <div class="ai-tab-text">
          <span class="ai-tab-title">กำลังวิเคราะห์ภาพด้วย AI</span>
          <span class="ai-tab-progress">{{ aiProgress.current }}/{{ aiProgress.total }}</span>
        </div>
        <button class="ai-tab-cancel" @click="handleCancelAnalysis" title="ยกเลิก">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      <div class="ai-tab-filename">{{ aiProgress.filename }}</div>
      <div class="ai-tab-bar">
        <div class="ai-tab-bar-fill" :style="{ width: aiProgress.total > 0 ? ((aiProgress.current / aiProgress.total) * 100) + '%' : '0%' }"></div>
      </div>
    </div>

    <footer class="bottom-status-bar" v-else-if="!lastSummary.has_run">
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
        <button 
          class="status-pill pill-ai clickable" 
          @click="showAiResultsModal = true"
          title="คลิกเพื่อดูคำอธิบายและ Tag จาก AI ทั้งหมด"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
          </svg>
          AI วิเคราะห์ {{ lastSummary.moved_files }}/{{ lastSummary.moved_files }} (คลิกดูคำอธิบาย & Tag)
        </button>
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

    <!-- AI Analysis Results Modal (Part C: ฟีเจอร์ข้อ 12-14) -->
    <div v-if="showAiResultsModal" class="modal-backdrop" @click.self="showAiResultsModal = false">
      <div class="modal-card modal-large">
        <div class="modal-header">
          <div class="modal-title-row">
            <span class="modal-icon">🤖</span>
            <div>
              <h3>ผลการวิเคราะห์ภาพด้วย AI (Gemini Flash)</h3>
              <p class="modal-subtitle">บันทึกคำอธิบายและแท็กลง SQLite แล้ว (พบ {{ analyzedPhotosList.length }} ภาพ)</p>
            </div>
          </div>
          <button class="btn-close" @click="showAiResultsModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="analyzedPhotosList.length === 0" class="empty-modal-text">
            ยังไม่มีภาพที่วิเคราะห์ในโฟลเดอร์ปลายทางนี้
          </div>
          <div v-else class="ai-results-grid">
            <div v-for="item in analyzedPhotosList" :key="item.filename" class="ai-result-card">
              <div class="ai-card-header">
                <span class="ai-file-name">📄 {{ item.filename }}</span>
              </div>
              <div class="ai-card-desc">
                <span class="ai-label">คำอธิบายจาก AI:</span>
                <p class="ai-desc-text">{{ item.description || '(ไม่มีคำอธิบาย - ข้ามตามเงื่อนไขข้อ 14)' }}</p>
              </div>
              <div class="ai-card-tags" v-if="item.tags && item.tags.length > 0">
                <span class="ai-label">Tags:</span>
                <div class="ai-tag-pills">
                  <span v-for="tag in item.tags" :key="tag" class="ai-tag-chip">
                    #{{ tag }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showAiResultsModal = false">ปิด</button>
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

    <!-- Loading Spinner Modal (Only for isProcessing, not AI) -->
    <div v-if="isProcessing && !isAnalyzing" class="modal-backdrop loading-backdrop">
      <div class="loading-container">
        <div class="spinner"></div>
        <p class="loading-text">กำลังประมวลผล...</p>
        <p class="loading-subtext">รอสักครู่</p>
      </div>
    </div>

    <!-- Error Alert Modal -->
    <div v-if="aiError" class="modal-backdrop error-backdrop" @click.self="aiError = null">
      <div class="error-card">
        <div class="error-header">
          <span class="error-icon">❌</span>
          <h3>เกิดข้อผิดพลาด</h3>
          <button class="btn-close-error" @click="aiError = null">✕</button>
        </div>
        <div class="error-body">
          <p class="error-message">{{ aiError }}</p>
          <div class="error-suggestion">
            <strong>สาเหตุที่เป็นไปได้:</strong>
            <ul>
              <li>API Key หมดเงิน หรือสิ้นสุดการใช้งาน (Rate Limit)</li>
              <li>ปัญหาการเชื่อมต่ออินเทอร์เน็ต</li>
              <li>Server ของ Gemini AI ชั่วคราวขาด</li>
              <li>ไฟล์ภาพมีขนาดใหญ่เกินไป</li>
            </ul>
          </div>
        </div>
        <div class="error-footer">
          <button class="btn-error-close" @click="aiError = null">ปิด</button>
        </div>
      </div>
    </div>

    <!-- Success Notification Modal -->
    <div v-if="showSuccessModal" class="modal-backdrop success-backdrop" @click.self="showSuccessModal = false">
      <div class="success-card">
        <div class="success-icon-top">✅</div>
        <h3 class="success-title">สำรองสำเร็จ!</h3>
        <p class="success-msg">{{ successMessage }}</p>
        <div class="success-celebration">🎉</div>
        <button class="btn-success-ok" @click="showSuccessModal = false">ตกลง</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useBackup } from '~/composables/useBackup'

// Props & Emits
const props = defineProps<{
  sourcePathProp?: string
  destPathProp?: string
}>()

const emit = defineEmits<{
  (e: 'update:dest', path: string): void
  (e: 'update:source', path: string): void
}>()

const folderInputRef = ref<HTMLInputElement | null>(null)

// Consume all backend API operations and reactive state from useBackup composable
const {
  sourceType,
  sourcePath,
  destPath,
  destCount,
  isProcessing,
  isAnalyzing,
  aiProgress,
  sourceFiles,
  selectedFiles,
  thumbnailMap,
  destFilesList,
  showMissingModal,
  showDestFilesModal,
  showAiResultsModal,
  showSuccessModal,
  successMessage,
  analyzedPhotosList,
  aiError,
  integrityResult,
  lastSummary,
  sourceTotalBytes,
  displayMissingPreview,
  allSelected,
  isSelected,
  toggleSelect,
  toggleSelectAll,
  formatTotalSize,
  selectSourceFolder,
  handleBrowserFolderUpload,
  selectDestFolder,
  listDestFiles,
  moveSelectedPhotos,
  moveAllPhotos,
  initBackupState
} = useBackup()

// Sync props if provided
if (props.sourcePathProp) {
  sourcePath.value = props.sourcePathProp
}
if (props.destPathProp) {
  destPath.value = props.destPathProp
}

watch(sourcePath, (newVal) => emit('update:source', newVal))
watch(destPath, (newVal) => emit('update:dest', newVal))

// Action bridges for template buttons
function handleSelectSource() {
  selectSourceFolder(folderInputRef.value)
}

function handleSelectDest() {
  selectDestFolder()
}

function handleOpenDestList() {
  listDestFiles()
}

function handleMoveSelected() {
  moveSelectedPhotos()
}

function handleMoveAll() {
  moveAllPhotos()
}

function handleCancelAnalysis() {
  if (confirm('คุณต้องการยกเลิกการวิเคราะห์ภาพด้วย AI หรือไม่?')) {
    isAnalyzing.value = false
    aiProgress.value = { current: 0, total: 0, filename: '' }
  }
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
  const theme = artThemes[index % artThemes.length]!
  return { background: theme.bg }
}

function getSunStyle(index: number) {
  const theme = artThemes[index % artThemes.length]!
  return { background: theme.sun }
}

function getWaveColor(index: number) {
  const theme = artThemes[index % artThemes.length]!
  return theme.wave
}

onMounted(() => {
  initBackupState()
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

.thumb-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 0.2s ease;
}

.photo-card:hover .thumb-img {
  transform: scale(1.05);
}

.thumb-art-fallback {
  width: 100%;
  height: 100%;
  position: relative;
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

.pill-ai.clickable {
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid rgba(147, 51, 234, 0.3);
}

.pill-ai.clickable:hover {
  background: #E9D5FF;
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(147, 51, 234, 0.15);
}

.ai-results-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 380px;
  overflow-y: auto;
  padding: 4px;
}

.ai-result-card {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 10px;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ai-card-header {
  font-weight: 600;
  font-size: 13px;
  color: #1E293B;
}

.ai-label {
  font-size: 11px;
  font-weight: 600;
  color: #64748B;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 2px;
  display: block;
}

.ai-desc-text {
  margin: 0;
  font-size: 12.5px;
  color: #334155;
  line-height: 1.45;
  background: #FFFFFF;
  padding: 8px 10px;
  border-radius: 6px;
  border: 1px solid #E2E8F0;
}

.ai-tag-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.ai-tag-chip {
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 11px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 9999px;
  border: 1px solid #E0E7FF;
}

/* Loading Spinner Modal */
.loading-backdrop {
  background: rgba(15, 23, 42, 0.5);
  backdrop-filter: blur(3px);
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  background: #FFFFFF;
  border-radius: 16px;
  padding: 40px 32px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.15);
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #E5E7EB;
  border-top-color: #4F46E5;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading-text {
  font-size: 16px;
  font-weight: 600;
  color: #1F2937;
  margin: 0;
}

.loading-subtext {
  font-size: 13px;
  color: #6B7280;
  margin: 0;
}

/* Error Alert Modal */
.error-backdrop {
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(4px);
}

.error-card {
  background: #FFFFFF;
  border-radius: 16px;
  width: 480px;
  max-width: 90vw;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  overflow: hidden;
  border-left: 4px solid #EF4444;
}

.error-header {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  border-bottom: 1px solid #FEE2E2;
  background: #FEF2F2;
}

.error-icon {
  font-size: 20px;
  flex-shrink: 0;
}

.error-header h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: #DC2626;
  flex: 1;
}

.btn-close-error {
  background: transparent;
  border: none;
  font-size: 16px;
  color: #DC2626;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.error-body {
  padding: 18px 20px;
}

.error-message {
  background: #FEE2E2;
  border: 1px solid #FECACA;
  border-radius: 8px;
  padding: 12px 14px;
  margin: 0 0 14px 0;
  font-size: 13px;
  color: #991B1B;
  font-family: 'Inter', monospace;
  line-height: 1.5;
  word-break: break-word;
}

.error-suggestion {
  background: #FFFBEB;
  border: 1px solid #FDE68A;
  border-radius: 8px;
  padding: 12px 14px;
  font-size: 12px;
  color: #78350F;
}

.error-suggestion strong {
  display: block;
  margin-bottom: 6px;
  color: #92400E;
}

.error-suggestion ul {
  margin: 0;
  padding-left: 18px;
}

.error-suggestion li {
  margin: 4px 0;
  line-height: 1.4;
}

.error-footer {
  padding: 12px 20px;
  background: #F9FAFB;
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid #FEE2E2;
}

.btn-error-close {
  padding: 8px 18px;
  border-radius: 8px;
  border: 1px solid #DC2626;
  background: #DC2626;
  color: #FFFFFF;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-error-close:hover {
  background: #B91C1C;
  border-color: #B91C1C;
}

/* Success Modal */
.success-backdrop {
  background: rgba(15, 23, 42, 0.5);
  backdrop-filter: blur(4px);
}

.success-card {
  background: #FFFFFF;
  border-radius: 20px;
  width: 420px;
  max-width: 90vw;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  overflow: hidden;
  border-top: 6px solid #10B981;
  padding: 32px 28px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.success-icon-top {
  font-size: 56px;
  animation: bounce 0.6s ease;
}

@keyframes bounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

.success-title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #059669;
}

.success-msg {
  margin: 0;
  font-size: 14px;
  color: #374151;
  line-height: 1.6;
  background: #ECFDF5;
  border: 1px solid #A7F3D0;
  border-radius: 10px;
  padding: 14px 16px;
  width: 100%;
  box-sizing: border-box;
}

.success-celebration {
  font-size: 32px;
  margin: 8px 0;
  animation: celebrate 0.8s ease infinite alternate;
}

@keyframes celebrate {
  0% {
    transform: scale(1) rotate(0deg);
  }
  100% {
    transform: scale(1.1) rotate(10deg);
  }
}

.btn-success-ok {
  padding: 10px 28px;
  border-radius: 10px;
  border: none;
  background: linear-gradient(135deg, #10B981 0%, #059669 100%);
  color: #FFFFFF;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  margin-top: 8px;
}

.btn-success-ok:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(16, 185, 129, 0.4);
}

.btn-success-ok:active {
  transform: translateY(0);
}

/* Cancel Analysis Button */
.btn-cancel-analysis {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 10px;
  border: 1px solid #DC2626;
  background: #FFFFFF;
  color: #DC2626;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-top: 12px;
  box-shadow: 0 2px 6px rgba(220, 38, 38, 0.15);
}

.btn-cancel-analysis:hover {
  background: #FEF2F2;
  border-color: #B91C1C;
  color: #B91C1C;
  transform: translateY(-1px);
  box-shadow: 0 4px 10px rgba(220, 38, 38, 0.2);
}

.btn-cancel-analysis:active {
  transform: translateY(0);
  box-shadow: 0 2px 6px rgba(220, 38, 38, 0.15);
}

/* AI Analysis Status Tab (Bottom Right) */
.ai-status-tab {
  position: fixed;
  bottom: 24px;
  right: 24px;
  width: 360px;
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 14px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1), 0 4px 8px rgba(0, 0, 0, 0.05);
  padding: 14px 16px;
  z-index: 900;
  animation: slideInUp 0.3s ease;
}

@keyframes slideInUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.ai-tab-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.ai-tab-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #6366F1 0%, #4F46E5 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.ai-spinner {
  animation: sparkle 2s ease-in-out infinite;
  filter: drop-shadow(0 0 3px rgba(255, 255, 255, 0.8));
}

@keyframes sparkle {
  0%, 100% {
    transform: scale(1) rotate(0deg);
    opacity: 1;
  }
  50% {
    transform: scale(1.1) rotate(180deg);
    opacity: 0.8;
  }
}

.ai-tab-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.ai-tab-title {
  font-size: 13px;
  font-weight: 600;
  color: #111827;
}

.ai-tab-progress {
  font-size: 11px;
  font-weight: 500;
  color: #6B7280;
  font-family: 'Inter', monospace;
}

.ai-tab-cancel {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  color: #6B7280;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s;
  flex-shrink: 0;
}

.ai-tab-cancel:hover {
  background: #FEF2F2;
  border-color: #FECACA;
  color: #DC2626;
}

.ai-tab-filename {
  font-size: 11px;
  color: #4B5563;
  margin-bottom: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: monospace;
  padding-left: 2px;
}

.ai-tab-bar {
  height: 6px;
  background: #F3F4F6;
  border-radius: 9999px;
  overflow: hidden;
  position: relative;
}

.ai-tab-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #6366F1 0%, #8B5CF6 100%);
  border-radius: 9999px;
  transition: width 0.3s ease;
  position: relative;
  overflow: hidden;
}

.ai-tab-bar-fill::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4), transparent);
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}
</style>
