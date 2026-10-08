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
    <div class="search-panel-card" :class="{ 'compact-mode': activeTab === 'tag' }">
      <div class="search-nav-row">
        <!-- Switch between Search by Description vs Search by Tag -->
        <div class="search-type-tabs">
          <button 
            class="tab-btn" 
            :class="{ active: activeTab === 'desc' }"
            @click="switchTab('desc')"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" :stroke="activeTab === 'desc' ? '#4F46E5' : '#6B7280'" stroke-width="2">
              <line x1="8" y1="6" x2="21" y2="6"/>
              <line x1="8" y1="12" x2="21" y2="12"/>
              <line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/>
              <line x1="3" y1="12" x2="3.01" y2="12"/>
              <line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
            <span>ค้นหาด้วยคำอธิบาย</span>
          </button>
          
          <button 
            class="tab-btn" 
            :class="{ active: activeTab === 'tag' }"
            @click="switchTab('tag')"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" :stroke="activeTab === 'tag' ? '#4F46E5' : '#6B7280'" stroke-width="2">
              <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
              <line x1="7" y1="7" x2="7.01" y2="7"/>
            </svg>
            <span>ค้นหาด้วย Tag</span>
          </button>
        </div>

        <!-- Tag mode active tag badge (middle) & clear all (right) -->
        <template v-if="activeTab === 'tag'">
          <div class="selected-tag-info" v-if="selectedTag">
            <span class="label-selected">Tag ที่เลือก:</span>
            <span class="active-tag-chip">
              <span>{{ selectedTag }}</span>
              <button class="btn-clear-tag" @click="selectedTag = ''" title="ยกเลิกการเลือก">✕</button>
            </span>
          </div>

          <button class="btn-reset-filter" @click="selectedTag = ''">
            ล้างทั้งหมด
          </button>
        </template>
      </div>

      <!-- Description Search Input Box (Figma Screen 2) -->
      <div v-if="activeTab === 'desc'" class="search-input-box">
        <div class="search-input-wrapper">
          <svg class="search-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input 
            type="text" 
            v-model="searchQuery" 
            placeholder="ชายหาด พระอาทิตย์ตก" 
            class="search-input"
          />
          <button v-if="searchQuery" class="btn-clear" @click="searchQuery = ''">✕</button>
        </div>
        <button class="btn-submit-search">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <span>ค้นหา</span>
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

    <!-- Main Content Layout -->
    <div class="gallery-content-layout" :class="{ 'with-sidebar': activeTab === 'tag' }">
      <!-- Tag Cloud Sidebar Card (Figma Screen 2 Search-by-tag) -->
      <aside v-if="activeTab === 'tag'" class="tag-cloud-card">
        <div class="tag-cloud-header">
          <h3 class="tag-cloud-title">Tag Cloud</h3>
          <span class="tag-cloud-count">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2">
              <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
              <line x1="7" y1="7" x2="7.01" y2="7"/>
            </svg>
            64 tags
          </span>
        </div>

        <div class="cloud-words">
          <span 
            v-for="item in tagCloudList" 
            :key="item.name"
            class="cloud-word"
            :class="{ selected: selectedTag === item.name }"
            :style="selectedTag === item.name ? {} : { fontSize: item.size + 'px', color: item.color }"
            @click="selectedTag = item.name"
          >
            {{ item.name }}
          </span>
        </div>

        <div class="cloud-footnote">
          ขนาดตัวอักษร = จำนวนรูปที่มี Tag นั้น
        </div>
      </aside>

      <!-- Photo Cards Results Container -->
      <section class="results-container">
        <!-- Results count & Sorting bar -->
        <div class="results-header">
          <h3 class="results-count">
            <template v-if="activeTab === 'desc'">
              พบ <strong>12 ภาพ</strong> สำหรับ "{{ searchQuery }}"
            </template>
            <template v-else>
              พบ <strong>36 ภาพ</strong> ที่มี Tag "{{ selectedTag || 'ทั้งหมด' }}"
            </template>
          </h3>

          <div class="sort-selector" v-if="activeTab === 'desc'">
            <span>เรียงตาม: ความเกี่ยวข้อง</span>
          </div>
          <div class="sort-selector" v-else>
            <span>แสดง 6 จาก 36</span>
          </div>
        </div>

        <!-- 8 Cards for Description mode OR 6 Cards for Tag mode -->
        <div class="gallery-cards-grid" :class="{ 'grid-tag-mode': activeTab === 'tag' }">
          <div 
            v-for="(card, i) in displayedPhotos" 
            :key="i" 
            class="gallery-card"
          >
            <!-- Thumbnail Graphic matching exact Figma wave colors -->
            <NuxtLink :to="'/gallery/' + encodeURIComponent(card.filename || 'beach_trip')" class="card-thumb-link">
              <div class="card-thumb" :style="card.thumbBg">
                <div class="card-sun" :style="card.sunStyle"></div>
                <svg class="card-wave" viewBox="0 0 240 90" preserveAspectRatio="none">
                  <path d="M0,45 C75,20 150,68 240,35 L240,90 L0,90 Z" :fill="card.waveColor" opacity="0.45"/>
                  <path d="M0,60 C80,30 160,75 240,48 L240,90 L0,90 Z" :fill="card.waveColor"/>
                </svg>
              </div>
            </NuxtLink>

            <!-- Card Body -->
            <div class="card-body">
              <!-- AI description only in Description mode -->
              <NuxtLink :to="'/gallery/' + encodeURIComponent(card.filename || 'beach_trip')" class="card-desc-link">
                <p class="ai-desc" v-if="activeTab === 'desc'">
                  {{ card.desc }}
                </p>
              </NuxtLink>
              
              <!-- Tag chips -->
              <div class="card-tags-row">
                <span 
                  v-for="t in card.tags" 
                  :key="t" 
                  class="tag-item"
                  :class="{ highlight: selectedTag === t }"
                  @click.stop="activeTab = 'tag'; selectedTag = t"
                >
                  {{ t }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Load more button in Tag mode -->
        <div class="load-more-row" v-if="activeTab === 'tag'">
          <button class="btn-load-more">โหลดเพิ่ม</button>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const route = useRoute()

// Default to 'tag' mode matching Screen 2 Search by Tag provided by user
const activeTab = ref<'desc' | 'tag'>('tag')
const searchQuery = ref('ชายหาด พระอาทิตย์ตก')
const selectedTag = ref('ชายหาด')

function switchTab(tab: 'desc' | 'tag') {
  activeTab.value = tab
}

onMounted(() => {
  if (route.query.tab === 'desc') {
    activeTab.value = 'desc'
  } else if (route.query.tab === 'tag') {
    activeTab.value = 'tag'
  }
  if (route.query.tag && typeof route.query.tag === 'string') {
    selectedTag.value = route.query.tag
  }
})

const suggestions = ['แมวนอนบนโซฟา', 'ภูเขาหิมะ', 'อาหารญี่ปุ่น', 'ครอบครัวปิกนิก']

const tagCloudList = [
  { name: 'ทะเล', size: 32, color: '#0284C7' },
  { name: 'ชายหาด', size: 28, color: '#312E81' },
  { name: 'ภูเขา', size: 26, color: '#059669' },
  { name: 'พระอาทิตย์ตก', size: 21, color: '#EA580C' },
  { name: 'ต้นไม้', size: 21, color: '#16A34A' },
  { name: 'ท้องฟ้า', size: 18, color: '#6366F1' },
  { name: 'อาหาร', size: 22, color: '#B45309' },
  { name: 'แมว', size: 16, color: '#E11D48' },
  { name: 'ครอบครัว', size: 16, color: '#0D9488' },
  { name: 'เมือง', size: 16, color: '#64748B' },
  { name: 'กลางคืน', size: 15, color: '#6366F1' },
  { name: 'ดอกไม้', size: 15, color: '#EC4899' },
  { name: 'หิมะ', size: 14, color: '#0284C7' },
  { name: 'เดินป่า', size: 14, color: '#15803D' },
  { name: 'รถยนต์', size: 14, color: '#64748B' },
  { name: 'งานเลี้ยง', size: 15, color: '#7C3AED' },
  { name: 'เด็ก', size: 15, color: '#EA580C' },
  { name: 'ท่าเรือ', size: 14, color: '#0284C7' }
]

// 8 Cards data for Description search & 6 cards data for Tag search
const allCards = [
  {
    filename: 'beach_trip.png',
    desc: 'หาดทรายขาวกับท้องฟ้าสีส้มยามพระอาทิตย์ตก มีเรือเล็กลอยอยู่ไกล ๆ',
    tags: ['ชายหาด', 'พระอาทิตย์ตก'],
    thumbBg: { background: 'linear-gradient(180deg, #F97316 0%, #FDBA74 100%)' },
    sunStyle: { background: '#FEF08A' },
    waveColor: '#7C2D12'
  },
  {
    filename: 'sunset.jpg',
    desc: 'คลื่นซัดฝั่งตอนเย็น ท้องฟ้าเป็นสีชมพูอมม่วง',
    tags: ['ชายหาด', 'ทะเล'],
    thumbBg: { background: 'linear-gradient(180deg, #38BDF8 0%, #BAE6FD 100%)' },
    sunStyle: { background: '#FFFFFF' },
    waveColor: '#0284C7'
  },
  {
    filename: 'family.jpg',
    desc: 'ครอบครัวนั่งเล่นบนชายหาดพร้อมร่มสีสันสดใส',
    tags: ['ชายหาด', 'ครอบครัว'],
    thumbBg: { background: 'linear-gradient(180deg, #BFDBFE 0%, #E0E7FF 100%)' },
    sunStyle: { background: '#FFFFFF' },
    waveColor: '#94A3B8'
  },
  {
    filename: 'IMG_2041.jpg',
    desc: 'เงาของต้นมะพร้าวทอดยาวบนผืนทรายตอนใกล้ค่ำ',
    tags: ['ชายหาด', 'ต้นไม้'],
    thumbBg: { background: 'linear-gradient(180deg, #FDBA74 0%, #FED7AA 100%)' },
    sunStyle: { background: '#FEF08A' },
    waveColor: '#C2410C'
  },
  {
    filename: 'sea_view.png',
    desc: 'ทะเลสีฟ้าใสและท้องฟ้าโปร่ง เหมาะแก่การพักผ่อน',
    tags: ['ชายหาด', 'ท้องฟ้า'],
    thumbBg: { background: 'linear-gradient(180deg, #38BDF8 0%, #BAE6FD 100%)' },
    sunStyle: { background: '#FFFFFF' },
    waveColor: '#0284C7'
  },
  {
    filename: 'mountain.jpg',
    desc: 'คนเดินเล่นริมหาดยามเช้า แสงแดดอ่อน ๆ',
    tags: ['ชายหาด', 'เดินเล่น'],
    thumbBg: { background: 'linear-gradient(180deg, #A7F3D0 0%, #D1FAE5 100%)' },
    sunStyle: { background: '#FACC15' },
    waveColor: '#4D7C0F'
  },
  {
    filename: 'hike_sunset.jpg',
    desc: 'ภูเขาสีม่วงตัดกับท้องฟ้าสีส้มตอนพระอาทิตย์ตก',
    tags: ['ภูเขา', 'พระอาทิตย์ตก'],
    thumbBg: { background: 'linear-gradient(180deg, #818CF8 0%, #C7D2FE 100%)' },
    sunStyle: { background: '#FEF08A' },
    waveColor: '#4338CA'
  },
  {
    filename: 'harbor.png',
    desc: 'ท่าเรือไม้ยื่นลงทะเลช่วงเย็น',
    tags: ['ทะเล', 'ท่าเรือ'],
    thumbBg: { background: 'linear-gradient(180deg, #38BDF8 0%, #BAE6FD 100%)' },
    sunStyle: { background: '#FFFFFF' },
    waveColor: '#0369A1'
  }
]

const displayedPhotos = computed(() => {
  if (activeTab.value === 'tag') {
    // Show 6 cards in Tag mode matching Screen 2 Tag Search
    return allCards.slice(0, 6)
  }
  // Show 8 cards in Description search mode
  return allCards
})
</script>

<style scoped>
.gallery-view {
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
.view-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}

.view-title {
  font-size: 26px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px 0;
}

.view-subtitle {
  font-size: 13.5px;
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

.search-panel-card.compact-mode {
  padding: 12px 18px;
  min-height: 48px;
}

.search-nav-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.search-panel-card:not(.compact-mode) .search-nav-row {
  margin-bottom: 14px;
}

.search-type-tabs {
  display: flex;
  background: #F1F5F9;
  padding: 3px;
  border-radius: 10px;
  gap: 4px;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 7px 16px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #6B7280;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease-in-out;
}

.tab-btn.active {
  background: #FFFFFF;
  color: #111827;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}

.selected-tag-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.label-selected {
  font-size: 13px;
  color: #6B7280;
}

.active-tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #4F46E5;
  color: #FFFFFF;
  font-size: 13px;
  font-weight: 600;
  padding: 5px 12px;
  border-radius: 9999px;
}

.btn-clear-tag {
  background: transparent;
  border: none;
  color: #FFFFFF;
  font-size: 11px;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
}

.btn-reset-filter {
  background: transparent;
  border: none;
  color: #4F46E5;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.btn-reset-filter:hover {
  text-decoration: underline;
}

/* Description Search Input */
.search-input-box {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}

.search-input-wrapper {
  flex: 1;
  display: flex;
  align-items: center;
  background: #F8FAFC;
  border: 1.5px solid #4F46E5;
  border-radius: 12px;
  padding: 0 14px;
  height: 46px;
  gap: 10px;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-family: inherit;
  font-size: 14px;
  color: #111827;
  font-weight: 500;
}

.btn-clear {
  background: transparent;
  border: none;
  color: #6B7280;
  font-size: 14px;
  cursor: pointer;
}

.btn-submit-search {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #4F46E5;
  color: #FFFFFF;
  border: none;
  border-radius: 12px;
  padding: 0 24px;
  font-family: inherit;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.15s;
}

.btn-submit-search:hover {
  background: #4338CA;
}

/* Suggestions chips */
.suggestions-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.suggest-label {
  font-size: 12.5px;
  color: #6B7280;
}

.suggest-chip {
  background: #F1F5F9;
  border: none;
  border-radius: 9999px;
  padding: 4px 12px;
  font-family: inherit;
  font-size: 12px;
  color: #475569;
  cursor: pointer;
  transition: all 0.15s;
}

.suggest-chip:hover {
  background: #E2E8F0;
  color: #1E293B;
}

/* Layout: Full width or With Tag Cloud Sidebar */
.gallery-content-layout {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

/* Tag Cloud Sidebar */
.tag-cloud-card {
  width: 320px;
  flex-shrink: 0;
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 14px;
  padding: 18px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}

.tag-cloud-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.tag-cloud-title {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
  margin: 0;
}

.tag-cloud-count {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 9999px;
}

.cloud-words {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px 14px;
  min-height: 220px;
  padding-bottom: 14px;
  border-bottom: 1px solid #F3F4F6;
}

.cloud-word {
  cursor: pointer;
  font-weight: 600;
  transition: transform 0.15s;
  line-height: 1.2;
}

.cloud-word:hover {
  transform: scale(1.08);
}

.cloud-word.selected {
  background: #312E81 !important;
  color: #FFFFFF !important;
  font-size: 26px !important;
  font-weight: 700 !important;
  padding: 6px 22px;
  border-radius: 9999px;
  box-shadow: 0 4px 12px rgba(49, 46, 129, 0.28);
}

.cloud-footnote {
  margin-top: 14px;
  font-size: 11.5px;
  color: #94A3B8;
  text-align: left;
}

/* Results Content */
.results-container {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.results-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.results-count {
  font-size: 14.5px;
  font-weight: 500;
  color: #111827;
  margin: 0;
}

.results-count strong {
  font-weight: 700;
}

.sort-selector {
  font-size: 12.5px;
  color: #6B7280;
}

/* Photo Cards Grid */
.gallery-cards-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.gallery-cards-grid.grid-tag-mode {
  grid-template-columns: repeat(3, 1fr);
  gap: 18px;
}

.gallery-card {
  background: #FFFFFF;
  border-radius: 14px;
  border: 1px solid #E5E7EB;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: transform 0.15s, box-shadow 0.15s;
}

.gallery-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 14px rgba(0,0,0,0.05);
}

.card-thumb-link, .card-desc-link {
  text-decoration: none;
  color: inherit;
  display: block;
}

/* Thumbnail Visuals */
.card-thumb {
  position: relative;
  width: 100%;
  height: 128px;
  overflow: hidden;
}

.card-sun {
  position: absolute;
  top: 22px;
  left: 50%;
  transform: translateX(-50%);
  width: 30px;
  height: 30px;
  border-radius: 50%;
  box-shadow: 0 0 14px rgba(255, 255, 255, 0.45);
}

.card-wave {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 64px;
}

/* Card Body */
.card-body {
  padding: 12px 14px 14px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.ai-desc {
  font-size: 12.5px;
  color: #1F2937;
  line-height: 1.45;
  margin: 0;
  min-height: 36px;
}

.card-tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: auto;
}

.tag-item {
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 11px;
  font-weight: 500;
  padding: 3px 10px;
  border-radius: 9999px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.tag-item:hover, .tag-item.highlight {
  background: #E0E7FF;
  font-weight: 600;
}

/* Load more button */
.load-more-row {
  margin-top: 20px;
  display: flex;
}

.btn-load-more {
  background: #FFFFFF;
  border: 1px solid #D1D5DB;
  border-radius: 8px;
  padding: 8px 18px;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  color: #374151;
  cursor: pointer;
  transition: background-color 0.15s;
}

.btn-load-more:hover {
  background: #F9FAFB;
}
</style>
