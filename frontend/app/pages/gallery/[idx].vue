<template>
  <div class="photo-detail-page">
    <!-- Loading / error states bound to real API status -->
    <div v-if="photoApi.loading.value && !photo" class="detail-state detail-loading">
      <span class="state-spinner"></span>
      <span>กำลังโหลดข้อมูลรูปภาพ...</span>
    </div>
    <div v-else-if="photoApi.error.value && !photo" class="detail-state detail-error">
      <span>{{ photoApi.error.value }}</span>
      <button class="btn-state-retry" @click="loadDetail()">ลองอีกครั้ง</button>
    </div>

    <!-- Header with Back Button -->
    <template v-if="photo">
      <header class="detail-header">
        <div class="header-left">
          <h1 class="photo-filename">{{ photo.file_name }}</h1>
          <p class="header-subtitle">รายละเอียดรูปภาพ</p>
        </div>

        <button type="button" class="btn-back" @click="handleBack">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round">
            <line x1="19" y1="12" x2="5" y2="12" />
            <polyline points="12 19 5 12 12 5" />
          </svg>
          <span>ย้อนกลับ</span>
        </button>
      </header>

      <!-- Main 2-Column Content Layout -->
      <div class="detail-content-layout">
        <!-- Left Column: Big Preview + Thumbnails Strip -->
        <section class="preview-column">
          <!-- Big Picture Preview: original from /image/{id}, gradient art as fallback -->
          <div class="main-preview-card">
            <img v-if="!previewFailed" :src="previewUrl" class="preview-img" alt="" @error="previewFailed = true" />
            <template v-else>
              <div class="preview-sun"></div>
              <svg class="preview-wave preview-wave-back" viewBox="0 0 560 200" preserveAspectRatio="none">
                <path d="M0,90 C120,40 260,130 560,70 L560,200 L0,200 Z" fill="#94A3B8" />
              </svg>
              <svg class="preview-wave preview-wave-front" viewBox="0 0 560 200" preserveAspectRatio="none">
                <path d="M0,130 C150,70 320,140 560,95 L560,200 L0,200 Z" fill="#64748B" />
              </svg>
            </template>
          </div>

          <!-- Thumbnails Strip below preview: other photos from the gallery API -->
          <div class="thumbnails-strip">
            <div v-for="thumb in strip" :key="thumb.id" class="thumb-strip-item"
              :class="{ active: thumb.id === photoId }" @click="selectPhoto(thumb.id)">
              <div class="thumb-mini-art">
                <img :src="thumb.thumb_url || thumbUrl(thumb.id)" class="thumb-mini-img" alt="" loading="lazy" />
              </div>
            </div>
            <span v-if="strip.length === 0 && !photoApi.loading" class="strip-empty">ไม่มีรูปภาพอื่นในโฟลเดอร์นี้</span>
          </div>
        </section>

        <!-- Right Column: AI Description, Tags, Metadata, Actions -->
        <aside class="sidebar-column">
          <!-- Backup Status Tag -->
          <div class="status-badge-row">
            <span class="badge-backed-up">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"
                stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span>สำรองแล้ว</span>
            </span>
          </div>

          <!-- AI Description Card (Figma Screen) -->
          <div class="info-card ai-desc-card">
            <div class="card-header-row">
              <div class="card-title-badge">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#4F46E5" stroke-width="2">
                  <path
                    d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                </svg>
                <span>คำอธิบายจาก AI</span>
              </div>

              <div class="header-action-group">
                <button class="btn-ai-live" :disabled="isAnalyzingLive" @click="handleRunAiAnalysis"
                  title="ส่งภาพให้ Gemini วิเคราะห์คำอธิบายและ Tag ใหม่">
                  <span v-if="!isAnalyzingLive">✨ วิเคราะห์ด้วย AI</span>
                  <span v-else>⏳ กำลังวิเคราะห์...</span>
                </button>
                <button v-if="isAnalyzingLive" class="btn-cancel-ai" @click="handleCancelAnalysis"
                  title="ยกเลิกการวิเคราะห์">
                  ✕ ยกเลิก
                </button>
                <!-- <button class="btn-edit" :disabled="isSavingDesc" @click="toggleEditDesc">
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                    stroke-linecap="round" stroke-linejoin="round">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                  </svg>
                  <span>{{ isEditingDesc ? (isSavingDesc ? 'กำลังบันทึก...' : 'เสร็จสิ้น') : 'แก้ไข' }}</span>
                </button> -->
              </div>
            </div>

            <div v-if="!isEditingDesc" class="ai-desc-text">
              {{ photo.description || 'ยังไม่มีคำอธิบาย' }}
            </div>
            <div v-else class="edit-desc-box">
              <textarea v-model="descDraft" class="desc-textarea" rows="3"></textarea>
            </div>
          </div>

          <!-- Tags Card -->
          <div class="info-card tags-card">
            <div class="card-title-row">
              <div class="card-title-badge">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#4B5563" stroke-width="2">
                  <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" />
                  <line x1="7" y1="7" x2="7.01" y2="7" />
                </svg>
                <span class="text-dark">Tags</span>
              </div>
            </div>

            <div class="tags-chips-container">
              <span v-for="tag in photo.tags" :key="tag.id" class="tag-pill-item">
                <span class="tag-source-badge" :class="tag.source === 'user' ? 'src-user' : 'src-ai'">{{ tag.source ===
                  'user' ?
                  'คุณ' : 'AI' }}</span>
                <span>{{ tag.name }}</span>
                <!-- <button class="btn-remove-tag" :disabled="photoApi.loading" @click="removeTag(tag)" -->
                <!-- title="ลบแท็ก">✕</button> -->
              </span>

              <!-- <button v-if="!isAddingTag" class="btn-add-tag" @click="openAddTag()">
              <span>+ เพิ่ม Tag</span>
            </button> -->
              <!-- <div v-else class="add-tag-box">
              <input
                v-model="newTag"
                class="tag-input"
                placeholder="พิมพ์ชื่อ Tag แล้วกดเพิ่ม"
                @input="onTagInput()"
                @keyup.enter="submitTag()"
              />
              <div v-if="tagSuggestions.length > 0" class="tag-suggest-list">
                <button
                  v-for="s in tagSuggestions"
                  :key="s"
                  class="tag-suggest-item"
                  @click="pickSuggestion(s)"
                >
                  {{ s }}
                </button>
              </div>
              <div class="add-tag-actions">
                <button class="btn-tag-confirm" :disabled="photoApi.loading || !newTag.trim()" @click="submitTag()">เพิ่ม</button>
                <button class="btn-tag-cancel" :disabled="photoApi.loading" @click="isAddingTag = false">ยกเลิก</button>
              </div>
            </div> -->
            </div>
          </div>

          <!-- File Metadata Card (ข้อมูลไฟล์) -->
          <div class="info-card file-metadata-card">
            <h3 class="card-title-heading">ข้อมูลไฟล์</h3>
            <div class="meta-list">
              <div class="meta-row">
                <span class="meta-label">ที่เก็บ</span>
                <span class="meta-value path-val" :title="photo.full_path">{{ photo.full_path }}</span>
              </div>
              <div class="meta-row">
                <span class="meta-label">ขนาด</span>
                <span class="meta-value">{{ formatBytes(photo.size_bytes) }}</span>
              </div>
              <div class="meta-row">
                <span class="meta-label">สำรองเมื่อ</span>
                <span class="meta-value">{{ formatBackedUpAt(photo.backed_up_at) }}</span>
              </div>
              <div class="meta-row">
                <span class="meta-label">ต้นทาง</span>
                <span class="meta-value">{{ photo.destination_path }}</span>
              </div>
            </div>
          </div>

          <!-- Delete Action Button -->
          <div class="delete-action-row">
            <button class="btn-delete-dest" @click="handleDelete">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                stroke-linecap="round" stroke-linejoin="round">
                <polyline points="3 6 5 6 21 6" />
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
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
import { ref, computed, onMounted, watch } from 'vue'
import { useBackup } from '~/composables/useBackup'
import { usePhotoApi, imageUrl, thumbUrl, formatBackedUpAt } from '~/composables/usePhotoApi'
import type { PhotoDetail, PhotoListItem, TagItem } from '~/composables/usePhotoApi'
import { formatBytes } from '~/utils/formatters'

