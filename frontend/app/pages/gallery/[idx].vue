<template>
  <div class="photo-detail-page">
    <!-- Notice Banner that this is a Mockup of a friend's part -->
    <div class="mock-notice">
      <div class="notice-badge">UI Mockup</div>
      <span class="notice-text">
        หน้านี้เป็นภาพตัวอย่าง (ส่วนของเพื่อน): หน้ารายละเอียดภาพ คำอธิบาย AI, ข้อมูลไฟล์ และการแก้ไข Tag
      </span>
    </div>

    <!-- Header with Back Button -->
    <template v-if="currentPhoto">
    <header class="detail-header">
      <div class="header-left">
        <h1 class="photo-filename">{{ currentPhoto.filename }}</h1>
        <p class="header-subtitle">รายละเอียดรูปภาพ</p>
      </div>

      <NuxtLink to="/gallery" class="btn-back">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="19" y1="12" x2="5" y2="12"/>
          <polyline points="12 19 5 12 12 5"/>
        </svg>
        <span>กลับไปแกลเลอรี่</span>
      </NuxtLink>
    </header>

    <!-- Main 2-Column Content Layout -->
    <div class="detail-content-layout">
      <!-- Left Column: Big Preview + Thumbnails Strip -->
      <section class="preview-column">
        <!-- Big Picture Preview Container -->
        <div class="main-preview-card" :style="currentPhoto.bgStyle">
          <div class="preview-sun" :style="currentPhoto.sunStyle"></div>
          
          <svg class="preview-wave preview-wave-back" viewBox="0 0 560 200" preserveAspectRatio="none">
            <path d="M0,90 C120,40 260,130 560,70 L560,200 L0,200 Z" :fill="currentPhoto.waveBack" />
          </svg>
          <svg class="preview-wave preview-wave-front" viewBox="0 0 560 200" preserveAspectRatio="none">
            <path d="M0,130 C150,70 320,140 560,95 L560,200 L0,200 Z" :fill="currentPhoto.waveFront" />
          </svg>
        </div>

        <!-- Thumbnails Strip below preview -->
        <div class="thumbnails-strip">
          <div 
            v-for="(thumb, i) in galleryStrip" 
            :key="thumb.filename"
            class="thumb-strip-item"
            :class="{ active: currentPhotoIndex === i }"
            @click="selectPhoto(i)"
          >
            <div class="thumb-mini-art" :style="thumb.bgStyle">
              <div class="thumb-mini-sun" :style="thumb.sunStyle"></div>
              <svg class="thumb-mini-wave" viewBox="0 0 100 40" preserveAspectRatio="none">
                <path d="M0,18 C25,8 55,26 100,14 L100,40 L0,40 Z" :fill="thumb.waveBack" />
                <path d="M0,26 C30,14 65,28 100,19 L100,40 L0,40 Z" :fill="thumb.waveFront" />
              </svg>
            </div>
          </div>
        </div>
      </section>

      <!-- Right Column: AI Description, Tags, Metadata, Actions -->
      <aside class="sidebar-column">
        <!-- Backup Status Tag -->
        <div class="status-badge-row">
          <span class="badge-backed-up">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
            <span>สำรองแล้ว</span>
          </span>
        </div>

        <!-- AI Description Card (Figma Screen) -->
        <div class="info-card ai-desc-card">
          <div class="card-header-row">
            <div class="card-title-badge">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2">
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
              </svg>
              <span>คำอธิบายจาก AI</span>
            </div>

            <div class="header-action-group">
              <button class="btn-ai-live" :disabled="isAnalyzingLive" @click="handleRunAiAnalysis" title="ส่งภาพให้ Gemini วิเคราะห์คำอธิบายและ Tag ใหม่">
                <span v-if="!isAnalyzingLive">✨ วิเคราะห์ด้วย AI</span>
                <span v-else>⏳ กำลังวิเคราะห์...</span>
              </button>
              <button
                v-if="isAnalyzingLive"
                class="btn-cancel-ai"
                @click="handleCancelAnalysis"
                title="ยกเลิกการวิเคราะห์"
              >
                ✕ ยกเลิก
              </button>
              <button class="btn-edit" @click="isEditingDesc = !isEditingDesc">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                  <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                </svg>
                <span>{{ isEditingDesc ? 'เสร็จสิ้น' : 'แก้ไข' }}</span>
              </button>
            </div>
          </div>

          <div v-if="!isEditingDesc" class="ai-desc-text">
            {{ currentPhoto.description }}
          </div>
          <div v-else class="edit-desc-box">
            <textarea v-model="currentPhoto.description" class="desc-textarea" rows="3"></textarea>
          </div>
        </div>

        <!-- Tags Card -->
        <div class="info-card tags-card">
          <div class="card-title-row">
            <div class="card-title-badge">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#4B5563" stroke-width="2">
                <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
                <line x1="7" y1="7" x2="7.01" y2="7"/>
              </svg>
              <span class="text-dark">Tags</span>
            </div>
          </div>

          <div class="tags-chips-container">
            <span 
              v-for="(tag, idx) in currentPhoto.tags" 
              :key="tag" 
              class="tag-pill-item"
            >
              <span>{{ tag }}</span>
              <button class="btn-remove-tag" @click="removeTag(idx)" title="ลบแท็ก">✕</button>
            </span>

            <button class="btn-add-tag" @click="promptAddTag">
              <span>+ เพิ่ม Tag</span>
            </button>
          </div>
        </div>

        <!-- File Metadata Card (ข้อมูลไฟล์) -->
        <div class="info-card file-metadata-card">
          <h3 class="card-title-heading">ข้อมูลไฟล์</h3>
          <div class="meta-list">
            <div class="meta-row">
              <span class="meta-label">ที่เก็บ</span>
              <span class="meta-value path-val" :title="currentPhoto.path">{{ currentPhoto.path }}</span>
            </div>
            <div class="meta-row">
              <span class="meta-label">ขนาด</span>
              <span class="meta-value">{{ currentPhoto.size }}</span>
            </div>
            <div class="meta-row">
              <span class="meta-label">สำรองเมื่อ</span>
              <span class="meta-value">{{ currentPhoto.backupDate }}</span>
            </div>
            <div class="meta-row">
              <span class="meta-label">ต้นทาง</span>
              <span class="meta-value">{{ currentPhoto.source }}</span>
            </div>
          </div>
        </div>

        <!-- Delete Action Button -->
        <div class="delete-action-row">
          <button class="btn-delete-dest" @click="handleDelete">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="3 6 5 6 21 6"/>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
            </svg>
            <span>ลบจากปลายทาง</span>
          </button>
        </div>
      </aside>
    </div>
    </template> <!-- end v-if="currentPhoto" -->
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface GalleryPhoto {
  filename: string
  description: string
  tags: string[]
  path: string
  size: string
  backupDate: string
  source: string
  bgStyle: { background: string }
  sunStyle: { background: string }
  waveBack: string
  waveFront: string
}

