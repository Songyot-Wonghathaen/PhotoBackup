export interface PhotoItem {
  filename: string
  size_bytes: number
  is_duplicate: boolean
  full_path?: string
  load_failed?: boolean
}

export interface FileItem {
  filename: string
  size_bytes: number
  mod_time: string
  full_path: string
  is_image: boolean
}

export interface MoveSummary {
  total_files: number
  moved_files: number
  skipped_files: number
  failed_files: number
  duration_ms: number
  duration_formatted: string
  has_run: boolean
  moved_list?: string[]
  skipped_list?: string[]
  failed_list?: string[]
}

export interface IntegrityResult {
  total_db: number
  total_disk: number
  missing_in_disk: string[]
  matched_count: number
  is_exact_match: boolean
  alert_message?: string
}
