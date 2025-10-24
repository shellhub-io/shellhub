<template>
  <WindowDialog
    v-model="showDialog"
    @close="$emit('close')"
    :threshold="threshold"
    :force-fullscreen="forceFullscreen"
    :title="title"
    :description="description"
    :icon="icon"
    :icon-color="iconColor"
    :show-close-button="showCloseButton"
    :show-footer="showFooter"
  >
    <!-- Content -->
    <v-form @submit.prevent="handleConfirm">
      <slot />
    </v-form>

    <!-- Footer with form-specific functionality -->
    <template #footer>
      <!-- Buttons with alert system -->
      <v-window
        v-if="hasButtons"
        v-model="footerWindow"
        class="w-100"
        direction="vertical"
      >
        <!-- Default buttons window -->
        <v-window-item :value="0" class="w-100">
          <div class="d-flex align-center w-100">
            <!-- Helper text on the left -->
            <div
              v-if="footerHelperText"
              class="text-caption text-medium-emphasis"
            >
              <span>{{ footerHelperText }}</span>
              <a
                v-if="footerHelperLink && footerHelperLinkText"
                :href="footerHelperLink"
                :target="footerHelperTarget"
                class="text-decoration-none text-primary ml-1"
                rel="noopener noreferrer"
              >
                {{ footerHelperLinkText }}
                <v-icon size="12" class="ml-1">mdi-open-in-new</v-icon>
              </a>
            </div>
            <v-spacer />
            <!-- Buttons on the right -->
            <div class="d-flex">
              <slot name="footer-right">
                <v-btn
                  v-if="cancelText"
                  @click="$emit('cancel')"
                  variant="tonal"
                  :data-test="cancelDataTest || 'cancel-btn'"
                  class="mr-2"
                >
                  {{ cancelText }}
                </v-btn>
                <v-btn
                  v-if="confirmText"
                  :color="confirmColor || 'primary'"
                  :variant="confirmDisabled || showAlert ? 'outlined' : 'flat'"
                  @click="handleConfirm"
                  :disabled="confirmDisabled || showAlert"
                  :loading="confirmLoading"
                  :data-test="confirmDataTest || 'confirm-btn'"
                >
                  {{ confirmText }}
                </v-btn>
              </slot>
            </div>
          </div>
        </v-window-item>

        <!-- Alert window -->
        <v-window-item :value="1" class="w-100">
          <v-alert
            v-if="alertMessage"
            :text="alertMessage"
            :type="alertType || 'error'"
            variant="tonal"
            class="mb-0"
            data-test="form-dialog-alert"
          >
            <template #append>
              <v-btn
                variant="tonal"
                :color="alertType || 'error'"
                size="small"
                @click="hideAlert"
                data-test="alert-got-it-btn"
              >
                {{ alertButtonText }}
              </v-btn>
            </template>
          </v-alert>
        </v-window-item>
      </v-window>

      <!-- Fallback: just helper text centered if no buttons -->
      <div
        v-else-if="footerHelperText"
        class="text-caption text-medium-emphasis mx-auto"
      >
        <span>{{ footerHelperText }}</span>
        <a
          v-if="footerHelperLink && footerHelperLinkText"
          :href="footerHelperLink"
          :target="footerHelperTarget"
          class="text-decoration-none text-primary ml-1"
          rel="noopener noreferrer"
        >
          {{ footerHelperLinkText }}
          <v-icon size="12" class="ml-1">mdi-open-in-new</v-icon>
        </a>
      </div>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import WindowDialog from "./WindowDialog.vue";
import { type BaseDialogProps } from "./BaseDialog.vue";

interface Props extends BaseDialogProps {
  // Titlebar props
  title?: string;
  description?: string;
  icon?: string;
  iconColor?: string;
  showCloseButton?: boolean;
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
  // Alert props
  alertMessage?: string;
  alertType?: "error" | "success" | "warning" | "info";
  alertButtonText?: string;
  // Footer helper text
  footerHelperText?: string;
  footerHelperLinkText?: string;
  footerHelperLink?: string;
  footerHelperTarget?: "_blank" | "_self";
}

const props = withDefaults(defineProps<Props>(), {
  threshold: "sm",
  title: "",
  description: "",
  icon: "",
  iconColor: "primary",
  showCloseButton: true,
  showFooter: true,
  confirmText: "",
  confirmColor: "primary",
  confirmDataTest: "",
  cancelText: "Close",
  cancelDataTest: "",
  alertMessage: "",
  alertType: "error",
  alertButtonText: "Got it",
  footerHelperText: "",
  footerHelperLinkText: "",
  footerHelperLink: "",
  footerHelperTarget: "_blank",
});

const emit = defineEmits(["close", "confirm", "cancel", "alert-dismissed"]);
const showDialog = defineModel<boolean>({ required: true });

// Alert state
const showAlert = computed(() => !!props.alertMessage);

// Window control
const footerWindow = ref(showAlert.value ? 1 : 0); // 0 = buttons, 1 = alert

// Check if dialog has buttons configured
const hasButtons = computed(() => !!(props.confirmText || props.cancelText));

// Watch for alert message changes
watch(() => props.alertMessage, (newMessage) => {
  if (newMessage) {
    footerWindow.value = 1; // Show alert
  }
});

// Hide alert and go back to buttons
const hideAlert = () => {
  footerWindow.value = 0;
  emit("alert-dismissed");
};

// Handle confirm - don't emit if alert is showing
const handleConfirm = () => {
  if (!showAlert.value) {
    emit("confirm");
  }
};

defineExpose({ footerWindow });
</script>