const route = useRoute()
const isEditingDesc = ref(false)

const currentPhotoIndex = ref(0)

const galleryStrip = ref([
  {
    filename: 'beach_trip.png',
    description: 'ชายหาดทรายขาวกับทะเลสีฟ้าใสในวันที่ท้องฟ้าโปร่ง มีเมฆบาง ๆ และภูเขาอยู่ไกล ๆ เหมาะกับบรรยากาศท่องเที่ยวพักผ่อน',
    tags: ['ชายหาด', 'ทะเล', 'ท้องฟ้า', 'ภูเขา', 'ท่องเที่ยว', 'พักผ่อน'],
    path: '~/Pictures/PhotoBackup/beach_trip.png',
    size: '3.4 MB',
    backupDate: '3 ต.ค. 2569 07:21',
    source: 'Flash Drive · /DCIM',
    bgStyle: { background: 'linear-gradient(180deg, #38BDF8 0%, #BAE6FD 100%)' },
    sunStyle: { background: '#FFFFFF' },
    waveBack: '#0EA5E9',
    waveFront: '#0369A1'
  },
  {
    filename: 'sunset.jpg',
    description: 'หาดทรายขาวกับท้องฟ้าสีส้มยามพระอาทิตย์ตก มีเรือเล็กลอยอยู่ไกล ๆ ทิวทัศน์เงียบสงบ',
    tags: ['ชายหาด', 'พระอาทิตย์ตก', 'ทะเล'],
    path: '~/Pictures/PhotoBackup/sunset.jpg',
    size: '4.2 MB',
    backupDate: '3 ต.ค. 2569 07:22',
    source: 'Flash Drive · /DCIM',
    bgStyle: { background: 'linear-gradient(180deg, #F97316 0%, #FDBA74 100%)' },
    sunStyle: { background: '#FEF08A' },
    waveBack: '#7C2D12',
    waveFront: '#431407'
  },
  {
    filename: 'family.jpg',
    description: 'ครอบครัวนั่งเล่นบนชายหาดพร้อมร่มสีสันสดใส บรรยากาศอบอุ่น',
    tags: ['ชายหาด', 'ครอบครัว', 'พักผ่อน'],
    path: '~/Pictures/PhotoBackup/family.jpg',
    size: '2.8 MB',
    backupDate: '3 ต.ค. 2569 07:23',
    source: 'Flash Drive · /DCIM',
    bgStyle: { background: 'linear-gradient(180deg, #BFDBFE 0%, #E0E7FF 100%)' },
    sunStyle: { background: '#FFFFFF' },
    waveBack: '#CBD5E1',
    waveFront: '#94A3B8'
  },
  {
    filename: 'IMG_2041.jpg',
    description: 'เงาของต้นมะพร้าวทอดยาวบนผืนทรายตอนใกล้ค่ำ แดดร่มลมตก',
    tags: ['ชายหาด', 'ต้นไม้', 'ธรรมชาติ'],
    path: '~/Pictures/PhotoBackup/IMG_2041.jpg',
    size: '3.1 MB',
    backupDate: '3 ต.ค. 2569 07:24',
    source: 'Flash Drive · /DCIM',
    bgStyle: { background: 'linear-gradient(180deg, #FDBA74 0%, #FED7AA 100%)' },
    sunStyle: { background: '#FEF08A' },
    waveBack: '#C2410C',
    waveFront: '#9A3412'
  },
  {
    filename: 'mountain.jpg',
    description: 'ทุ่งหญ้าเขียวขจีริมเชิงเขา อากาศบริสุทธิ์ยามสาย',
    tags: ['ภูเขา', 'ต้นไม้', 'เดินเล่น'],
    path: '~/Pictures/PhotoBackup/mountain.jpg',
    size: '3.9 MB',
    backupDate: '3 ต.ค. 2569 07:25',
    source: 'Flash Drive · /DCIM',
    bgStyle: { background: 'linear-gradient(180deg, #A7F3D0 0%, #D1FAE5 100%)' },
    sunStyle: { background: '#FDE047' },
    waveBack: '#65A30D',
    waveFront: '#3F6212'
  }
])

