<template>
  <div class="gallery-view">
    <!-- Notice Banner that this is a Mockup of a friend's part -->
    <div class="mock-notice">
      <div class="notice-badge">UI Mockup</div>
      <span class="notice-text">
        หน้านี้เป็นภาพตัวอย่าง (ส่วนของเพื่อน): ระบบวิเคราะห์คำอธิบาย AI และ ค้นหาคลังภาพจากคำอธิบาย
      </span>
    </div>

    <!-- Header -->
    <header class="view-header">
      <div class="header-text">
        <h1 class="view-title">แกลเลอรี่ & ค้นหา</h1>
        <p class="view-subtitle">
          ค้นหาภาพจากคำอธิบายที่ AI สร้างให้ หรือค้นหาจาก Tag
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
        <!-- Switch between Search by Description vs Search by Tag via NuxtLink -->
        <div class="search-type-tabs">
          <NuxtLink to="/gallery" class="tab-btn active">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2">
              <line x1="8" y1="6" x2="21" y2="6"/>
              <line x1="8" y1="12" x2="21" y2="12"/>
              <line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/>
              <line x1="3" y1="12" x2="3.01" y2="12"/>
              <line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
            <span>ค้นหาด้วยคำอธิบาย</span>
          </NuxtLink>
          
          <NuxtLink to="/gallery/tags" class="tab-btn">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#6B7280" stroke-width="2">
              <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/>
              <line x1="7" y1="7" x2="7.01" y2="7"/>
            </svg>
            <span>ค้นหาด้วย Tag</span>
          </NuxtLink>
        </div>
      </div>

      <!-- Description Search Input Box (Figma Screen 2) -->
      <div class="search-input-box">
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
      <div class="suggestions-row">
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

    <!-- Main Content Results Container -->
    <section class="results-container">
      <!-- Results count & Sorting bar -->
      <div class="results-header">
        <h3 class="results-count">
          พบ <strong>12 ภาพ</strong> สำหรับ "{{ searchQuery }}"
        </h3>

        <div class="sort-selector">
          <span>เรียงตาม: ความเกี่ยวข้อง</span>
        </div>
      </div>

      <!-- 8 Cards for Description mode -->
      <div class="gallery-cards-grid">
        <div 
          v-for="(card, i) in allCards" 
          :key="i" 
          class="gallery-card"
        >
          <!-- Thumbnail Graphic matching exact Figma wave colors -->
          <NuxtLink :to="'/gallery/' + encodeURIComponent(card.filename)" class="card-thumb-link">
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
            <!-- AI description -->
            <NuxtLink :to="'/gallery/' + encodeURIComponent(card.filename)" class="card-desc-link">
              <p class="ai-desc">
                {{ card.desc }}
              </p>
            </NuxtLink>
            
            <!-- Tag chips linking to /gallery/tags -->
            <div class="card-tags-row">
              <NuxtLink 
                v-for="t in card.tags" 
                :key="t" 
                :to="'/gallery/tags?tag=' + encodeURIComponent(t)"
                class="tag-item"
              >
                {{ t }}
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const searchQuery = ref('ชายหาด พระอาทิตย์ตก')

const suggestions = ['แมวนอนบนโซฟา', 'ภูเขาหิมะ', 'อาหารญี่ปุ่น', 'ครอบครัวปิกนิก']

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
  padding: 16px 20px;
  margin-bottom: 24px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
  box-sizing: border-box;
}

.search-nav-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.search-type-tabs {
  display: flex;
  align-items: center;
  background: #F1F5F9;
  padding: 4px;
  border-radius: 10px;
  gap: 4px;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #6B7280;
  font-family: 'Prompt', 'Inter', sans-serif;
  font-size: 13.5px;
  font-weight: 500;
  cursor: pointer;
  text-decoration: none;
  transition: all 0.15s ease-in-out;
}

.tab-btn.active {
  background: #FFFFFF;
  color: #4F46E5;
  font-weight: 600;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
}

.search-input-box {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 14px;
}

.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  flex: 1;
}

.search-icon {
  position: absolute;
  left: 14px;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 11px 38px 11px 42px;
  border-radius: 10px;
  border: 1px solid #D1D5DB;
  font-family: 'Prompt', 'Inter', sans-serif;
  font-size: 14px;
  color: #111827;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.search-input:focus {
  border-color: #4F46E5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.15);
}

.btn-clear {
  position: absolute;
  right: 12px;
  background: none;
  border: none;
  color: #9CA3AF;
  font-size: 13px;
  cursor: pointer;
  padding: 4px;
}

.btn-clear:hover {
  color: #4B5563;
}

.btn-submit-search {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #4F46E5;
  color: #FFFFFF;
  border: none;
  padding: 11px 22px;
  border-radius: 10px;
  font-family: 'Prompt', 'Inter', sans-serif;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-submit-search:hover {
  background: #4338CA;
}

.suggestions-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  flex-wrap: wrap;
}

.suggest-label {
  font-size: 12.5px;
  color: #6B7280;
}

.suggest-chip {
  background: #F1F5F9;
  border: none;
  color: #475569;
  font-family: 'Prompt', 'Inter', sans-serif;
  font-size: 12px;
  padding: 4px 12px;
  border-radius: 9999px;
  cursor: pointer;
  transition: all 0.15s;
}

.suggest-chip:hover {
  background: #E2E8F0;
  color: #1E293B;
}

/* Results Container */
.results-container {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.results-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.results-count {
  font-size: 17px;
  font-weight: 600;
  color: #111827;
  margin: 0;
}

.results-count strong {
  color: #111827;
  font-weight: 700;
}

.sort-selector {
  font-size: 13px;
  color: #6B7280;
}

/* Gallery Cards Grid */
.gallery-cards-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.gallery-card {
  background: #FFFFFF;
  border-radius: 14px;
  border: 1px solid #E5E7EB;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.03);
  display: flex;
  flex-direction: column;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.gallery-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0,0,0,0.06);
}

.card-thumb-link {
  text-decoration: none;
  display: block;
}

.card-thumb {
  height: 150px;
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-sun {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  position: absolute;
  top: 22px;
  box-shadow: 0 0 16px rgba(255, 255, 255, 0.4);
}

.card-wave {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 60px;
}

.card-body {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
}

.card-desc-link {
  text-decoration: none;
  color: inherit;
}

.ai-desc {
  font-size: 13px;
  color: #374151;
  line-height: 1.45;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag-item {
  display: inline-block;
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 11px;
  font-weight: 500;
  padding: 3px 8px;
  border-radius: 6px;
  cursor: pointer;
  text-decoration: none;
  transition: all 0.15s;
}

.tag-item:hover {
  background: #E0E7FF;
}
</style>
