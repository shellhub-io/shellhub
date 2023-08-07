<template>
  <v-card class="bg-v-theme-surface" data-test="devices-list-chooser">
    <DataTable
      :headers="headers"
      :items="devices"
      :itemsPerPage="itemsPerPage"
      :nextPage="next"
      :previousPage="prev"
      :loading="loading"
      :totalCount="numberDevices"
      :actualPage="page"
      :enable-items-per-page="false"
      :comboboxOptions="[5]"
      @changeItemsPerPage="changeItemsPerPage"
      @clickNextPage="next"
      @clickPreviousPage="prev"
      @clickSortableIcon="sortByItem"
      data-test="devices-dataTable"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in devices" :key="i">
          <td class="pa-0 text-center">
            <v-checkbox
              v-if="props.isAllDevices"
              v-model="selected"
              class="mt-5 ml-5"
              density="compact"
              :value="item"
            />
          </td>
          <td class="text-center">
            <router-link
              :to="{ name: 'detailsDevice', params: { id: item.uid } }"
            >
              {{ item.name }}
            </router-link>
          </td>

          <td class="text-center" v-if="item.info">
            <DeviceIcon
              :icon="item.info.id"
              data-test="deviceIcon-component"
            />
            {{ item.info.pretty_name }}
          </td>

          <td class="text-center">
            <v-chip>
              <v-tooltip location="bottom">
                <template v-slot:activator="{ props }">
                  <span
                    v-bind="props"
                    @click="copyText(address(item))"
                    @keypress="copyText(address(item))"
                    class="hover-text"
                  >
                    {{ address(item) }}
                  </span>
                </template>
                <span>Copy ID</span>
              </v-tooltip>
            </v-chip>
          </td>
        </tr>
      </template>
    </DataTable>
  </v-card>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import axios, { AxiosError } from "axios";
import {
  INotificationsCopy,
  INotificationsError,
} from "../../interfaces/INotifications";
import { useStore } from "../../store";
import DataTable from "../DataTable.vue";
import DeviceIcon from "./DeviceIcon.vue";
import handleError from "@/utils/handleError";
import { IDevice } from "@/interfaces/IDevice";

const props = defineProps(["action", "isAllDevices"]);

const store = useStore();

const headers = [
  {
    text: "",
    value: "selected",
  },
  {
    text: "Hostname",
    value: "hostname",
  },
  {
    text: "Operating System",
    value: "info.pretty_name",
  },
  {
    text: "SSHID",
    value: "namespace",
  },
];

const loading = ref(false);

const filter = ref("");

const itemsPerPage = ref(5);

const page = ref(1);

const devices = computed(
  () => store.getters["devices/getDevicesForUserToChoose"],
);

const numberDevices = computed(
  () => store.getters["devices/getNumberForUserToChoose"],
);

const selected = computed({
  get() {
    return store.getters["devices/getDevicesSelected"];
  },
  set(value) {
    store.commit("devices/setDevicesSelected", value);
  },
});

onMounted(() => {
  store.dispatch("devices/getDevicesMostUsed");
});

const getDevices = async (perPagaeValue: number, pageValue: number) => {
  try {
    loading.value = true;

    const hasDevices = await store.dispatch(
      "devices/setDevicesForUserToChoose",
      {
        perPage: perPagaeValue,
        page: pageValue,
        filter: filter.value,
        sortStatusField: store.getters["devices/getSortStatusField"],
        sortStatusString: store.getters["devices/getSortStatusString"],
      },
    );

    if (!hasDevices) {
      page.value--;
    }

    loading.value = false;
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        store.dispatch("snackbar/showSnackbarErrorAssociation");
        handleError(error);
      }
    } else {
      store.dispatch(
        "snackbar/showSnackbarErrorLoading",
        INotificationsError.deviceList,
      );
      handleError(error);
    }
  }
};

const sortByItem = async (field: string) => {
  let sortStatusString = store.getters["devices/getSortStatusString"];
  const sortStatusField = store.getters["devices/getSortStatusField"];

  if (field !== sortStatusField && sortStatusField) {
    if (sortStatusString === "asc") {
      sortStatusString = "desc";
    } else {
      sortStatusString = "asc";
    }
  }

  if (sortStatusString === "") {
    sortStatusString = "asc";
  } else if (sortStatusString === "asc") {
    sortStatusString = "desc";
  } else {
    sortStatusString = "asc";
  }
  await store.dispatch("devices/setSortStatus", {
    sortStatusField: field,
    sortStatusString,
  });
  await getDevices(itemsPerPage.value, page.value);
};

const next = async () => {
  await getDevices(itemsPerPage.value, ++page.value);
};

const prev = async () => {
  try {
    if (page.value > 1) await getDevices(itemsPerPage.value, --page.value);
  } catch (error) {
    store.dispatch("snackbar/setSnackbarErrorDefault");
  }
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, async () => {
  await getDevices(itemsPerPage.value, page.value);
});

watch(selected, (newValue, oldValue) => {
  if (newValue.length > 3) {
    selected.value = oldValue;
  }
});

const address = (item: IDevice) => `${item.namespace}.${item.name}@${window.location.hostname}`;

const copyText = (value: string | undefined) => {
  if (value) {
    navigator.clipboard.writeText(value);
    store.dispatch(
      "snackbar/showSnackbarCopy",
      INotificationsCopy.deviceSSHID,
    );
  }
};
</script>
