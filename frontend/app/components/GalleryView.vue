<template>
  <div class="gallery-view">
    <!-- Notice Banner that this is a Mockup of a friend's part -->
    <div class="mock-notice">
      <div class="notice-badge">UI Mockup</div>
      <span class="notice-text">
        หน้านี้เป็นภาพตัวอย่าง (ส่วนของเพื่อน): ระบบวิเคราะห์คำอธิบาย AI, แท็ก และ ค้นหาคลังภาพ
      </span>
    </div>

    <!-- Header -->
    <header class="view-header">
      <div class="header-text">
        <h1 class="view-title">แกลเลอรี่ & ค้นหา</h1>
        <p class="view-subtitle">
          {{ activeTab === 'desc' ? 'ค้นหาภาพจากคำอธิบายที่ AI สร้างให้ หรือจาก Tag' : 'คลิก Tag ใน Tag Cloud เพื่อกรองรูปภาพ' }}
        </p>
      </div>

      <div class="header-count-pill">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="3" width="18" height="18" rx="2"/>
          <circle cx="8.5" cy="8.5" r="1.5"/>
          <polyline points="21 15 16 10 5 21"/>
        </svg>
        <span>รูปทั้งหมด 342</span>
      </div>
    </header>

    <!-- Search Controls Bar -->
    <div class="search-panel-card">
      <div class="search-nav-row">
        <!-- Switch between Search by Description vs Search by Tag -->
        <div class="search-type-tabs">
          <button 
            class="tab-btn" 
            :class="{ active: activeTab === 'desc' }"
            @click="activeTab = 'desc'"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="8" y1="6" x2="21" y2="6"/>
              <line x1="8" y1="12" x2="21" y2="12"/>
              <line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/>
              <line x1="3" y1="12" x2="3.01" y2="12"/>
              <line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
            ค้นหาด้วยคำอธิบาย
          </button>
          <button 
            class="tab-btn" 
            :class="{ active: activeTab === 'tag' }"
            @click="activeTab = 'tag'"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
              <line x1="7" y1="7" x2="7.01" y2="7"/>
            </svg>
            ค้นหาด้วย Tag
          </button>
        </div>

        <!-- Tag mode active tag badge -->
        <div v-if="activeTab === 'tag'" class="selected-tag-info">
          <span>Tag ที่เลือก:</span>
          <span class="active-tag-chip">
            {{ selectedTag }}
            <button class="btn-clear-tag" @click="selectedTag = ''">✕</button>
          </span>
          <button class="btn-reset-filter" @click="selectedTag = ''">ล้างทั้งหมด</button>
        </div>
      </div>

      <!-- Description Search Input Box (Figma Screen 2) -->
      <div v-if="activeTab === 'desc'" class="search-input-box">
        <div class="search-input-wrapper">
          <svg class="search-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#6B7280" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input 
            type="text" 
            v-model="searchQuery" 
            placeholder="ค้นหาภาพ เช่น ชายหาด พระอาทิตย์ตก" 
            class="search-input"
          />
          <button v-if="searchQuery" class="btn-clear" @click="searchQuery = ''">✕</button>
        </div>
        <button class="btn-submit-search">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          ค้นหา
        </button>
      </div>

      <!-- Quick Suggestion chips -->
      <div v-if="activeTab === 'desc'" class="suggestions-row">
        <span class="suggest-label">ลองค้นหา:</span>
        <button 
          v-for="s in suggestions" 
          :key="s" 
          class="suggest-chip"
          @click="searchQuery = s"
        >
          {{ s }}
        </button>
      </div>
    </div>

    <!-- Main Content Grid -->
    <div class="gallery-content-layout" :class="{ 'with-sidebar': activeTab === 'tag' }">
      <!-- Tag Cloud Sidebar Card (Figma Screen 3) -->
      <div v-if="activeTab === 'tag'" class="tag-cloud-card">
        <div class="tag-cloud-header">
          <h3 class="tag-cloud-title">Tag Cloud</h3>
          <span class="tag-cloud-count">🏷️ 64 tags</span>
        </div>

        <div class="cloud-words">
          <span 
            v-for="item in tagCloudList" 
            :key="item.name"
            class="cloud-word"
            :class="{ selected: selectedTag === item.name }"
            :style="{ fontSize: item.size + 'px', color: item.color }"
            @click="selectedTag = item.name"
          >
            {{ item.name }}
          </span>
        </div>

        <div class="cloud-footnote">
          ขนาดตัวอักษร = จำนวนรูปที่มี Tag นั้น
        </div>
      </div>

      <!-- Photo Cards Results Grid -->
      <div class="results-container">
        <div class="results-header">
          <h3 class="results-count">
            พบ <strong>{{ activeTab === 'tag' ? '36 ภาพ' : '12 ภาพ' }}</strong>
            <span v-if="activeTab === 'desc' && searchQuery"> สำหรับ "{{ searchQuery }}"</span>
            <span v-else-if="activeTab === 'tag' && selectedTag"> ที่มี Tag "{{ selectedTag }}"</span>
          </h3>

          <div class="sort-selector" v-if="activeTab === 'desc'">
            <span>เรียงตาม: ความเกี่ยวข้อง</span>
          </div>
          <div class="sort-selector" v-else>
            <span>แสดง 6 จาก 36</span>
          </div>
        </div>

        <div class="gallery-cards-grid" :class="{ compact: activeTab === 'tag' }">
          <div 
            v-for="(card, i) in currentPhotoResults" 
            :key="i" 
            class="gallery-card"
          >
            <!-- Thumbnail Graphic matching Figma -->
            <div class="card-thumb" :style="getGalleryArtBg(i)">
              <div class="card-sun" :style="getGallerySun(i)"></div>
              <svg class="card-wave" viewBox="0 0 200 80" preserveAspectRatio="none">
                <path d="M0,40 C60,15 130,65 200,30 L200,80 L0,80 Z" :fill="getGalleryWave(i)"/>
              </svg>
            </div>

            <!-- Card Content (AI description + tags) -->
            <div class="card-body">
              <p class="ai-desc" v-if="activeTab === 'desc'">
                {{ card.desc }}
              </p>
              <div class="card-tags-row">
                <span 
                  v-for="tag in card.tags" 
                  :key="tag" 
                  class="tag-item"
                  :class="{ highlight: selectedTag === tag }"
                  @click="activeTab = 'tag'; selectedTag = tag"
                >
                  {{ tag }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="load-more-row" v-if="activeTab === 'tag'">
          <button class="btn-load-more">โหลดเพิ่ม</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const activeTab = ref<'desc' | 'tag'>('desc')
const searchQuery = ref('ชายหาด พระอาทิตย์ตก')
const selectedTag = ref('ชายหาด')

const suggestions = ['แมวนอนบนโซฟา', 'ภูเขาหิมะ', 'อาหารญี่ปุ่น', 'ครอบครัวปิกนิก']

const tagCloudList = [
  { name: 'ทะเล', size: 28, color: '#0284C7' },
  { name: 'ชายหาด', size: 26, color: '#4338CA' },
  { name: 'ภูเขา', size: 24, color: '#059669' },
  { name: 'พระอาทิตย์ตก', size: 20, color: '#EA580C' },
  { name: 'ต้นไม้', size: 19, color: '#16A34A' },
  { name: 'ท้องฟ้า', size: 18, color: '#6366F1' },
  { name: 'อาหาร', size: 22, color: '#B45309' },
  { name: 'แมว', size: 16, color: '#E11D48' },
  { name: 'ครอบครัว', size: 16, color: '#0D9488' },
  { name: 'เมือง', size: 14, color: '#64748B' },
  { name: 'กลางคืน', size: 14, color: '#6366F1' },
  { name: 'ดอกไม้', size: 15, color: '#EC4899' },
  { name: 'หิมะ', size: 13, color: '#0284C7' },
  { name: 'เดินป่า', size: 13, color: '#15803D' },
  { name: 'รถยนต์', size: 13, color: '#64748B' },
  { name: 'งานเลี้ยง', size: 14, color: '#7C3AED' },
  { name: 'เด็ก', size: 14, color: '#F59E0B' },
  { name: 'ท่าเรือ', size: 13, color: '#0284C7' }
]

const galleryMockData = [
  {
    desc: 'หาดทรายขาวกับท้องฟ้าสีส้มยามพระอาทิตย์ตก มีเรือเล็กลอยอยู่ไกล ๆ',
    tags: ['ชายหาด', 'พระอาทิตย์ตก'],
    art: { bg: 'linear-gradient(180deg, #FB923C 0%, #FED7AA 100%)', sun: '#FEF08A', wave: '#9A3412' }
  },
  {
    desc: 'คลื่นซัดฝั่งตอนเย็น ท้องฟ้าเป็นสีชมพูอมม่วง',
    tags: ['ทะเล', 'ท้องฟ้า'],
    art: { bg: 'linear-gradient(180deg, #38BDF8 0%, #BAE6FD 100%)', sun: '#FFFFFF', wave: '#0284C7' }
  },
  {
    desc: 'ครอบครัวนั่งเล่นบนชายหาดพร้อมร่มสีสันสดใส',
    tags: ['ครอบครัว', 'ชายหาด'],
    art: { bg: 'linear-gradient(180deg, #93C5FD 0%, #DBEAFE 100%)', sun: '#FFFFFF', wave: '#64748B' }
  },
  {
    desc: 'เงาของต้นมะพร้าวทอดยาวบนผืนทรายตอนใกล้ค่ำ',
    tags: ['ต้นไม้', 'พระอาทิตย์ตก'],
    art: { bg: 'linear-gradient(180deg, #F97316 0%, #FDBA74 100%)', sun: '#FEF08A', wave: '#C2410C' }
  },
  {
    desc: 'ทะเลสีฟ้าใสและท้องฟ้าโปร่ง เหมาะแก่การพักผ่อน',
    tags: ['ทะเล', 'ท้องฟ้า'],
    art: { bg: 'linear-gradient(180deg, #38BDF8 0%, #E0F2FE 100%)', sun: '#FFFFFF', wave: '#0284C7' }
  },
  {
    desc: 'คนเดินเล่นริมหาดยามเช้า แสงแดดอ่อน ๆ',
    tags: ['ชายหาด', 'เดินเล่น'],
    art: { bg: 'linear-gradient(180deg, #6EE7B7 0%, #A7F3D0 100%)', sun: '#FDE047', wave: '#059669' }
  },
  {
    desc: 'ภูเขาสีม่วงตัดกับท้องฟ้าสีส้มตอนพระอาทิตย์ตก',
    tags: ['ภูเขา', 'พระอาทิตย์ตก'],
    art: { bg: 'linear-gradient(180deg, #818CF8 0%, #C7D2FE 100%)', sun: '#FEF08A', wave: '#4338CA' }
  },
  {
    desc: 'ท่าเรือไม้ยื่นลงทะเลช่วงเย็น',
    tags: ['ทะเล', 'ท่าเรือ'],
    art: { bg: 'linear-gradient(180deg, #38BDF8 0%, #BAE6FD 100%)', sun: '#FFFFFF', wave: '#0369A1' }
  }
]

const currentPhotoResults = computed(() => {
  if (activeTab.value === 'tag') {
    return galleryMockData.slice(0, 6)
  }
  return galleryMockData
})

function getGalleryArtBg(i: number) {
  return { background: galleryMockData[i % galleryMockData.length].art.bg }
}

function getGallerySun(i: number) {
  return { background: galleryMockData[i % galleryMockData.length].art.sun }
}

function getGalleryWave(i: number) {
  return galleryMockData[i % galleryMockData.length].art.wave
}
</script>

<style scoped>
.gallery-view {
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
.view-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
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

.header-count-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 12px;
  font-weight: 600;
  padding: 6px 14px;
  border-radius: 9999px;
}

/* Search Panel Card */
.search-panel-card {
  background: #FFFFFF;
  border-radius: 14px;
  border: 1px solid #E5E7EB;
  padding: 14px 18px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}

.search-nav-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.search-type-tabs {
  display: flex;
  background: #F3F4F6;
  padding: 3px;
  border-radius: 10px;
  gap: 4px;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #4B5563;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.tab-btn.active {
  background: #FFFFFF;
  color: #111827;
  font-weight: 600;
  box-shadow: 0 1px 2px rgba(0,0,0,0.06);
}

.selected-tag-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  color: #4B5563;
}

