<template>
  <v-dialog
    v-if="hasAuthorization"
    v-model="show"
    max-width="900px"
    min-width="45vw"
    data-test="dialog"
  >
    <v-card class="bg-v-theme-surface" data-test="deviceChooser-dialog">
      <v-card-title class="text-headline bg-primary" data-test="title">
        Update account or select three devices
      </v-card-title>

      <v-card-text>
        <p class="ml-2 text-body-2" data-test="subtext">
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
        <v-tabs align-tabs="center" color="primary" data-test="v-tabs">
          <v-tab
            v-for="item in items"
            :key="item.title"
            :data-test="item.title + '-tab'"
            @click="doAction(item.action)"
          >
            {{ item.title }}
          </v-tab>
        </v-tabs>
      </div>
      <v-card-text class="mb-2 pb-0">
        <DeviceListChooser
          :action="action"
          :isAllDevices="isAllDevices"
          data-test="deviceListChooser-component"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn data-test="close-btn" @click="close()"> Close </v-btn>

        <v-tooltip :disabled="!disableTooltipOrButton" top>
          <template v-slot:activator="{ props }">
            <span v-on="props">
              <v-btn
                v-bind="props"
                :disabled="disableTooltipOrButton"
                data-test="accept-btn"
                @click="accept()"
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
import { ref, computed, onMounted, onUnmounted } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "../../store";
import DeviceListChooser from "./DeviceListChooser.vue";
import hasPermision from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import { INotificationsSuccess, INotificationsError } from "../../interfaces/INotifications";
import handleError from "../../utils/handleError";
import { IDevice } from "../../interfaces/IDevice";

const store = useStore();
const filter = ref("");
const action = ref("chooser");
const items = ref([
  {
    title: "Suggested Devices",
    action: "suggestedDevices",
  },
  {
    title: "All devices",
    action: "allDevices",
  },
]);

const hostname = ref(window.location.hostname);

const isAllDevices = computed(() => action.value === items.value[1].action);

const show = computed({
  get() {
    return store.getters["devices/getDeviceChooserStatus"];
  },

  set(value) {
    store.dispatch("devices/setDeviceChooserStatus", value);
  },
});

const hasDevice = computed(() => (
  store.getters["stats/stats"].registered_devices > 0
        || store.getters["stats/stats"].pending_devices > 0
        || store.getters["stats/stats"].rejected_devices > 0
));

const url = () => `${window.location.protocol}//${hostname.value}/settings/billing`;

const disableTooltipOrButton = computed(() => (
  (store.getters["devices/getDevicesSelected"].length <= 0
          || store.getters["devices/getDevicesSelected"].length > 3)
        && action.value !== items.value[0].action
));

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermision(authorizer.role[role], actions.device.chooser);
  }

  return false;
});

const doAction = async (actionParam: string) => {
  action.value = actionParam;
  if (action.value === items.value[0].action) {
    store.dispatch("devices/getDevicesMostUsed");
  } else if (action.value === items.value[1].action) {
    const data = {
      perPage: 5,
      page: 1,
      filter: store.getters["devices/getFilter"],
      status: "accepted",
      sortStatusField: null,
      sortStatusString: "asc",
    };

    try {
      await store.dispatch("devices/setDevicesForUserToChoose", data);
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError;
        if (axiosError.response?.status === 403) {
          store.dispatch("snackbar/showSnackbarErrorAssociation");
        } else {
          store.dispatch(
            "snackbar/showSnackbarErrorLoading",
            INotificationsError.deviceList,
          );
        }
      }
      handleError(error);
    }
  } else {
    store.dispatch(
      "snackbar/showSnackbarErrorLoading",
      INotificationsError.deviceList,
    );
  }
};

const close = () => {
  store.dispatch("devices/setDeviceChooserStatus", false);
};

const sendDevicesChoice = async (devices: Array<IDevice>) => {
  const choices: Array<string> = [];
  devices.forEach((device) => {
    choices.push(device.uid);
  });

  try {
    await store.dispatch("devices/postDevicesChooser", { choices });
    store.dispatch("snackbar/showSnackbarSuccessAction", INotificationsSuccess.deviceChooser);

    store.dispatch("devices/refresh");

    close();
  } catch (error: unknown) {
    store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.deviceChooser);
    handleError(error);
  }

  store.dispatch("stats/get");
};

const accept = () => {
  if (action.value === items.value[0].action) {
    sendDevicesChoice(store.getters["devices/getDevicesForUserToChoose"]);
  } else {
    sendDevicesChoice(store.getters["devices/getDevicesSelected"]);
  }
};

const searchDevices = () => {
  let encodedFilter = "";

  if (filter.value) {
    const filterToEncodeBase64 = [
      {
        type: "property",
        params: { name: "name", operator: "contains", value: filter.value },
      },
    ];
    encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));
  }

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

onMounted(async () => {
  action.value = items.value[0].action;
  try {
    await store.dispatch("stats/get");
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) store.dispatch("snackbar/showSnackbarErrorAssociation");
    } else {
      store.dispatch("snackbar/showSnackbarErrorDefault");
    }
    handleError(error);
  }
});

onUnmounted(async () => {
  await store.dispatch("devices/setFilter", null);
});
</script>
