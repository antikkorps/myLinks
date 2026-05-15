<script setup lang="ts">
definePageMeta({ auth: false })

const { email, password, error, isLoading, submit } = useLoginForm()
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 p-4">
    <form
      class="w-full max-w-sm bg-white rounded-lg shadow p-8 space-y-6"
      @submit.prevent="submit"
    >
      <h1 class="text-2xl font-semibold text-center">Connexion</h1>

      <div class="space-y-2">
        <label for="email" class="block text-sm font-medium">Email</label>
        <InputText
          id="email"
          v-model="email"
          type="email"
          autocomplete="email"
          required
          class="w-full"
        />
      </div>

      <div class="space-y-2">
        <label for="password" class="block text-sm font-medium">Mot de passe</label>
        <Password
          id="password"
          v-model="password"
          :feedback="false"
          toggle-mask
          autocomplete="current-password"
          required
          input-class="w-full"
          class="w-full"
        />
      </div>

      <Message v-if="error" severity="error" :closable="false">
        {{ error }}
      </Message>

      <Button
        type="submit"
        label="Se connecter"
        :loading="isLoading"
        class="w-full"
      />
    </form>
  </div>
</template>
