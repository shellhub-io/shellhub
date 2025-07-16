<template>
  <div>
    <DataTable
      v-model:page="page"
      v-model:itemsPerPage="itemsPerPage"
      :headers="computedHeaders"
      :items
      :totalCount="numberDevices"
      :loading
      :itemsPerPageOptions="[10, 20, 50, 100]"
      @update:sort="sortByItem"
      data-test="items-list"
    >
      <template
        v-if="status === 'accepted'"
        v-slot:rows
      >
        <tr
          v-for="(item, i) in items"
          :key="i"
        >
          <td class="text-center">
            <TerminalConnectButton
              :deviceUid="item.uid"
              :sshid="sshidAddress(item)"
              :online="item.online"
              data-test="terminal-connect-btn"
            />
          </td>
          <td class="text-center">{{ item.name }}</td>
          <td class="text-center">
            <DeviceIcon
              :icon="item.info.id"
              class="mr-2"
              data-test="deviceIcon-component"
            />
            <span>{{ item.info.pretty_name }}</span>
          </td>
          <td class="text-center">
            <CopyWarning
              :copied-item="'Device SSHID'"
              :bypass="shouldOpenTerminalHelper()"
            >
              <template #default="{ copyText }">
                <v-chip data-test="sshid-chip">
                  <v-tooltip location="bottom">
                    <template v-slot:activator="{ props }">
                      <span
                        v-bind="props"
                        @click="handleSshidClick(item, copyText)"
                        @keypress.enter="handleSshidClick(item, copyText)"
                        class="hover-text"
                      >
                        {{ sshidAddress(item) }}
                      </span>
                    </template>
                    <span>{{ shouldOpenTerminalHelper() ? "Show connection instructions" : "Copy ID" }}</span>
                  </v-tooltip>
                </v-chip>
              </template>
            </CopyWarning>
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
                  <v-chip
                    size="small"
                    v-bind="props"
                    class="mr-1"
                    data-test="tag-chip"
                  >
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
            <v-menu
              location="bottom"
              scrim
              eager
              data-test="v-menu"
            >
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  variant="plain"
                  class="border rounded bg-v-theme-background"
                  density="comfortable"
                  size="default"
                  icon="mdi-format-list-bulleted"
                />
              </template>
              <v-list
                class="bg-v-theme-surface"
                lines="two"
                density="compact"
              >
                <v-list-item
                  @click="redirectToDevice(item.uid)"
                  data-test="mdi-information-list-item"
                >
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
                        :variant
                        :uid="item.uid"
                        :hasAuthorization="hasAuthorizationRemove()"
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
      <template
        v-else
        v-slot:rows
      >
        <tr
          v-for="(item, i) in items"
          :key="i"
        >
          <td class="text-center">
            <router-link
              :to="{ name: 'DeviceDetails', params: { identifier: item.uid } }"
              :data-test="item.uid + '-field'"
            >
              {{ item.name }}
            </router-link>
          </td>
          <td class="text-center">
            <DeviceIcon
              :icon="item.info.id"
              class="mr-2"
              data-test="device-icon"
            />
            <span>{{ item.info.pretty_name }}</span>
          </td>
          <td class="text-center">
            {{ formatFullDateTime(item.last_seen) }}
          </td>

          <td class="text-center">
            <v-menu
              location="bottom"
              scrim
              eager
            >
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  variant="plain"
                  class="border rounded bg-v-theme-background"
                  density="comfortable"
                  size="default"
                  icon="mdi-format-list-bulleted"
                  data-test="device-table-actions"
                />
              </template>
              <v-list
                class="bg-v-theme-surface"
                lines="two"
                density="compact"
              >
                <DeviceActionButton
                  :uid="item.uid"
                  :name="item.name"
                  :variant
                  :isInNotification="false"
                  action="accept"
                  :show="showDeviceAcceptButton"
                  @update="refreshDevices()"
                  data-test="DeviceActionButtonAccept-component"
                />
                <DeviceActionButton
                  :uid="item.uid"
                  :variant
                  :action="status === 'pending' ? 'reject' : 'remove'"
                  :isInNotification="false"
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
    <TerminalHelper
      v-if="showTerminalHelper"
      v-model="showTerminalHelper"
      :sshid="selectedSshid"
      :user-id="userId"
      :show-checkbox="true"
      data-test="terminal-helper-component"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed, PropType } from "vue";
