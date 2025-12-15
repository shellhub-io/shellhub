<template>
  <FormDialog
    v-if="canChooseDevices"
    v-model="showDialog"
    title="Update account or select three devices"
    icon="mdi-developer-board"
    confirm-text="Accept"
    cancel-text="Close"
    :confirm-disabled="disableButton"
    confirm-data-test="accept-btn"
    cancel-data-test="close-btn"
    threshold="md"
    data-test="device-chooser-dialog"
    @close="close"
    @confirm="accept"
    @cancel="close"
  >
    <div class="px-6 pt-4">
      <p
        class="text-body-2 mb-4"
        data-test="subtext"
      >
        You currently have no subscription to the
        <router-link :to="billingUrl">
          premium plan
        </router-link> and the free version is limited to
        3 devices. To unlock access to all devices, you can subscribe to the
        <router-link :to="billingUrl">
          premium plan
        </router-link>. If you want to continue on
        the free plan, you need to select three devices.
      </p>

      <div
        v-if="isAllDevicesTab && hasDevices"
        class="mb-4"
      >
        <v-text-field
          v-model.trim="filter"
          label="Search by hostname"
          variant="outlined"
          color="primary"
          single-line
          hide-details
          append-inner-icon="mdi-magnify"
          density="comfortable"
          data-test="search-text"
          @keyup="searchDevices"
        />
      </div>

      <v-tabs
        v-model="tab"
        align-tabs="center"
        color="primary"
        class="mb-4"
        data-test="v-tabs"
      >
        <v-tab
          v-for="(item, id) in tabItems"
          :key="id"
          :value="id"
          :disabled="item.disabled"
          :data-test="item.title + '-tab'"
        >
          {{ item.title }} Devices
        </v-tab>
      </v-tabs>

      <v-window
        v-model="tab"
        @update:model-value="handleTabChange"
      >
        <v-window-item
          v-for="(item, id) in tabItems"
          :key="id"
          :value="id"
        >
          <DeviceListChooser
            :is-selectable="item.selectable"
            data-test="device-list-chooser-component"
          />
        </v-window-item>
      </v-window>
    </div>
  </FormDialog>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import axios, { AxiosError } from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import DeviceListChooser from "./DeviceListChooser.vue";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useDevicesStore from "@/store/modules/devices";
import { IDevice } from "@/interfaces/IDevice";

const devicesStore = useDevicesStore();
const snackbar = useSnackbar();

const showDialog = computed(() => devicesStore.showDeviceChooser);
const disableSuggestedTab = ref(devicesStore.suggestedDevices.length <= 0);
const tab = ref(disableSuggestedTab.value ? "all" : "suggested");
const filter = ref("");
const billingUrl = "/settings/billing";
const isAllDevicesTab = computed(() => tab.value === "all");
const hasDevices = computed(() => devicesStore.devices.length > 0);
const disableButton = computed(() => (
  (devicesStore.selectedDevices.length <= 0
    || devicesStore.selectedDevices.length > 3)
  && tab.value === "all"
));

const canChooseDevices = hasPermission("device:choose");

const tabItems = ref({
  suggested: {
    title: "Suggested",
    selectable: false,
    disabled: disableSuggestedTab,
  },
  all: {
    title: "All",
    selectable: true,
    disabled: false,
  },
});

const close = () => { devicesStore.showDeviceChooser = false; };

const searchDevices = async () => {
  const filterToEncodeBase64 = [{
    type: "property",
    params: { name: "name", operator: "contains", value: filter.value },
  }];

  const encodedFilter = filter.value ? btoa(JSON.stringify(filterToEncodeBase64)) : undefined;

  try {
    await devicesStore.fetchDeviceList({
      perPage: 5,
      filter: encodedFilter,
    });
  } catch {
    snackbar.showError("An error occurred while searching for devices.");
  }
};

const sendDevicesChoice = async (devices: Array<IDevice>) => {
  try {
    await devicesStore.sendDeviceChoices(devices);
    snackbar.showSuccess("Devices selected successfully.");
    await devicesStore.fetchDeviceList();
    close();
  } catch (error: unknown) {
    snackbar.showError("An error occurred while selecting devices.");
    handleError(error);
  }
};

const handleTabChange = async () => {
  if (tab.value === "suggested") await devicesStore.fetchMostUsedDevices();
  else await devicesStore.fetchDeviceList({ perPage: 5 });
};

const accept = async () => {
  if (tab.value === "suggested") await sendDevicesChoice(devicesStore.suggestedDevices);
  else await sendDevicesChoice(devicesStore.selectedDevices);
};

onMounted(async () => {
  try {
    await devicesStore.fetchMostUsedDevices();
    await devicesStore.fetchDeviceList({ perPage: 5 });
  } catch (error: unknown) {
    const axiosError = error as AxiosError;
    switch (axios.isAxiosError(error)) {
      case axiosError.response?.status === 403: snackbar.showError("You don't have this kind of authorization."); break;
      default: snackbar.showError("An error occurred."); break;
    }
    handleError(error);
  }
});
</script>
