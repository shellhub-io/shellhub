<template>
  <DataTable
    :headers="computedHeaders"
    :items="items"
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
    data-test="items-list"
  >
    <template v-if="status === 'accepted'" v-slot:rows>
      <tr v-for="(item, i) in items" :key="i">
        <td class="text-center">
          <TerminalDialog
            :enable-connect-button="true"
            :uid="item.uid"
            :online="item.online"
            data-test="terminalDialog-component"
          />
        </td>
        <td class="text-center">{{ item.name }}</td>
        <td class="text-center">
          <DeviceIcon :icon="item.info.id" class="mr-2" data-test="deviceIcon-component" />
          <span>{{ item.info.pretty_name }}</span>
        </td>
        <td class="text-center">
          <v-chip data-test="sshid-chip">
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
                <v-chip size="small" v-bind="props" v-on="props" class="mr-1" data-test="tag-chip">
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
          <v-menu location="bottom" scrim eager data-test="v-menu">
            <template v-slot:activator="{ props }">
              <v-chip v-bind="props" density="comfortable" size="small">
                <v-icon>mdi-dots-horizontal</v-icon>
              </v-chip>
            </template>
            <v-list class="bg-v-theme-surface" lines="two" density="compact">
              <v-list-item @click="redirectToDevice(item.uid)" data-test="mdi-information-list-item">
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
                      :variant="props.variant === 'device' ? 'device' : 'container' "
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
    <template v-else v-slot:rows>
      <tr v-for="(item, i) in items" :key="i">
        <td class="text-center">
          <router-link
            :to="{ name: 'detailsDevice', params: { id: item.uid } }"
            :data-test="item.uid + '-field'"
          >
            {{ item.name }}
          </router-link>
        </td>
        <td class="text-center">
          <DeviceIcon :icon="item.info.id" class="mr-2" data-test="device-icon" />
          <span>{{ item.info.pretty_name }}</span>
        </td>
        <td class="text-center">
          {{ formatDate(item.last_seen) }}
        </td>

        <td class="text-center">
          <v-menu location="bottom" scrim eager>
            <template v-slot:activator="{ props }">
              <v-chip density="comfortable" size="small" data-test="sshid-chip">
                <v-icon v-bind="props">mdi-dots-horizontal</v-icon>
              </v-chip>
            </template>
            <v-list class="bg-v-theme-surface" lines="two" density="compact">
              <DeviceActionButton
                :uid="item.uid"
                :name="item.name"
                :variant="props.variant === 'device' ? 'device' : 'container' "
                :notificationStatus="false"
                action="accept"
                :show="showDeviceAcceptButton"
                @update="refreshDevices()"
                data-test="DeviceActionButtonAccept-component"
              />
              <DeviceActionButton
                :uid="item.uid"
                :variant="props.variant === 'device' ? 'device' : 'container' "
                :action="status === 'pending' ? 'reject' : 'remove'"
                :notificationStatus="false"
                :show="showDeviceRejectButton"
                @update="refreshDevices()"
                data-test="deviceActionButtonReject-component"
              />
            </v-list>
          </v-menu>
        </td>
      </tr>
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed, PropType } from "vue";
import { useRouter } from "vue-router";
import { store } from "../../store";
import { actions, authorizer } from "@/authorizer";
import DataTable from "../DataTable.vue";
import DeviceIcon from "../Devices/DeviceIcon.vue";
import DeviceActionButton from "../Devices/DeviceActionButton.vue";
import DeviceDelete from "../Devices/DeviceDelete.vue";
import TagFormUpdate from "../Tags/TagFormUpdate.vue";
import TerminalDialog from "../Terminal/TerminalDialog.vue";
import { INotificationsCopy } from "../../interfaces/INotifications";
import { IDevice, IDeviceMethods } from "../../interfaces/IDevice";
import hasPermission from "@/utils/permission";
import showTag from "../../utils/tag";
import { displayOnlyTenCharacters } from "../../utils/string";
import handleError from "../../utils/handleError";
import { formatDate } from "../../utils/formateDate";

const props = defineProps({
  storeMethods: {
    type: Object as PropType<IDeviceMethods>,
    required: true,
  },
  status: {
    type: String as PropType<"accepted" | "pending" | "rejected">,
    required: true,
  },
  header: {
    type: String as PropType<"primary" | "secondary">,
    required: true,
  },
  variant: {
    type: String as PropType<"device" | "container">,
    required: true,
  },
});

const { fetchDevices, getFilter, getDevicesList, getSortStatusField, getSortStatusString, getNumberDevices } = props.storeMethods;

const router = useRouter();
const loading = ref(false);
const filter = computed(() => getFilter());
const items = computed(() => getDevicesList());
const numberDevices = computed(() => getNumberDevices());
const showDeviceAcceptButton = ref(false);
const showDeviceRejectButton = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const status = computed(() => props.status);

const headers = [
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
];

const headersSecondary = [
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
];

const computedHeaders = computed(() => props.header === "primary" ? headers : headersSecondary);

onMounted(async () => {
  try {
    loading.value = true;
    await fetchDevices({
      perPage: itemsPerPage.value,
      page: page.value,
      filter: filter.value,
      status: status.value,
      sortStatusField: "",
      sortStatusString: "",
    });
  } catch (error: unknown) {
    handleError(error);
  } finally {
    loading.value = false;
  }
});

const getDevices = async (perPagaeValue: number, pageValue: number, filter: string) => {
  try {
    loading.value = true;
    await fetchDevices({
      perPage: perPagaeValue,
      page: pageValue,
      status: props.status,
      filter,
      sortStatusField: getSortStatusField(),
      sortStatusString: getSortStatusString(),
    });
    loading.value = false;
  } catch (error: unknown) {
    handleError(error);
  }
};

const sortByItem = async (field: string) => {
  let sortStatusString = getSortStatusString();
  const sortStatusField = getSortStatusField();

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
  await fetchDevices({ sortStatusField: field, sortStatusString });
  await getDevices(itemsPerPage.value, page.value, filter.value);
};

const next = async () => {
  await getDevices(itemsPerPage.value, ++page.value, filter.value);
};

const prev = async () => {
  try {
    if (page.value > 1) await getDevices(itemsPerPage.value, --page.value, filter.value);
  } catch (error: unknown) {
    handleError(error);
  }
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, async () => {
  await getDevices(itemsPerPage.value, page.value, filter.value);
});

const redirectToDevice = (deviceId: string) => {
  router.push({ name: "detailsDevice", params: { id: deviceId } });
};

const sshidAddress = (item: IDevice) => `${item.namespace}.${item.name}@${window.location.hostname}`;

const copyText = (value: string | undefined) => {
  if (value) {
    navigator.clipboard.writeText(value);
    store.dispatch("snackbar/showSnackbarCopy", INotificationsCopy.deviceSSHID);
  }
};

const refreshDevices = () => {
  getDevices(itemsPerPage.value, page.value, filter.value);
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

defineExpose({ page, getSortStatusField, getSortStatusString });
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
