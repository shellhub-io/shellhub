<template>
  <FormDialog
    v-model="showDialog"
    @close="close"
    @cancel="close"
    @confirm="addWebEndpoint"
    title="Create Device Web Endpoint"
    icon="mdi-lan"
    confirm-text="Create Web Endpoint"
    cancel-text="Close"
    :confirm-disabled="hasErrors"
    :alert-message="alertText"
    confirm-data-test="create-tunnel-btn"
    cancel-data-test="close-btn"
    data-test="tunnel-create-dialog"
  >
    <v-container>
      <v-card-text class="pa-0">
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
              @update:search="fetchDevices"
              data-test="web-endpoint-autocomplete"
            >
              <template #item="{ item, props }">
                <v-list-item v-bind="props">
                  <div>
                    <DeviceIcon :icon="item.raw.info.id" class="mr-2" />
                    <span class="text-body-1">{{ item.raw.name }}</span>
                  </div>
                </v-list-item>
              </template>

              <template #selection="{ item }">
                <div class="d-flex align-center">
                  <DeviceIcon :icon="item.raw.info.id" class="mr-2" />
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
  </FormDialog>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import DeviceIcon from "@/components/Devices/DeviceIcon.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useDevicesStore from "@/store/modules/devices";
import useWebEndpointsStore from "@/store/modules/web_endpoints";
import { IDevice } from "@/interfaces/IDevice";

const props = defineProps<{
  uid?: string;
  useDevicesList: boolean;
}>();

const emit = defineEmits(["update"]);
const devicesStore = useDevicesStore();
const webEndpointsStore = useWebEndpointsStore();
const snackbar = useSnackbar();
const showDialog = defineModel({ default: false });
const alertText = ref("");

const selectedDevice = ref<IDevice | null>(null);
const deviceOptions = ref<IDevice[]>([]);
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
  yup.string().required(),
  { initialValue: "127.0.0.1" },
);

const { value: port, errorMessage: portError, resetField: resetPort } = useField<number>(
  "port",
  yup.number().integer().max(65535).required(),
  { initialValue: undefined },
);

const { value: customTimeout, errorMessage: customTimeoutError, resetField: resetCustomTimeout } = useField<number>(
  "customTimeout",
  yup.number().integer().min(1).max(9223372036)
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

  if (props.useDevicesList) return formInvalid || !selectedDevice.value;
  return formInvalid;
});

const resetFields = () => {
  resetPort();
  resetHost();
  selectedTimeout.value = -1;
  resetCustomTimeout();
  alertText.value = "";
  selectedDevice.value = null;
};

const close = () => {
  resetFields();
  showDialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const fetchDevices = async (searchQuery?: string) => {
  loadingDevices.value = true;

  const filter = searchQuery
    ? btoa(JSON.stringify([
      { type: "property", params: { name: "name", operator: "contains", value: searchQuery } },
    ]))
    : undefined;

  try {
    await devicesStore.fetchDeviceList({ filter });
    deviceOptions.value = devicesStore.devices;
  } catch (error) {
    snackbar.showError("Failed to load devices.");
    handleError(error);
  } finally {
    loadingDevices.value = false;
  }
};

const addWebEndpoint = async () => {
  if (hasErrors.value) return;

  const deviceUid = props.useDevicesList ? selectedDevice.value?.uid : props.uid;

  try {
    await webEndpointsStore.createWebEndpoint({
      uid: deviceUid as string,
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
        return;
      }
    }
    snackbar.showError("Failed to create Web Endpoint.");
    handleError(error);
  }
};

onMounted(async () => {
  if (props.useDevicesList) await fetchDevices();
});
</script>