.active-tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #4F46E5;
  color: #FFFFFF;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 9999px;
}

.btn-clear-tag {
  background: transparent;
  border: none;
  color: #FFFFFF;
  font-size: 10px;
  cursor: pointer;
  padding: 0;
}

.btn-reset-filter {
  background: transparent;
  border: none;
  color: #4F46E5;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
}

.btn-reset-filter:hover {
  text-decoration: underline;
}

/* Search Input Box */
.search-input-box {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}

.search-input-wrapper {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 14px;
}

.search-input {
  width: 100%;
  height: 42px;
  border: 1.5px solid #4F46E5;
  border-radius: 10px;
  padding: 0 38px 0 42px;
  font-size: 13.5px;
  font-family: inherit;
  outline: none;
  box-sizing: border-box;
}

.btn-clear {
  position: absolute;
  right: 12px;
  background: transparent;
  border: none;
  color: #9CA3AF;
  cursor: pointer;
  font-size: 14px;
}

.btn-submit-search {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #4F46E5;
  color: #FFFFFF;
  border: none;
  border-radius: 10px;
  padding: 0 20px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(79, 70, 229, 0.25);
}

.suggestions-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #6B7280;
}

.suggest-label {
  font-weight: 500;
}

.suggest-chip {
  background: #F3F4F6;
  border: 1px solid #E5E7EB;
  border-radius: 9999px;
  padding: 4px 12px;
  font-size: 11.5px;
  color: #4B5563;
  cursor: pointer;
  transition: all 0.15s;
}

