<template>
  <v-list-item
    v-bind="$attrs"
    @click="showDialog = true"
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

  <BaseDialog v-model="showDialog" @click:outside="close()">
    <v-card data-test="tunnel-create-dialog" class="bg-v-theme-surface">
      <v-card-title class="bg-primary" data-test="create-dialog-title">
        Create Device Tunnel
      </v-card-title>
      <v-container>
        <v-alert
          v-if="alertText"
          type="error"
          :text="alertText"
          data-test="tunnel-create-alert"
        />
        <v-card-text>
          <p class="mb-2" data-test="tunnel-create-text">
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
          <v-row>
            <v-col>
              <v-select
                v-model="selectedTimeout"
                :items="predefinedTimeouts"
                item-title="text"
                item-value="value"
                label="Timeout (in seconds)"
                variant="outlined"
                data-test="timeout-combobox"
              />
            </v-col>
          </v-row>
          <v-row v-if="selectedTimeout === 'custom'">
            <v-col>
              <v-text-field
                v-model.number="customTimeout"
                :error-messages="customTimeoutError"
                label="Custom Timeout (in seconds)"
                type="number"
                variant="outlined"
                data-test="custom-timeout"
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
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";

const props = defineProps<{ uid: string }>();
const emit = defineEmits(["update"]);
const store = useStore();
const snackbar = useSnackbar();
const showDialog = ref(false);
const alertText = ref();

const predefinedTimeouts = ref([
  { value: -1, text: "Unlimited Timeout" },
  { value: 60, text: "1 minute" },
  { value: 300, text: "5 minutes" },
  { value: 900, text: "15 minutes" },
  { value: 3600, text: "1 hour" },
  { value: 86400, text: "1 day" },
  { value: 604800, text: "1 week" },
  { value: 2624016, text: "1 month" },
  { value: "custom", text: "Custom Expiration" },
]);

const { value: host, errorMessage: hostError, resetField: resetHost } = useField<string>(
  "host",
  yup
    .string()
    .required(),
  { initialValue: "127.0.0.1" },
);

const { value: port, errorMessage: portError, resetField: resetPort } = useField<number>(
  "port",
  yup
    .number()
    .integer()
    .max(65535)
    .required(),
  { initialValue: undefined },
);

const { value: customTimeout, errorMessage: customTimeoutError, resetField: resetCustomTimeout } = useField<number>(
  "customTimeout",
  yup
    .number()
    .integer()
    .min(1)
    .max(9223372036)
    .required(),
  { initialValue: 60 },
);

const selectedTimeout = ref<number | "custom">(-1);
const timeout = computed(() => (selectedTimeout.value === "custom" ? customTimeout.value : selectedTimeout.value));

const hasAuthorizationCreateTunnel = () => {
  const role = store.getters["auth/role"];
  return !!role && hasPermission(authorizer.role[role], actions.tunnel.create);
};

const hasErrors = () => !!(portError.value || hostError.value || customTimeoutError.value || !port.value || !host.value || !timeout.value);

const resetFields = () => { resetPort(); resetHost(); selectedTimeout.value = -1; resetCustomTimeout(); };
const close = () => { resetFields(); showDialog.value = false; };
const update = () => { emit("update"); close(); };

const addTunnel = async () => {
  if (!hasErrors()) {
    try {
      await store.dispatch("tunnels/create", { uid: props.uid, host: host.value, port: port.value, ttl: timeout.value });
      snackbar.showSuccess("Tunnel created successfully.");
      update();
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        if ((error as AxiosError).response?.status === 403) {
          alertText.value = "This device has reached the maximum allowed number of tunnels";
        } else {
          snackbar.showError("Failed to create tunnel.");
          handleError(error);
        }
      }
    }
  }
};
</script>
