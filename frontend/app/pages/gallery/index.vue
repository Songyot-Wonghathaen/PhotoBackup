<template>
  <div class="gallery-view">
    <!-- Header -->
    <header class="view-header">
      <div class="header-text">
        <h1 class="view-title">แกลเลอรี่ & ค้นหา</h1>
        <p class="view-subtitle">
          {{ activeTab === 'desc' ? 'ค้นหาภาพจากคำอธิบายที่ AI สร้างให้ หรือจาก Tag' : 'คลิก Tag ใน Tag Cloud เพื่อกรองรูปภาพ' }}
        </p>
      </div>

      <div class="header-pills-row">
        <!-- AI Progress Indicator -->
        <div v-if="isAiAnalyzing || (total > 0 && analyzedTotal < total)" class="header-count-pill pill-ai-live">
          <span class="pulse-dot"></span>
          <span>AI วิเคราะห์แล้ว {{ Math.max(aiProgress.current, analyzedTotal) }}/{{ total }} รูป</span>
        </div>
        <div v-else-if="total > 0 && analyzedTotal >= total" class="header-count-pill pill-ai-done">
          <span>✨ AI วิเคราะห์ครบแล้ว {{ total }}/{{ total }}</span>
        </div>

        <div class="header-count-pill">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2" />
            <circle cx="8.5" cy="8.5" r="1.5" />
            <polyline points="21 15 16 10 5 21" />
          </svg>
          <span>รูปทั้งหมด {{ total }}</span>
        </div>
      </div>
    </header>

    <!-- Search Controls Bar -->
    <div class="search-panel-card" :class="{ 'compact-mode': activeTab === 'tag' }">
      <div class="search-nav-row">
        <!-- Switch between Search by Description vs Search by Tag -->
        <div class="search-type-tabs">
          <button class="tab-btn" :class="{ active: activeTab === 'desc' }" @click="switchTab('desc')">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none"
              :stroke="activeTab === 'desc' ? '#4F46E5' : '#6B7280'" stroke-width="2">
              <line x1="8" y1="6" x2="21" y2="6" />
              <line x1="8" y1="12" x2="21" y2="12" />
              <line x1="8" y1="18" x2="21" y2="18" />
              <line x1="3" y1="6" x2="3.01" y2="6" />
              <line x1="3" y1="12" x2="3.01" y2="12" />
              <line x1="3" y1="18" x2="3.01" y2="18" />
            </svg>
            <span>ค้นหาด้วยคำอธิบาย</span>
          </button>

          <button class="tab-btn" :class="{ active: activeTab === 'tag' }" @click="switchTab('tag')">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none"
              :stroke="activeTab === 'tag' ? '#4F46E5' : '#6B7280'" stroke-width="2">
              <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" />
              <line x1="7" y1="7" x2="7.01" y2="7" />
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
              <button class="btn-clear-tag" @click="clearTag()" title="ยกเลิกการเลือก">✕</button>
            </span>
          </div>

          <button class="btn-reset-filter" @click="clearTag()">
            ล้างทั้งหมด
          </button>
        </template>
      </div>

      <!-- Description Search Input Box (Figma Screen 2) -->
      <div v-if="activeTab === 'desc'" class="search-input-box">
        <div class="search-input-wrapper">
          <svg class="search-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#4F46E5"
            stroke-width="2">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input type="text" v-model="searchQuery" placeholder="ชายหาด พระอาทิตย์ตก" class="search-input" />
          <button v-if="searchQuery" class="btn-clear" @click="searchQuery = ''">✕</button>
        </div>
        <button class="btn-submit-search" :disabled="photoApi.loading.value" @click="search(1, false)">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <span>ค้นหา</span>
        </button>
      </div>

      <!-- Quick Suggestion chips -->
      <div v-if="activeTab === 'desc'" class="suggestions-row">
        <span class="suggest-label">ลองค้นหา:</span>
        <button v-for="s in suggestions" :key="s" class="suggest-chip" @click="searchQuery = s">
          {{ s }}
        </button>
      </div>

      <!-- AI Expanded Terms Indicator -->
      <div v-if="activeTab === 'desc' && expandedTerms.length > 1 && searchQuery.trim()" class="ai-expanded-row">
        <span class="ai-expanded-badge">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon>
          </svg>
          AI แตกคำค้นหาให้:
        </span>
        <div class="ai-term-chips">
          <span v-for="term in expandedTerms" :key="term" class="ai-term-chip">{{ term }}</span>
        </div>
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
              <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" />
              <line x1="7" y1="7" x2="7.01" y2="7.01" />
            </svg>
            {{ tagCloudList.length }} tags
          </span>
        </div>

        <div class="cloud-words">
          <span v-for="item in tagCloudList" :key="item.name" class="cloud-word"
            :class="{ selected: selectedTag === item.name }"
            :style="selectedTag === item.name ? {} : { fontSize: item.size + 'px', color: item.color }"
            @click="selectTag(item)">
            {{ item.name }}
          </span>
          <span v-if="tagCloudList.length === 0 && !photoApi.loading" class="cloud-empty">
            ยังไม่มี Tag ในโฟลเดอร์ปลายทางนี้
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
              พบ <strong>{{ total }} ภาพ</strong> สำหรับ "{{ searchQuery }}"
            </template>
            <template v-else-if="selectedTag">
              พบ <strong>{{ total }} ภาพ</strong> ที่มี Tag "{{ selectedTag }}"
            </template>
            <template v-else>
              รูปภาพในโฟลเดอร์ <strong>{{ total }} ภาพ</strong> (วิเคราะห์คำอธิบายแล้ว {{ Math.max(aiProgress.current, analyzedTotal) }} ภาพ)
            </template>
          </h3>

          <div class="sort-selector" v-if="activeTab === 'desc'">
            <span>เรียงตาม: ความเกี่ยวข้อง</span>
          </div>
          <div class="sort-selector" v-else>
            <span>แสดง {{ cards.length }} จาก {{ total }}</span>
          </div>
        </div>

        <!-- Loading is shown only while the API request is in progress. -->
        <div v-if="photoApi.loading && cards.length === 0" class="gallery-state gallery-loading">
          <span class="state-spinner"></span>
          <span>กำลังโหลดรูปภาพ</span>
        </div>

        <!-- One shared result-card template for description and tag searches. -->
        <div v-else class="gallery-cards-grid">
          <template v-if="cards.length > 0">
            <div v-for="(card, i) in displayedPhotos" :key="card.id || i" class="gallery-card">
              <!-- Thumbnail from the asset server (/thumb/{id}), gradient art as fallback -->
              <NuxtLink :to="'/gallery/' + card.id" class="card-thumb-link">
                <div class="card-thumb">
                  <img v-if="card.thumb_url" :src="card.thumb_url" class="card-thumb-img" alt="" loading="lazy"
                    @error="onThumbError(card)" />
                  <template v-else>
                    <div class="card-sun"></div>
                    <svg class="card-wave" viewBox="0 0 240 90" preserveAspectRatio="none">
                      <path d="M0,45 C75,20 150,68 240,35 L240,90 L0,90 Z" fill="#94A3B8" opacity="0.45" />
                      <path d="M0,60 C80,30 160,75 240,48 L240,90 L0,90 Z" fill="#64748B" />
                    </svg>
                  </template>
                </div>
              </NuxtLink>

              <!-- Card Body -->
              <div class="card-body">
                <!-- Keep the same content position in both search modes. -->
                <NuxtLink :to="'/gallery/' + card.id" class="card-desc-link">
                  <p class="ai-desc">
                    {{ card.desc || 'ไม่มีคำอธิบาย' }}
                  </p>
                </NuxtLink>

                <!-- Tag chips -->
                <div class="card-tags-row">
                  <span v-for="t in card.tags" :key="t" class="tag-item" :class="{ highlight: selectedTag === t }"
                    @click.stop="selectTagByName(t)">
                    {{ t }}
                  </span>
                </div>
              </div>
            </div>
          </template>

          <div v-else class="gallery-card gallery-empty-card">
            <svg class="empty-image-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"
              aria-hidden="true">
              <rect x="3" y="4" width="18" height="16" rx="2" />
              <circle cx="8.5" cy="9" r="1.5" />
              <path d="m21 15-4.5-4.5L7 20" />
            </svg>
            <span>{{ emptyResultMessage }}</span>
          </div>
        </div>

        <!-- Load more button while more pages are available -->
        <div class="load-more-row" v-if="hasMore && !photoApi.error.value">
          <button class="btn-load-more" :disabled="photoApi.loading.value" @click="loadMore()">
            {{ photoApi.loading.value ? 'กำลังโหลด...' : 'โหลดเพิ่ม' }}
          </button>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useBackup } from '~/composables/useBackup'
