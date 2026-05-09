<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-50 p-8">
    <div class="max-w-md w-full space-y-4 p-8 bg-white rounded-xl shadow">
      <h1 class="text-2xl font-bold">myLinks</h1>
      <p class="text-slate-600">PrimeVue + Tailwind smoke test.</p>
      <Button label="Click me" @click="count++" />
      <p class="text-sm">Count: {{ count }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
const count = ref(0)
</script>
