<template>
  <div>
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
      data-test="devices-list"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in devices" :key="i">
          <td class="text-center">{{ item.name }}</td>
          <td class="d-flex align-center justify-center">
            <DeviceIcon :icon="item.info.id" class="mr-2" />
            <span>{{ item.info.pretty_name }}</span>
          </td>
          <td class="text-center">
            <v-chip>
              <v-tooltip location="bottom">
                <template v-slot:activator="{ props }">
                  <span
                    v-bind="props"
                    @click="copyText(sshidAddress(item))"
                    @keypress="copyText(sshidAddress(item))"
                    class="hover-text"
                  >
                    {{ sshidAddress(item) }}
                  </span>
                </template>
                <span>Copy ID</span>
              </v-tooltip>
            </v-chip>
          </td>

          <td class="text-center">
            <v-menu location="bottom" scrim eager>
              <template v-slot:activator="{ props }">
                <v-chip density="comfortable" size="small">
                  <v-icon v-bind="props">mdi-dots-horizontal</v-icon>
                </v-chip>
              </template>
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <DeviceActionButton
                  :uid="item.uid"
                  action="accept"
                  :show.sync="showDeviceAcceptButton"
                  data-test="DeviceActionButtonAccept-component"
                  @update="refreshDevices"
                />

                <DeviceActionButton
                  :uid="item.uid"
                  action="reject"
                  :show="showDeviceRejectButton"
                  data-test="deviceActionButtonReject-component"
                  @update="refreshDevices"
                />
              </v-list>
            </v-menu>
          </td>
        </tr>
      </template>
    </DataTable>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, watch, computed } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "../../store";
import DataTable from "../DataTable.vue";
import DeviceIcon from "./DeviceIcon.vue";
import { formatDate } from "../../utils/formateDate";
import { displayOnlyTenCharacters } from "../../utils/string";
import showTag from "../../utils/tag";
import {
  INotificationsCopy,
  INotificationsError,
} from "../../interfaces/INotifications";
import DeviceActionButton from "./DeviceActionButton.vue";

export default defineComponent({
  setup() {
    const store = useStore();
    const router = useRouter();
    const loading = ref(false);
    const filter = ref("");
    const itemsPerPage = ref(10);
    const page = ref(1);
    const showDeviceAcceptButton = ref(false);
    const showDeviceRejectButton = ref(false);

    const devices = computed(() => store.getters["devices/list"]);
    const numberDevices = computed<number>(
      () => store.getters["devices/getNumberDevices"]
    );

    onMounted(async () => {
      try {
        loading.value = true;
        await store.dispatch("devices/fetch", {
          page: page.value,
          perPage: itemsPerPage.value,
          filter: "",
          status: "pending",
          sortStatusField: "",
          sortStatusString: "",
        });
      } catch {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.devicePending
        );
      } finally {
        loading.value = false;
      }
    });

    const getDevices = async (perPagaeValue: number, pageValue: number) => {
      try {
        loading.value = true;

        const hasDevices = await store.dispatch("devices/fetch", {
          page: pageValue,
          perPage: perPagaeValue,
          filter: filter.value,
          status: "pending",
          sortStatusField: store.getters["devices/sortStatusField"],
          sortStatusString: store.getters["devices/sortStatusString"],
        });

        if (!hasDevices) {
          page.value--;
        }

        loading.value = false;
      } catch (error) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.devicePending
        );
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

    const goToNamespace = (namespace: string) => {
      router.push({ name: "namespaceDetails", params: { id: namespace } });
    };

    const redirectToDevice = (deviceId: string) => {
      router.push({ name: "deviceDetails", params: { id: deviceId } });
    };

    const sshidAddress = (item: any) =>
      `${item.namespace}.${item.name}@${window.location.hostname}`;

    const copyText = (value: string | undefined) => {
      if (value) {
        navigator.clipboard.writeText(value);
        store.dispatch(
          "snackbar/showSnackbarCopy",
          INotificationsCopy.tenantId
        );
      }
    };

    const changeDeviceAcceptButton = () => {
      showDeviceAcceptButton.value = !showDeviceAcceptButton.value;
    };

    const changeDeviceRejectButton = () => {
      showDeviceRejectButton.value = !showDeviceRejectButton.value;
    };

    const refreshDevices = () => {
      getDevices(itemsPerPage.value, page.value);
    };

    return {
      headers: [
        {
          text: "Hostname",
          value: "name",
          sortable: true,
        },
        {
          text: "Operating System",
          value: "operating_system",
        },
        {
          text: "Request Time",
          value: "request_time",
        },
        {
          text: "Actions",
          value: "actions",
        },
      ],
      itemsPerPage,
      page,
      loading,
      devices,
      numberDevices,
      next,
      prev,
      sortByItem,
      showTag,
      displayOnlyTenCharacters,
      formatDate,
      goToNamespace,
      changeItemsPerPage,
      redirectToDevice,
      sshidAddress,
      copyText,
      refreshDevices,
      showDeviceAcceptButton,
      showDeviceRejectButton,
      changeDeviceAcceptButton,
      changeDeviceRejectButton,
    };
  },
  components: { DataTable, DeviceIcon, DeviceActionButton },
});
</script>

<style scoped>
.hover-text {
  cursor: pointer;
  animation: fadeIn 0.5s;
}

.hover-text:hover,
.hover-text:focus {
  text-decoration: underline;
}
</style>
