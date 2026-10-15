import { ref, computed } from 'vue'
import type { PhotoItem, FileItem, MoveSummary, IntegrityResult } from '~/models'

export type { PhotoItem, FileItem, MoveSummary, IntegrityResult }

// Global Shared Reactive State for Backup System (Feature B)
const sourcePath = ref('')
const destPath = ref('')
const destCount = ref(0)
const sourceType = ref<'usb' | 'phone'>('usb')
const isProcessing = ref(false)

const sourceFiles = ref<PhotoItem[]>([])
const selectedFiles = ref<string[]>([])
const thumbnailMap = ref<Record<string, string>>({})
const destFilesList = ref<FileItem[]>([])

const showMissingModal = ref(false)
const showDestFilesModal = ref(false)
const showAiResultsModal = ref(false)
const showSuccessModal = ref(false)
const successMessage = ref('')

const isAnalyzing = ref(false)
const aiProgress = ref({ current: 0, total: 0, filename: '' })
const aiError = ref<string | null>(null)
const analyzedPhotosList = ref<Array<{ filename: string; description: string; tags: string[]; path?: string }>>([])

const integrityResult = ref<IntegrityResult | null>(null)
const lastSummary = ref<MoveSummary>({
  total_files: 0,
  moved_files: 0,
  skipped_files: 0,
  failed_files: 0,
  duration_ms: 0,
  duration_formatted: '-',
  has_run: false
})

