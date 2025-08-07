<template>
  <BaseDialog v-model="dialog" max-width="450" @click:outside="close()">
    <v-card data-test="tunnel-create-dialog" class="bg-v-theme-surface">
      <v-card-title class="bg-primary" data-test="create-dialog-title">
        Create Device Web Endpoint
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
            Configure the host and port to create a tunnel to your device.
          </p>
          <v-row>
            <v-col sm="8" class="pb-0">
              <v-text-field
                v-model="host"
                class="mt-1"
                label="Host"
                :error-messages="hostError"
                variant="underlined"
                data-test="host-text"
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

          <v-row class="mt-1" v-if="props.useDevicesList">
            <v-col>
              <v-autocomplete
                v-model="selectedDevice"
                :items="deviceOptions"
                :loading="loadingDevices"
                item-title="info.pretty_name"
                item-value="uid"
                label="Select Device"
                variant="outlined"
                return-object
                hide-details
                @click:control="() => fetchDevices()"
                @update:search="fetchDevices"
                data-test="web-endpoint-autocomplete"
              >
                <template #item="{ item, props }">
                  <v-list-item
                    v-bind="props"
                  >
                    <div>
                      <DeviceIcon
                        :icon="item.raw.info.id"
                        class="mr-2"
                      />
                      <span class="text-body-1">{{ item.raw.name }}</span>
                    </div>

                  </v-list-item>
                </template>

                <template #selection="{ item }">
                  <div class="d-flex align-center">
                    <DeviceIcon
                      :icon="item.raw.info.id"
                      class="mr-2"
                    />
                    <span class="text-body-1">{{ item.raw.name }}</span>
                  </div>
                </template>
              </v-autocomplete>

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
        <v-btn :disabled="hasErrors" color="primary" data-test="create-tunnel-btn" @click="addWebEndpoint()">
          Create Web Endpoint
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
import DeviceIcon from "@/components/Devices/DeviceIcon.vue";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";

interface DeviceOption {
  uid: string;
  name: string;
  info: {
    id: string;
    pretty_name: string;
  };
  [key: string]: unknown;
}

const props = defineProps({
  uid: { type: String, required: false, default: "" },
  useDevicesList: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits(["update"]);
const store = useStore();
const snackbar = useSnackbar();
const dialog = defineModel({ default: false });
const alertText = ref();

const selectedDevice = ref<DeviceOption | null>(null);
const deviceOptions = ref<DeviceOption[]>([]);
const loadingDevices = ref(false);

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

const hasErrors = computed(() => {
  const formInvalid = !!portError.value
    || !!hostError.value
    || !!customTimeoutError.value
    || !port.value
    || !host.value
    || !timeout.value;

  if (props.useDevicesList) {
    return formInvalid || !selectedDevice.value;
  }

  return formInvalid;
});

const resetFields = () => { resetPort(); resetHost(); selectedTimeout.value = -1; resetCustomTimeout(); };
const close = () => { resetFields(); dialog.value = false; };
const update = () => { emit("update"); close(); };

const fetchDevices = async (val?: string) => {
  if (!val && deviceOptions.value.length > 0) return;

  loadingDevices.value = true;

  const filter = val
    ? btoa(JSON.stringify([
      { type: "property", params: { name: "name", operator: "contains", value: val } },
    ]))
    : "";

  try {
    await store.dispatch("devices/search", {
      page: 1,
      perPage: 10,
      filter,
      status: "accepted",
    });

    deviceOptions.value = store.getters["devices/list"];
  } catch {
    snackbar.showError("Failed to load devices.");
  } finally {
    loadingDevices.value = false;
  }
};

const addWebEndpoint = async () => {
  if (hasErrors.value) return;

  const deviceUid = props.useDevicesList
    ? selectedDevice.value?.uid
    : props.uid;

  try {
    await store.dispatch("webEndpoints/create", {
      uid: deviceUid,
      host: host.value,
      port: port.value,
      ttl: timeout.value,
    });

    snackbar.showSuccess("Web Endpoint created successfully.");
    update();
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      if ((error as AxiosError).response?.status === 403) {
        alertText.value = "This device has reached the maximum allowed number of Web Endpoints";
      } else {
        snackbar.showError("Failed to create Web Endpoint.");
        handleError(error);
      }
    }
  }
};

</script>
