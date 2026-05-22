import { defineStore } from "pinia"
import type { Folder } from "~/types/folders"

export interface CreateFolderInput {
  name: string
}

export const useFoldersStore = defineStore("folders", () => {
  const folders = ref<Folder[]>([])
  const isLoading = ref(false)

  async function fetchAllFolders(): Promise<void> {
    const { $api } = useNuxtApp()
    isLoading.value = true
    try {
      folders.value = await $api<Folder[]>("/folders")
    } finally {
      isLoading.value = false
    }
  }

  async function createFolder(input: CreateFolderInput): Promise<void> {
    const { $api } = useNuxtApp()
    const created = await $api<Folder>("/folders", {
      method: "POST",
      body: input,
    })
    folders.value.push(created)
  }

  async function updateFolder(id: string, input: CreateFolderInput): Promise<void> {
    const { $api } = useNuxtApp()
    const updated = await $api<Folder>(`/folders/${id}`, {
      method: "PUT",
      body: input,
    })
    const idx = folders.value.findIndex((folder) => folder.id === id)
    if (idx !== -1) folders.value[idx] = updated
  }

  async function deleteFolder(id: string): Promise<void> {
    const { $api } = useNuxtApp()
    await $api(`/folders/${id}`, {
      method: "DELETE",
    })
    const idx = folders.value.findIndex((folder) => folder.id === id)
    if (idx !== -1) folders.value.splice(idx, 1)
  }

  return { folders, isLoading, fetchAllFolders, createFolder, updateFolder, deleteFolder }
})
