<template>
  <v-card class="pa-6 bg-v-theme-surface text-center h-100 border">
    <v-avatar
      color="primary"
      size="64"
      class="mb-4"
    >
      <v-icon
        size="40"
        :icon
      />
    </v-avatar>

    <v-card-title class="text-overline text-medium-emphasis mb-2">
      {{ title }}
    </v-card-title>
    <v-card-subtitle class="text-h2 font-weight-bold mb-4">
      {{ stat }}
    </v-card-subtitle>

    <v-btn
      variant="text"
      color="primary"
      size="small"
      :to="path"
      block
      :text="buttonLabel"
    />
  </v-card>
</template>

<script setup lang="ts">
import { StatCardItem } from "@/interfaces/IStats";

defineProps<StatCardItem>();
</script>
