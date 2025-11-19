<template>
  <div>
    <v-tooltip
      location="bottom"
      class="text-center"
      :disabled="canGenerateApiKey"
    >
      <template #activator="{ props }">
        <div v-bind="props">
          <v-btn
            :disabled="!canGenerateApiKey"
            color="primary"
            variant="elevated"
            data-test="api-key-generate-main-btn"
            @click="showDialog = true"
          >
            Generate key
          </v-btn>
        </div>
      </template>
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <FormDialog
      v-model="showDialog"
      title="New API Key"
      description="Generate API key for this namespace"
      icon="mdi-key-outline"
      confirm-text="Generate Api Key"
      :confirm-disabled="!isFormValid"
      confirm-data-test="add-btn"
      cancel-data-test="close-btn"
      data-test="api-key-generate-dialog"
      :alert-message="errorMessage"
      alert-type="error"
      footer-helper-text="Learn more about"
      footer-helper-link-text="API Keys"
      footer-helper-link="https://docs.shellhub.io/user-guides/settings/namespace/api-keys/"
      @close="close"
      @confirm="handleSubmit"
      @cancel="close"
      @alert-dismissed="errorMessage = ''"
    >
      <ApiKeyForm
        ref="formRef"
        mode="create"
        :can-manage-roles="canGenerateApiKey"
        @submit="generateKey"
        @update:valid="isFormValid = $event"
      />
    </FormDialog>

    <ApiKeySuccess
      v-model="showSuccessDialog"
      :api-key="generatedApiKey"
      @update:model-value="handleSuccessDialogUpdate"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import axios from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import ApiKeyForm from "./ApiKeyForm.vue";
import ApiKeySuccess from "./ApiKeySuccess.vue";
import useApiKeysStore from "@/store/modules/api_keys";
import { BasicRole } from "@/interfaces/INamespace";
import { IApiKeyCreate } from "@/interfaces/IApiKey";

const emit = defineEmits(["update"]);
const snackbar = useSnackbar();
const apiKeyStore = useApiKeysStore();

const showDialog = ref(false);
const showSuccessDialog = ref(false);
const errorMessage = ref("");
const generatedApiKey = ref("");
const isFormValid = ref(false);
const formRef = ref<InstanceType<typeof ApiKeyForm>>();
const canGenerateApiKey = hasPermission("apiKey:create");

const wasSuccessful = ref(false);

const handleGenerateKeyError = (error: unknown) => {
  snackbar.showError("Failed to generate API Key.");

  if (axios.isAxiosError(error)) {
    const status = error.response?.status;
    switch (status) {
      case 400:
        errorMessage.value = "Please provide a name for the API key.";
        break;
      case 401:
        errorMessage.value = "You are not authorized to create an API key.";
        break;
      case 409:
        errorMessage.value = "An API key with the same name already exists.";
        break;
      default:
        errorMessage.value = "An error occurred while generating your API key. Please try again later.";
        handleError(error);
    }
  } else {
    handleError(error);
  }
};

const handleSubmit = () => {
  formRef.value?.submitForm();
};

const generateKey = async (formData: { name: string; expires_in?: number; role: BasicRole }) => {
  try {
    generatedApiKey.value = await apiKeyStore.generateApiKey(formData as IApiKeyCreate);
    wasSuccessful.value = true;
    showDialog.value = false;
    showSuccessDialog.value = true;
  } catch (error: unknown) {
    handleGenerateKeyError(error);
  }
};

const handleSuccessDialogUpdate = (value: boolean) => {
  showSuccessDialog.value = value;

  if (!value && wasSuccessful.value) {
    emit("update");
    wasSuccessful.value = false;
    generatedApiKey.value = "";
  }
};

const close = () => {
  showDialog.value = false;
  showSuccessDialog.value = false;
  errorMessage.value = "";
  wasSuccessful.value = false;
  generatedApiKey.value = "";
  formRef.value?.reset();
};

defineExpose({ generateKey, showDialog, errorMessage, close });
</script>