const route = useRoute()
const router = useRouter()
const { destPath } = useBackup()
const photoApi = usePhotoApi()

const photoId = computed(() => Number(route.params.idx))
const photo = ref<PhotoDetail | null>(null)
const strip = ref<PhotoListItem[]>([])

function handleBack() {
  if (typeof window !== 'undefined' && window.history.length > 1) {
    router.back()
  } else {
    router.push('/gallery')
  }
}

// Description editing (Feature 19)
const isEditingDesc = ref(false)
const descDraft = ref('')
const isSavingDesc = ref(false)

// Real image loading with fallback to the gradient art
const previewFailed = ref(false)
const previewBase64 = ref('')
const previewUrl = computed(() => {
  if (previewBase64.value) return previewBase64.value
  return photo.value ? (photo.value.thumb_url || imageUrl(photo.value.id)) : ''
})

// Tag editing with autocomplete (Feature 19)
const isAddingTag = ref(false)
const newTag = ref('')
const tagSuggestions = ref<string[]>([])

async function loadDetail() {
  if (!photoId.value || Number.isNaN(photoId.value)) return
  const detail = await photoApi.getPhotoDetail(photoId.value)
  if (detail) {
    photo.value = detail
    previewFailed.value = false
    previewBase64.value = ''
    const getThumbFn = (window as any)?.go?.main?.App?.GetThumbnailBase64
    if (detail.full_path && typeof getThumbFn === 'function') {
      try {
        const b64 = await getThumbFn(detail.full_path)
        if (b64) {
          previewBase64.value = b64
          previewFailed.value = false
        }
      } catch {
        // ignore
      }
    }
  }
}

