<template>
  <FormDialog
    v-model="showDialog"
    title="Create Device Web Endpoint"
    icon="mdi-lan"
    confirm-text="Create Web Endpoint"
    cancel-text="Close"
    :confirm-disabled="hasErrors"
    :alert-message="alertText"
    confirm-data-test="create-tunnel-btn"
    cancel-data-test="close-btn"
    data-test="tunnel-create-dialog"
    @close="close"
    @cancel="close"
    @confirm="addWebEndpoint"
  >
    <v-container>
      <v-card-text class="pa-0">
        <p
          class="mb-2"
          data-test="tunnel-create-text"
        >
          Configure the host and port to create a tunnel to your device.
        </p>

        <v-row>
          <v-col
            sm="8"
            class="pb-0"
          >
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

          <p class="mt-7 pa-0">:</p>

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

        <v-row
          v-if="props.useDevicesList"
          class="mt-1"
        >
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
              data-test="web-endpoint-autocomplete"
              @update:search="onSearchUpdate"
            >
              <template #item="{ item, props }">
                <v-list-item v-bind="props">
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
              hide-details
              data-test="timeout-combobox"
            />
          </v-col>
        </v-row>

        <v-row v-if="selectedTimeout === 'custom'">
          <v-col>
            <v-text-field
              v-model.number="customTimeout"
              :error-messages="customTimeoutErrorMsg"
              label="Custom Timeout (in seconds)"
              type="number"
              hide-details
              variant="outlined"
              data-test="custom-timeout"
            />
          </v-col>
        </v-row>

        <v-divider class="my-4" />
        <div class="text-subtitle-1">TLS</div>

        <v-row>
          <v-col
            cols="12"
            md="6"
          >
            <v-checkbox
              v-model="tlsEnabled"
              label="Enable TLS (HTTPS)"
              hint="Use HTTPS when creating the web endpoint"
              persistent-hint
              data-test="tls-enabled-checkbox"
              @update:model-value="onTlsEnabledChange(tlsEnabled)"
            />
          </v-col>

          <v-expand-transition v-show="tlsEnabled">
            <v-col>
              <v-checkbox
                v-model="tlsVerify"
                label="Verify certificate"
                hint="Validate the server certificate using the Domain below"
                persistent-hint
                data-test="tls-verify-checkbox"
              />
            </v-col>
          </v-expand-transition>
        </v-row>

        <v-expand-transition>
          <div
            v-show="tlsEnabled"
            data-test="tls-accordion"
          >
            <v-row>
              <v-col>
                <v-text-field
                  v-model="tlsDomain"
                  :error-messages="tlsDomainError"
                  label="TLS Domain"
                  hint="Example: example.com or device.local"
                  persistent-hint
                  variant="outlined"
                  data-test="tls-domain-text"
                />
              </v-col>
            </v-row>
          </div>
        </v-expand-transition>
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
import { IWebEndpointsCreate } from "@/interfaces/IWebEndpoints";

const props = defineProps<{
  uid?: string;
  useDevicesList: boolean;
}>();

const emit = defineEmits(["update"]);
const devicesStore = useDevicesStore();
const webEndpointsStore = useWebEndpointsStore();
const snackbar = useSnackbar();
const showDialog = defineModel<boolean>({ required: true });
const alertText = ref("");

// IPv4 / IPv6 regexes
const ipv4Regex
  = /^(25[0-5]|2[0-4]\d|1?\d{1,2})\.(25[0-5]|2[0-4]\d|1?\d{1,2})\.(25[0-5]|2[0-4]\d|1?\d{1,2})\.(25[0-5]|2[0-4]\d|1?\d{1,2})$/;

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

const {
  value: host,
  errorMessage: hostError,
  resetField: resetHost,
} = useField<string>(
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

const {
  value: port,
  errorMessage: portError,
  resetField: resetPort,
} = useField<number>(
  "port",
  yup
    .number()
    .typeError("Port is a number between 1 and 65535")
    .integer()
    .min(1)
    .max(65535)
    .required(),
  { initialValue: undefined },
);

const {
  value: customTimeout,
  errorMessage: customTimeoutErrorMsg,
  resetField: resetCustomTimeout,
} = useField<number>(
  "customTimeout",
  yup.number().integer().min(1).max(9223372036).required(),
  { initialValue: 60 },
);

const selectedTimeout = ref<number | "custom">(-1);
const timeout = computed(() =>
  selectedTimeout.value === "custom" ? customTimeout.value : selectedTimeout.value,
);

const tlsEnabled = ref<boolean>(false);
const tlsVerify = ref<boolean>(false);

const isFQDN = (value: string): boolean => {
  if (!value) return false;

  const cleaned = value.trim();

  const fqdnRegex = /^([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?(\.[a-zA-Z]{1}[a-zA-Z0-9]{0,62})\.?$/;

  return fqdnRegex.test(cleaned);
};

const {
  value: tlsDomain,
  errorMessage: tlsDomainError,
  resetField: resetTlsDomain,
} = useField<string>(
  "tlsDomain",
  yup
    .string()
    .trim()
    .when([], {
      is: () => tlsEnabled.value,
      then: (schema) =>
        schema
          .required("Domain is required when TLS is enabled")
          .test(
            "valid-domain",
            "Enter a valid FQDN (e.g., example.com or device.local)",
            (value) => isFQDN(value || ""),
          ),
      otherwise: (schema) => schema,
    }),
  { initialValue: "" },
);

const tlsDomainNormalized = computed(() =>
  tlsDomain.value.trim().replace(/^\[|\]$/g, ""),
);

const onTlsEnabledChange = (enabled: boolean) => {
  if (!enabled) {
    tlsVerify.value = false;
    tlsDomain.value = "";
  }
};

const hasErrors = computed(() => {
  const baseErrors
    = !!portError.value
      || !!hostError.value
      || !port.value
      || !host.value
      || timeout.value === undefined
      || timeout.value === null;

  const customTimeoutInvalid
    = selectedTimeout.value === "custom" && !!customTimeoutErrorMsg.value;

  const tlsError = (tlsEnabled.value && !!tlsDomainError.value) || (tlsEnabled.value && !tlsDomain.value);

  const deviceError = props.useDevicesList && !selectedDevice.value;

  return baseErrors || customTimeoutInvalid || tlsError || deviceError;
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

  tlsEnabled.value = false;
  tlsVerify.value = false;
  resetTlsDomain();
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

const update = async () => {
  emit("update");
  await close();
};

const fetchDevices = async (searchQuery?: string) => {
  loadingDevices.value = true;

  const query = (searchQuery ?? deviceSearch.value ?? "").trim();
  const filter = query
    ? Buffer.from(
      JSON.stringify([
        {
          type: "property",
          params: { name: "name", operator: "contains", value: query },
        },
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
    const payload: IWebEndpointsCreate = {
      uid: deviceUid as string,
      host: host.value,
      port: port.value,
      ttl: timeout.value,
    };

    if (tlsEnabled.value) {
      payload.tls = {
        enabled: true,
        verify: tlsVerify.value,
        domain: tlsDomainNormalized.value || "",
      };
    }

    await webEndpointsStore.createWebEndpoint(payload);

    snackbar.showSuccess("Web Endpoint created successfully.");
    await update();
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      if ((error as AxiosError).response?.status === 403) {
        alertText.value
          = "This device has reached the maximum allowed number of Web Endpoints";
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