import { usePhotoApi, toGalleryCard, toTagCloudItems } from '~/composables/usePhotoApi'
import type { GalleryCard, TagCloudItem } from '~/composables/usePhotoApi'

const route = useRoute()
const { destPath } = useBackup()
const photoApi = usePhotoApi()

// Default to 'tag' mode matching Screen 2 Search by Tag provided by user
const activeTab = ref<'desc' | 'tag'>('tag')
const searchQuery = ref('')
const selectedTag = ref('')
const selectedTagId = ref<number | null>(null)
const selectedTagExists = ref<boolean | null>(null)

// Discards responses from outdated requests (last write wins)
let searchSeq = 0
let debounceTimer: ReturnType<typeof setTimeout> | null = null
let lastAiRefreshAt = 0
let aiRefreshTimer: ReturnType<typeof setTimeout> | null = null
const PAGE_SIZE = 24
const thumbCache = new Map<string, string>()

const tagCloudList = ref<TagCloudItem[]>([])
const cards = ref<GalleryCard[]>([])
const total = ref(0)
const analyzedTotal = ref(0)
const page = ref(1)
const totalPages = ref(0)
const aiProgress = ref({ current: 0, total: 0 })
const isAiAnalyzing = ref(false)
const expandedTerms = ref<string[]>([])

