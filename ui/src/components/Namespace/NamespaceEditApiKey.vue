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

  <v-dialog max-width="450" v-model="showDialog">
    <v-card class="bg-v-theme-surface" min-height="300">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="title">
        Edit Api Key Name
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-3">
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
      </v-card-text>

      <v-card-actions>
        <v-btn variant="text" @click="showDialog = false" data-test="close-btn"> Close </v-btn>
        <v-spacer />
        <v-btn color="success" variant="flat" data-test="edit-btn" @click="edit()">
          Edit key
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";

const props = defineProps({
  keyName: {
    type: String,
    required: true,
  },
  keyId: {
    type: String,
    required: true,
  },
  hasAuthorization: {
    type: Boolean,
    required: true,
  },
  disabled: {
    type: Boolean,
    required: true,
  },
});
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();
const keyGetter = computed(() => props.keyName);
const tenant = computed(() => localStorage.getItem("tenant"));

const {
  value: keyInput,
  errorMessage: keyInputError,
} = useField<string | undefined>("name", yup.string().required(), {
  initialValue: keyGetter.value,
});

const open = () => {
  showDialog.value = true;
  keyInput.value = keyGetter.value;
};

const update = () => {
  emit("update");
  showDialog.value = false;
};

const edit = async () => {
  try {
    await store.dispatch("auth/editApiKey", {
      tenant: tenant.value,
      name: keyInput.value,
      id: props.keyId,
    });
    update();
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.editKey,
    );
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.editKey,
    );
    handleError(error);
  }
};
</script>
