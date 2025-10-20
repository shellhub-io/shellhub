<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        tag="button"
        v-bind="props"
        @click="showDialog = true"
        tabindex="0"
        data-test="open-dialog-icon"
        icon="mdi-account-lock-open"
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
            This action will enable local authentication to this user and generate a new password.
            Do you want to enable?
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
          <p class="text-justify">A new password has been generated for this user. Please copy and share it securely:</p>
          <CopyWarning :macro="generatedPassword" copied-item="Password">
            <template #default="{ copyText }">
              <v-text-field
                class="mt-2"
                v-model="generatedPassword"
                readonly
                @click="copyText(generatedPassword)"
                @keypress.enter="copyText(generatedPassword)"
                prepend-inner-icon="mdi-key"
                append-inner-icon="mdi-content-copy"
                data-test="generated-password-field"
              />
            </template>
          </CopyWarning>
        </v-window-item>
      </v-window>
    </v-card-text>

    <template #footer>
      <v-spacer />
      <v-btn v-if="step === 1" @click="close" data-test="cancel-btn">Cancel</v-btn>
      <v-btn v-if="step === 1" color="primary" @click="proceedToSecondStep" data-test="enable-btn">Enable</v-btn>
      <v-btn v-if="step === 2" @click="close" data-test="close-btn">Close</v-btn>
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
  } catch (error) { snackbar.showError("Failed to reset user password. Please try again."); }
};

defineExpose({ showDialog, step });
</script>
