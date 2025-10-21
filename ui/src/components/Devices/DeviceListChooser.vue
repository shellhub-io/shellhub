<template>
  <v-card class="bg-v-theme-surface" data-test="devices-list-chooser">
    <DataTable
      v-model:page="page"
      v-model:itemsPerPage="itemsPerPage"
      :headers
      :items="devices"
      :totalCount="devices.length"
      :loading
      data-test="devices-dataTable"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in devices" :key="i">
          <td class="pa-0 text-center">
            <v-checkbox
              v-if="props.isSelectable"
              v-model="selectedDevices"
              class="mt-5 ml-5"
              density="compact"
              :value="item"
            />
          </td>
          <td class="text-center">
            <router-link
              :to="{ name: 'DeviceDetails', params: { identifier: item.uid } }"
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
              <span
                class="hover-text"
              >
                {{ getSshid(item) }}
              </span>
            </v-chip>
          </td>
        </tr>
      </template>
    </DataTable>
  </v-card>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import axios, { AxiosError } from "axios";
import DataTable from "../Tables/DataTable.vue";
import DeviceIcon from "./DeviceIcon.vue";
import handleError from "@/utils/handleError";
import { IDevice } from "@/interfaces/IDevice";
import useSnackbar from "@/helpers/snackbar";
import useDevicesStore from "@/store/modules/devices";

const props = defineProps(["isSelectable"]);
const snackbar = useSnackbar();
const devicesStore = useDevicesStore();

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
const itemsPerPage = ref(5);
const page = ref(1);
const devices = computed(() => devicesStore.devices);
const selectedDevices = computed({
  get() {
    return devicesStore.selectedDevices;
  },
  set(value) {
    devicesStore.selectedDevices = value;
  },
});

const getDevices = async () => {
  try {
    loading.value = true;

    await devicesStore.fetchDeviceList({
      perPage: itemsPerPage.value,
      page: page.value,
    });
  } catch (error: unknown) {
    const axiosError = error as AxiosError;
    switch (axios.isAxiosError(error)) {
      case axiosError.response?.status === 403: {
        snackbar.showError("You do not have permission to access this resource.");
        break;
      }
      default: {
        snackbar.showError("An error occurred while fetching devices.");
        break;
      }
    }
    handleError(error);
  }
  loading.value = false;
};

watch([page, itemsPerPage], async () => {
  await getDevices();
});

watch(selectedDevices, (newValue, oldValue) => {
  if (newValue.length > 3) {
    selectedDevices.value = oldValue;
  }
});

const getSshid = (item: IDevice) => `${item.namespace}.${item.name}@${window.location.hostname}`;
</script>
