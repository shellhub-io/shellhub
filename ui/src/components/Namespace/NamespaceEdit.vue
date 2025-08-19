<template>
  <BaseDialog v-model="showDialog" @click:outside="close">
    <v-card data-test="password-change-card" class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="title">
        Change Connection Announcement
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1">
        <div class="mt-4 pl-4 pr-4">
          <v-textarea
            v-model="connectionAnnouncement"
            label="Connection Announcement"
            :error-messages="connectionAnnouncementError"
            data-test="connection-announcement-text"
            variant="underlined"
            hint="A connection announcement is a custom message written
      during a session when a connection is established on a device
      within the namespace."
            persistent-hint
            required
            auto-grow
            max-rows="25"
          />
        </div>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="close">
          Cancel
        </v-btn>

        <v-btn
          color="primary"
          variant="text"
          data-test="change-connection-btn"
          :disabled="!!connectionAnnouncementError"
          @click="updateAnnouncement()"
        >
          Save Announcement
        </v-btn>

      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useField } from "vee-validate";
import axios from "axios";
import * as yup from "yup";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const namespace = computed(() => namespacesStore.currentNamespace);
const { tenantId } = authStore;
const showDialog = defineModel({ default: false });
const emit = defineEmits(["update"]);

const {
  value: connectionAnnouncement,
  errorMessage: connectionAnnouncementError,
  setErrors: setConnectionAnnouncementError,
} = useField<string>(
  "Connection Announcement",
  yup
    .string()
    .max(4096, "Your message should be 1-4096 characters long"),
  {
    initialValue: "",
  },
);

const close = () => {
  connectionAnnouncement.value = namespace.value.settings.connection_announcement || "";
  showDialog.value = false;
};

onMounted(async () => {
  if (!authStore.isLoggedIn) return;
  try {
    await namespacesStore.fetchNamespace(tenantId);
    connectionAnnouncement.value = namespace.value.settings.connection_announcement || "";
  } catch (error) {
    handleError(error);
  }
});

const handleUpdateNameError = (error: unknown): void => {
  if (axios.isAxiosError(error)) {
    switch (error.response?.status) {
      case 400:
        setConnectionAnnouncementError("This message is not valid");
        break;
      default:
        snackbar.showError("An error occurred while updating the connection announcement.");
        handleError(error);
    }
  }

  snackbar.showError("An error occurred while updating the connection announcement.");
  handleError(error);
};

const updateAnnouncement = async () => {
  try {
    await namespacesStore.editNamespace({
      tenant_id: tenantId,
      settings: {
        connection_announcement: connectionAnnouncement.value,
      },
    });

    await namespacesStore.fetchNamespaceList();

    emit("update");
    snackbar.showSuccess("Connection announcement updated successfully.");

    showDialog.value = false;
  } catch (error) {
    handleUpdateNameError(error);
  }
};

defineExpose({ showDialog });
</script>