import { useRouter } from "vue-router";
import { store } from "@/store";
import { actions, authorizer } from "@/authorizer";
import DataTable from "../DataTable.vue";
import DeviceIcon from "../Devices/DeviceIcon.vue";
import DeviceActionButton from "../Devices/DeviceActionButton.vue";
import DeviceDelete from "../Devices/DeviceDelete.vue";
import TagFormUpdate from "../Tags/TagFormUpdate.vue";
import TerminalConnectButton from "../Terminal/TerminalConnectButton.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import TerminalHelper from "../Terminal/TerminalHelper.vue";
import { IDevice, IDeviceMethods } from "@/interfaces/IDevice";
import hasPermission from "@/utils/permission";
import showTag from "@/utils/tag";
import { displayOnlyTenCharacters } from "@/utils/string";
import handleError from "@/utils/handleError";
import { formatFullDateTime } from "@/utils/date";
import { IContainerMethods } from "@/interfaces/IContainer";

const props = defineProps({
  storeMethods: {
    type: Object as PropType<IDeviceMethods | IContainerMethods>,
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

const { fetchDevices, setSort, getFilter, getList, getSortStatusField, getSortStatusString, getNumber } = props.storeMethods;

const router = useRouter();
const loading = ref(false);
const filter = computed(() => getFilter());
const items = computed(() => getList());
const numberDevices = computed(() => getNumber());
const showDeviceAcceptButton = ref(false);
const showDeviceRejectButton = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const status = computed(() => props.status);
const showTerminalHelper = ref(false);
const selectedSshid = ref("");
const userId = computed(() => store.getters["auth/id"]);

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
    text: props.variant === "device" ? "Operating System" : "Image",
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
    text: props.variant === "device" ? "Operating System" : "Image",
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

const getDevices = async (perPageValue: number, pageValue: number, filter: string) => {
  try {
    loading.value = true;
    await fetchDevices({
      perPage: perPageValue,
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

const getSortOrder = () => {
  const currentOrder = getSortStatusString();
  if (currentOrder === "asc") return "desc";
  return "asc";
};

const sortByItem = async (field: string) => {
  setSort({
    sortStatusField: field,
    sortStatusString: getSortOrder(),
  });
  await getDevices(itemsPerPage.value, page.value, filter.value);
};

watch([page, itemsPerPage], async () => {
  await getDevices(itemsPerPage.value, page.value, filter.value);
});

const redirectToDevice = (deviceId: string) => {
  router.push({ name: "DeviceDetails", params: { identifier: deviceId } });
};

const sshidAddress = (item: IDevice) => `${item.namespace}.${item.name}@${window.location.hostname}`;

const openTerminalHelper = (item: IDevice) => {
  selectedSshid.value = sshidAddress(item);
  showTerminalHelper.value = true;
};

const shouldOpenTerminalHelper = () => {
  try {
    const dispensedUsers = JSON.parse(localStorage.getItem("dispenseTerminalHelper") || "[]");
    return !dispensedUsers.includes(userId.value);
  } catch {
    return true;
  }
};

const handleSshidClick = (item: IDevice, copyFn: (text: string) => void) => {
  if (shouldOpenTerminalHelper()) {
    openTerminalHelper(item);
    return;
  }
  copyFn(sshidAddress(item));
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

defineExpose({ page, getSortStatusField, getSortStatusString, showTerminalHelper, openTerminalHelper });
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
