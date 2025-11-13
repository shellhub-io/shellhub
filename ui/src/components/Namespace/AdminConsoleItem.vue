<template>
  <template v-if="compact">
    <AdminConsoleItemIcon />
    <span class="text-body-1">Admin Console</span>
  </template>

  <v-list-item
    v-else
    lines="two"
  >
    <template #prepend><AdminConsoleItemIcon /></template>
    <v-list-item-title>Admin Console</v-list-item-title>
    <v-list-item-subtitle>
      <div class="d-flex align-center text-caption text-capitalize">
        <div class="d-flex align-center ga-1 flex-1-0">
          <v-icon
            size="x-small"
            icon="mdi-shield-crown"
          />
          <span>Super Admin</span>
        </div>
        <div class="d-flex align-center ga-1 flex-1-0">
          <v-icon
            size="x-small"
            icon="mdi-server"
          />
          <span>Instance</span>
        </div>
      </div>
    </v-list-item-subtitle>
  </v-list-item>
</template>

<script setup lang="ts">
import AdminConsoleItemIcon from "./AdminConsoleItemIcon.vue";
const { compact } = defineProps<{ compact?: boolean }>();
</script>