const currentPhoto = computed<GalleryPhoto | undefined>(() => {
  return galleryStrip.value[currentPhotoIndex.value] ?? galleryStrip.value[0]
})

function selectPhoto(index: number) {
  currentPhotoIndex.value = index
}

function removeTag(idx: number) {
  if (!currentPhoto.value) return
  currentPhoto.value.tags.splice(idx, 1)
}

function promptAddTag() {
  if (!currentPhoto.value) return
  const newTag = prompt('กรุณาระบุ Tag ใหม่:')
  if (newTag && newTag.trim()) {
    currentPhoto.value.tags.push(newTag.trim())
  }
}

const isAnalyzingLive = ref(false)

async function handleRunAiAnalysis() {
  if (!currentPhoto.value) return
  isAnalyzingLive.value = true
  try {
    let result = null
    const path = currentPhoto.value.path || currentPhoto.value.filename
    if (typeof (window as any)?.go?.main?.App?.AnalyzePhoto === 'function') {
      result = await (window as any).go.main.App.AnalyzePhoto(path)
    } else if (typeof (window as any)?.go?.service?.BackupService?.AnalyzePhoto === 'function') {
      result = await (window as any).go.service.BackupService.AnalyzePhoto(path)
    }
    if (result) {
      if (result.description) currentPhoto.value.description = result.description
      if (result.tags && result.tags.length > 0) currentPhoto.value.tags = result.tags
      alert(`AI วิเคราะห์ภาพสำเร็จ!\n\nคำอธิบาย: ${result.description}\nแท็ก: ${result.tags.join(', ')}`)
    } else {
      alert('ไม่สามารถวิเคราะห์ภาพได้ หรือยังไม่ได้ตั้งค่า API Key')
    }
  } catch (err: any) {
    alert('เกิดข้อผิดพลาดในการวิเคราะห์ AI: ' + (err?.message || err))
  } finally {
    isAnalyzingLive.value = false
  }
}

function handleCancelAnalysis() {
  isAnalyzingLive.value = false
  alert('ยกเลิกการวิเคราะห์เรียบร้อยแล้ว')
}

function handleDelete() {
  if (!currentPhoto.value) return
  if (confirm(`คุณต้องการลบ ${currentPhoto.value.filename} จากโฟลเดอร์ปลายทางหรือไม่?`)) {
    alert(`จำลองการลบ ${currentPhoto.value.filename} สำเร็จ`)
  }
}
</script>

<style scoped>
.photo-detail-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 24px 28px 20px 28px;
  box-sizing: border-box;
  background-color: #F4F5FA;
  overflow-y: auto;
  user-select: none;
  font-family: 'Prompt', 'Inter', sans-serif;
}

/* Mock Notice Banner */
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

/* Header */
.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}

.photo-filename {
  font-size: 26px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px 0;
  font-family: 'Inter', 'Prompt', sans-serif;
}

.header-subtitle {
  font-size: 13.5px;
  color: #6B7280;
  margin: 0;
}

.btn-back {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  color: #111827;
  font-family: inherit;
  font-size: 13.5px;
  font-weight: 600;
  padding: 8px 18px;
  border-radius: 10px;
  text-decoration: none;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
  transition: all 0.15s ease;
}

