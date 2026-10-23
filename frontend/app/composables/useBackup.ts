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

export interface BackupQueueItem {
  id: string
  name: string
  type: 'backup' | 'move'
  targets: string[]
  isAll: boolean
  totalCount: number
  status: 'waiting' | 'running' | 'completed' | 'failed'
  createdAt: number
}

// FIFO Queue State: Ensures subsequent backups/moves wait until the current one finishes
const backupQueue = ref<BackupQueueItem[]>([])
const activeQueueItem = ref<BackupQueueItem | null>(null)
let queueWorkerRunning = false
let resolveCurrentBatchAi: (() => void) | null = null

// Fallback timer: closes the AI popup if no progress event arrives within 3 minutes
let aiTimeoutHandle: ReturnType<typeof setTimeout> | null = null

function startAiTimeout() {
  if (aiTimeoutHandle) clearTimeout(aiTimeoutHandle)
  aiTimeoutHandle = setTimeout(() => {
    isAnalyzing.value = false
    aiTimeoutHandle = null
  }, 3 * 60 * 1000)
}

function clearAiTimeout() {
  if (aiTimeoutHandle) {
    clearTimeout(aiTimeoutHandle)
    aiTimeoutHandle = null
  }
}

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

let opStartTime = 0
let timerInterval: any = null

function formatDurationPretty(ms: number): string {
  const totalSeconds = ms / 1000
  if (totalSeconds < 60) {
    return `${totalSeconds.toFixed(1)} วินาที`
  }
  const mins = Math.floor(totalSeconds / 60)
  const secs = (totalSeconds % 60).toFixed(1)
  return `${mins} นาที ${secs} วินาที`
}

function startLiveTimer() {
  if (timerInterval) clearInterval(timerInterval)
  timerInterval = setInterval(() => {
    if (!opStartTime || !isAnalyzing.value) {
      clearInterval(timerInterval)
      timerInterval = null
      return
    }
    const elapsed = Date.now() - opStartTime
    lastSummary.value.duration_formatted = `${formatDurationPretty(elapsed)} (กำลังวิเคราะห์...)`
  }, 500)
}

