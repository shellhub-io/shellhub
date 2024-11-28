<template>
  <SettingDrawer v-model="showNavigationDrawer" />

  <div>
    <router-view />
  </div>
</template>

<script setup lang="ts">
import SettingDrawer from "@/components/Setting/SettingDrawer.vue";

const showNavigationDrawer = defineModel();
</script>