.btn-back:hover {
  background: #F9FAFB;
  border-color: #D1D5DB;
}

/* 2-Column Content Layout */
.detail-content-layout {
  display: flex;
  gap: 24px;
  align-items: flex-start;
}

/* Left Column: Preview + Thumbnails */
.preview-column {
  flex: 1.4;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.main-preview-card {
  position: relative;
  width: 100%;
  height: 380px;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.04);
}

.preview-sun {
  position: absolute;
  top: 60px;
  left: 36%;
  width: 72px;
  height: 72px;
  border-radius: 50%;
  box-shadow: 0 0 24px rgba(255, 255, 255, 0.6);
}

.preview-wave {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 150px;
}

.preview-wave-front {
  height: 105px;
}

/* Thumbnails Strip */
.thumbnails-strip {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 4px;
}

.thumb-strip-item {
  width: 98px;
  height: 66px;
  border-radius: 10px;
  border: 1.5px solid #E5E7EB;
  overflow: hidden;
  cursor: pointer;
  box-sizing: border-box;
  transition: all 0.15s ease;
  background: #FFFFFF;
}

.thumb-strip-item.active {
  border: 2px solid #4F46E5;
  box-shadow: 0 0 0 2px rgba(79, 70, 229, 0.2);
}

.thumb-mini-art {
  position: relative;
  width: 100%;
  height: 100%;
}

.thumb-mini-sun {
  position: absolute;
  top: 14px;
  left: 38%;
  width: 16px;
  height: 16px;
  border-radius: 50%;
}

.thumb-mini-wave {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 28px;
}

/* Right Column: Cards */
.sidebar-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.status-badge-row {
  display: flex;
  margin-bottom: -4px;
}

.badge-backed-up {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #DCFCE7;
  color: #15803D;
  font-size: 12px;
  font-weight: 600;
  padding: 5px 12px;
  border-radius: 9999px;
}

.info-card {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 14px;
  padding: 18px 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
}

/* AI Desc Card */
.ai-desc-card {
  background: #EEF2FF;
  border-color: #C7D2FE;
}

.card-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.card-title-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 700;
  color: #4F46E5;
}

.card-title-badge .text-dark {
  color: #111827;
}

.card-title-heading {
  font-size: 14.5px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 14px 0;
}

.header-action-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-ai-live {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: linear-gradient(135deg, #6366F1 0%, #4F46E5 100%);
  border: none;
  color: #FFFFFF;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 5px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(79, 70, 229, 0.25);
}

.btn-ai-live:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-1px);
}

.btn-ai-live:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-edit {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  color: #111827;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 5px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-edit:hover {
  background: #F8FAFC;
}

.ai-desc-text {
  font-size: 13px;
  line-height: 1.6;
  color: #312E81;
  font-weight: 400;
}

.desc-textarea {
  width: 100%;
  border: 1px solid #A5B4FC;
  border-radius: 8px;
  padding: 8px 10px;
  font-family: inherit;
  font-size: 13px;
  color: #1E1B4B;
  box-sizing: border-box;
  outline: none;
  resize: vertical;
}

/* Tags Card */
.tags-chips-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.tag-pill-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #F1F5F9;
  border: 1px solid #E2E8F0;
  color: #334155;
  font-size: 12px;
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 9999px;
}

.btn-remove-tag {
  background: transparent;
  border: none;
  color: #64748B;
  font-size: 11px;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
}

.btn-remove-tag:hover {
  color: #DC2626;
}

.btn-add-tag {
  background: transparent;
  border: 1px dashed #CBD5E1;
  color: #64748B;
  font-family: inherit;
  font-size: 12px;
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 9999px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-add-tag:hover {
  background: #F8FAFC;
  border-color: #94A3B8;
  color: #1E293B;
}

/* File Metadata */
.meta-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.meta-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12.5px;
}

.meta-label {
  color: #6B7280;
  font-weight: 500;
}

.meta-value {
  color: #111827;
  font-weight: 600;
}

.meta-value.path-val {
  font-family: monospace;
  font-size: 11.5px;
  color: #374151;
  max-width: 220px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Delete Action Row */
.delete-action-row {
  display: flex;
}

.btn-delete-dest {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #FFFFFF;
  border: 1px solid #FECACA;
  color: #DC2626;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  padding: 10px 18px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-delete-dest:hover {
  background: #FEF2F2;
  border-color: #F87171;
}

/* Cancel AI Analysis Button */
.btn-cancel-ai {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: #FFFFFF;
  border: 1px solid #FECACA;
  color: #DC2626;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 5px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-cancel-ai:hover {
  background: #FEF2F2;
  border-color: #F87171;
}
</style>
