<script setup lang="ts">
import { RouterView } from 'vue-router'
import NavigationMenu from '@/components/NavigationMenu.vue'
</script>

<template>
  <v-app>
    <NavigationMenu />
    <v-main>
      <RouterView />
    </v-main>
  </v-app>
</template>
