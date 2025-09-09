<template>
  <v-card class="bg-v-theme-surface mx-auto py-6 border text-center position-relative">
    <v-skeleton-loader
      class="bg-v-theme-surface position-absolute w-100 h-100 opacity-10 border skeleton-loader"
      type="table-tbody"
    />
    <v-card-title class="d-flex flex-column justify-center pa-5 ga-6 position-relative card-title">
      <div>
        <v-icon size="x-large" data-test="message-icon" :icon="icon" />
      </div>
      <div class="text-sm-h5 text-wrap font-weight-medium" data-test="message-title">
        Looks like you don't have any {{ item }}
      </div>
    </v-card-title>

    <v-card-text
      class="pb-0 px-4 py-1 d-flex flex-column ga-2 text-subtitle-1 position-relative card-content"
      data-test="message-content"
    >
      <slot name="content" />
    </v-card-text>

    <v-card-actions class="justify-center pt-8 pb-0 position-relative card-actions">
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

<style scoped>
.skeleton-loader {
  top: 0;
  left: 0;
  z-index: 0;
}

.card-title,
.card-content,
.card-actions {
  z-index: 1;
}
</style>
