<template>
  <v-list-item v-bind="$attrs" @click="open()">
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon data-test="rename-icon"> mdi-pencil </v-icon>
      </div>

      <v-list-item-title data-test="rename-title"> Rename </v-list-item-title>
    </div>
  </v-list-item>

  <BaseDialog v-model="showDialog" data-test="device-rename-dialog" @click:outside="close">
    <v-card class="bg-v-theme-surface" data-test="device-rename-card">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="text-title">
        Rename Device
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1">
        <v-text-field
          v-model="newName"
          label="Hostname"
          :error-messages="newNameError"
          :messages="messages"
          required
          variant="underlined"
          data-test="rename-field"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn data-test="close-btn" variant="text" @click="close()">
          Close
        </v-btn>

        <v-btn
          data-test="rename-btn"
          color="primary darken-1"
          variant="text"
          @click="rename()"
        >
          Rename
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
import BaseDialog from "../BaseDialog.vue";
import useDevicesStore from "@/store/modules/devices";

const props = defineProps<{
  uid: string,
  name: string
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
} = useField<string | undefined>("name", yup.string().required(), {
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

defineExpose({ showDialog });
</script>