const suggestions = ['แมวนอนบนโซฟา', 'ภูเขาหิมะ', 'อาหารญี่ปุ่น', 'ครอบครัวปิกนิก']

const hasMore = computed(() => cards.value.length < total.value)
const displayedPhotos = computed(() => cards.value)
const emptyResultMessage = computed(() => {
  if (activeTab.value === 'tag') {
    return selectedTag.value && selectedTagExists.value === false
      ? 'ไม่พบผลการค้นหา'
      : 'ไม่มีรูปภาพ'
  }
  return searchQuery.value.trim() ? 'ไม่พบผลการค้นหา' : 'ไม่มีรูปภาพ'
})

function switchTab(tab: 'desc' | 'tag') {
  if (activeTab.value === tab) return
  activeTab.value = tab
  expandedTerms.value = []
  search(1, false)
}

// Feature 16: load tags with photo counts for this destination
async function loadCloud() {
  if (!destPath.value) return
  const cloud = await photoApi.getTagCloud(destPath.value, 50)
  photoApi.clearLoading()
  if (cloud) {
    tagCloudList.value = toTagCloudItems(cloud)
    const pending = tagCloudList.value.find((t) => t.name === selectedTag.value)
    if (pending && !selectedTagId.value) {
      selectedTagId.value = pending.id
    }
  }
}

async function hydrateCardThumbnails(items: GalleryCard[]) {
  const getThumbFn = (window as any)?.go?.main?.App?.GetThumbnailBase64
  if (typeof getThumbFn !== 'function') return

  for (const card of items) {
    if (!card.full_path) continue
    const cached = thumbCache.get(card.full_path)
    if (cached) {
      card.thumb_url = cached
      continue
    }
    try {
      const b64 = await getThumbFn(card.full_path)
      if (b64) {
        thumbCache.set(card.full_path, b64)
        card.thumb_url = b64
      }
    } catch {
      // ignore individual thumb error
    }
  }
}

