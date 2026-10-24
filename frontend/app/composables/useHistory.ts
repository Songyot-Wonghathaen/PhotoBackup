import { ref, computed } from 'vue'
import type { PhotoHistoryItem, HistoryStats, IntegrityResult } from '~/models'
import { useBackup } from './useBackup'

// Shared reactive state for History View (Feature E)
const currentDest = ref('')
const rows = ref<PhotoHistoryItem[]>([])
const isChecking = ref(false)
const isDeleting = ref(false)
const isLoading = ref(false)
const lastCheckDuration = ref('12.4 ms')
const activeFilter = ref<'all' | 'passed' | 'missing' | 'deleted'>('all')
const searchQuery = ref('')
const alertNotice = ref<{ type: 'success' | 'warning' | 'info'; message: string } | null>(null)

// Modal states
const selectedPhotoForDetail = ref<PhotoHistoryItem | null>(null)
const photoToDelete = ref<PhotoHistoryItem | null>(null)
const showDeleteAllModal = ref(false)
const showChangeDestModal = ref(false)
const manualDestInput = ref('')

const gradientPalettes = [
  'linear-gradient(135deg, #38BDF8, #0284C7)',
  'linear-gradient(135deg, #60A5FA, #2563EB)',
  'linear-gradient(135deg, #34D399, #059669)',
  'linear-gradient(135deg, #FB923C, #EA580C)',
  'linear-gradient(135deg, #FBBF24, #D97706)',
  'linear-gradient(135deg, #4ADE80, #16A34A)',
  'linear-gradient(135deg, #A855F7, #7C3AED)',
  'linear-gradient(135deg, #EC4899, #BE185D)',
  'linear-gradient(135deg, #14B8A6, #0F766E)',
  'linear-gradient(135deg, #6366F1, #4338CA)'
]

function getGradient(name: string, id: number): string {
  let hash = id
  for (let i = 0; i < name.length; i++) {
    hash = (hash << 5) - hash + name.charCodeAt(i)
    hash |= 0
  }
  const idx = Math.abs(hash) % gradientPalettes.length
  return gradientPalettes[idx] ?? 'linear-gradient(135deg, #38BDF8, #0284C7)'
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  if (bytes >= 1024 * 1024 * 1024) {
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
  }
  if (bytes >= 1024 * 1024) {
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  }
  if (bytes >= 1024) {
    return (bytes / 1024).toFixed(0) + ' KB'
  }
  return bytes + ' B'
}

const thaiMonths = [
  'ม.ค.', 'ก.พ.', 'มี.ค.', 'เม.ย.', 'พ.ค.', 'มิ.ย.',
  'ก.ค.', 'ส.ค.', 'ก.ย.', 'ต.ค.', 'พ.ย.', 'ธ.ค.'
]

function formatThaiDate(dateStr: string | Date | undefined): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return String(dateStr)
  const day = d.getDate()
  const month = thaiMonths[d.getMonth()] ?? ''
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  return `${day} ${month} ${hours}:${minutes}`
}

