<template>
  <v-row class="ma-0 px-3 align-center">
    <v-col
      cols="12"
      sm="6"
      md="6"
      class="d-flex ga-3 align-center align-sm-start justify-center justify-sm-start text-center text-sm-start"
    >
      <div class="text-center text-sm-start">
        <div
          v-if="title || $slots.title"
          class="text-subtitle-1"
          :data-test="titleTestId"
        >
          <v-icon
            v-if="icon"
            class="mr-3"
            :data-test="iconTestId"
          >
            {{ icon }}
          </v-icon>
          <slot name="title">{{ title }}</slot>
        </div>
        <div
          v-if="subtitle || $slots.subtitle"
          class="text-body-2 mt-1"
          :data-test="subtitleTestId"
        >
          <slot name="subtitle">{{ subtitle }}</slot>
        </div>
      </div>
    </v-col>
    <v-col
      cols="12"
      sm="6"
      md="6"
      class="px-0 d-flex align-center justify-center justify-sm-end"
    >
      <div class="w-100 d-flex justify-center justify-sm-end">
        <slot />
      </div>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
defineProps<{
  icon?: string;
  iconTestId?: string;
  title?: string;
  titleTestId?: string;
  subtitle?: string;
  subtitleTestId?: string;
}>();
</script>
