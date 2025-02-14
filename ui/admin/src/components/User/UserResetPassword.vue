<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        tag="a"
        dark
        v-bind="props"
        @click="dialog = true"
        tabindex="0"
        data-test="open-dialog-icon"
      >mdi-account-lock-open
      </v-icon>
    </template>
    <span>Enable Local Authentication</span>
  </v-tooltip>

  <v-dialog v-model="dialog" @click:outside="close" max-width="500">
    <v-card>
      <v-card-title class="text-h5 pb-2">Enable Local Authentication</v-card-title>
      <v-divider />
      <v-window v-model="step">
        <v-window-item value="step-1">
          <v-card-text class="pb-0" data-test="step-1-description">
            This action will enable local authentication to this user and generate a new password.
          </v-card-text>

          <v-card-text data-test="step-1-confirmation">
            Do you want to enable?
          </v-card-text>

          <v-card-actions>
            <v-btn @click="close" data-test="cancel-btn">Cancel</v-btn>
            <v-spacer />
            <v-btn color="primary" @click="proceedToSecondStep" data-test="enable-btn">Enable</v-btn>
          </v-card-actions>
        </v-window-item>

        <v-window-item value="step-2">
          <v-card-text>
            <v-alert
              class="mb-2"
              type="warning"
              variant="tonal"
              text="Users are strongly encouraged to change this password after their first successful authentication"
              data-test="password-warning"
            />
            A new password has been generated for this user. Please copy it and share it securely:
            <v-text-field
              class="mt-4"
              v-model="generatedPassword"
              readonly
              @click="copyText(generatedPassword)"
              @keypress="copyText(generatedPassword)"
              variant="outlined"
              dense
              prepend-inner-icon="mdi-key"
              append-inner-icon="mdi-clipboard-text"
              data-test="generated-password-field"
            />
          </v-card-text>

          <v-card-actions>
            <v-btn @click="close" data-test="close-btn">Close</v-btn>
          </v-card-actions>
        </v-window-item>
      </v-window>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useStore } from "../../store";
import { INotificationsCopy } from "../../interfaces/INotifications";

const props = defineProps({
  userId: {
    type: String,
    default: "",
    required: true,
  },
});

const emit = defineEmits(["update"]);

const store = useStore();
const dialog = ref(false);
const step = ref("step-1");
const generatedPassword = computed(() => store.getters["users/getGeneratedPassword"]);

const close = () => {
  dialog.value = false;
  step.value = "step-1";
  emit("update");
};

const copyText = (value: string | undefined) => {
  if (value) {
    navigator.clipboard.writeText(value);
    store.dispatch("snackbar/showSnackbarCopy", INotificationsCopy.tenantId);
  }
};

const proceedToSecondStep = async () => {
  try {
    await store.dispatch("users/resetUserPassword", props.userId);
    step.value = "step-2";
  } catch (error) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      "Failed to reset user password. Please try again.",
    );
  }
};

defineExpose({ dialog, step });
</script>