.suggest-chip:hover {
  background: #E5E7EB;
  color: #111827;
}

/* Layout */
.gallery-content-layout {
  display: flex;
  gap: 16px;
  flex: 1;
}

.gallery-content-layout.with-sidebar {
  display: grid;
  grid-template-columns: 320px 1fr;
}

/* Tag Cloud Card */
.tag-cloud-card {
  background: #FFFFFF;
  border-radius: 14px;
  border: 1px solid #E5E7EB;
  padding: 18px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
  display: flex;
  flex-direction: column;
}

.tag-cloud-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.tag-cloud-title {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
  margin: 0;
}

.tag-cloud-count {
  font-size: 11.5px;
  font-weight: 600;
  color: #4F46E5;
  background: #EEF2FF;
  padding: 3px 8px;
  border-radius: 9999px;
}

.cloud-words {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 14px;
  align-items: center;
  flex: 1;
}

.cloud-word {
  cursor: pointer;
  font-weight: 700;
  transition: transform 0.15s;
  padding: 2px 6px;
  border-radius: 8px;
}

.cloud-word:hover {
  transform: scale(1.1);
}

.cloud-word.selected {
  background: #4F46E5 !important;
  color: #FFFFFF !important;
  border-radius: 9999px;
  padding: 4px 14px;
}

