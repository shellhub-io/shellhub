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

  <BaseDialog v-model="showDialog">
    <v-card class="bg-v-theme-surface" min-height="300">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="title">
        Edit Api Key Name
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-3">
        <v-text-field
          v-model="keyName"
          class="mb-5"
          label="key name"
          prepend-icon="mdi-key-outline"
          :error-messages="keyNameError"
          required
          variant="underlined"
          data-test="key-name-text"
          messages="Please note that the new name must be unique
          and not already in use by another key."
        />

        <RoleSelect
          v-if="hasAuthorization"
          v-model="selectedRole"
          data-test="namespace-generate-role"
        />
      </v-card-text>

      <v-card-actions>
        <v-btn variant="text" @click="showDialog = false" data-test="close-btn"> Close </v-btn>
        <v-spacer />
        <v-btn color="success" variant="flat" data-test="edit-btn" @click="edit()" :disabled="!!keyNameError">
          Edit key
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "@/components/BaseDialog.vue";
import RoleSelect from "@/components/Team/RoleSelect.vue";
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
const apiKeyStore = useApiKeysStore();
const snackbar = useSnackbar();
const selectedRole = ref<BasicRole>(props.keyRole as BasicRole);

const {
  value: keyName,
  errorMessage: keyNameError,
  setErrors: setKeyNameError,
} = useField<string>(
  "name",
  yup
    .string()
    .required()
    .min(3)
    .max(20)
    .matches(/^(?!.*\s).*$/, "This field cannot contain any blank spaces"),
  {
    initialValue: props.keyName,
  },
);

const open = () => {
  showDialog.value = true;
  keyName.value = props.keyName;
  selectedRole.value = props.keyRole as BasicRole;
};

const update = () => {
  emit("update");
  showDialog.value = false;
};

const edit = async () => {
  try {
    await apiKeyStore.editApiKey({
      key: props.keyName,
      name: keyName.value === props.keyName ? undefined : keyName.value, // Only send name if it has changed
      role: selectedRole.value,
    });

    update();
    snackbar.showSuccess("Api Key edited successfully.");
  } catch (error: unknown) {
    snackbar.showError("Failed to edit Api Key.");
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 409) {
        setKeyNameError("An API key with the same name already exists.");
      } else {
        setKeyNameError("An error occurred while editing your API key");
        handleError(error);
      }
      return;
    }
    handleError(error);
  }
};

defineExpose({ keyNameError });
</script>
