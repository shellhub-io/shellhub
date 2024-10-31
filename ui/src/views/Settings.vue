<template>
  <SettingDrawer v-model="showNavigationDrawer" />

  <v-card
    flat
    border
    class="bg-v-theme-surface"
  >
    <v-toolbar
      density="compact"
      class="bg-v-theme-surface"
    >
      <v-app-bar-nav-icon
        class="hidden-lg-and-up"
        @click.stop="showNavigationDrawer = !showNavigationDrawer"
        aria-label="Toggle Menu"
        density="comfortable"
      />
    </v-toolbar>
    <router-view />
  </v-card>
</template>

<script setup lang="ts">
import SettingDrawer from "@/components/Setting/SettingDrawer.vue";

const showNavigationDrawer = defineModel();
</script>
