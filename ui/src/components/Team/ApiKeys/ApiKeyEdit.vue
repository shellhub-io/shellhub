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

        <v-row>
          <v-col>
            <v-text-field
              v-model="keyInput"
              label="key name"
              prepend-icon="mdi-key-outline"
              :error-messages="keyInputError"
              required
              variant="underlined"
              data-test="key-name-text"
              messages="Please note that the new name must be unique
          and not already in use by another key."
            />

          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <RoleSelect
              v-if="hasAuthorization"
              v-model="selectedRole"
              data-test="namespace-generate-role"
            />
          </v-col>
        </v-row>
      </v-card-text>

      <v-card-actions>
        <v-btn variant="text" @click="showDialog = false" data-test="close-btn"> Close </v-btn>
        <v-spacer />
        <v-btn color="success" variant="flat" data-test="edit-btn" @click="edit()">
          Edit key
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "@/components/BaseDialog.vue";
import RoleSelect from "@/components/Team/RoleSelect.vue";
import { BasicRole } from "@/interfaces/INamespace";

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
const store = useStore();
const snackbar = useSnackbar();
const keyGetter = computed(() => props.keyName);
const {
  value: keyInput,
  errorMessage: keyInputError,
  setErrors: setKeyInputError,
} = useField<string | undefined>(
  "name",
  yup
    .string()
    .required()
    .min(3)
    .max(20)
    .matches(/^(?!.*\s).*$/, "This field cannot contain any blankspaces"),
  {
    initialValue: keyGetter.value,
  },
);

const open = () => {
  showDialog.value = true;
  keyInput.value = keyGetter.value;
};

const update = () => {
  emit("update");
  showDialog.value = false;
};

const selectedRole = ref<BasicRole>(props.keyRole as BasicRole);

const edit = async () => {
  const equalName = props.keyName === keyInput.value;
  const equalRole = props.keyRole === selectedRole.value;
  const payload: { key: string; name?: string; role?: string } = { key: props.keyName };

  if (!equalName) {
    payload.name = keyInput.value;
  }
  if (!equalRole) {
    payload.role = selectedRole.value;
  }

  try {
    await store.dispatch("apiKeys/editApiKey", payload);
    update();
    snackbar.showSuccess("Api Key edited successfully.");
  } catch (error: unknown) {
    snackbar.showError("Failed to edit Api Key.");
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 409:
          setKeyInputError("An API key with the same name already exists.");
          break;
        default:
          setKeyInputError("An error occurred while editing your API key");
          handleError(error);
      }
      return;
    }
    handleError(error);
  }
};

defineExpose({ keyInputError });
</script>
