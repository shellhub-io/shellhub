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
          <td class="d-flex justify-center">
            <TerminalDialog
              :enable-connect-button="true"
              :uid="item.uid"
              :online="item.online"
              data-test="terminalDialog-component"
            />
          </td>
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
            <div v-if="item.tags[0]">
              <v-tooltip
                v-for="(tag, index) in item.tags"
                :key="index"
                location="bottom"
                :disabled="!showTag(tag)"
              >
                <template #activator="{ props }">
                  <v-chip size="small" v-bind="props" v-on="props" class="mr-1">
                    {{ displayOnlyTenCharacters(tag) }}
                  </v-chip>
                </template>

                <span v-if="showTag(tag)">
                  {{ tag }}
                </span>
              </v-tooltip>
            </div>

            <div v-else>
              <v-chip size="small"> No tags </v-chip>
            </div>
          </td>

          <td class="text-center">
            <v-menu location="bottom" scrim eager>
              <template v-slot:activator="{ props }">
                <v-chip v-bind="props" density="comfortable" size="small">
                  <v-icon>mdi-dots-horizontal</v-icon>
                </v-chip>
              </template>
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <v-list-item @click="redirectToDevice(item.uid)">
                  <div class="d-flex align-center">
                    <div class="mr-2">
                      <v-icon> mdi-information </v-icon>
                    </div>

                    <v-list-item-title data-test="mdi-information-list-item">
                      Details
                    </v-list-item-title>
                  </div>
                </v-list-item>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationFormUpdate()"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <TagFormUpdate
                        :device-uid="item.uid"
                        :tagsList="item.tags"
                        :notHasAuthorization="!hasAuthorizationFormUpdate()"
                        @update="refreshDevices"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationRemove()"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <DeviceDelete
                        :uid="item.uid"
                        :notHasAuthorization="!hasAuthorizationRemove()"
                        @update="refreshDevices"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>
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
import { formatDate } from "../../utils/formateDate";
import { displayOnlyTenCharacters } from "../../utils/string";
import showTag from "../../utils/tag";
import DataTable from "../DataTable.vue";
import DeviceIcon from "./DeviceIcon.vue";
import DeviceDelete from "./DeviceDelete.vue";
import TagFormUpdate from "../Tags/TagFormUpdate.vue";
import {
  INotificationsCopy,
  INotificationsError,
} from "../../interfaces/INotifications";
import TerminalDialog from "../Terminal/TerminalDialog.vue";
import { actions, authorizer } from "../../authorizer";
import hasPermission from "@/utils/permission";

export default defineComponent({
  setup() {
    const store = useStore();
    const router = useRouter();
    const loading = ref(false);
    const filter = ref("");
    const itemsPerPage = ref(10);
    const page = ref(1);

    const deviceDeleteShow = ref([]);

    const devices = computed(() => store.getters["devices/list"]);
    const numberDevices = computed<number>(
      () => store.getters["devices/getNumberDevices"],
    );

    onMounted(async () => {
      try {
        loading.value = true;
        await store.dispatch("devices/fetch", {
          perPage: itemsPerPage.value,
          page: page.value,
          filter: "",
          status: "accepted",
          sortStatusField: "",
          sortStatusString: "",
        });
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deviceList,
        );
        throw new Error(error);
      } finally {
        loading.value = false;
      }
    });

    const getDevices = async (perPagaeValue: number, pageValue: number) => {
      try {
        loading.value = true;

        const hasDevices = await store.dispatch("devices/fetch", {
          perPage: perPagaeValue,
          page: pageValue,
          status: "accepted",
          filter: filter.value,
          sortStatusField: store.getters["devices/sortStatusField"],
          sortStatusString: store.getters["devices/sortStatusString"],
        });

        if (!hasDevices) {
          page.value--;
        }

        loading.value = false;
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deviceList,
        );
        throw new Error(error);
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
      } catch (error: any) {
        store.dispatch("snackbar/setSnackbarErrorDefault");
        throw new Error(error);
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
      router.push({ name: "detailsDevice", params: { id: deviceId } });
    };

    const sshidAddress = (item: any) => `${item.namespace}.${item.name}@${window.location.hostname}`;

    const copyText = (value: string | undefined) => {
      if (value) {
        navigator.clipboard.writeText(value);
        store.dispatch(
          "snackbar/showSnackbarCopy",
          INotificationsCopy.deviceSSHID,
        );
      }
    };

    const refreshDevices = () => {
      getDevices(itemsPerPage.value, page.value);
    };

    const hasAuthorizationFormUpdate = () => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(authorizer.role[role], actions.tag.deviceUpdate);
      }

      return false;
    };

    const hasAuthorizationRemove = () => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(authorizer.role[role], actions.device.remove);
      }

      return false;
    };

    return {
      headers: [
        {
          text: "Online",
          value: "online",
          sortable: true,
        },
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
          text: "SSHID",
          value: "sshid",
        },
        {
          text: "Tags",
          value: "tags",
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
      deviceDeleteShow,
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
      hasAuthorizationFormUpdate,
      hasAuthorizationRemove,
    };
  },
  components: {
    DataTable,
    DeviceIcon,
    DeviceDelete,
    TagFormUpdate,
    TerminalDialog,
  },
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
