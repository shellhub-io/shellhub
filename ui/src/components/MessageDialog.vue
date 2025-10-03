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
        v-if="showTitlebar"
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
        <slot name="titlebar-actions" />
        <v-btn
          v-if="showCloseButton"
          icon="mdi-close"
          variant="text"
          @click="$emit('close')"
          data-test="close-btn-toolbar"
        />
      </v-toolbar>

      <!-- Content -->
      <v-card-text class="text-center py-8">
        <v-icon
          v-if="icon && !showTitlebar"
          :color="iconColor"
          :size="iconSize"
          class="mb-4"
        >
          {{ icon }}
        </v-icon>

        <div v-if="title && !showTitlebar" class="text-h5 mb-4">
          {{ title }}
        </div>

        <div v-if="description && !showTitlebar" class="text-body-2 text-medium-emphasis mb-6">
          {{ description }}
        </div>

        <slot />
      </v-card-text>

      <!-- Footer -->
      <v-toolbar
        v-if="showFooter"
        color="primary"
        class="bg-v-theme-surface border-t px-6 py-2"
      >
        <!-- Block Layout: Full width buttons -->
        <v-row class="ma-0 ga-3">
          <v-col v-if="cancelText" class="pa-0">
            <v-btn
              @click="$emit('cancel')"
              variant="tonal"
              :data-test="cancelDataTest || 'cancel-btn'"
              block
            >
              {{ cancelText }}
            </v-btn>
          </v-col>
          <v-col v-if="confirmText" class="pa-0">
            <v-btn
              :color="confirmColor || 'primary'"
              :variant="confirmDisabled ? 'outlined' : 'flat'"
              @click="$emit('confirm')"
              :disabled="confirmDisabled"
              :loading="confirmLoading"
              :data-test="confirmDataTest || 'confirm-btn'"
              block
            >
              {{ confirmText }}
            </v-btn>
          </v-col>
        </v-row>
      </v-toolbar>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import BaseDialog, { type BaseDialogProps } from "./BaseDialog.vue";

interface Props extends BaseDialogProps {
  // Titlebar props
  showTitlebar?: boolean;
  showCloseButton?: boolean;
  // Content props
  title?: string;
  description?: string;
  icon?: string;
  iconColor?: string;
  iconSize?: string | number;
  // Footer props
  showFooter?: boolean;
  // Button props
  confirmText?: string;
  confirmColor?: string;
  confirmDisabled?: boolean;
  confirmLoading?: boolean;
  confirmDataTest?: string;
  cancelText?: string;
  cancelDataTest?: string;
}

withDefaults(defineProps<Props>(), {
  threshold: "sm",
  title: "",
  description: "",
  icon: "",
  showTitlebar: false,
  showCloseButton: true,
  iconColor: "success",
  iconSize: 48,
  showFooter: true,
  confirmText: "",
  confirmColor: "primary",
  confirmDataTest: "",
  cancelText: "",
  cancelDataTest: "",
});

defineEmits(["close", "confirm", "cancel"]);
const showDialog = defineModel<boolean>({ required: true });
</script>
