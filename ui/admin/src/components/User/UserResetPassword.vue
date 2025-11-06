<template>
  <v-tooltip
    bottom
    anchor="bottom"
  >
    <template #activator="{ props }">
      <v-icon
        tag="button"
        v-bind="props"
        tabindex="0"
        data-test="open-dialog-icon"
        icon="mdi-account-lock-open"
        @click="showDialog = true"
      />
    </template>
    <span>Enable Local Authentication</span>
  </v-tooltip>

  <WindowDialog
    v-model="showDialog"
    title="Enable Local Authentication"
    icon="mdi-account-lock-open"
    @close="close"
  >
    <v-card-text class="pa-6">
      <v-window v-model="step">
        <v-window-item :value="1">
          <p>
            This action will enable local authentication to this user and
            generate a new password. Do you want to enable?
          </p>
        </v-window-item>
        <v-window-item :value="2">
          <v-alert
            class="mb-4"
            type="warning"
            variant="tonal"
            text="Users are strongly encouraged to change this password after their first successful authentication"
            data-test="password-warning"
          />
          <p class="text-justify">
            A new password has been generated for this user. Please copy and
            share it securely:
          </p>
          <CopyWarning
            :macro="generatedPassword"
            copied-item="Password"
          >
            <template #default="{ copyText }">
              <v-text-field
                v-model="generatedPassword"
                class="mt-2"
                readonly
                prepend-inner-icon="mdi-key"
                append-inner-icon="mdi-content-copy"
                data-test="generated-password-field"
                @click="copyText(generatedPassword)"
                @keypress.enter="copyText(generatedPassword)"
              />
            </template>
          </CopyWarning>
        </v-window-item>
      </v-window>
    </v-card-text>

    <template #footer>
      <v-spacer />
      <v-btn
        v-if="step === 1"
        data-test="cancel-btn"
        @click="close"
      >
        Cancel
      </v-btn>
      <v-btn
        v-if="step === 1"
        color="primary"
        data-test="enable-btn"
        @click="proceedToSecondStep"
      >
        Enable
      </v-btn>
      <v-btn
        v-if="step === 2"
        data-test="close-btn"
        @click="close"
      >
        Close
      </v-btn>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import useUsersStore from "@admin/store/modules/users";
import useSnackbar from "@/helpers/snackbar";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";

const props = defineProps<{ userId: string }>();

const emit = defineEmits(["update"]);

const snackbar = useSnackbar();
const usersStore = useUsersStore();
const showDialog = ref(false);
const step = ref(1);
const generatedPassword = ref("");

const close = () => {
  showDialog.value = false;
  step.value = 1;
  emit("update");
};

const proceedToSecondStep = async () => {
  try {
    generatedPassword.value = await usersStore.resetUserPassword(props.userId);
    step.value = 2;
  } catch {
    snackbar.showError("Failed to reset user password. Please try again.");
  }
};

defineExpose({ showDialog, step });
</script>
