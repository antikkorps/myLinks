<script setup lang="ts">
definePageMeta({ layout: "auth", auth: false })

const { email, password, firstname, lastname, error, isLoading, submit } =
  useRegisterForm()
</script>

<template>
  <form
    class="w-full max-w-sm bg-white rounded-lg shadow p-8 space-y-6"
    @submit.prevent="submit"
  >
    <h1 class="text-2xl font-semibold text-center">Créer un compte</h1>

    <div class="space-y-2">
      <label for="firstname" class="block text-sm font-medium">Prénom</label>
      <InputText
        id="firstname"
        v-model="firstname"
        type="text"
        autocomplete="given-name"
        required
        class="w-full"
      />
    </div>

    <div class="space-y-2">
      <label for="lastname" class="block text-sm font-medium">Nom</label>
      <InputText
        id="lastname"
        v-model="lastname"
        type="text"
        autocomplete="family-name"
        required
        class="w-full"
      />
    </div>

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
        autocomplete="new-password"
        required
        minlength="8"
        input-class="w-full"
        class="w-full"
      />
    </div>

    <Message v-if="error" severity="error" :closable="false">
      {{ error }}
    </Message>

    <Button type="submit" label="Créer mon compte" :loading="isLoading" class="w-full" />

    <p class="text-sm text-center text-gray-600">
      Déjà un compte ?
      <NuxtLink to="/login" class="text-blue-600 hover:underline">
        Se connecter
      </NuxtLink>
    </p>
  </form>
</template>
