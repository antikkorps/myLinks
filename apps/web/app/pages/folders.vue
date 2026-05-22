<script setup lang="ts">
import { FetchError } from "ofetch"
import { useConfirm } from "primevue/useconfirm"
import type { Folder } from "~/types/folders"

const store = useFoldersStore()
const confirm = useConfirm()

const showCreate = ref(false)
const createName = ref("")
const isCreating = ref(false)
const createError = ref<string | null>(null)

const editingFolder = ref<Folder | null>(null)
const editName = ref("")
const isUpdating = ref(false)
const updateError = ref<string | null>(null)

onMounted(() => {
  store.fetchAllFolders()
})

function mapError(err: unknown): string {
  if (err instanceof FetchError) {
    if (err.statusCode === 400) return errorMessages.invalid_name ?? "Invalid name"
    if (err.statusCode === 404) return errorMessages.not_found ?? "Not found"
  }
  return errorMessages.network_error ?? "Network error"
}

function openCreate() {
  createName.value = ""
  createError.value = null
  showCreate.value = true
}

async function onCreate() {
  createError.value = null
  isCreating.value = true
  try {
    await store.createFolder({ name: createName.value.trim() })
    showCreate.value = false
  } catch (err) {
    createError.value = mapError(err)
  } finally {
    isCreating.value = false
  }
}

function openEdit(folder: Folder) {
  editingFolder.value = folder
  editName.value = folder.name
  updateError.value = null
}

async function onUpdate() {
  if (!editingFolder.value) return
  updateError.value = null
  isUpdating.value = true
  try {
    await store.updateFolder(editingFolder.value.id, { name: editName.value.trim() })
    editingFolder.value = null
  } catch (err) {
    updateError.value = mapError(err)
  } finally {
    isUpdating.value = false
  }
}

function confirmDelete(folder: Folder) {
  confirm.require({
    message: `Delete "${folder.name}" ?`,
    header: "Confirm deletion",
    icon: "pi pi-exclamation-triangle",
    acceptProps: { severity: "danger", label: "Delete" },
    rejectProps: { severity: "secondary", label: "Cancel" },
    accept: async () => {
      try {
        await store.deleteFolder(folder.id)
      } catch (err) {
        console.error("delete folder failed", err)
      }
    },
  })
}
</script>

<template>
  <div class="max-w-3xl mx-auto p-6">
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-semibold">Folders</h1>
      <Button label="New folder" icon="pi pi-plus" @click="openCreate" />
    </div>

    <div v-if="store.isLoading" class="text-gray-500">Loading...</div>

    <div v-else-if="store.folders.length === 0" class="text-gray-500">
      No folders yet.
    </div>

    <div v-else class="flex flex-col gap-2">
      <Card v-for="folder in store.folders" :key="folder.id">
        <template #content>
          <div class="flex items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <i class="pi pi-folder text-gray-500" />
              <span class="font-medium">{{ folder.name }}</span>
            </div>
            <div class="flex gap-1">
              <Button
                icon="pi pi-pencil"
                severity="secondary"
                text
                rounded
                aria-label="Rename folder"
                @click="openEdit(folder)"
              />
              <Button
                icon="pi pi-trash"
                severity="danger"
                text
                rounded
                aria-label="Delete folder"
                @click="confirmDelete(folder)"
              />
            </div>
          </div>
        </template>
      </Card>
    </div>

    <Dialog
      v-model:visible="showCreate"
      modal
      header="New folder"
      :style="{ width: '28rem' }"
      :closable="!isCreating"
    >
      <form class="flex flex-col gap-4" @submit.prevent="onCreate">
        <div class="flex flex-col gap-1">
          <label for="create-name" class="text-sm font-medium">Name</label>
          <InputText
            id="create-name"
            v-model="createName"
            autofocus
            required
            maxlength="100"
            :disabled="isCreating"
          />
        </div>
        <Message v-if="createError" severity="error" :closable="false">
          {{ createError }}
        </Message>
        <div class="flex justify-end gap-2">
          <Button
            type="button"
            label="Cancel"
            severity="secondary"
            :disabled="isCreating"
            @click="showCreate = false"
          />
          <Button
            type="submit"
            label="Create"
            :loading="isCreating"
            :disabled="!createName.trim()"
          />
        </div>
      </form>
    </Dialog>

    <Dialog
      :visible="editingFolder !== null"
      modal
      header="Rename folder"
      :style="{ width: '28rem' }"
      :closable="!isUpdating"
      @update:visible="(v) => { if (!v) editingFolder = null }"
    >
      <form class="flex flex-col gap-4" @submit.prevent="onUpdate">
        <div class="flex flex-col gap-1">
          <label for="edit-name" class="text-sm font-medium">Name</label>
          <InputText
            id="edit-name"
            v-model="editName"
            autofocus
            required
            maxlength="100"
            :disabled="isUpdating"
          />
        </div>
        <Message v-if="updateError" severity="error" :closable="false">
          {{ updateError }}
        </Message>
        <div class="flex justify-end gap-2">
          <Button
            type="button"
            label="Cancel"
            severity="secondary"
            :disabled="isUpdating"
            @click="editingFolder = null"
          />
          <Button
            type="submit"
            label="Save"
            :loading="isUpdating"
            :disabled="!editName.trim()"
          />
        </div>
      </form>
    </Dialog>

    <ConfirmDialog />
  </div>
</template>