// Other photos for the thumbnails strip
async function loadStrip() {
  if (!destPath.value) return
  const result = await photoApi.searchPhotos(destPath.value, { query: '', tag_ids: [], page: 1, page_size: 8 })
  if (result) {
    strip.value = result.items || []
    const getThumbFn = (window as any)?.go?.main?.App?.GetThumbnailBase64
    if (typeof getThumbFn === 'function') {
      for (const item of strip.value) {
        if (!item.full_path) continue
        try {
          const b64 = await getThumbFn(item.full_path)
          if (b64) item.thumb_url = b64
        } catch {
          // ignore
        }
      }
    }
  }
}

function selectPhoto(id: number) {
  if (id !== photoId.value) navigateTo(`/gallery/${id}`)
}

// Feature 19: edit description, save through the backend and refresh from its result
function toggleEditDesc() {
  if (isEditingDesc.value) {
    saveDescription()
    return
  }
  descDraft.value = photo.value?.description || ''
  isEditingDesc.value = true
}

async function saveDescription() {
  if (!photo.value || isSavingDesc.value) return
  const next = descDraft.value.trim()
  const previous = photo.value.description
  isSavingDesc.value = true
  try {
    photo.value = { ...photo.value, description: next }
    await photoApi.updateDescription(photo.value.id, next)
    await loadDetail()
  } catch (err: any) {
    photo.value = { ...photo.value, description: previous }
    alert('บันทึกคำอธิบายไม่สำเร็จ: ' + (err?.message || 'ข้อผิดพลาดไม่ทราบสาเหตุ'))
  } finally {
    isSavingDesc.value = false
    isEditingDesc.value = false
  }
}

// Feature 19: add tag with autocomplete suggestions, update UI from backend return
function openAddTag() {
  isAddingTag.value = true
  newTag.value = ''
  tagSuggestions.value = []
}

async function onTagInput() {
  const q = newTag.value.trim()
  if (!q) {
    tagSuggestions.value = []
    return
  }
  const list = await photoApi.suggestTags(q, 8)
  if (list) {
    const owned = new Set((photo.value?.tags || []).map((t) => t.id))
    tagSuggestions.value = list.filter((t) => !owned.has(t.id)).map((t) => t.name)
  }
}

function pickSuggestion(name: string) {
  newTag.value = name
  tagSuggestions.value = []
}

async function submitTag() {
  const name = newTag.value.trim()
  if (!name || !photo.value) return
  const updated = await photoApi.addTag(photo.value.id, name)
  if (updated) {
    photo.value = updated
    newTag.value = ''
    isAddingTag.value = false
    tagSuggestions.value = []
  } else {
    alert('เพิ่ม Tag ไม่สำเร็จ: ' + (photoApi.error.value || 'ข้อผิดพลาดไม่ทราบสาเหตุ'))
  }
}

// Feature 19: remove tag with rollback when the backend rejects it
async function removeTag(tag: TagItem) {
  if (!photo.value) return
  const previous = photo.value.tags
  photo.value = { ...photo.value, tags: previous.filter((t) => t.id !== tag.id) }
  try {
    await photoApi.removeTag(photo.value.id, tag.id)
    await loadDetail()
  } catch (err: any) {
    photo.value = { ...photo.value, tags: previous }
    alert('ลบ Tag ไม่สำเร็จ: ' + (err?.message || 'ข้อผิดพลาดไม่ทราบสาเหตุ'))
  }
}

