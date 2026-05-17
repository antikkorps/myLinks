export interface Link {
  id: string
  url: string
  title: string | null
  description: string | null
  created_at: string
  updated_at: string
  folder_id: string | null
  image: string | null
  user_id: string
  tags: Tag[]
}

export interface Tag {
  id: string
  name: string
  created_at: string
  updated_at: string
  user_id: string
}
