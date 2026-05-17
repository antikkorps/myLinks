import { defineStore } from "pinia"
import type { Link } from "~/types/link"

export interface CreateLinkInput {
  url: string
  title?: string
  description?: string
  folder_id?: string
  image?: string
  tags?: string[]
}

export const useLinksStore = defineStore("links", () => {
  const links = ref<Link[]>([])
  const isLoading = ref(false)

  async function fetchAllLinks(): Promise<void> {
    const { $api } = useNuxtApp()
    isLoading.value = true
    try {
      links.value = await $api<Link[]>("/links")
    } finally {
      isLoading.value = false
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

  return { links, isLoading, fetchAllLinks, createLink }
})
