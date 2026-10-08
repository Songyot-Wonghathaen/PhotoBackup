export interface Tag {
  id: number
  name: string
}

export interface PhotoTag {
  photo_id: number
  tag_id: number
  source: string
  tag?: Tag
}

export interface Photo {
  id: number
  destination_path: string
  file_name: string
  rel_path: string
  size_bytes: number
  sha256: string
  backed_up_at: string
  status: 'active' | 'missing' | 'deleted' | string
  description?: string
  photo_tags?: PhotoTag[]
}
