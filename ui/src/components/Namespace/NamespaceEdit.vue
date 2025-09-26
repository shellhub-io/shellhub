<template>
  <FormDialog
    v-model="showDialog"
    @close="close"
    @confirm="updateAnnouncement"
    @cancel="close"
    title="Change Connection Announcement"
    icon="mdi-bullhorn"
    confirm-text="Save Announcement"
    :confirm-disabled="!!connectionAnnouncementError"
    :confirm-loading="isLoading"
    cancel-text="Cancel"
    confirm-data-test="change-connection-btn"
    cancel-data-test="close-btn"
  >
    <v-card-text class="pa-6">
      <v-textarea
        v-model="connectionAnnouncement"
        label="Connection Announcement"
        :error-messages="connectionAnnouncementError"
        data-test="connection-announcement-text"
        hint="A connection announcement is a custom message written during a session
        when a connection is established on a device within the namespace."
        persistent-hint
        required
        auto-grow
        max-rows="25"
      />
    </v-card-text>
  </FormDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useField } from "vee-validate";
import axios from "axios";
import * as yup from "yup";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import FormDialog from "../FormDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const namespace = computed(() => namespacesStore.currentNamespace);
const { tenantId } = authStore;
const showDialog = defineModel({ default: false });
const emit = defineEmits(["update"]);
const isLoading = ref(false);

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

const handleUpdateAnnouncementError = (error: unknown): void => {
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
  isLoading.value = true;
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
    handleUpdateAnnouncementError(error);
  } finally {
    isLoading.value = false;
  }
};

defineExpose({ showDialog });
</script>
