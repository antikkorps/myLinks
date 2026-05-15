<script setup lang="ts">
const store = useAuthStore()
const user = computed(() => store.user?.email ?? "")

async function logout() {
  await store.logout()
  await navigateTo("/login")
}
</script>

<template>
  <div>
    <h1>Home</h1>
    <p>Welcome {{ user }}</p>
    <button @click="logout">Logout</button>
  </div>
</template>
