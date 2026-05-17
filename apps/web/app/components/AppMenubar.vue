<template>
  <Menubar :model="items">
    <template #start>
      <div class="text-xl font-extrabold text-primary">MyLinks</div>
    </template>
    <template #item="{ item, props, hasSubmenu, root }">
      <a v-ripple class="flex items-center" v-bind="props.action">
        <span>{{ item.label }}</span>
        <Badge
          v-if="item.badge"
          :class="{ 'ml-auto': !root, 'ml-2': root }"
          :value="item.badge"
        />
        <span
          v-if="item.shortcut"
          class="ml-auto border border-surface rounded bg-emphasis text-muted-color text-xs p-1"
          >{{ item.shortcut }}</span
        >
        <i
          v-if="hasSubmenu"
          :class="[
            'pi pi-angle-down ml-auto',
            { 'pi-angle-down': root, 'pi-angle-right': !root },
          ]"
        ></i>
      </a>
    </template>
    <template #end>
      <div class="flex items-center gap-2">
        <span v-if="store.user" class="font-medium">
          {{ store.user.first_name }} {{ store.user.last_name }}
        </span>

        <Button label="Logout" severity="secondary" @click="handleLogout" />
      </div>
    </template>
  </Menubar>
</template>

<script setup lang="ts">
import { ref } from "vue"

const store = useAuthStore()

const handleLogout = async () => {
  await store.logout()
  await navigateTo("/login")
}

const items = ref([
  {
    label: "Home",
    icon: "pi pi-home",
    command: () => navigateTo("/"),
  },
  {
    label: "Links",
    icon: "pi pi-link",
    command: () => navigateTo("/links"),
  },
])
</script>