// Feature 15/17: run the search; append=true loads the next page
async function search(newPage: number, append: boolean) {
  if (!destPath.value) return

  // A tag name that does not exist must not fall back to an unfiltered search.
  if (activeTab.value === 'tag' && selectedTag.value && selectedTagExists.value === false) {
    if (!append) cards.value = []
    total.value = 0
    totalPages.value = 0
    return
  }
  const seq = ++searchSeq
  const params = {
    query: activeTab.value === 'desc' ? searchQuery.value.trim() : '',
    tag_ids: activeTab.value === 'tag' && selectedTagId.value ? [selectedTagId.value] : [],
    page: newPage,
    page_size: PAGE_SIZE
  }
  const result = await photoApi.searchPhotos(destPath.value, params)
  // Safety: ensure loading is cleared even if race condition occurs
  photoApi.clearLoading()
  if (seq !== searchSeq) return // stale response: thrown away
  if (!result) {
    if (!append) cards.value = []
    total.value = 0
    totalPages.value = 0
    return
  }
  const mapped = (result.items || []).map((item) => {
    const c = toGalleryCard(item)
    const cached = c.full_path ? thumbCache.get(c.full_path) : undefined
    if (cached) c.thumb_url = cached
    return c
  })
  cards.value = append ? [...cards.value, ...mapped] : mapped
  total.value = result.total || 0
  if (!append) {
    expandedTerms.value = result.expanded_terms || []
  }
  if (typeof result.analyzed_total === 'number') {
    analyzedTotal.value = result.analyzed_total
  }
  page.value = result.page || newPage
  totalPages.value = result.total_pages || 0

  const sliceStart = append ? Math.max(0, cards.value.length - mapped.length) : 0
  void hydrateCardThumbnails(cards.value.slice(sliceStart))
}

// 300ms debounce for the description search box
watch(searchQuery, () => {
  if (activeTab.value !== 'desc') return
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => search(1, false), 300)
})

// Feature 17: click a cloud tag (AND filter) or a card tag chip
function selectTag(item: TagCloudItem) {
  activeTab.value = 'tag'
  selectedTag.value = item.name
  selectedTagId.value = item.id
  selectedTagExists.value = true
  search(1, false)
}

async function resolveSelectedTag(name: string): Promise<void> {
  const cloudMatch = tagCloudList.value.find((tag) => tag.name === name)
  if (cloudMatch) {
    selectedTagId.value = cloudMatch.id
    selectedTagExists.value = true
    return
  }

  // SuggestTags queries the existing tags table, including tags whose photos
  // are no longer active, so the empty state can remain accurate.
  const matches = await photoApi.suggestTags(name, 50)
  photoApi.clearLoading()
  const exact = matches?.find((tag) => tag.name === name)
  selectedTagId.value = exact?.id ?? null
  selectedTagExists.value = Boolean(exact)
}

async function selectTagByName(name: string) {
  const found = tagCloudList.value.find((t) => t.name === name)
  if (found) {
    selectTag(found)
  } else {
    activeTab.value = 'tag'
    selectedTag.value = name
    await resolveSelectedTag(name)
    await search(1, false)
  }
}

function clearTag() {
  selectedTag.value = ''
  selectedTagId.value = null
  selectedTagExists.value = null
  if (activeTab.value === 'tag') search(1, false)
}

function loadMore() {
  if (hasMore.value && !photoApi.loading.value) {
    search(page.value + 1, true)
  }
}

// Real thumbnails come from the asset server or Wails IPC fallback; fall back to gradient art only if both fail
async function onThumbError(card: GalleryCard) {
  if (card.full_path && !card.thumb_url.startsWith('data:')) {
    const cached = thumbCache.get(card.full_path)
    if (cached) {
      card.thumb_url = cached
      return
    }
    const getThumbFn = (window as any)?.go?.main?.App?.GetThumbnailBase64
    if (typeof getThumbFn === 'function') {
      try {
        const b64 = await getThumbFn(card.full_path)
        if (b64) {
          thumbCache.set(card.full_path, b64)
          card.thumb_url = b64
          return
        }
      } catch {
        // fall through to clear thumb_url
      }
    }
  }
  card.thumb_url = ''
}

