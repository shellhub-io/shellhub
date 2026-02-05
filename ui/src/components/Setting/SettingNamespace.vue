<!-- eslint-disable vue/max-len -->
<template>
  <v-container fluid>
    <NamespaceDelete
      v-model="namespaceDelete"
      :tenant="tenantId"
      @billing-in-debt="billingInDebt = true"
    />
    <NamespaceLeave
      v-model="namespaceLeave"
      :tenant="tenantId"
    />
    <ConnectionAnnouncementEdit
      v-model="editAnnouncement"
      @update="getNamespace"
    />
    <PageHeader
      icon="mdi-cloud-braces"
      title="Namespace"
      overline="Settings"
      description="Manage the namespace settings including name, type, and access controls."
      icon-color="primary"
      data-test="card"
    >
      <template #actions>
        <v-btn
          v-if="!editDataStatus"
          :disabled="!canRenameNamespace"
          color="primary"
          variant="elevated"
          data-test="edit-namespace-btn"
          @click="editDataStatus = true"
        >
          Edit Namespace
        </v-btn>
        <template v-else>
          <v-btn
            color="primary"
            variant="text"
            class="mr-2"
            data-test="cancel-edit-btn"
            @click="cancel"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            variant="flat"
            data-test="save-changes-btn"
            :disabled="!!nameError"
            @click="updateName"
          >
            Save Changes
          </v-btn>
        </template>
      </template>
    </PageHeader>
    <SettingsSection data-test="profile-details-list">
      <SettingsRow
        icon="mdi-cloud-braces"
        icon-test-id="name-icon"
        title="Name"
        title-test-id="name-title"
        data-test="profile-details-item"
      >
        <v-text-field
          v-model="name"
          :error-messages="nameError"
          :disabled="!editDataStatus"
          :readonly="!editDataStatus"
          required
          :reverse="smAndUp"
          :hide-details="!nameError"
          density="compact"
          :variant="editDataStatus ? 'outlined' : 'plain'"
          data-test="name-input"
        />
      </SettingsRow>
      <v-divider />
      <SettingsRow
        icon="mdi-shape-outline"
        icon-test-id="type-icon"
        title="Type"
        title-test-id="type-title"
        data-test="type-details-item"
      >
        <v-chip
          class="text-capitalize"
          data-test="type-chip"
        >
          <v-icon
            size="small"
            class="mr-1"
            :icon="namespaceTypeIcon"
          />
          {{ namespace.type || "team" }}
        </v-chip>
      </SettingsRow>
      <v-divider />
      <SettingsRow
        icon="mdi-identifier"
        icon-test-id="tenant-icon"
        title="Tenant ID"
        title-test-id="tenant-title"
        data-test="tenant-details-item"
      >
        <v-chip>
          <v-tooltip location="top">
            <template #activator="props">
              <CopyWarning :copied-item="'Tenant ID'">
                <template #default="{ copyText }">
                  <span
                    v-bind="props"
                    class="hover-text"
                    data-test="tenant-copy-btn"
                    @click="copyText(tenantId)"
                    @keypress="copyText(tenantId)"
                  >
                    {{ tenantId }}
                    <v-icon icon="mdi-content-copy" />
                  </span>
                </template>
              </CopyWarning>
            </template>
            <span data-test="tenant-tooltip">Copy ID</span>
          </v-tooltip>
        </v-chip>
      </SettingsRow>
      <v-divider />
      <SettingsRow
        icon="mdi-bullhorn-variant-outline"
        icon-test-id="announcement-icon"
        title="Connection Announcement"
        title-test-id="announcement-title"
        subtitle="A connection announcement is a custom message written during a
         session when a connection is established on a device within the namespace."
        subtitle-test-id="announcement-subtitle"
        data-test="announcement-item"
      >
        <v-btn
          class="ml-4"
          variant="text"
          color="primary"
          data-test="edit-announcement-btn"
          @click="editAnnouncement = true"
        >
          Edit Announcement
        </v-btn>
      </SettingsRow>
      <v-divider />
      <SettingsRow
        icon="mdi-play-box-outline"
        icon-test-id="record-icon"
        title="Session Record"
        title-test-id="record-title"
        subtitle="Session record is a feature that allows you to check logged activity when connecting to a device."
        subtitle-test-id="record-description"
        data-test="record-item"
      >
        <SettingSessionRecording
          :tenant-id
          class="mr-sm-4"
          data-test="session-recording-setting-component"
        />
      </SettingsRow>
      <v-divider />
      <SettingsRow
        icon="mdi-delete"
        icon-test-id="delete-leave-icon"
        :title="isOwner ? 'Delete Namespace' : 'Leave Namespace'"
        title-test-id="delete-leave-title"
        :subtitle="isOwner ? 'After deleting a namespace, there is no going back. Be sure.' : 'After leaving a namespace, you will need to be invited again to access it.'"
        :subtitle-test-id="isOwner ? 'delete-description' : 'leave-description'"
        data-test="delete-leave-item"
      >
        <v-btn
          v-if="isOwner"
          class="ml-4"
          variant="text"
          color="error"
          :disabled="billingInDebt"
          data-test="delete-namespace-btn"
          @click="namespaceDelete = true"
        >
          Delete
        </v-btn>
        <v-btn
          v-else
          variant="text"
          color="error"
          data-test="leave-namespace-btn"
          @click="namespaceLeave = true"
        >
          Leave
        </v-btn>
      </SettingsRow>
    </SettingsSection>
  </v-container>
