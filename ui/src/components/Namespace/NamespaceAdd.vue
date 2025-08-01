<template>
  <BaseDialog v-model="showDialog" @click:outside="close">
    <v-card data-test="namespace-add-card" class="bg-v-theme-surface rounded" rounded>
      <template v-if="!isCommunityVersion">
        <v-card-title class="bg-primary d-flex justify-space-between align-center text-h5 pa-4">
          New Namespace
          <v-btn
            icon="mdi-close"
            variant="text"
            @click="close"
          />
        </v-card-title>
        <v-container>
          <v-card-text class="pb-0">
            <v-text-field
              v-model="namespaceName"
              label="Namespace"
              :error-messages="namespaceNameError"
              variant="outlined"
              data-test="username-text"
              :persistent-hint="true"
            />
          </v-card-text>
          <v-card-text>
            <v-row class="pl-3 pt-0">
              <v-col>
                <ul>
                  <li>The namespace you choose here will be used for in the SSHID of your devices.</li>
                  <li>The namespace can contain only lowercase alphanumeric characters and hyphens.</li>
                  <li>It cannot begin or end with a hyphen ("-").</li>
                  <li>The namespace must be a minimum of 3 characters and a maximum of 63 characters.</li>
                  <li>The namespace cannot be changed after creation.</li>
                </ul>
              </v-col>
            </v-row>
          </v-card-text>
          <v-divider />
          <v-card-actions>
            <v-spacer />
            <v-btn data-test="close-btn" @click="close">Close</v-btn>
            <v-btn color="primary" variant="outlined" data-test="add-btn" @click="addNamespace" :disabled="!fieldMeta.valid">Submit</v-btn>
          </v-card-actions>
        </v-container>
      </template>
      <template v-else>
        <v-card-title class="bg-primary">Add a namespace using the CLI</v-card-title>
        <v-card-text class="mt-4 mb-0 pb-1 mb-4">
          <p class="text-body-2">
            In the Community Edition of ShellHub, namespaces must be added using the administration CLI.
            For detailed instructions on how to add namespaces, please refer to the documentation at the ShellHub
            Administration Guide.
          </p>
          <div id="cli-instructions" class="mt-3 text-body-2">
            <p class="text-caption mb-0 mt-3" data-test="openContentSecond-text">
              Check the
              <a
                :href="'https://docs.shellhub.io/self-hosted/administration'"
                target="_blank"
                rel="noopener noreferrer"
              >ShellHub Administration Guide</a>
              for more information.
            </p>
          </div>
        </v-card-text>
      </template>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import { envVariables } from "@/envVariables";
import BaseDialog from "../BaseDialog.vue";

const store = useStore();
const snackbar = useSnackbar();
const showDialog = defineModel({ default: false });
const isCommunityVersion = computed(() => envVariables.isCommunity);

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
    await store.dispatch("namespaces/switchNamespace", { tenant_id: tenantId });
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
  try {
    const response = await store.dispatch("namespaces/post", namespaceName.value);
    await changeNamespace(response.data.tenant_id);
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
  }
};
</script>