export function useBackup() {
  // Computed values
  const sourceTotalBytes = computed(() => {
    return sourceFiles.value.reduce((acc, f) => acc + f.size_bytes, 0)
  })

  const displayMissingPreview = computed(() => {
    if (!integrityResult.value || !integrityResult.value.missing_in_disk) return ''
    return integrityResult.value.missing_in_disk.join(', ')
  })

  const allSelected = computed(() => {
    return sourceFiles.value.length > 0 && selectedFiles.value.length === sourceFiles.value.length
  })

  // Selection helpers
  function isSelected(filename: string): boolean {
    return selectedFiles.value.includes(filename)
  }

  function toggleSelect(filename: string) {
    const index = selectedFiles.value.indexOf(filename)
    if (index > -1) {
      selectedFiles.value.splice(index, 1)
    } else {
      selectedFiles.value.push(filename)
    }
  }

  function toggleSelectAll() {
    if (allSelected.value) {
      selectedFiles.value = []
    } else {
      selectedFiles.value = sourceFiles.value.map(f => f.filename)
    }
  }

  function formatTotalSize(bytes: number): string {
    if (!bytes || bytes <= 0) return '0 B'
    if (bytes > 1024 * 1024 * 1024) {
      return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
    }
    if (bytes > 1024 * 1024) {
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    }
    return (bytes / 1024).toFixed(1) + ' KB'
  }

  // --- Backend API Operations via Wails IPC (window.go.main.App) ---

  // Feature 4 & 11: Select Source Folder
  async function selectSourceFolder(folderInputEl?: HTMLInputElement | null) {
    try {
      if (typeof (window as any)?.go?.main?.App?.SelectSourceFolder === 'function') {
        const selected = await (window as any).go.main.App.SelectSourceFolder()
        if (selected) {
          sourcePath.value = selected
          await refreshSourceFiles()
        }
      } else if (folderInputEl) {
        // Browser fallback
        folderInputEl.click()
      }
    } catch (err) {
      console.error('Failed to select source folder:', err)
    }
  }

  // Browser folder upload handler
  function handleBrowserFolderUpload(event: Event) {
    const target = event.target as HTMLInputElement | null
    if (!target || !target.files || target.files.length === 0) return
    const files = Array.from(target.files)
    const firstFile = files[0]
    if (!firstFile) return

    const relativePath = (firstFile as any).webkitRelativePath as string | undefined
    const firstPath = relativePath || firstFile.name || ''
    const parts = firstPath.split('/')
    const folderName: string = (parts.length > 0 && parts[0]) ? parts[0] : 'โฟลเดอร์ที่เลือก'
    sourcePath.value = folderName

    const validExts = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp', '.tif', '.heic', '.raw', '.svg']
    const images = files.filter(f => {
      const name = f.name.toLowerCase()
      return validExts.some(ext => name.endsWith(ext))
    })

    const destLower = new Set(destFilesList.value.map(f => f.filename.toLowerCase()))

    sourceFiles.value = images.map(f => {
      const url = URL.createObjectURL(f)
      thumbnailMap.value[url] = url
      return {
        filename: f.name,
        size_bytes: f.size,
        full_path: url,
        is_duplicate: destLower.has(f.name.toLowerCase()),
        load_failed: false
      }
    })
    selectedFiles.value = sourceFiles.value.filter(f => !f.is_duplicate).map(f => f.filename)
  }

  // Feature 5 & 21: Select Destination Folder
  async function selectDestFolder() {
    try {
      if (typeof (window as any)?.go?.main?.App?.SelectDestFolder === 'function') {
        const res = await (window as any).go.main.App.SelectDestFolder()
        if (res) {
          const dest = await (window as any).go.main.App.GetDestPath()
          destPath.value = dest
          integrityResult.value = res
          destCount.value = res.total_disk || 0
        }
      } else {
        const chosen = prompt('กำหนดโฟลเดอร์ปลายทางสำหรับจัดเก็บรูปภาพ (เช่น D:/PhotoBackup):', destPath.value || 'D:/PhotoBackup')
        if (chosen) {
          destPath.value = chosen
          destCount.value = destFilesList.value.length
        }
      }
    } catch (err) {
      console.error('Failed to select dest folder:', err)
    }
  }

  // Feature 10: List Destination Actual Files (/list dest)
  async function listDestFiles() {
    if (!destPath.value) {
      alert('กรุณาเลือกโฟลเดอร์ปลายทางก่อน!')
      return
    }
    try {
      if (typeof (window as any)?.go?.main?.App?.ListDestFileItems === 'function') {
        const items = await (window as any).go.main.App.ListDestFileItems(destPath.value)
        destFilesList.value = items || []
      }
    } catch (err) {
      console.error('Failed to list dest files:', err)
    }
    showDestFilesModal.value = true
  }

  // Feature 4, 7, 11: Refresh Source Files and Detect Duplicates
  async function refreshSourceFiles() {
    try {
      if (typeof (window as any)?.go?.main?.App?.ListSourceFileItems === 'function') {
        const items = await (window as any).go.main.App.ListSourceFileItems()
        if (items && Array.isArray(items)) {
          const imageItems = items.filter((item: any) => item.is_image)

          let destFiles: string[] = []
          if (destPath.value && typeof (window as any)?.go?.main?.App?.ListDestPhotos === 'function') {
            try {
              destFiles = await (window as any).go.main.App.ListDestPhotos(destPath.value) || []
            } catch (_) {}
          }
          const destLower = new Set((destFiles || []).map((f: string) => f.toLowerCase()))

          sourceFiles.value = imageItems.map((item: any) => ({
            filename: item.filename,
            size_bytes: item.size_bytes,
            full_path: item.full_path,
            is_duplicate: destLower.has(item.filename.toLowerCase()),
            load_failed: false
          }))
          selectedFiles.value = sourceFiles.value.filter(f => !f.is_duplicate).map(f => f.filename)

          // Load real thumbnails via Wails IPC
          const pathsToLoad = sourceFiles.value.map(f => f.full_path).filter((p): p is string => Boolean(p))
          loadThumbnails(pathsToLoad)
        } else {
          sourceFiles.value = []
          selectedFiles.value = []
        }
      }
    } catch (err) {
      console.error('Failed to list source items:', err)
      sourceFiles.value = []
      selectedFiles.value = []
    }
  }

  // Load Real Thumbnails via Wails IPC (Base64)
  async function loadThumbnails(paths: string[]) {
    if (!paths || paths.length === 0) return
    try {
      if (typeof (window as any)?.go?.main?.App?.GetPhotoThumbnails === 'function') {
        const thumbs = await (window as any).go.main.App.GetPhotoThumbnails(paths)
        if (thumbs) {
          thumbnailMap.value = { ...thumbnailMap.value, ...thumbs }
        }
      }
    } catch (err) {
      console.error('Failed to load thumbnails:', err)
    }
  }

  // Feature 6, 7, 8, 9, 11: Move Selected Photos
  async function moveSelectedPhotos() {
    if (!sourcePath.value) {
      alert('กรุณาเลือกโฟลเดอร์ต้นทาง (Flash Drive / โทรศัพท์) ก่อนทำการย้ายไฟล์!')
      return
    }
    if (!destPath.value) {
      alert('กรุณาเลือกโฟลเดอร์ปลายทางสำหรับเก็บภาพก่อนทำการย้ายไฟล์!')
      return
    }
    if (selectedFiles.value.length === 0) {
      alert('กรุณาเลือกรูปภาพอย่างน้อย 1 ไฟล์ที่ต้องการย้าย!')
      return
    }

    isProcessing.value = true
    try {
      if (typeof (window as any)?.go?.main?.App?.MovePhotos === 'function') {
        const summary = await (window as any).go.main.App.MovePhotos(selectedFiles.value, false)
        lastSummary.value = {
          total_files: summary.total_files || selectedFiles.value.length,
          moved_files: summary.moved_files || 0,
          skipped_files: summary.skipped_files || 0,
          failed_files: summary.failed_files || 0,
          duration_ms: summary.duration_ms || 0,
          duration_formatted: summary.duration_formatted || `${summary.duration_ms} ms`,
          has_run: true
        }
        aiError.value = null // Clear any previous errors
        await refreshSourceFiles()
        if (typeof (window as any)?.go?.main?.App?.CheckIntegrity === 'function') {
          integrityResult.value = await (window as any).go.main.App.CheckIntegrity(destPath.value)
          destCount.value = integrityResult.value?.total_disk || 0
        }
        await fetchDbPhotos()
        successMessage.value = `สำรองรูปภาพที่เลือกสำเร็จ ${summary.moved_files} ไฟล์ (ข้ามไฟล์ซ้ำ ${summary.skipped_files} ไฟล์) ใช้เวลา ${summary.duration_formatted}`
        showSuccessModal.value = true
      } else {
        // Browser preview fallback
        const count = selectedFiles.value.length
        destCount.value += count
        lastSummary.value = {
          total_files: count,
          moved_files: count,
          skipped_files: 0,
          failed_files: 0,
          duration_ms: 18.5,
          duration_formatted: '18.50 ms',
          has_run: true
        }
        sourceFiles.value = sourceFiles.value.filter(f => !selectedFiles.value.includes(f.filename))
        selectedFiles.value = []
        successMessage.value = `จำลองการย้ายไฟล์สำเร็จ ${count} รูป (เวลา ${lastSummary.value.duration_formatted})`
        showSuccessModal.value = true
      }
    } catch (err: any) {
      const errorMsg = err?.message || String(err)
      aiError.value = errorMsg
      alert('เกิดข้อผิดพลาดในการย้าย: ' + errorMsg)
    } finally {
      isProcessing.value = false
    }
  }

  // Feature 6, 7, 8, 9, 11: Move All Photos
  async function moveAllPhotos() {
    if (!sourcePath.value) {
      alert('กรุณาเลือกโฟลเดอร์ต้นทาง (Flash Drive / โทรศัพท์) ก่อนทำการสำรองข้อมูล!')
      return
    }
    if (!destPath.value) {
      alert('กรุณาเลือกโฟลเดอร์ปลายทางสำหรับเก็บภาพก่อนทำการสำรองข้อมูล!')
      return
    }
    if (sourceFiles.value.length === 0) {
      alert('ไม่พบรูปภาพในโฟลเดอร์ต้นทางที่จะสำรอง!')
      return
    }

    isProcessing.value = true
    try {
      if (typeof (window as any)?.go?.main?.App?.MovePhotos === 'function') {
        const summary = await (window as any).go.main.App.MovePhotos([], true)
        lastSummary.value = {
          total_files: summary.total_files || sourceFiles.value.length,
          moved_files: summary.moved_files || 0,
          skipped_files: summary.skipped_files || 0,
          failed_files: summary.failed_files || 0,
          duration_ms: summary.duration_ms || 0,
          duration_formatted: summary.duration_formatted || `${summary.duration_ms} ms`,
          has_run: true
        }
        aiError.value = null // Clear any previous errors
        await refreshSourceFiles()
        if (typeof (window as any)?.go?.main?.App?.CheckIntegrity === 'function') {
          integrityResult.value = await (window as any).go.main.App.CheckIntegrity(destPath.value)
          destCount.value = integrityResult.value?.total_disk || 0
        }
        await fetchDbPhotos()
        successMessage.value = `สำรองรูปภาพทั้งหมดสำเร็จ ${summary.moved_files} ไฟล์ (ข้ามไฟล์ซ้ำ ${summary.skipped_files} ไฟล์) ใช้เวลา ${summary.duration_formatted}`
        showSuccessModal.value = true
      } else {
        // Browser preview fallback
        const count = sourceFiles.value.filter(f => !f.is_duplicate).length
        const skipped = sourceFiles.value.filter(f => f.is_duplicate).length
        destCount.value += count
        lastSummary.value = {
          total_files: sourceFiles.value.length,
          moved_files: count,
          skipped_files: skipped,
          failed_files: 0,
          duration_ms: 32.4,
          duration_formatted: '32.40 ms',
          has_run: true
        }
        sourceFiles.value = sourceFiles.value.filter(f => f.is_duplicate)
        selectedFiles.value = []
        alert(`จำลองการสำรองทั้งหมดสำเร็จ ${count} รูป (ข้ามไฟล์ซ้ำ ${skipped} รูป) เวลา ${lastSummary.value.duration_formatted}`)
      }
    } catch (err: any) {
      const errorMsg = err?.message || String(err)
      aiError.value = errorMsg
      alert('เกิดข้อผิดพลาดในการย้าย: ' + errorMsg)
    } finally {
      isProcessing.value = false
    }
  }

  // Fetch photos from DB including AI descriptions and tags
  async function fetchDbPhotos() {
    if (!destPath.value) return []
    try {
      if (typeof (window as any)?.go?.main?.App?.ListDBPhotos === 'function') {
        const dbPhotos = await (window as any).go.main.App.ListDBPhotos(destPath.value)
        if (dbPhotos && Array.isArray(dbPhotos)) {
          analyzedPhotosList.value = dbPhotos.map((p: any) => {
            const tags = (p.photo_tags || []).map((pt: any) => pt.tag?.name).filter(Boolean)
            return {
              filename: p.file_name,
              description: p.description || '',
              tags: tags,
              path: destPath.value ? `${destPath.value}/${p.file_name}` : p.file_name
            }
          })
          return dbPhotos
        }
      }
    } catch (err) {
      console.error('Failed to fetch DB photos:', err)
    }
    return []
  }

  // Directly analyze any photo via Gemini Vision AI
  async function analyzeSinglePhoto(filePath: string) {
    isAnalyzing.value = true
    try {
      let result = null
      if (typeof (window as any)?.go?.main?.App?.AnalyzePhoto === 'function') {
        result = await (window as any).go.main.App.AnalyzePhoto(filePath)
      } else if (typeof (window as any)?.go?.service?.BackupService?.AnalyzePhoto === 'function') {
        result = await (window as any).go.service.BackupService.AnalyzePhoto(filePath)
      }
      return result
    } catch (err) {
      console.error('Failed to analyze photo:', err)
      throw err
    } finally {
      isAnalyzing.value = false
    }
  }

  // Initialize from live Go backend if available
  async function initBackupState() {
    if (typeof (window as any)?.go?.main?.App?.GetSourcePath === 'function') {
      const src = await (window as any).go.main.App.GetSourcePath()
      if (src) {
        sourcePath.value = src
        await refreshSourceFiles()
      }
    }
    if (typeof (window as any)?.go?.main?.App?.GetDestPath === 'function') {
      const dst = await (window as any).go.main.App.GetDestPath()
      if (dst) {
        destPath.value = dst
        integrityResult.value = await (window as any).go.main.App.CheckIntegrity(dst)
        destCount.value = integrityResult.value?.total_disk || 0
        await fetchDbPhotos()
      }
    }

    // Listen to AI analysis progress events
    if (typeof (window as any)?.runtime?.EventsOn === 'function') {
      (window as any).runtime.EventsOn('ai_analysis_progress', (data: any) => {
        aiProgress.value = {
          current: data.current || 0,
          total: data.total || 0,
          filename: data.filename || ''
        }
        isAnalyzing.value = data.current < data.total
      })
    }
  }

  return {
    // State
    sourceType,
    sourcePath,
    destPath,
    destCount,
    isProcessing,
    sourceFiles,
    selectedFiles,
    thumbnailMap,
    destFilesList,
    showMissingModal,
    showDestFilesModal,
    showAiResultsModal,
    showSuccessModal,
    successMessage,
    isAnalyzing,
    aiProgress,
    aiError,
    analyzedPhotosList,
    integrityResult,
    lastSummary,

    // Computed
    sourceTotalBytes,
    displayMissingPreview,
    allSelected,

    // Helpers
    isSelected,
    toggleSelect,
    toggleSelectAll,
    formatTotalSize,

    // API Actions
    selectSourceFolder,
    handleBrowserFolderUpload,
    selectDestFolder,
    listDestFiles,
    refreshSourceFiles,
    loadThumbnails,
    moveSelectedPhotos,
    moveAllPhotos,
    fetchDbPhotos,
    analyzeSinglePhoto,
    initBackupState
  }
}
