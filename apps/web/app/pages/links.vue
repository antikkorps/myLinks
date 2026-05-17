<script setup lang="ts">
import { FetchError } from "ofetch"
import { useConfirm } from "primevue/useconfirm"
import type { CreateLinkInput } from "~/stores/links"
import type { Link } from "~/types/link"

const store = useLinksStore()
const confirm = useConfirm()

const showCreate = ref(false)
const isCreating = ref(false)
const createError = ref<string | null>(null)

const editingLink = ref<Link | null>(null)
const isUpdating = ref(false)
const updateError = ref<string | null>(null)

onMounted(() => {
  store.fetchAllLinks()
})

function openCreate() {
  createError.value = null
  showCreate.value = true
}

async function onCreate(input: CreateLinkInput) {
  createError.value = null
  isCreating.value = true
  try {
    await store.createLink(input)
    showCreate.value = false
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 400) {
      createError.value = errorMessages.url_required ?? "URL required"
    } else {
      createError.value = errorMessages.network_error ?? "Network error"
    }
  } finally {
    isCreating.value = false
  }
}

function openEdit(link: Link) {
  updateError.value = null
  editingLink.value = link
}

function editingInitialValue(link: Link): CreateLinkInput {
  return {
    url: link.url,
    title: link.title ?? undefined,
    description: link.description ?? undefined,
    image: link.image ?? undefined,
    folder_id: link.folder_id ?? undefined,
    tags: link.tags.map((t) => t.name),
  }
}

async function onUpdate(input: CreateLinkInput) {
  if (!editingLink.value) return
  updateError.value = null
  isUpdating.value = true
  try {
    await store.updateLink(editingLink.value.id, input)
    editingLink.value = null
  } catch (err) {
    if (err instanceof FetchError && err.statusCode === 400) {
      updateError.value = errorMessages.url_required ?? "URL required"
    } else {
      updateError.value = errorMessages.network_error ?? "Network error"
    }
  } finally {
    isUpdating.value = false
  }
}

function confirmDelete(link: {
  id: string
  title: string | null
  url: string
}) {
  confirm.require({
    message: `Delete "${link.title ?? link.url}" ?`,
    header: "Confirm deletion",
    icon: "pi pi-exclamation-triangle",
    acceptProps: { severity: "danger", label: "Delete" },
    rejectProps: { severity: "secondary", label: "Cancel" },
    accept: async () => {
      try {
        await store.deleteLink(link.id)
      } catch (err) {
        console.error("delete link failed", err)
      }
    },
  })
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
          <div class="flex items-start justify-between gap-2">
            <a
              :href="link.url"
              target="_blank"
              rel="noopener noreferrer"
              class="hover:underline"
            >
              {{ link.title ?? link.url }}
            </a>
            <div class="flex gap-1">
              <Button
                icon="pi pi-pencil"
                severity="secondary"
                text
                rounded
                aria-label="Edit link"
                @click="openEdit(link)"
              />
              <Button
                icon="pi pi-trash"
                severity="danger"
                text
                rounded
                aria-label="Delete link"
                @click="confirmDelete(link)"
              />
            </div>
          </div>
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
      :closable="!isCreating"
    >
      <LinkForm
        :is-submitting="isCreating"
        :server-error="createError"
        @submit="onCreate"
        @cancel="showCreate = false"
      />
    </Dialog>

    <Dialog
      :visible="editingLink !== null"
      modal
      header="Edit link"
      :style="{ width: '32rem' }"
      :closable="!isUpdating"
      @update:visible="(v) => { if (!v) editingLink = null }"
    >
      <LinkForm
        v-if="editingLink"
        :initial-value="editingInitialValue(editingLink)"
        :is-submitting="isUpdating"
        :server-error="updateError"
        submit-label="Save"
        @submit="onUpdate"
        @cancel="editingLink = null"
      />
    </Dialog>

    <ConfirmDialog />
  </div>
</template>
