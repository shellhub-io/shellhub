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
    <NamespaceEdit
      v-model="editAnnouncement"
      @update="getNamespace"
    />
    <v-card
      variant="flat"
      class="bg-transparent"
      data-test="card"
    >
      <v-card-item>
        <v-list-item
          class="pa-0"
          data-test="card-header"
        >
          <template #title>
            <h1 data-test="card-title">
              Namespace
            </h1>
          </template>
          <template #subtitle>
            <span data-test="card-subtitle">Manage the namespace settings</span>
          </template>
          <template #append>
            <div class="mr-4">
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
            </div>
          </template>
        </v-list-item>
      </v-card-item>
      <v-card-text class="pt-4">
        <v-list
          border
          rounded
          class="bg-background pa-0"
          data-test="profile-details-list"
        >
          <v-card-item
            style="grid-template-columns: max-content 1.5fr 2fr"
            data-test="profile-details-item"
          >
            <template #prepend>
              <v-icon data-test="name-icon">
                mdi-cloud-braces
              </v-icon>
            </template>
            <template #title>
              <span
                class="text-subtitle-1"
                data-test="name-title"
              >Name</span>
            </template>
            <template #append>
              <v-text-field
                v-model="name"
                :error-messages="nameError"
                :disabled="!editDataStatus"
                :readonly="!editDataStatus"
                required
                :hide-details="!nameError"
                density="compact"
                :variant="editDataStatus ? 'outlined' : 'plain'"
                data-test="name-input"
              />
            </template>
          </v-card-item>
          <v-divider />
          <v-card-item
            style="grid-template-columns: max-content 1.5fr 2fr"
            data-test="tenant-details-item"
          >
            <template #prepend>
              <v-icon data-test="tenant-icon">
                mdi-identifier
              </v-icon>
            </template>
            <template #title>
              <span
                class="text-subtitle-1"
                data-test="tenant-title"
              >Tenant ID</span>
            </template>
            <template #append>
              <v-chip class="ml-1">
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
            </template>
          </v-card-item>
          <v-divider />
          <v-card-item
            style="grid-template-columns: max-content 1.5fr 2fr"
            data-test="announcement-item"
          >
            <template #title>
              <v-icon
                data-test="announcement-icon"
                size="18"
                class="pl-1 mr-3"
              >
                mdi-bullhorn-variant-outline
              </v-icon>
              <span
                class="text-subtitle-1"
                data-test="announcement-title"
              >Connection Announcement</span>
            </template>
            <v-card-text class="pt-1 pl-0">
              <span data-test="announcement-subtitle">A connection announcement is a custom message written
                during a session when a connection is established on a device
                within the namespace.</span>
            </v-card-text>
            <template #append>
              <v-btn
                class="ml-4"
                variant="text"
                color="primary"
                data-test="edit-announcement-btn"
                @click="editAnnouncement = true"
              >
                Edit Announcement
              </v-btn>
            </template>
          </v-card-item>
          <v-divider />
          <v-row class="ma-0">
            <v-card
              flat
              class="bg-background"
              data-test="record-item"
            >
              <template #title>
                <v-icon
                  data-test="record-icon"
                  size="18"
                  class="pl-1 mr-3"
                >
                  mdi-play-box-outline
                </v-icon>
                <span
                  class="text-subtitle-1"
                  data-test="record-title"
                >Session Record</span>
                <v-card-text
                  class="pl-0 pt-1"
                  data-test="record-description"
                >
                  Session record is a feature that allows you to check logged activity
                  when connecting to a device.
                </v-card-text>
              </template>
            </v-card>
            <v-col class="d-flex align-center justify-end bg-background">
              <SettingSessionRecording
                :tenant-id
                data-test="session-recording-setting-component"
              />
            </v-col>
          </v-row>
          <v-divider />
          <v-card-item
            style="grid-template-columns: max-content 1.5fr 2fr"
            data-test="delete-leave-item"
          >
            <template #title>
              <v-icon
                data-test="delete-leave-icon"
                size="18"
                class="pl-1 mr-3"
              >
                mdi-delete
              </v-icon>
              <span
                v-if="isOwner"
                class="text-subtitle-1"
                data-test="delete-leave-title"
              >Delete Namespace</span>
              <span
                v-else
                class="text-subtitle-1"
                data-test="delete-leave-title"
              >Leave Namespace</span>
            </template>
            <v-card-text class="pt-1 pl-0">
              <span
                v-if="isOwner"
                data-test="delete-description"
              >After deleting a namespace, there is no going back. Be sure. </span>
              <span
                v-else
                data-test="leave-description"
              >After leaving a namespace, you will need to be invited again to access it.</span>
            </v-card-text>
            <template #append>
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
            </template>
          </v-card-item>
        </v-list>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { onMounted, computed, ref } from "vue";
import axios, { AxiosError } from "axios";
import * as yup from "yup";
import { useField } from "vee-validate";
import hasPermission from "@/utils/permission";
import SettingSessionRecording from "./SettingSessionRecording.vue";
import NamespaceDelete from "../Namespace/NamespaceDelete.vue";
import NamespaceEdit from "../Namespace/NamespaceEdit.vue";
import handleError from "@/utils/handleError";
import NamespaceLeave from "../Namespace/NamespaceLeave.vue";
import useSnackbar from "@/helpers/snackbar";
import CopyWarning from "@/components/User/CopyWarning.vue";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
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
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        snackbar.showError("You are not authorized to access this resource.");
      }
    } else {
      snackbar.showError("Failed to load namespace.");
      handleError(error);
    }
  }
};

const handleUpdateNameError = (error: unknown): void => {
  if (axios.isAxiosError(error)) {
    switch (error.response?.status) {
      case 400:
        setNameError("This name is not valid");
        break;
      case 409:
        setNameError("Name used already");
        break;
      default:
        snackbar.showError("Failed to update name.");
        handleError(error);
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
</style>
