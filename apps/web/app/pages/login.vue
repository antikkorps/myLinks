<script setup lang="ts">
definePageMeta({ layout: "auth", auth: false, middleware: "guest" })

const route = useRoute()
const justRegistered = computed(() => route.query.registered === "1")

const { email, password, error, isLoading, submit } = useLoginForm()
</script>

<template>
  <form
    class="w-full max-w-sm bg-white rounded-lg shadow p-8 space-y-6"
    @submit.prevent="submit"
  >
    <h1 class="text-2xl font-semibold text-center">Connexion</h1>

    <Message v-if="justRegistered" severity="success" :closable="false">
      Compte créé, vous pouvez vous connecter.
    </Message>

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

    <Button type="submit" label="Se connecter" :loading="isLoading" class="w-full" />

    <p class="text-sm text-center text-gray-600">
      Pas encore de compte ?
      <NuxtLink to="/register" class="text-blue-600 hover:underline">
        Créer un compte
      </NuxtLink>
    </p>
  </form>
</template>
