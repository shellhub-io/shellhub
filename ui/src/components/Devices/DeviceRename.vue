<template>
  <v-list-item
    v-bind="$attrs"
    data-test="rename-device-button"
    @click="open()"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon data-test="rename-icon">
          mdi-pencil
        </v-icon>
      </div>

      <v-list-item-title data-test="rename-title">
        Rename
      </v-list-item-title>
    </div>
  </v-list-item>

  <FormDialog
    v-model="showDialog"
    title="Rename Device"
    icon="mdi-pencil"
    confirm-text="Rename"
    cancel-text="Close"
    confirm-data-test="rename-btn"
    cancel-data-test="close-btn"
    data-test="device-rename-dialog"
    @close="close"
    @confirm="rename"
    @cancel="close"
  >
    <div class="px-6 pt-6 pb-4">
      <v-text-field
        v-model="newName"
        label="Hostname"
        :error-messages="newNameError"
        :messages="messages"
        required
        variant="outlined"
        data-test="rename-field"
      />
    </div>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useDevicesStore from "@/store/modules/devices";

const props = defineProps<{
  uid: string,
  name: string
}>();

const emit = defineEmits<{
  update: [];
}>();

const showDialog = ref(false);
const snackbar = useSnackbar();
const messages = ref(
  "Examples: (foobar, foo-bar-ba-z-qux, foo-example, 127-0-0-1)",
);
const devicesStore = useDevicesStore();
const {
  value: newName,
  errorMessage: newNameError,
  setErrors: setNewNameError,
} = useField<string>("name", yup.string().required(), {
  initialValue: props.name,
});

const open = () => {
  newName.value = props.name;
  showDialog.value = true;
};

const close = () => {
  setNewNameError("");
  showDialog.value = false;
};

const rename = async () => {
  try {
    await devicesStore.renameDevice({
      uid: props.uid,
      name: { name: newName.value },
    });

    close();
    emit("update");
    snackbar.showSuccess("Device renamed successfully.");
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 400) {
        setNewNameError("The characters being used are invalid");
      } else if (error.response?.status === 409) {
        setNewNameError("The name already exists in the namespace");
      }
      handleError(error);
    } else {
      snackbar.showError("Failed to rename the device.");
      handleError(error);
    }
  }
};
</script>
