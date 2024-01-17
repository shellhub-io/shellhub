<template>
  <v-dialog
    v-if="hasAuthorization"
    v-model="show"
    max-width="900px"
    min-width="45vw"
    data-test="dialog"
  >
    <v-card
      class="bg-v-theme-surface"
      data-test="deviceChooser-dialog"
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
          <a :href="url()"> premium plan </a> and the free version is limited to
          3 devices. To unlock access to all devices, you can subscribe to the
          <a :href="url()"> premium plan </a>. Case, If you want to continue on
          the free plan, you need to select three devices.
        </p>
      </v-card-text>
      <div v-if="isAllDevices && hasDevice" class="pa-5">
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
        <v-window v-model="tab">
          <v-window-item
            v-for="(item, id) in tabItems"
            :key="id"
            :value="id"
          >
            <DeviceListChooser
              :isSelectable="item.selectable"
              data-test="deviceListChooser-component"
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
            <span v-on="props">
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
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import { actions, authorizer } from "../../authorizer";
import DeviceListChooser from "./DeviceListChooser.vue";
import hasPermision from "../../utils/permission";
import handleError from "../../utils/handleError";
import { INotificationsSuccess, INotificationsError } from "../../interfaces/INotifications";

const store = useStore();

const show = computed({
  get() {
    return store.getters["devices/getDeviceChooserStatus"];
  },

  set(value) {
    store.dispatch("devices/setDeviceChooserStatus", value);
  },
});

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];

  if (role !== "") {
    return hasPermision(authorizer.role[role], actions.device.chooser);
  }

  return false;
});

const url = () => `${window.location.protocol}//${window.location.hostname}/settings/billing`;

const close = () => {
  store.dispatch("devices/setDeviceChooserStatus", false);
};

const filter = ref("");

const searchDevices = () => {
  let encodedFilter = "";

  if (!filter.value) {
    return;
  }

  const filterToEncodeBase64 = [
    {
      type: "property",
      params: { name: "name", operator: "contains", value: filter.value },
    },
  ];

  encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));

  try {
    store.dispatch("devices/search", {
      page: store.getters["devices/getPage"],
      perPage: store.getters["devices/getPerPage"],
      filter: encodedFilter,
      status: store.getters["devices/getStatus"],
    });
  } catch {
    store.dispatch("snackbar/showSnackbarErrorDefault");
  }
};

const sendDevicesChoice = async (devices: Array<string>) => {
  try {
    await store.dispatch("devices/postDevicesChooser", { devices });
    store.dispatch("snackbar/showSnackbarSuccessAction", INotificationsSuccess.deviceChooser);
    store.dispatch("devices/refresh");
    close();
  } catch (error: unknown) {
    store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.deviceChooser);
    handleError(error);
  }

  store.dispatch("stats/get");
};

const disableTab = ref(false);

// Current selected tab
const tab = ref("");

// This variable is used to define the items for the `<v-tabs>`.
const tabItems = ref({
  // TODO: 'automatically' seems a better name to use
  suggested: {
    title: "Suggested",
    selectable: false,
    disabled: disableTab,
  },

  // TODO: 'manually' seems a better name to use
  all: {
    title: "All",
    selectable: true,
    disabled: false,
  },
});

const isAllDevices = computed(() => tab.value === "all");

const hasDevice = computed(() => (
  store.getters["stats/stats"].registered_devices > 0
        || store.getters["stats/stats"].pending_devices > 0
        || store.getters["stats/stats"].rejected_devices > 0
));

// Watch for changes in the current tab to load the relevant data into the data table
watch(tab, async (tabId) => {
  switch (tabItems.value[tabId]) {
    case tabItems.value.suggested: {
      await store.dispatch("devices/getDevicesMostUsed");
      disableTab.value = store.getters["devices/getDevicesForUserToChoose"].length <= 0;
      // Set tab.value to "all" if disabled.value is true
      if (disableTab.value) {
        tab.value = "all";
      }
      break;
    }
    case tabItems.value.all: {
      await store.dispatch("devices/setDevicesForUserToChoose", {
        perPage: 5,
        page: 1,
        filter: store.getters["devices/getFilter"],
        status: "accepted",
        sortStatusField: null,
        sortStatusString: "asc",
      });
      break;
    }
    default:
      break;
  }
});

const disableButton = computed(() => (
  (store.getters["devices/getDevicesSelected"].length <= 0
    || store.getters["devices/getDevicesSelected"].length > 3)
    && tab.value !== "suggested"
));

const accept = async () => {
  switch (tabItems.value[tab.value]) {
    case tabItems.value.suggested: {
      sendDevicesChoice(store.getters["devices/getDevicesForUserToChoose"]);
      break;
    }
    case tabItems.value.all: {
      sendDevicesChoice(store.getters["devices/getDevicesSelected"]);
      break;
    }
    default:
      break;
  }
};

onMounted(async () => {
  try {
    await store.dispatch("stats/get");
  } catch (error: unknown) {
    const axiosError = error as AxiosError;
    switch (axios.isAxiosError(error)) {
      case axiosError.response?.status === 403: store.dispatch("snackbar/showSnackbarErrorAssociation"); break;
      default: store.dispatch("snackbar/showSnackbarErrorDefault"); break;
    }
    handleError(error);
  }
});

onUnmounted(async () => {
  await store.dispatch("devices/setFilter", null);
});
</script>