const isAnalyzingLive = ref(false)

async function handleRunAiAnalysis() {
  if (!photo.value) return
  isAnalyzingLive.value = true
  try {
    let result = null
    const path = photo.value.full_path || photo.value.file_name
    if (typeof (window as any)?.go?.main?.App?.AnalyzePhoto === 'function') {
      result = await (window as any).go.main.App.AnalyzePhoto(path)
    } else if (typeof (window as any)?.go?.service?.BackupService?.AnalyzePhoto === 'function') {
      result = await (window as any).go.service.BackupService.AnalyzePhoto(path)
    }
    if (result) {
      if (result.description && photo.value.id) {
        photo.value = { ...photo.value, description: result.description }
        await photoApi.updateDescription(photo.value.id, result.description)
      }
      if (result.tags && Array.isArray(result.tags) && result.tags.length > 0 && photo.value.id) {
        for (const tag of result.tags) {
          try {
            await photoApi.addTag(photo.value.id, tag)
          } catch (tErr) {
            console.warn('Failed to add tag:', tag, tErr)
          }
        }
      }
      await loadDetail()
      alert(`AI วิเคราะห์ภาพสำเร็จ!\n\nคำอธิบาย: ${result.description || '(ไม่มี)'}\nแท็ก: ${(result.tags || []).join(', ')}`)
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
  if (!photo.value) return
  if (confirm(`คุณต้องการลบ ${photo.value.file_name} จากโฟลเดอร์ปลายทางหรือไม่?`)) {
    alert('การลบไฟล์ในปลายทางจัดการได้ที่หน้าประวัติ & ตรวจสอบ')
  }
}

onMounted(async () => {
  await loadDetail()
  await loadStrip()
})

watch(photoId, async () => {
  await loadDetail()
  await loadStrip()
})
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

/* Real data states: loading / error */
.detail-state {
  background: #FFFFFF;
  border: 1px dashed #C7D2FE;
  border-radius: 14px;
  padding: 32px 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 13.5px;
  color: #6B7280;
  font-weight: 500;
  box-sizing: border-box;
}

.detail-state.detail-error {
  border-color: #FECACA;
  color: #DC2626;
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

.btn-state-retry {
  background: #4F46E5;
  border: none;
  color: #FFFFFF;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 6px 14px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.btn-state-retry:hover {
  background: #4338CA;
}

.strip-empty {
  font-size: 12px;
  color: #94A3B8;
  align-self: center;
  padding-left: 4px;
}

/* Real preview image from /image/{id} */
.preview-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
  background: #F8FAFC;
}

/* Real strip thumbnails from /thumb/{id} */
.thumb-mini-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  background: #F1F5F9;
}

/* Tag source badges (AI vs user) */
.tag-source-badge {
  font-size: 9.5px;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 9999px;
  letter-spacing: 0.3px;
}

.tag-source-badge.src-ai {
  background: #EEF2FF;
  color: #4F46E5;
}

.tag-source-badge.src-user {
  background: #DCFCE7;
  color: #15803D;
}

/* Inline add-tag box with autocomplete */
.add-tag-box {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.tag-input {
  border: 1px solid #CBD5E1;
  border-radius: 9999px;
  padding: 5px 12px;
  font-family: inherit;
  font-size: 12px;
  color: #1E293B;
  width: 180px;
  outline: none;
  background: #FFFFFF;
}

.tag-input:focus {
  border-color: #4F46E5;
}

.tag-suggest-list {
  position: absolute;
  top: 34px;
  left: 0;
  z-index: 5;
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 10px;
  box-shadow: 0 6px 14px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
  min-width: 180px;
  overflow: hidden;
}

.tag-suggest-item {
  background: transparent;
  border: none;
  font-family: inherit;
  font-size: 12px;
  color: #334155;
  text-align: left;
  padding: 7px 12px;
  cursor: pointer;
}

.tag-suggest-item:hover {
  background: #EEF2FF;
  color: #4338CA;
}

.add-tag-actions {
  display: flex;
  gap: 6px;
}

.btn-tag-confirm,
.btn-tag-cancel {
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  padding: 5px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-tag-confirm {
  background: #4F46E5;
  border: none;
  color: #FFFFFF;
}

.btn-tag-confirm:hover:not(:disabled) {
  background: #4338CA;
}

.btn-tag-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-tag-cancel {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  color: #374151;
}

.btn-tag-cancel:hover:not(:disabled) {
  background: #F9FAFB;
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
  cursor: pointer;
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