</template>

<script setup lang="ts">
import { onMounted, computed, ref } from "vue";
import axios from "axios";
import * as yup from "yup";
import { useField } from "vee-validate";
import { useDisplay } from "vuetify";
import hasPermission from "@/utils/permission";
import SettingSessionRecording from "./SettingSessionRecording.vue";
import NamespaceDelete from "../Namespace/NamespaceDelete.vue";
import ConnectionAnnouncementEdit from "../Namespace/ConnectionAnnouncementEdit.vue";
import PageHeader from "../PageHeader.vue";
import handleError from "@/utils/handleError";
import NamespaceLeave from "../Namespace/NamespaceLeave.vue";
import SettingsRow from "./SettingsRow.vue";
import SettingsSection from "./SettingsSection.vue";
import useSnackbar from "@/helpers/snackbar";
import CopyWarning from "@/components/User/CopyWarning.vue";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const { smAndUp } = useDisplay();
const namespace = computed(() => namespacesStore.currentNamespace);
const isOwner = computed(() => namespace.value.owner === localStorage.getItem("id"));
const { tenantId } = authStore;
const billingInDebt = ref(false);
const namespaceLeave = ref(false);
const namespaceDelete = ref(false);
const editDataStatus = ref(false);
const editAnnouncement = ref(false);

const {
  value: name,
  errorMessage: nameError,
  setErrors: setNameError,
} = useField<string>(
  "name",
  yup
    .string()
    .min(3, "Your namespace should be 3-30 characters long")
    .max(30, "Your namespace should be 3-30 characters long")
    .required()
    .matches(/^[^.]*$/, "The name must not contain dots"),
  {
    initialValue: namespace.value.name,
  },
);

const cancel = () => {
  name.value = namespace.value.name;
  editDataStatus.value = false;
};

const getNamespace = async () => {
  try {
    await namespacesStore.fetchNamespace(tenantId);
  } catch (error: unknown) {
    if (axios.isAxiosError(error) && error.response?.status === 403) {
      snackbar.showError("You are not authorized to access this resource.");
      return;
    }
    snackbar.showError("Failed to load namespace.");
    handleError(error);
  }
};
const handleUpdateNameError = (error: unknown): void => {
  const nameFieldErrorMap: Record<number, string> = {
    400: "This name is not valid",
    409: "Name used already",
  };

  if (axios.isAxiosError(error)) {
    const errorMessage = nameFieldErrorMap[error.response?.status as number];

    if (errorMessage) {
      setNameError(errorMessage);
      return;
    }
  }

  snackbar.showError("Failed to update name.");
  handleError(error);
};

const updateName = async () => {
  if (nameError.value) return;

  try {
    await namespacesStore.editNamespace({
      tenant_id: tenantId,
      name: name.value,
    });

    await namespacesStore.fetchNamespaceList();

    await getNamespace();
    snackbar.showSuccess("Namespace name updated successfully.");
    editDataStatus.value = false;
  } catch (error) {
    handleUpdateNameError(error);
  }
};

const canRenameNamespace = hasPermission("namespace:rename");

const namespaceTypeIcon = computed(() => namespace.value.type === "personal" ? "mdi-account" : "mdi-account-group");

onMounted(async () => {
  if (tenantId) await getNamespace();
});
</script>

<style scoped>
.hover-text {
  cursor: pointer;
}

.hover-text:hover {
  text-decoration: underline;
}

:deep(.v-field--variant-plain) {
  --v-field-padding-start: 16px;
  --v-field-padding-end: 16px;
  --v-field-padding-bottom: 8px;
}

@media (max-width: 600px) {
  :deep(.v-text-field .v-field__input) {
    text-align: center;
  }
}

</style>