export function useHistory() {
  const { destPath: backupDestPath } = useBackup()

  // Computed Stats
  const stats = computed<HistoryStats>(() => {
    let passed = 0
    let missing = 0
    let deleted = 0

    for (const r of rows.value) {
      if (r.status === 'passed') passed++
      else if (r.status === 'missing') missing++
      else if (r.status === 'deleted') deleted++
    }

    return {
      total: rows.value.length,
      passed,
      missing,
      deleted
    }
  })

  // Filtered rows by status filter and search query
  const filteredRows = computed(() => {
    let list = rows.value

    if (activeFilter.value !== 'all') {
      list = list.filter(r => r.status === activeFilter.value)
    }

    const q = searchQuery.value.trim().toLowerCase()
    if (q) {
      list = list.filter(r => {
        const matchName = r.name.toLowerCase().includes(q)
        const matchDesc = (r.description || '').toLowerCase().includes(q)
        const matchTags = (r.tags || []).some(t => t.toLowerCase().includes(q))
        return matchName || matchDesc || matchTags
      })
    }

    return list
  })

  // Transform raw DB photo record to PhotoHistoryItem
  function mapPhotoRecord(p: any): PhotoHistoryItem {
    const id = p.id || Math.floor(Math.random() * 10000)
    const name = p.file_name || p.filename || 'photo.jpg'
    const sizeBytes = p.size_bytes || 0
    const rawDate = p.backed_up_at || p.created_at
    const desc = p.description || ''
    const tags: string[] = (p.photo_tags || []).map((pt: any) => pt.tag?.name || pt.name).filter(Boolean)
    const hasAi = Boolean(desc.trim() || tags.length > 0)

    let status: 'passed' | 'missing' | 'deleted' = 'passed'
    if (p.status === 'missing') {
      status = 'missing'
    } else if (p.status === 'deleted') {
      status = 'deleted'
    } else {
      status = 'passed'
    }

    const fullPath = currentDest.value ? `${currentDest.value}/${name}` : name

    return {
      id,
      name,
      size: formatBytes(sizeBytes),
      size_bytes: sizeBytes,
      date: formatThaiDate(rawDate),
      raw_date: rawDate,
      ai: hasAi ? 'วิเคราะห์แล้ว' : 'ยังไม่วิเคราะห์',
      description: desc,
      tags,
      status,
      gradient: getGradient(name, id),
      full_path: fullPath,
      sha256: p.sha256
    }
  }

  // Load photos from DB for given destination
  async function loadHistory(targetDest?: string) {
    const dst = targetDest ?? currentDest.value
    if (!dst) return

    isLoading.value = true
    try {
      let rawList: any[] = []

      if (typeof (window as any)?.go?.main?.App?.ListAllDBPhotos === 'function') {
        rawList = await (window as any).go.main.App.ListAllDBPhotos(dst)
      } else if (typeof (window as any)?.go?.service?.BackupService?.ListAllDB === 'function') {
        rawList = await (window as any).go.service.BackupService.ListAllDB(dst)
      } else if (typeof (window as any)?.go?.main?.App?.ListDBPhotos === 'function') {
        rawList = await (window as any).go.main.App.ListDBPhotos(dst)
      }

      if (rawList && Array.isArray(rawList)) {
        rows.value = rawList.map(mapPhotoRecord)
      }
    } catch (err: any) {
      console.error('Failed to load history photos:', err)
      alertNotice.value = {
        type: 'warning',
        message: 'ไม่สามารถโหลดประวัติจากฐานข้อมูลได้: ' + (err?.message || String(err))
      }
    } finally {
      isLoading.value = false
    }
  }

  // Check integrity (Feature 21)
  async function checkIntegrity(targetDest?: string) {
    if (isChecking.value) return
    const dst = targetDest ?? currentDest.value
    if (!dst) {
      alertNotice.value = {
        type: 'warning',
        message: 'กรุณาเลือกโฟลเดอร์ปลายทางก่อนตรวจสอบความถูกต้อง'
      }
      return
    }

    isChecking.value = true
    const startTime = performance.now()

    try {
      let result: IntegrityResult | null = null

      if (typeof (window as any)?.go?.main?.App?.CheckIntegrity === 'function') {
        result = await (window as any).go.main.App.CheckIntegrity(dst)
      } else if (typeof (window as any)?.go?.service?.BackupService?.CheckIntegrity === 'function') {
        result = await (window as any).go.service.BackupService.CheckIntegrity(dst)
      } else {
        // Fallback simulation for browser development
        await new Promise(resolve => setTimeout(resolve, 400))
        const missingCount = rows.value.filter(r => r.status === 'missing').length
        result = {
          total_db: rows.value.length,
          total_disk: rows.value.filter(r => r.status !== 'missing').length,
          missing_in_disk: rows.value.filter(r => r.status === 'missing').map(r => r.name),
          matched_count: rows.value.filter(r => r.status === 'passed').length,
          is_exact_match: missingCount === 0,
          duration_formatted: '10.2 ms'
        }
      }

      const elapsed = (performance.now() - startTime).toFixed(1)
      lastCheckDuration.value = result?.duration_formatted || `${elapsed} ms`

      // Refresh records to show updated status from DB
      await loadHistory(dst)

      if (result && result.missing_in_disk && result.missing_in_disk.length > 0) {
        alertNotice.value = {
          type: 'warning',
          message: `พบไฟล์สูญหาย ${result.missing_in_disk.length} ไฟล์: มีบันทึกใน DB แต่ไม่พบในโฟลเดอร์จริง!`
        }
      } else {
        alertNotice.value = {
          type: 'success',
          message: `ตรวจสอบความถูกต้องเรียบร้อย: ไฟล์ในฐานข้อมูลและในโฟลเดอร์ตรงกันทั้งหมด (${lastCheckDuration.value})`
        }
      }
    } catch (err: any) {
      console.error('Integrity check error:', err)
      const elapsed = (performance.now() - startTime).toFixed(1)
      lastCheckDuration.value = `${elapsed} ms`
      alertNotice.value = {
        type: 'warning',
        message: 'เกิดข้อผิดพลาดในการตรวจสอบ: ' + (err?.message || String(err))
      }
    } finally {
      isChecking.value = false
    }
  }

  // Change destination folder (Feature 24)
  async function changeDest() {
    try {
      if (typeof (window as any)?.go?.main?.App?.SelectDestFolder === 'function') {
        const res = await (window as any).go.main.App.SelectDestFolder()
        // If native dialog returned, update currentDest from backend GetDestPath
        const newDest = await (window as any).go.main.App.GetDestPath()
        if (newDest && newDest !== currentDest.value) {
          currentDest.value = newDest
          backupDestPath.value = newDest
          if (res?.duration_formatted) {
            lastCheckDuration.value = res.duration_formatted
          }
          await loadHistory(newDest)
          alertNotice.value = {
            type: 'info',
            message: `เปลี่ยนโฟลเดอร์ปลายทางเป็น: ${newDest}`
          }
          return
        }
      } else {
        // Show manual path modal in browser preview
        manualDestInput.value = currentDest.value
        showChangeDestModal.value = true
      }
    } catch (err: any) {
      console.error('Failed to change destination:', err)
      manualDestInput.value = currentDest.value
      showChangeDestModal.value = true
    }
  }

  // Manually apply new destination path
  async function applyManualDest(path: string) {
    const trimmed = path.trim()
    if (!trimmed) return
    currentDest.value = trimmed
    backupDestPath.value = trimmed
    showChangeDestModal.value = false

    try {
      if (typeof (window as any)?.go?.main?.App?.SetDestPath === 'function') {
        const res = await (window as any).go.main.App.SetDestPath(trimmed)
        if (res?.duration_formatted) {
          lastCheckDuration.value = res.duration_formatted
        }
      }
    } catch (e) {
      console.warn('SetDestPath warning:', e)
    }

    await loadHistory(trimmed)
    await checkIntegrity(trimmed)
  }

  // Delete a single file (Feature 23)
  async function confirmDeleteSingle(photo: PhotoHistoryItem) {
    photoToDelete.value = photo
  }

  async function executeDeleteSingle() {
    if (!photoToDelete.value) return
    const target = photoToDelete.value
    isDeleting.value = true

    try {
      if (typeof (window as any)?.go?.main?.App?.DeletePhotosFromDest === 'function') {
        await (window as any).go.main.App.DeletePhotosFromDest(currentDest.value, [target.name], false)
      } else if (typeof (window as any)?.go?.main?.App?.DeletePhotos === 'function') {
        await (window as any).go.main.App.DeletePhotos([target.name], false)
      } else if (typeof (window as any)?.go?.service?.BackupService?.Delete === 'function') {
        await (window as any).go.service.BackupService.Delete([target.name], false)
      }

      // Update in-memory row status to 'deleted'
      const found = rows.value.find(r => r.id === target.id || r.name === target.name)
      if (found) {
        found.status = 'deleted'
      }

      alertNotice.value = {
        type: 'info',
        message: `ลบไฟล์ "${target.name}" เรียบร้อย และบันทึกสถานะ 'ลบแล้ว' ในฐานข้อมูล`
      }
    } catch (err: any) {
      console.error('Failed to delete photo:', err)
      alertNotice.value = {
        type: 'warning',
        message: 'เกิดข้อผิดพลาดในการลบไฟล์: ' + (err?.message || String(err))
      }
    } finally {
      isDeleting.value = false
      photoToDelete.value = null
    }
  }

  // Delete all files in destination (Feature 25)
  async function executeDeleteAll() {
    isDeleting.value = true
    try {
      let count = 0
      if (typeof (window as any)?.go?.main?.App?.DeletePhotosFromDest === 'function') {
        count = await (window as any).go.main.App.DeletePhotosFromDest(currentDest.value, [], true)
      } else if (typeof (window as any)?.go?.main?.App?.DeletePhotos === 'function') {
        count = await (window as any).go.main.App.DeletePhotos([], true)
      } else if (typeof (window as any)?.go?.service?.BackupService?.Delete === 'function') {
        count = await (window as any).go.service.BackupService.Delete([], true)
      }

      // Mark all rows as deleted
      for (const r of rows.value) {
        r.status = 'deleted'
      }

      alertNotice.value = {
        type: 'info',
        message: `ลบไฟล์ทั้งหมดในปลายทางเรียบร้อย (${count} ไฟล์) และอัปเดตสถานะในฐานข้อมูลเป็น 'ลบแล้ว'`
      }
    } catch (err: any) {
      console.error('Failed to delete all photos:', err)
      alertNotice.value = {
        type: 'warning',
        message: 'เกิดข้อผิดพลาดในการลบไฟล์ทั้งหมด: ' + (err?.message || String(err))
      }
    } finally {
      isDeleting.value = false
      showDeleteAllModal.value = false
    }
  }

  // Initialize history state on component mount
  async function initHistory() {
    // 1. Get current destination from backend or fallback to backupDestPath
    if (!currentDest.value) {
      if (typeof (window as any)?.go?.main?.App?.GetDestPath === 'function') {
        const dst = await (window as any).go.main.App.GetDestPath()
        if (dst) {
          currentDest.value = dst
        }
      }
      if (!currentDest.value && backupDestPath.value) {
        currentDest.value = backupDestPath.value
      }
      if (!currentDest.value) {
        currentDest.value = '~/Pictures/PhotoBackup'
      }
    }

    // 2. Load DB records
    await loadHistory()

    // 3. If there are DB records or a real destination, perform check
    if (currentDest.value && currentDest.value !== '~/Pictures/PhotoBackup') {
      await checkIntegrity()
    }
  }

  return {
    // State
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

    // Modals state
    selectedPhotoForDetail,
    photoToDelete,
    showDeleteAllModal,
    showChangeDestModal,
    manualDestInput,

    // Actions
    initHistory,
    loadHistory,
    checkIntegrity,
    changeDest,
    applyManualDest,
    confirmDeleteSingle,
    executeDeleteSingle,
    executeDeleteAll
  }
}
