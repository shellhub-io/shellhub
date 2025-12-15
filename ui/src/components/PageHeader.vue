<template>
  <v-card
    class="pa-6 bg-transparent mb-6 position-relative overflow-hidden border-b mx-n8 mt-n8"
    elevation="0"
    rounded="0"
  >
    <div
      class="d-flex align-start justify-space-between position-relative"
      style="z-index: 1;"
    >
      <!-- Left side: Icon + Title + Description -->
      <div class="d-flex align-start flex-grow-1">
        <v-avatar
          :color="iconColor"
          size="48"
          class="mr-4"
        >
          <v-icon size="32">{{ icon }}</v-icon>
        </v-avatar>
        <div>
          <div
            v-if="overline"
            class="text-overline text-medium-emphasis mb-1"
          >
            {{ overline }}
          </div>
          <div class="text-h5 font-weight-bold mb-2">{{ title }}</div>
          <div
            v-if="description"
            class="text-body-2 text-medium-emphasis"
          >
            {{ description }}
          </div>
          <div
            v-if="$slots.description"
            class="text-body-2 text-medium-emphasis"
          >
            <slot name="description" />
          </div>
        </div>
      </div>

      <!-- Right side: Actions -->
      <div
        v-if="$slots.actions"
        class="ml-4"
      >
        <slot name="actions" />
      </div>
    </div>

    <!-- Extra content below -->
    <div
      v-if="$slots.default"
      class="mt-4 position-relative"
      style="z-index: 1;"
    >
      <slot />
    </div>
  </v-card>
</template>

<script setup lang="ts">
defineProps<{
  icon: string;
  title: string;
  description?: string;
  overline?: string;
  iconColor?: string;
}>();
</script>
