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
            <v-select
              v-if="hasAuthorization"
              v-model="selectedRole"
              label="Key Role"
              :items="itemsRoles"
              :item-props="true"
              variant="outlined"
              return-object
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
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
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
  keyRole: {
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
const isOwner = computed(() => store.getters["auth/role"] === "owner");
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

const itemsRoles = [
  {
    title: "observer",
    value: "observer",
  },
  {
    title: "operator",
    value: "operator",
  },
  {
    title: "administrator",
    value: "administrator",
    disabled: !props.hasAuthorization || !isOwner.value,
  },
];

const selectedRole = ref(itemsRoles.find((role) => role.value === props.keyRole) || itemsRoles[0]);

const edit = async () => {
  const equalName = props.keyName === keyInput.value;
  const equalRole = props.keyRole === selectedRole.value.value;
  const payload: { key: string; name?: string; role?: string } = { key: props.keyName };

  if (!equalName) {
    payload.name = keyInput.value;
  }
  if (!equalRole) {
    payload.role = selectedRole.value.value;
  }

  try {
    await store.dispatch("apiKeys/editApiKey", payload);
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
