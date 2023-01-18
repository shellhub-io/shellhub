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
      :comboboxOptions="[10, 20, 50, 100]"
      @changeItemsPerPage="changeItemsPerPage"
      @clickNextPage="next"
      @clickPreviousPage="prev"
      @clickSortableIcon="sortByItem"
      data-test="devices-dataTable"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in devices" :key="i">
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

<script lang="ts">
import {
  INotificationsCopy,
  INotificationsError,
} from "../../interfaces/INotifications";
import { defineComponent, ref, computed, onMounted, watch } from "vue";
import { useStore } from "../../store";
import DataTable from "../DataTable.vue";
import DeviceIcon from "./DeviceIcon.vue";

export default defineComponent({
  props: {
    action: {
      type: String,
      required: true,
    },
  },
  setup() {
    const store = useStore();
    const hostname = ref(window.location.hostname);
    const singleSelect = ref(true);
    const loading = ref(false);
    const filter = ref("");
    const itemsPerPage = ref(10);
    const page = ref(1);

    const devices = computed(
      () => store.getters["devices/getDevicesForUserToChoose"]
    );
    const numberDevices = computed(
      () => store.getters["devices/getNumberForUserToChoose"]
    );
    const selected = computed({
      get() {
        return store.getters["devices/getDevicesSelected"];
      },
      set(value) {
        store.commit("devices/setDevicesSelected", value);
      },
    });
    const disableShowSelect = computed(() => !(numberDevices.value === 3));

    onMounted(() => {
      store.dispatch("devices/getDevicesMostUsed");
    });

    const getDevices = async (perPagaeValue: number, pageValue: number) => {
      try {
        loading.value = true;

        const hasDevices = await store.dispatch(
          "devices/setDevicesForUserToChoosesetDevicesForUserToChoose",
          {
            perPage: perPagaeValue,
            page: pageValue,
            filter: filter.value,
            sortStatusField: store.getters["devices/sortStatusField"],
            sortStatusString: store.getters["devices/sortStatusString"],
          }
        );

        if (!hasDevices) {
          page.value--;
        }

        loading.value = false;
      } catch (error: any) {
        if (error.response.status === 403) {
          store.dispatch("snackbar/showSnackbarErrorAssociation");
          throw new Error(error);
        } else {
          store.dispatch(
            "snackbar/showSnackbarErrorLoading",
            INotificationsError.deviceList
          );
          throw new Error(error);
        }
      }
    };

    const sortByItem = async (field: string) => {
      let sortStatusString = store.getters["devices/sortStatusString"];
      if (sortStatusString === "") {
        sortStatusString = "asc";
      } else if (sortStatusString === "asc") {
        sortStatusString = "desc";
      } else {
        sortStatusString = "";
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

    const address = (item: any) =>
      `${item.namespace}.${item.name}@${window.location.hostname}`;

    const copyText = (value: string | undefined) => {
      if (value) {
        navigator.clipboard.writeText(value);
        store.dispatch(
          "snackbar/showSnackbarCopy",
          INotificationsCopy.deviceSSHID
        );
      }
    };

    const refreshUsers = () => {
      getDevices(itemsPerPage.value, page.value);
    };

    return {
      headers: [
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
      ],
      devices,
      numberDevices,
      selected,
      disableShowSelect,
      loading,
      itemsPerPage,
      page,
      hostname,
      singleSelect,
      next,
      prev,
      changeItemsPerPage,
      sortByItem,
      address,
      copyText,
    };
  },
  components: { DataTable, DeviceIcon },
});
</script>
