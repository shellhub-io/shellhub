<template>
  <v-list-item @click="open()" :disabled="!hasAuthorization || props.disabled">
    <div class="d-flex align-center">

      <div class="d-flex align-center">
        <div class="mr-2" data-test="edit-icon">
          <v-icon>mdi-pencil</v-icon>
        </div>

        <v-list-item-title data-test="edit-main-btn-title"> Edit </v-list-item-title>
      </div>
    </div>
  </v-list-item>

  <FormDialog
    v-model="showDialog"
    @close="showDialog = false"
    @confirm="handleSubmit"
    @cancel="showDialog = false"
    @alert-dismissed="errorMessage = ''"
    title="Edit API Key"
    description="Update the name and role for this API key"
    icon="mdi-pencil"
    confirm-text="Save Changes"
    :confirm-disabled="!isFormValid"
    confirm-data-test="edit-btn"
    cancel-data-test="close-btn"
    data-test="edit-dialog"
    :alert-message="errorMessage"
    alert-type="error"
  >
    <ApiKeyForm
      ref="formRef"
      mode="edit"
      :initial-key-name="keyName"
      :initial-role="keyRole"
      :can-manage-roles="hasAuthorization"
      @submit="editKey"
      @update:valid="isFormValid = $event"
    />
  </FormDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import axios from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import FormDialog from "@/components/FormDialog.vue";
import ApiKeyForm from "./ApiKeyForm.vue";
import { BasicRole } from "@/interfaces/INamespace";
import useApiKeysStore from "@/store/modules/api_keys";

defineOptions({
  inheritAttrs: false,
});

const props = defineProps<{
  keyName: string;
  keyRole: string;
  hasAuthorization: boolean;
  disabled: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const errorMessage = ref("");
const isFormValid = ref(false);
const formRef = ref();
const apiKeyStore = useApiKeysStore();
const snackbar = useSnackbar();

const open = () => {
  showDialog.value = true;
  errorMessage.value = "";
};

const handleSubmit = () => {
  formRef.value?.submitForm();
};

const editKey = async (formData: { name: string; role: BasicRole }) => {
  try {
    await apiKeyStore.editApiKey({
      key: props.keyName,
      name: formData.name === props.keyName ? undefined : formData.name, // Only send name if it has changed
      role: formData.role,
    });

    emit("update");
    showDialog.value = false;
    snackbar.showSuccess("API Key edited successfully.");
  } catch (error: unknown) {
    snackbar.showError("Failed to edit API Key.");
    if (axios.isAxiosError(error)) {
      const status = error.response?.status;
      if (status === 409) {
        errorMessage.value = "An API key with the same name already exists.";
      } else {
        errorMessage.value = "An error occurred while editing your API key.";
        handleError(error);
      }
      return;
    }
    handleError(error);
  }
};

defineExpose({ errorMessage });
</script>
