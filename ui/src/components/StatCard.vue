<template>
  <v-hover v-slot="{ isHovering, props }">
    <v-card
      class="pa-1 mb-2 bg-v-theme-surface border"
      v-bind="props"
      :elevation="isHovering ? 24 : 8"
      density="compact"
      variant="flat"
    >
      <v-card-item class="pa-4 pt-5">
        <v-card-title class="text-overline mb-4" :title-test="title">
          {{ title }}
        </v-card-title>
        <v-card-text class="text-h5 pl-0">
          {{ stat }}
        </v-card-text>
        <v-card-subtitle>
          {{ content }}
        </v-card-subtitle>
        <template v-slot:append>
          <v-icon :icon size="x-large" class="pb-4" />
        </template>
      </v-card-item>

      <v-card-actions>
        <v-btn
          v-if="buttonLabel !== 'Add Device'"
          :to="path"
          color="primary"
          variant="elevated"
          size="small"
          tabindex="0"
          class="text-subtitle"
          :text="buttonLabel"
        />
        <DeviceAdd v-else size="small" />
      </v-card-actions>
    </v-card>
  </v-hover>
</template>

<script setup lang="ts">
import { StatCardItem } from "@/interfaces/IStats";
import DeviceAdd from "./Devices/DeviceAdd.vue";

defineProps<StatCardItem>();
</script>
