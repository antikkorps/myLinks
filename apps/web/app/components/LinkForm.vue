<script setup lang="ts">
import type { CreateLinkInput } from "~/stores/links"
import type { Folder } from "~/types/folders"

const props = defineProps<{
  initialValue?: CreateLinkInput
  folders?: Folder[]
  submitLabel?: string
  isSubmitting?: boolean
  serverError?: string | null
}>()

const emit = defineEmits<{
  submit: [input: CreateLinkInput]
  cancel: []
}>()

const url = ref(props.initialValue?.url ?? "")
const title = ref(props.initialValue?.title ?? "")
const description = ref(props.initialValue?.description ?? "")
const folderId = ref<string | null>(props.initialValue?.folder_id ?? null)
const tagsInput = ref(props.initialValue?.tags?.join(", ") ?? "")
const clientError = ref<string | null>(null)

const initialFolderId = props.initialValue?.folder_id ?? null

const displayedError = computed(
  () => clientError.value ?? props.serverError ?? null,
)

function onSubmit() {
  clientError.value = null
  const trimmedUrl = url.value.trim()
  if (!trimmedUrl) {
    clientError.value = errorMessages.url_required ?? "URL required"
    return
  }

  const tags = tagsInput.value
    .split(",")
    .map((t) => t.trim())
    .filter((t) => t.length > 0)

  const folderChanged = folderId.value !== initialFolderId
  const clearFolder = folderChanged && folderId.value === null && initialFolderId !== null

  emit("submit", {
    url: trimmedUrl,
    title: title.value.trim() || undefined,
    description: description.value.trim() || undefined,
    folder_id: folderId.value ?? undefined,
    clear_folder: clearFolder ? true : undefined,
    tags: tags.length > 0 ? tags : undefined,
  })
}
</script>

<template>
  <form class="flex flex-col gap-4" @submit.prevent="onSubmit">
    <div class="flex flex-col gap-1">
      <label for="link-url" class="font-medium">URL</label>
      <InputText
        id="link-url"
        v-model="url"
        type="url"
        required
        placeholder="https://example.com"
      />
    </div>

    <div class="flex flex-col gap-1">
      <label for="link-title" class="font-medium">Title</label>
      <InputText id="link-title" v-model="title" />
    </div>

    <div class="flex flex-col gap-1">
      <label for="link-description" class="font-medium">Description</label>
      <Textarea id="link-description" v-model="description" rows="3" />
    </div>

    <div class="flex flex-col gap-1">
      <label for="link-folder" class="font-medium">Folder</label>
      <Select
        id="link-folder"
        v-model="folderId"
        :options="folders ?? []"
        option-label="name"
        option-value="id"
        placeholder="No folder"
        show-clear
      />
    </div>

    <div class="flex flex-col gap-1">
      <label for="link-tags" class="font-medium">Tags</label>
      <InputText
        id="link-tags"
        v-model="tagsInput"
        placeholder="comma, separated, tags"
      />
    </div>

    <Message v-if="displayedError" severity="error" :closable="false">
      {{ displayedError }}
    </Message>

    <div class="flex justify-end gap-2">
      <Button
        type="button"
        label="Cancel"
        severity="secondary"
        :disabled="isSubmitting"
        @click="emit('cancel')"
      />
      <Button
        type="submit"
        :label="submitLabel ?? 'Create'"
        :loading="isSubmitting"
      />
    </div>
  </form>
</template>
