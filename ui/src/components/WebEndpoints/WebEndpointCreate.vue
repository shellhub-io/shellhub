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
              hint="IPv4 or IPv6 only"
              persistent-hint
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
              type="number"
              data-test="port-text"
            />
          </v-col>
        </v-row>

        <v-row class="mt-1" v-if="props.useDevicesList">
          <v-col>
            <v-autocomplete
              v-model="selectedDevice"
              v-model:search="deviceSearch"
              :items="deviceOptions"
              :loading="loadingDevices"
              item-title="info.pretty_name"
              item-value="uid"
              label="Select Device"
              variant="outlined"
              return-object
              hide-details
              :no-filter="true"
              @update:search="onSearchUpdate"
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
              hide-details
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
              hide-details
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
import type { IDevice } from "@/interfaces/IDevice";

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

// eslint-disable-next-line vue/max-len
const ipv4Regex = /^(25[0-5]|2[0-4]\d|1?\d{1,2})\.(25[0-5]|2[0-4]\d|1?\d{1,2})\.(25[0-5]|2[0-4]\d|1?\d{1,2})\.(25[0-5]|2[0-4]\d|1?\d{1,2})$/;

// eslint-disable-next-line vue/max-len
const ipv6Regex = /^((?:[0-9A-Fa-f]{1,4}:){7}[0-9A-Fa-f]{1,4}|(?:[0-9A-Fa-f]{1,4}:){1,7}:|(?:[0-9A-Fa-f]{1,4}:){1,6}:[0-9A-Fa-f]{1,4}|(?:[0-9A-Fa-f]{1,4}:){1,5}(?::[0-9A-Fa-f]{1,4}){1,2}|(?:[0-9A-Fa-f]{1,4}:){1,4}(?::[0-9A-Fa-f]{1,4}){1,3}|(?:[0-9A-Fa-f]{1,4}:){1,3}(?::[0-9A-Fa-f]{1,4}){1,4}|(?:[0-9A-Fa-f]{1,4}:){1,2}(?::[0-9A-Fa-f]{1,4}){1,5}|[0-9A-Fa-f]{1,4}:(?::[0-9A-Fa-f]{1,4}){1,6}|:(?::[0-9A-Fa-f]{1,4}){1,7}|fe80:(?::[0-9A-Fa-f]{0,4}){0,4}%[0-9A-Za-z]{1,}|::(?:ffff(?::0{1,4})?:)?(?:25[0-5]|2[0-4]\d|1?\d{1,2})(?:\.(?:25[0-5]|2[0-4]\d|1?\d{1,2})){3}|(?:[0-9A-Fa-f]{1,4}:){1,4}:(?:25[0-5]|2[0-4]\d|1?\d{1,2})(?:\.(?:25[0-5]|2[0-4]\d|1?\d{1,2})){3})$/;

const selectedDevice = ref<IDevice | null>(null);
const deviceOptions = ref<IDevice[]>([]);
const loadingDevices = ref(false);
const deviceSearch = ref("");

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
    .required("Host is required")
    .test("is-ipv4-or-ipv6", "Enter a valid IPv4 or IPv6 address", (value) => {
      const v = (value || "").trim();
      return ipv4Regex.test(v) || ipv6Regex.test(v);
    }),
  { initialValue: "127.0.0.1" },
);

const { value: port, errorMessage: portError, resetField: resetPort } = useField<number>(
  "port",
  yup.number().typeError("Port is a number between 1 and 65535").integer().min(1)
    .max(65535)
    .required(),
  { initialValue: undefined },
);

const {
  value: customTimeout,
  errorMessage: customTimeoutError,
  resetField: resetCustomTimeout,
} = useField<number>(
  "customTimeout",
  yup.number().integer().min(1).max(9223372036)
    .required(),
  { initialValue: 60 },
);

const selectedTimeout = ref<number | "custom">(-1);
const timeout = computed(() => selectedTimeout.value === "custom" ? customTimeout.value : selectedTimeout.value);

const hasErrors = computed(() => {
  const needsCustom = selectedTimeout.value === "custom";
  const formInvalid = !!portError.value
    || !!hostError.value
    || (needsCustom && !!customTimeoutError.value)
    || !port.value
    || !host.value
    || timeout.value === undefined
    || timeout.value === null;

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
  deviceSearch.value = "";
  deviceOptions.value = [];
};

const clearFilterAndRefetch = async () => {
  devicesStore.deviceListFilter = undefined;
  await devicesStore.fetchDeviceList({ filter: undefined });
  deviceOptions.value = devicesStore.devices;
};

const close = async () => {
  resetFields();
  showDialog.value = false;
  await clearFilterAndRefetch();
};

const update = () => {
  emit("update");
  close();
};

const fetchDevices = async (searchQuery?: string) => {
  loadingDevices.value = true;

  const query = (searchQuery ?? deviceSearch.value ?? "").trim();
  const filter = query
    ? Buffer.from(
      JSON.stringify([
        { type: "property", params: { name: "name", operator: "contains", value: query } },
      ]),
    ).toString("base64")
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

const onSearchUpdate = async (val: string) => {
  deviceSearch.value = val;
  const query = val.trim();
  if (!query) {
    await clearFilterAndRefetch();
    return;
  }
  await fetchDevices(query);
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
  if (props.useDevicesList) await clearFilterAndRefetch();
});
</script>
