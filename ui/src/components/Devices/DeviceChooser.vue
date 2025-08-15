<template>
  <BaseDialog
    v-if="hasAuthorization"
    v-model="showDialog"
    data-test="device-chooser-dialog"
  >
    <v-card
      class="bg-v-theme-surface"
      data-test="device-chooser-card"
    >
      <v-card-title
        class="text-headline bg-primary"
        data-test="title"
      >
        Update account or select three devices
      </v-card-title>

      <v-card-text>
        <p
          class="ml-2 text-body-2"
          data-test="subtext"
        >
          You currently have no subscription to the
          <router-link :to="billingUrl">premium plan</router-link> and the free version is limited to
          3 devices. To unlock access to all devices, you can subscribe to the
          <router-link :to="billingUrl">premium plan</router-link>. If you want to continue on
          the free plan, you need to select three devices.
        </p>
      </v-card-text>
      <div v-if="isAllDevicesTab && hasDevices" class="pa-5">
        <v-row>
          <v-col md="12" sm="12">
            <v-text-field
              label="Search by hostname"
              variant="underlined"
              color="primary"
              single-line
              hide-details
              v-model.trim="filter"
              v-on:keyup="searchDevices"
              append-inner-icon="mdi-magnify"
              density="comfortable"
              data-test="search-text"
            />
          </v-col>
        </v-row>
      </div>
      <div class="mt-2">
        <v-tabs
          v-model="tab"
          align-tabs="center"
          color="primary"
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
      </div>

      <v-card-text class="mb-2 pb-0">
        <v-window v-model="tab" @update:model-value="handleTabChange">
          <v-window-item
            v-for="(item, id) in tabItems"
            :key="id"
            :value="id"
          >
            <DeviceListChooser
              :isSelectable="item.selectable"
              data-test="device-list-chooser-component"
            />
          </v-window-item>
        </v-window>
      </v-card-text>
      <v-card-actions>
        <v-spacer />

        <v-btn
          data-test="close-btn"
          @click="close()"
        > Close </v-btn>
        <v-tooltip :disabled="!disableButton" top>
          <template v-slot:activator="{ props }">
            <span>
              <v-btn
                v-bind="props"
                :disabled="disableButton"
                @click="accept()"
                data-test="accept-btn"
              >
                Accept
              </v-btn>
            </span>
          </template>

          <span> You can select 3 devices or less. </span>
        </v-tooltip>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import axios, { AxiosError } from "axios";
import { actions, authorizer } from "@/authorizer";
import DeviceListChooser from "./DeviceListChooser.vue";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useDevicesStore from "@/store/modules/devices";
import { IDevice } from "@/interfaces/IDevice";

const authStore = useAuthStore();
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

const hasAuthorization = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.device.chooser);
});

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
