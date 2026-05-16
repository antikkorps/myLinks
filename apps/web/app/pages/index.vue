<script setup lang="ts">
const store = useAuthStore()
const user = computed(() => store.user?.email ?? "")
</script>

<template>
  <div class="max-w-3xl mx-auto p-6">
    <h1 class="text-2xl font-semibold">Home</h1>
    <p class="mt-2">Welcome {{ user }}</p>
  </div>
</template>
