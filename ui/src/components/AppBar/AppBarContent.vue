<template>
  <v-app-bar
    flat
    floating
    class="bg-v-theme-surface border-b-thin"
    data-test="app-bar"
  >
    <v-app-bar-nav-icon
      v-if="showMenuToggle"
      class="hidden-lg-and-up"
      aria-label="Toggle Menu"
      data-test="menu-toggle"
      @click.stop="emit('toggle-menu')"
    />

    <div class="d-flex align-center">
      <slot name="left" />
    </div>

    <v-spacer />

    <div class="d-flex align-center ga-4 mr-4">
      <v-tooltip
        v-if="showSupport"
        location="bottom"
        class="text-center"
      >
        <template #activator="{ props }">
          <v-btn
            v-bind="props"
            size="medium"
            color="primary"
            aria-label="community-help-icon"
            icon="mdi-help-circle"
            data-test="support-btn"
            @click="emit('support-click')"
          />
        </template>
        <span>Need assistance? Click here for support.</span>
      </v-tooltip>

      <slot name="right" />
    </div>
  </v-app-bar>
</template>

<script setup lang="ts">
defineProps<{
  showMenuToggle?: boolean;
  showSupport?: boolean;
}>();

const emit = defineEmits<{
  (e: "toggle-menu"): void;
  (e: "support-click"): void;
}>();
</script>
