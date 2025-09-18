<template>
  <BaseDialog
    v-model="showDialog"
    @close="$emit('close')"
    :threshold="threshold"
    :force-fullscreen="forceFullscreen"
  >
    <template #content>
      <!-- Titlebar -->
      <v-toolbar
        color="primary"
        class="bg-v-theme-surface border-b px-4 py-2"
      >
        <v-avatar
          v-if="icon"
          size="48"
          :color="iconColor"
          rounded="rounded"
          variant="tonal"
          class="border border-primary border-opacity-100 mr-3"
        >
          <v-icon size="24">{{ icon }}</v-icon>
        </v-avatar>
        <div v-if="title || description">
          <v-toolbar-title v-if="title" class="text-h6">{{ title }}</v-toolbar-title>
          <div v-if="description" class="text-caption text-medium-emphasis">{{ description }}</div>
        </div>
        <slot name="titlebar-content" />
        <v-spacer />
        <v-btn
          v-if="showCloseButton"
          icon="mdi-close"
          variant="text"
          @click="$emit('close')"
          data-test="close-btn-toolbar"
        />
        <slot name="titlebar-actions" />
      </v-toolbar>

      <!-- Content -->
      <slot />

      <!-- Footer -->
      <v-toolbar
        v-if="showFooter"
        color="primary"
        class="bg-v-theme-surface border-t px-6 py-2"
      >
        <slot name="footer" />
      </v-toolbar>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import BaseDialog, { type BaseDialogProps } from "./BaseDialog.vue";

interface Props extends BaseDialogProps {
  // Titlebar props
  title?: string;
  description?: string;
  icon?: string;
  iconColor?: string;
  showCloseButton?: boolean;
  // Footer props
  showFooter?: boolean;
}

withDefaults(defineProps<Props>(), {
  threshold: "sm",
  title: "",
  description: "",
  icon: "",
  iconColor: "primary",
  showCloseButton: true,
  showFooter: true,
});

defineEmits(["close"]);
const showDialog = defineModel<boolean>({ required: true });
</script>
