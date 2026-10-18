import { ref } from 'vue'

// Shapes mirror the generated wailsjs models (frontend/wailsjs/go/models.ts)
export interface TagItem {
  id: number
  name: string
  count: number
  source?: string
}

export interface PhotoListItem {
  id: number
  file_name: string
  full_path: string
  thumb_url: string
  description: string
  tags: string[]
  backed_up_at: string
}

export interface SearchParams {
  query: string
  tag_ids: number[]
  page: number
  page_size: number
}

export interface SearchResult {
  items: PhotoListItem[]
  total: number
  analyzed_total?: number
  page: number
  page_size: number
  total_pages: number
}

export interface PhotoDetail {
  id: number
  file_name: string
  destination_path: string
  full_path: string
  thumb_url: string
  description: string
  tags: TagItem[]
  size_bytes: number
  sha256: string
  status: string
  backed_up_at: string
}

// Tag cloud item shape the gallery UI consumes
export interface TagCloudItem {
  id: number
  name: string
  size: number
  color: string
}

// Gallery card shape the gallery UI consumes
export interface GalleryCard {
  id: number
  filename: string
  full_path: string
  desc: string
  tags: string[]
  thumb_url: string
}

// PhotoVault asset-server routes (B3)
export function imageUrl(id: number): string {
  return `/image/${id}`
}

export function thumbUrl(id: number): string {
  return `/thumb/${id}`
}

const TAG_COLORS = [
  '#0284C7', '#312E81', '#059669', '#EA580C', '#16A34A', '#6366F1',
  '#B45309', '#E11D48', '#0D9488', '#64748B', '#7C3AED', '#EC4899'
]

const TAG_SIZE_MIN = 13
const TAG_SIZE_MAX = 26

// toGalleryCard maps a backend list item to the gallery card shape
export function toGalleryCard(item: PhotoListItem): GalleryCard {
  return {
    id: item.id,
    filename: item.file_name,
    full_path: item.full_path || '',
    desc: item.description || '',
    tags: item.tags || [],
    thumb_url: item.thumb_url || thumbUrl(item.id)
  }
}

// toTagCloudItems scales font size by photo count on a log scale
// (font-size ~ log(count)) and cycles the existing color palette
export function toTagCloudItems(items: TagItem[]): TagCloudItem[] {
  const counts = items.map((i) => Math.max(1, i.count))
  const min = Math.min(...counts)
  const max = Math.max(...counts)
  const logMin = Math.log(min)
  const logMax = Math.log(max)

  return items.map((item, index) => {
    const ratio = max > min ? (Math.log(Math.max(1, item.count)) - logMin) / (logMax - logMin) : 0.5
    return {
      id: item.id,
      name: item.name,
      size: Math.round(TAG_SIZE_MIN + ratio * (TAG_SIZE_MAX - TAG_SIZE_MIN)),
      color: TAG_COLORS[index % TAG_COLORS.length]
    }
  })
}

// formatBackedUpAt renders the Go "2006-01-02 15:04:05" timestamp in Thai style
export function formatBackedUpAt(value: string): string {
  if (!value) return '-'
  const [datePart, timePart = ''] = value.split(' ')
  const [y, m, d] = datePart.split('-').map((n) => Number(n))
  if (!y || !m || !d) return value
  const months = ['ม.ค.', 'ก.พ.', 'มี.ค.', 'เม.ย.', 'พ.ค.', 'มิ.ย.', 'ก.ค.', 'ส.ค.', 'ก.ย.', 'ต.ค.', 'พ.ย.', 'ธ.ค.']
  const time = timePart.slice(0, 5)
  return `${d} ${months[m - 1]} ${y}${time ? ' ' + time : ''}`
}

// API layer wrapping the bound Wails GalleryController methods
export function usePhotoApi() {
  const loading = ref(false)
  const error = ref<string | null>(null)
  let pendingRequests = 0

  // A Wails binding normally always settles.  If the renderer loses its IPC
  // response (for example after a backend restart), keeping the page in a
  // permanent loading state is much worse than surfacing a retryable error.
  const REQUEST_TIMEOUT_MS = 15_000

  function available(): boolean {
    return typeof (window as any)?.go?.controller?.GalleryController?.SearchPhotos === 'function'
  }

  // Force clear loading state (call if UI gets stuck)
  function clearLoading() {
    pendingRequests = 0
    loading.value = false
  }

  async function run<T>(fn: () => Promise<T>, fallbackError: string): Promise<T | null> {
    if (!available()) {
      error.value = 'ไม่สามารถเชื่อมต่อ backend ได้ (GalleryController)'
      return null
    }
    pendingRequests += 1
    loading.value = true
    error.value = null
    try {
      // Use Promise.race with timeout to prevent hanging requests
      const result = await Promise.race([
        fn(),
        new Promise<never>((_, reject) => {
          window.setTimeout(() => reject(new Error('การเชื่อมต่อ backend ใช้เวลานานเกินไป กรุณาลองอีกครั้ง')), REQUEST_TIMEOUT_MS)
        })
      ])
      return result
    } catch (err: any) {
      error.value = err?.message || String(err) || fallbackError
      return null
    } finally {
      // Always decrement and update loading state, even if component discards the result
      pendingRequests = Math.max(0, pendingRequests - 1)
      loading.value = pendingRequests > 0
    }
  }

  // Feature 15 & 17: search by description words and/or tags (AND)
  async function searchPhotos(dest: string, params: SearchParams): Promise<SearchResult | null> {
    return run(() => (window as any).go.controller.GalleryController.SearchPhotos(dest, params), 'ค้นหารูปภาพไม่สำเร็จ')
  }

  // Feature 16: tag cloud ordered by photo count
  async function getTagCloud(dest: string, limit = 50): Promise<TagItem[] | null> {
    return run(() => (window as any).go.controller.GalleryController.GetTagCloud(dest, limit), 'ดึงข้อมูล Tag Cloud ไม่สำเร็จ')
  }

  // Feature 18: photo detail
  async function getPhotoDetail(id: number): Promise<PhotoDetail | null> {
    return run(() => (window as any).go.controller.GalleryController.GetPhotoDetail(id), 'ดึงข้อมูลรูปภาพไม่สำเร็จ')
  }

  // Feature 19: edit description
  async function updateDescription(id: number, description: string): Promise<void> {
    await run(() => (window as any).go.controller.GalleryController.UpdateDescription(id, description), 'บันทึกคำอธิบายไม่สำเร็จ')
  }

  // Feature 19: add tag (returns refreshed detail from backend)
  async function addTag(photoID: number, name: string): Promise<PhotoDetail | null> {
    return run(() => (window as any).go.controller.GalleryController.AddTag(photoID, name), 'เพิ่ม Tag ไม่สำเร็จ')
  }

  // Feature 19: remove tag
  async function removeTag(photoID: number, tagID: number): Promise<void> {
    await run(() => (window as any).go.controller.GalleryController.RemoveTag(photoID, tagID), 'ลบ Tag ไม่สำเร็จ')
  }

  // Autocomplete suggestions for the tag input
  async function suggestTags(prefix: string, limit = 8): Promise<TagItem[] | null> {
    return run(() => (window as any).go.controller.GalleryController.SuggestTags(prefix, limit), 'ดึงคำแนะนำ Tag ไม่สำเร็จ')
  }

  return {
    loading,
    error,
    available,
    clearLoading,
    searchPhotos,
    getTagCloud,
    getPhotoDetail,
    updateDescription,
    addTag,
    removeTag,
    suggestTags
  }
}
