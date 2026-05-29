import { defineStore } from "pinia"
import type { Link } from "~/types/link"

export interface CreateLinkInput {
  url: string
  title?: string
  description?: string
  folder_id?: string
  clear_folder?: boolean
  image?: string
  tags?: string[]
}

export const useLinksStore = defineStore("links", () => {
  const links = ref<Link[]>([])
  const isLoading = ref(false)

  // fetchAllLinks lists the user's links, optionally filtered by a search query.
  // Pass an AbortSignal to make a superseding search cancel this one: on abort
  // we leave isLoading true on purpose — the newer request now owns it, and
  // resetting here would flicker the loader off while it's still in flight.
  async function fetchAllLinks(q = "", signal?: AbortSignal): Promise<void> {
    const { $api } = useNuxtApp()
    isLoading.value = true
    try {
      links.value = await $api<Link[]>("/links", {
        query: q ? { q } : undefined,
        signal,
      })
      isLoading.value = false
    } catch (err) {
      if (!signal?.aborted) isLoading.value = false
      throw err
    }
  }

  async function createLink(input: CreateLinkInput): Promise<void> {
    const { $api } = useNuxtApp()
    const created = await $api<Link>("/links", {
      method: "POST",
      body: input,
    })
    links.value.push(created)
  }

  async function deleteLink(id: string): Promise<void> {
    const { $api } = useNuxtApp()
    await $api(`/links/${id}`, {
      method: "DELETE",
    })
    const idx = links.value.findIndex((link) => link.id === id)
    if (idx !== -1) links.value.splice(idx, 1)
  }

  async function updateLink(id: string, input: CreateLinkInput): Promise<void> {
    const { $api } = useNuxtApp()
    const updated = await $api<Link>(`/links/${id}`, {
      method: "PATCH",
      body: input,
    })
    const idx = links.value.findIndex((link) => link.id === id)
    if (idx !== -1) links.value[idx] = updated
  }

  return { links, isLoading, fetchAllLinks, createLink, deleteLink, updateLink }
})