function stopLiveTimer() {
  if (timerInterval) {
    clearInterval(timerInterval)
    timerInterval = null
  }
  if (opStartTime > 0) {
    const elapsed = Date.now() - opStartTime
    lastSummary.value.duration_formatted = formatDurationPretty(elapsed)
  }
}

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

  // Sequential Queue Runner: processes items in strict FIFO order
  async function runQueueWorker() {
    if (queueWorkerRunning) return
    queueWorkerRunning = true

    try {
      while (backupQueue.value.length > 0) {
        const item = backupQueue.value[0]
        if (!item) break

        item.status = 'running'
        activeQueueItem.value = item
        isProcessing.value = true
        opStartTime = Date.now()

        const actionName = item.type === 'move' ? 'ย้าย' : 'สำรอง'
        const isMove = item.type === 'move'

        try {
          const app = (window as any)?.go?.main?.App
          let summary: any = null

          if (isMove && typeof app?.MovePhotos === 'function') {
            summary = await app.MovePhotos(item.targets, item.isAll)
          } else if (!isMove && typeof app?.BackupPhotos === 'function') {
            summary = await app.BackupPhotos(item.targets, item.isAll)
          } else if (typeof app?.MovePhotos === 'function') {
            summary = await app.MovePhotos(item.targets, item.isAll)
          }

          if (summary) {
            lastSummary.value = {
              total_files: summary.total_files || item.totalCount,
              moved_files: summary.moved_files || 0,
              skipped_files: summary.skipped_files || 0,
              failed_files: summary.failed_files || 0,
              duration_ms: summary.duration_ms || 0,
              duration_formatted: summary.duration_formatted || `${summary.duration_ms} ms`,
              has_run: true
            }

            // If files were moved/backed up, wait for AI analysis to finish before moving to next queue item!
            if ((summary.moved_files || 0) > 0) {
              isAnalyzing.value = true
              aiProgress.value = { current: 0, total: summary.moved_files || 0, filename: '' }
              startLiveTimer()

              // Wait until AI analysis of this batch completes so next task never skips the line!
              await new Promise<void>((resolve) => {
                let resolved = false
                const done = () => {
                  if (!resolved) {
                    resolved = true
                    resolveCurrentBatchAi = null
                    resolve()
                  }
                }
                resolveCurrentBatchAi = done

                // Fallback timeout in case AI auto-analyze is disabled in settings
                setTimeout(() => {
                  if (!isAnalyzing.value || aiProgress.value.total === 0) {
                    done()
                  }
                }, 2000)
              })
            } else {
              stopLiveTimer()
            }

            aiError.value = null
            await refreshSourceFiles()

            if (typeof app?.CheckIntegrity === 'function') {
              integrityResult.value = await app.CheckIntegrity(destPath.value)
              destCount.value = integrityResult.value?.total_disk || 0
            }
            await fetchDbPhotos()

            const suffix = isMove ? ' (ลบออกจากต้นทางแล้ว)' : ' (เก็บไฟล์ต้นทางไว้)'
            const scopeText = item.isAll ? 'ทั้งหมด' : 'ที่เลือก'
            successMessage.value = `${actionName}รูปภาพ${scopeText}สำเร็จ ${summary.moved_files} ไฟล์${suffix} (ข้ามไฟล์ซ้ำ ${summary.skipped_files} ไฟล์) ใช้เวลา ${summary.duration_formatted}`
            showSuccessModal.value = true
          } else {
            // Browser preview fallback
            const count = item.totalCount
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
            if (isMove) {
              sourceFiles.value = sourceFiles.value.filter(f => !item.targets.includes(f.filename))
            }
            successMessage.value = `จำลองการ${actionName}ไฟล์สำเร็จ ${count} รูป (เวลา ${lastSummary.value.duration_formatted})`
            showSuccessModal.value = true
          }
          item.status = 'completed'
        } catch (err: any) {
          item.status = 'failed'
          const errorMsg = err?.message || String(err)
          aiError.value = errorMsg
          alert(`เกิดข้อผิดพลาดในการ${actionName}: ` + errorMsg)
        } finally {
          // Remove completed/failed item from queue
          backupQueue.value.shift()
        }
      }
    } finally {
      activeQueueItem.value = null
      isProcessing.value = false
      queueWorkerRunning = false
    }
  }

  // Helper for both Backup (Copy) and Move (Cut) - Queues work in FIFO order
  async function executeTransfer(isMove: boolean, isAll: boolean) {
    const actionName = isMove ? 'ย้าย' : 'สำรอง'
    if (!sourcePath.value) {
      alert(`กรุณาเลือกโฟลเดอร์ต้นทาง (Flash Drive / โทรศัพท์) ก่อนทำการ${actionName}ไฟล์!`)
      return
    }
    if (!destPath.value) {
      alert(`กรุณาเลือกโฟลเดอร์ปลายทางสำหรับเก็บภาพก่อนทำการ${actionName}ไฟล์!`)
      return
    }
    if (!isAll && selectedFiles.value.length === 0) {
      alert(`กรุณาเลือกรูปภาพอย่างน้อย 1 ไฟล์ที่ต้องการ${actionName}!`)
      return
    }
    if (isAll && sourceFiles.value.length === 0) {
      alert(`ไม่พบรูปภาพในโฟลเดอร์ต้นทางที่จะ${actionName}!`)
      return
    }

    if (isMove) {
      const countText = isAll ? `ทั้งหมด ${sourceFiles.value.length}` : `${selectedFiles.value.length}`
      if (!confirm(`คุณแน่ใจหรือไม่ว่าต้องการ "ย้าย" รูปภาพจำนวน ${countText} ไฟล์ไปยังปลายทาง?\n(ไฟล์ต้นทางในโฟลเดอร์ต้นทางจะถูกลบออกหลังจากย้ายสำเร็จ)`)) {
        return
      }
    }

    // Capture selected files and clear selection immediately so user can select more photos!
    const targetsToRun = isAll ? [] : [...selectedFiles.value]
    const count = isAll ? sourceFiles.value.length : targetsToRun.length
    selectedFiles.value = []

    const item: BackupQueueItem = {
      id: 'task_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
      name: isAll ? `${actionName}ทั้งหมด (${count} ไฟล์)` : `${actionName}ที่เลือก (${count} ไฟล์)`,
      type: isMove ? 'move' : 'backup',
      targets: targetsToRun,
      isAll,
      totalCount: count,
      status: 'waiting',
      createdAt: Date.now()
    }

    backupQueue.value.push(item)

    // Trigger queue runner if not already running
    if (!queueWorkerRunning) {
      runQueueWorker()
    }
  }

  // Backup (Copy) actions
  function backupSelectedPhotos() {
    return executeTransfer(false, false)
  }

  function backupAllPhotos() {
    return executeTransfer(false, true)
  }

  // Move (Cut/Transfer) actions
  function moveSelectedPhotos() {
    return executeTransfer(true, false)
  }

  function moveAllPhotos() {
    return executeTransfer(true, true)
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
          }).sort((a, b) => {
            if (a.description && !b.description) return -1
            if (!a.description && b.description) return 1
            return 0
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

    // Listen to AI analysis progress events from the Go background goroutine
    if (typeof (window as any)?.runtime?.EventsOn === 'function') {
      if (typeof (window as any)?.runtime?.EventsOff === 'function') {
        ;(window as any).runtime.EventsOff('ai_analysis_progress')
        ;(window as any).runtime.EventsOff('ai_analysis_error')
      }
      let lastFetchTime = 0
      ;(window as any).runtime.EventsOn('ai_analysis_progress', (data: any) => {
        clearAiTimeout()
        const cur = data?.current || 0
        const tot = data?.total || 0
        const isDone = Boolean(data?.done) || (tot > 0 && cur >= tot)

        aiProgress.value = {
          current: isDone ? tot : cur,
          total: tot,
          filename: data?.filename || ''
        }

        // Live update: refresh DB records periodically or if user has results modal open
        const now = Date.now()
        if (isDone || showAiResultsModal.value || (now - lastFetchTime > 1200)) {
          lastFetchTime = now
          fetchDbPhotos()
        }

        if (isDone) {
          stopLiveTimer()
          if (resolveCurrentBatchAi) {
            resolveCurrentBatchAi()
            resolveCurrentBatchAi = null
          }
          setTimeout(() => {
            if (!activeQueueItem.value) {
              isAnalyzing.value = false
            }
          }, 800)
        } else {
          isAnalyzing.value = true
          // Reset watchdog for each event so we only fire if Go goes silent
          startAiTimeout()
        }
      })

      // Listen for per-file AI errors emitted by the Go background goroutine
      ;(window as any).runtime.EventsOn('ai_analysis_error', (data: any) => {
        stopLiveTimer()
        const filename = data?.filename || 'ไฟล์ไม่ทราบชื่อ'
        const errMsg = data?.error || 'เกิดข้อผิดพลาดที่ไม่ทราบสาเหตุ'
        aiError.value = `วิเคราะห์ภาพ "${filename}" ไม่สำเร็จ: ${errMsg}`
        // Close the popup immediately on first error
        clearAiTimeout()
        isAnalyzing.value = false
        if (resolveCurrentBatchAi) {
          resolveCurrentBatchAi()
          resolveCurrentBatchAi = null
        }
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
    backupQueue,
    activeQueueItem,

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
    backupSelectedPhotos,
    backupAllPhotos,
    moveSelectedPhotos,
    moveAllPhotos,
    fetchDbPhotos,
    analyzeSinglePhoto,
    initBackupState,
    stopLiveTimer
  }
}
