<template>
  <v-list-item
    v-bind="$attrs"
    @click="dialog = true"
    :disabled="!hasAuthorizationCreateTunnel"
    data-test="tunnel-create-dialog-btn"
  >
    <div class="d-flex align-center">
      <div class="mr-2" data-test="create-icon">
        <v-icon>mdi-web-plus</v-icon>
      </div>

      <v-list-item-title> Create Tunnel </v-list-item-title>
    </div>
  </v-list-item>
  <v-dialog v-model="dialog" max-width="450" @click:outside="close()">
    <v-card data-test="tunnel-create-dialog" class="bg-v-theme-surface">
      <v-card-title class="bg-primary" data-test="create-dialog-title"> Create Device Tunnel </v-card-title>
      <v-container>
        <v-alert
          v-if="alertText"
          type="error"
          :text="alertText"
          data-test="tunnel-create-alert"
        />
        <v-card-text>
          <p class="text-grey-lighten-4 mb-2" data-test="tunnel-create-text">
            Configure the address and port to create a tunnel to your device.
          </p>
          <v-row>
            <v-col sm="8" class="pb-0">
              <v-text-field
                v-model="host"
                class="mt-1"
                label="Address"
                :error-messages="hostError"
                variant="underlined"
                data-test="address-text"
              />
            </v-col>
            <p class="mt-7 pa-0"> : </p>
            <v-col class="pb-0">
              <v-text-field
                v-model.number="port"
                label="Port"
                :error-messages="portError"
                variant="outlined"
                data-test="port-text"
              />
            </v-col>
          </v-row>
        </v-card-text>
      </v-container>
      <v-card-actions>
        <v-spacer />
        <v-btn data-test="close-btn" @click="close()"> Close </v-btn>
        <v-btn :disabled="hasErrors()" color="primary" data-test="create-tunnel-btn" @click="addTunnel()">
          Create Tunnel
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import { ref } from "vue";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import { useStore } from "@/store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";

const props = defineProps({
  uid: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["update"]);
const store = useStore();
const dialog = defineModel({ default: false });

const {
  value: host,
  errorMessage: hostError,
  resetField: resetHostRole,
} = useField<string>(
  "host",
  yup
    .string()
    .required(),
  {
    initialValue: "127.0.0.1",
  },
);

const {
  value: port,
  errorMessage: portError,
  resetField: resetPortRole,
} = useField<number>(
  "port",
  yup
    .number()
    .integer()
    .max(65535)
    .required(),
  {
    initialValue: undefined,
  },
);

const hasAuthorizationCreateTunnel = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.tunnel.create,
    );
  }
  return false;
};

const alertText = ref();

const hasErrors = () => !!(
  portError.value
  || hostError.value
  || !port.value
  || !host.value
);

const resetFields = () => {
  resetPortRole();
  resetHostRole();
};

const close = () => {
  resetFields();
  dialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const addTunnel = async () => {
  if (!hasErrors()) {
    try {
      await store.dispatch("tunnels/create", {
        uid: props.uid,
        host: host.value,
        port: port.value,
      });

      store.dispatch(
        "snackbar/showSnackbarSuccessAction",
        INotificationsSuccess.tunnelCreate,
      );
      update();
      resetFields();
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError;
        if (axiosError.response?.status === 403) {
          alertText.value = "This device has reached the maximum allowed number of tunnels";
        } else {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.tunnelCreate,
          );
          handleError(error);
        }
      }
    }
  }
};
</script>