function scheduleLiveAiRefresh(forceImmediate = false) {
  const now = Date.now()
  if (forceImmediate || now - lastAiRefreshAt >= 2000) {
    lastAiRefreshAt = now
    if (aiRefreshTimer) {
      clearTimeout(aiRefreshTimer)
      aiRefreshTimer = null
    }
    void loadCloud()
    if (page.value <= 1) {
      void search(1, false)
    }
    return
  }
  if (!aiRefreshTimer) {
    aiRefreshTimer = setTimeout(() => {
      aiRefreshTimer = null
      lastAiRefreshAt = Date.now()
      void loadCloud()
      if (page.value <= 1) {
        void search(1, false)
      }
    }, 2000 - (now - lastAiRefreshAt))
  }
}

onMounted(async () => {
  if (route.query.tab === 'desc') {
    activeTab.value = 'desc'
  } else if (route.query.tab === 'tag') {
    activeTab.value = 'tag'
  }
  if (route.query.tag && typeof route.query.tag === 'string') {
    selectedTag.value = route.query.tag
  }

  if (typeof window !== 'undefined' && (window as any).runtime?.EventsOn) {
    ;(window as any).runtime.EventsOn('ai_analysis_progress', (data: any) => {
      if (data) {
        aiProgress.value.current = data.current || 0
        aiProgress.value.total = data.total || 0
        isAiAnalyzing.value = (data.current || 0) < (data.total || 0)
        analyzedTotal.value = Math.max(analyzedTotal.value, data.current || 0)
      }
      const done = data && data.current >= data.total
      scheduleLiveAiRefresh(Boolean(done))
    })
  }

  await loadCloud()
  if (selectedTag.value && tagCloudList.value.length > 0) {
    const found = tagCloudList.value.find((t) => t.name === selectedTag.value)
    if (found) {
      selectedTagId.value = found.id
      selectedTagExists.value = true
      await search(1, false)
      return
    }
  }
  if (selectedTag.value) await resolveSelectedTag(selectedTag.value)
  await search(1, false)
})

onUnmounted(() => {
  if (aiRefreshTimer) {
    clearTimeout(aiRefreshTimer)
    aiRefreshTimer = null
  }
})

watch(destPath, async () => {
  await loadCloud()
  if (selectedTag.value) await resolveSelectedTag(selectedTag.value)
  await search(1, false)
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
  overflow-x: hidden;
  user-select: none;
  font-family: 'Prompt', 'Inter', sans-serif;
}

.header-pills-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-count-pill.pill-ai-live {
  background: #EEF2FF;
  border-color: #C7D2FE;
  color: #4338CA;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-count-pill.pill-ai-done {
  background: #F0FDF4;
  border-color: #BBF7D0;
  color: #15803D;
  font-weight: 600;
}

.pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #4F46E5;
  box-shadow: 0 0 0 0 rgba(79, 70, 229, 0.7);
  animation: pulse-ring 1.5s cubic-bezier(0.45, 0, 0.2, 1) infinite;
}

@keyframes pulse-ring {
  0% {
    box-shadow: 0 0 0 0 rgba(79, 70, 229, 0.7);
  }
  70% {
    box-shadow: 0 0 0 6px rgba(79, 70, 229, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(79, 70, 229, 0);
  }
}

/* Real data states: loading / error / empty */
.gallery-state {
  background: #FFFFFF;
  border: 1px dashed #C7D2FE;
  border-radius: 14px;
  padding: 28px 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 13.5px;
  color: #6B7280;
  font-weight: 500;
  box-sizing: border-box;
}

.gallery-state.gallery-error {
  border-color: #FECACA;
  color: #DC2626;
}

.gallery-state.gallery-empty {
  color: #64748B;
}

.state-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid #C7D2FE;
  border-top-color: #4F46E5;
  border-radius: 50%;
  animation: gallery-spin 0.8s linear infinite;
  flex-shrink: 0;
}

