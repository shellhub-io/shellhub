<template>
  <FormDialog
    v-model="showDialog"
    @close="close"
    @confirm="addNamespace"
    @cancel="close"
    :title="isCommunityVersion ? 'Add a namespace using the CLI' : 'New Namespace'"
    icon="mdi-folder-plus"
    :confirm-text="isCommunityVersion ? '' : 'Submit'"
    :confirm-disabled="isCommunityVersion || !fieldMeta.valid"
    :confirm-loading="isLoading"
    cancel-text="Close"
    confirm-data-test="add-btn"
    cancel-data-test="close-btn"
    :footer-helper-text="isCommunityVersion ? 'Learn more on' : ''"
    :footer-helper-link-text="isCommunityVersion ? 'ShellHub Administration Guide' : ''"
    :footer-helper-link="isCommunityVersion ? 'https://docs.shellhub.io/self-hosted/administration' : ''"
    data-test="namespace-add-card"
  >
    <v-card-text class="pa-6">
      <template v-if="!isCommunityVersion">
        <v-text-field
          v-model="namespaceName"
          label="Namespace"
          :error-messages="namespaceNameError"
          hide-details="auto"
          class="mt-1 mb-4"
        />
        <div class="text-body-2 text-justify">
          <ul class="pl-4">
            <li>The namespace you choose here will be used for in the SSHID of your devices.</li>
            <li>The namespace can contain only lowercase alphanumeric characters and hyphens.</li>
            <li>It cannot begin or end with a hyphen ("-").</li>
            <li>The namespace must be a minimum of 3 characters and a maximum of 63 characters.</li>
            <li>The namespace cannot be changed after creation.</li>
          </ul>
        </div>
      </template>
      <template v-else>
        <p class="text-body-2">
          In the Community Edition of ShellHub, namespaces must be added using the administration CLI.
          For detailed instructions on how to add namespaces, please refer to the documentation at the ShellHub
          Administration Guide.
        </p>
      </template>
    </v-card-text>
  </FormDialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import { envVariables } from "@/envVariables";
import useNamespacesStore from "@/store/modules/namespaces";

const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const showDialog = defineModel({ default: false });
const isCommunityVersion = computed(() => envVariables.isCommunity);
const isLoading = ref(false);

// Validation schema for namespace name
const namespaceSchema = yup
  .string()
  .required("Namespace is required")
  .min(3, "Namespace should be at least 3 characters")
  .max(30, "Namespace should be at most 30 characters")
  .matches(/^[a-z0-9]+(?:-[a-z0-9]+)*$/, "Invalid format");

// Form field for namespace name with validation
const {
  value: namespaceName,
  errorMessage: namespaceNameError,
  setErrors: setNamespaceNameError,
  resetField: resetNamespaceName,
  meta: fieldMeta,
} = useField<string>("namespaceName", namespaceSchema, { initialValue: "" });

// Close the dialog and reset the form
const close = () => {
  showDialog.value = false;
  resetNamespaceName();
};

// Change to the specified namespace
const changeNamespace = async (tenantId: string) => {
  try {
    await namespacesStore.switchNamespace(tenantId);
    window.location.reload();
  } catch (error) {
    snackbar.showError("An error occurred while switching namespaces.");
    handleError(error);
  }
};

// Handle unknown errors and display notifications
const handleErrorAndNotify = (error: unknown) => {
  snackbar.showError("An error occurred while creating the namespace.");
  handleError(error);
};

// Add a new namespace
const addNamespace = async () => {
  isLoading.value = true;
  try {
    const newNamespaceId = await namespacesStore.createNamespace(namespaceName.value);
    await changeNamespace(newNamespaceId);
    close();
    snackbar.showSuccess("Namespace created successfully");
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 400) {
        setNamespaceNameError("Your namespace should be 3-30 characters long");
      } else if (axiosError.response?.status === 403) {
        setNamespaceNameError("Update your plan to create more namespaces");
      } else if (axiosError.response?.status === 409) {
        setNamespaceNameError("Namespace already exists");
      } else {
        handleErrorAndNotify(error);
      }
    } else {
      handleErrorAndNotify(error);
    }
  } finally {
    isLoading.value = false;
  }
};
</script>