.cloud-footnote {
  margin-top: 16px;
  font-size: 11px;
  color: #94A3B8;
  border-top: 1px solid #F3F4F6;
  padding-top: 10px;
}

/* Results Grid */
.results-container {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.results-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.results-count {
  font-size: 14.5px;
  font-weight: 600;
  color: #1F2937;
  margin: 0;
}

.sort-selector {
  font-size: 12px;
  color: #6B7280;
}

.gallery-cards-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.gallery-cards-grid.compact {
  grid-template-columns: repeat(3, 1fr);
}

.gallery-card {
  background: #FFFFFF;
  border-radius: 12px;
  border: 1px solid #E5E7EB;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
  transition: all 0.15s ease;
}

.gallery-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0,0,0,0.06);
}

.card-thumb {
  height: 105px;
  position: relative;
  overflow: hidden;
}

.card-sun {
  width: 24px;
  height: 16px;
  border-radius: 50%;
  position: absolute;
  top: 18px;
  right: 36px;
  opacity: 0.95;
}

.card-wave {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 52px;
}

.card-body {
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ai-desc {
  font-size: 11.5px;
  color: #1F2937;
  margin: 0;
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-tags-row {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.tag-item {
  font-size: 10.5px;
  font-weight: 500;
  background: #EEF2FF;
  color: #4F46E5;
  padding: 2px 8px;
  border-radius: 9999px;
  cursor: pointer;
}

.tag-item.highlight {
  background: #4F46E5;
  color: #FFFFFF;
}

.load-more-row {
  margin-top: 16px;
}

.btn-load-more {
  background: #FFFFFF;
  border: 1px solid #D1D5DB;
  border-radius: 8px;
  padding: 8px 16px;
  font-size: 12px;
  font-weight: 500;
  color: #374151;
  cursor: pointer;
}

.btn-load-more:hover {
  background: #F9FAFB;
}
</style>
