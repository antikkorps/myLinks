<script setup lang="ts">
import { FetchError } from "ofetch"
import type { CreateLinkInput } from "~/stores/links"

const store = useLinksStore()

const showCreate = ref(false)
const isSubmitting = ref(false)
const serverError = ref<string | null>(null)

onMounted(() => {
  store.fetchAllLinks()
})

function openCreate() {
  serverError.value = null
  showCreate.value = true
}

async function onCreate(input: CreateLinkInput) {
  serverError.value = null
  isSubmitting.value = true
  try {
    await store.createLink(input)
    showCreate.value = false
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 400) {
      serverError.value = errorMessages.url_required ?? "URL required"
    } else {
      serverError.value = errorMessages.network_error ?? "Network error"
    }
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="max-w-3xl mx-auto p-6">
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-semibold">Links</h1>
      <Button label="New link" icon="pi pi-plus" @click="openCreate" />
    </div>

    <div v-if="store.isLoading" class="text-gray-500">Loading...</div>

    <div v-else-if="store.links.length === 0" class="text-gray-500">
      No links yet.
    </div>

    <div v-else class="flex flex-col gap-4">
      <Card v-for="link in store.links" :key="link.id">
        <template #title>
          <a
            :href="link.url"
            target="_blank"
            rel="noopener noreferrer"
            class="hover:underline"
          >
            {{ link.title ?? link.url }}
          </a>
        </template>
        <template #subtitle>
          <span class="text-sm text-gray-500 break-all">{{ link.url }}</span>
        </template>
        <template #content>
          <p v-if="link.description" class="mb-3">{{ link.description }}</p>
          <div v-if="link.tags.length > 0" class="flex flex-wrap gap-2">
            <Tag
              v-for="tag in link.tags"
              :key="tag.id"
              :value="tag.name"
              severity="secondary"
            />
          </div>
        </template>
      </Card>
    </div>

    <Dialog
      v-model:visible="showCreate"
      modal
      header="New link"
      :style="{ width: '32rem' }"
      :closable="!isSubmitting"
    >
      <LinkForm
        :is-submitting="isSubmitting"
        :server-error="serverError"
        @submit="onCreate"
        @cancel="showCreate = false"
      />
    </Dialog>
  </div>
</template>