@keyframes gallery-spin {
  to {
    transform: rotate(360deg);
  }
}

.cloud-empty {
  font-size: 12.5px;
  color: #94A3B8;
  font-weight: 500;
}

/* Real thumbnail from /thumb/{id} */
.card-thumb-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  background: #F1F5F9;
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
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
  box-sizing: border-box;
}

.search-panel-card.compact-mode {
  padding: 0 20px;
  height: 74px;
  min-height: 74px;
  display: flex;
  align-items: center;
}

.search-nav-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
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
  height: 41px;
  box-sizing: border-box;
  align-items: center;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 35px;
  padding: 0 16px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #6B7280;
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease-in-out;
  box-sizing: border-box;
}

.tab-btn.active {
  background: #FFFFFF;
  color: #111827;
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
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
  height: 31px;
  padding: 0 12px;
  border-radius: 15.5px;
  box-sizing: border-box;
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

/* AI Expanded Search Terms Row */
.ai-expanded-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 8px 12px;
  background: #F5F3FF;
  border: 1px solid #DDD6FE;
  border-radius: 8px;
  font-size: 12px;
}

.ai-expanded-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: #6D28D9;
  font-weight: 600;
}

.ai-term-chips {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.ai-term-chip {
  background: #EDE9FE;
  color: #5B21B6;
  font-weight: 500;
  padding: 2px 9px;
  border-radius: 6px;
  font-size: 12px;
}

/* Layout: Full width or With Tag Cloud Sidebar */
.gallery-content-layout {
  display: flex;
  gap: 20px;
  align-items: flex-start;
  min-width: 0;
}

/* Tag Cloud Sidebar */
.tag-cloud-card {
  width: 320px;
  flex-shrink: 0;
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 14px;
  padding: 18px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
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
  min-width: 0;
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
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
}

.gallery-card {
  background: #FFFFFF;
  border-radius: 14px;
  border: 1px solid #E5E7EB;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: transform 0.15s, box-shadow 0.15s;
  min-width: 0;
}

.gallery-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 14px rgba(0, 0, 0, 0.05);
}

.card-thumb-link,
.card-desc-link {
  text-decoration: none;
  color: inherit;
  display: block;
}

/* Thumbnail Visuals */
.card-thumb {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 10;
  height: auto;
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

.gallery-empty-card {
  min-height: 220px;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #64748B;
  font-size: 13px;
  font-weight: 500;
  border-style: dashed;
}

.empty-image-icon {
  width: 34px;
  height: 34px;
  color: #94A3B8;
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

.tag-item:hover,
.tag-item.highlight {
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

@media (max-width: 1023px) {
  .gallery-view {
    padding: 20px;
  }

  .gallery-content-layout {
    flex-direction: column;
  }

  .tag-cloud-card,
  .results-container {
    width: 100%;
    box-sizing: border-box;
  }

  .cloud-words {
    min-height: 0;
  }

  .gallery-cards-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 639px) {
  .gallery-view {
    padding: 16px;
  }

  .view-header,
  .results-header,
  .search-nav-row {
    align-items: stretch;
    flex-direction: column;
    gap: 10px;
  }

  .header-count-pill {
    align-self: flex-start;
  }

  .search-panel-card,
  .search-panel-card.compact-mode {
    height: auto;
    min-height: 0;
    padding: 14px;
  }

  .search-panel-card:not(.compact-mode) .search-nav-row {
    margin-bottom: 12px;
  }

  .search-type-tabs,
  .search-input-box {
    width: 100%;
  }

  .search-type-tabs {
    height: auto;
  }

  .tab-btn {
    flex: 1;
    justify-content: center;
    min-width: 0;
    padding: 0 8px;
    font-size: 12px;
  }

  .selected-tag-info {
    flex-wrap: wrap;
  }

  .search-input-box {
    flex-direction: column;
  }

  .search-input-wrapper {
    width: 100%;
    box-sizing: border-box;
  }

  .btn-submit-search {
    height: 42px;
    justify-content: center;
  }

  .gallery-cards-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .sort-selector {
    display: none;
  }
}
</style>
