<template>
  <v-card class="bg-v-theme-surface mx-auto py-6 border text-center">
    <v-card-title class="d-flex flex-column justify-center pa-5 ga-6">
      <div>
        <v-icon size="x-large" data-test="message-icon" :icon="icon" />
      </div>
      <div class="text-sm-h5 text-wrap font-weight-medium" data-test="message-title">
        Looks like you don't have any {{ item }}
      </div>
    </v-card-title>

    <v-card-text class="pb-0 px-4 py-1 d-flex flex-column ga-2 text-subtitle-1" data-test="message-content">
      <slot name="content" />
    </v-card-text>

    <v-card-actions class="justify-center pt-8 pb-0">
      <slot name="action" />
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
const { icon, item } = defineProps<{
  icon: string;
  item: string;
}>();
</script>
