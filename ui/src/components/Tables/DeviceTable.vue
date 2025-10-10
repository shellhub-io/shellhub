<template>
  <div>
    <DataTable
      v-model:page="page"
      v-model:itemsPerPage="itemsPerPage"
      :headers="computedHeaders"
      :items
      :totalCount="deviceCount"
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
              :device-uid="item.uid"
              :device-name="item.name"
              :sshid="getSshid(item)"
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
                        {{ getSshid(item) }}
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
                :disabled="!showTag(tag.name)"
              >
                <template #activator="{ props }">
                  <v-chip
                    size="small"
                    v-bind="props"
                    class="mr-1"
                    data-test="tag-chip"
                  >
                    {{ displayOnlyTenCharacters(tag.name) }}
                  </v-chip>
                </template>

                <span>
                  {{ tag.name }}
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
                  :disabled="canUpdateDeviceTag"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <TagFormUpdate
                        :device-uid="item.uid"
                        :tags-list="item.tags"
                        :has-authorization="canUpdateDeviceTag"
                        @update="refreshDevices"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="canRemoveDevice"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <DeviceDelete
                        :variant
                        :uid="item.uid"
                        :hasAuthorization="canRemoveDevice"
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
import { ref, onMounted, watch, computed } from "vue";
import { useRouter } from "vue-router";
import DataTable from "../DataTable.vue";
import DeviceIcon from "../Devices/DeviceIcon.vue";
import DeviceActionButton from "../Devices/DeviceActionButton.vue";
import DeviceDelete from "../Devices/DeviceDelete.vue";
import TagFormUpdate from "../Tags/TagFormUpdate.vue";
import TerminalConnectButton from "../Terminal/TerminalConnectButton.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import TerminalHelper from "../Terminal/TerminalHelper.vue";
import { IDevice, IDeviceMethods, DeviceStatus } from "@/interfaces/IDevice";
import hasPermission from "@/utils/permission";
import showTag from "@/utils/tag";
import { displayOnlyTenCharacters } from "@/utils/string";
import handleError from "@/utils/handleError";
import { formatFullDateTime } from "@/utils/date";
import { IContainerMethods } from "@/interfaces/IContainer";
import useAuthStore from "@/store/modules/auth";

const props = defineProps<{
  storeMethods: IDeviceMethods | IContainerMethods;
  status: DeviceStatus;
  header: "primary" | "secondary";
  variant: "device" | "container";
}>();

const { fetchDevices, getList, getCount } = props.storeMethods;
const authStore = useAuthStore();
const router = useRouter();
const loading = ref(false);
const items = computed(() => getList());
const deviceCount = computed(() => getCount());
const showDeviceAcceptButton = ref(false);
const showDeviceRejectButton = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const status = computed(() => props.status);
const sortField = ref();
const sortOrder = ref();
const showTerminalHelper = ref(false);
const selectedSshid = ref("");
const userId = authStore.id;

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

const getDevices = async () => {
  try {
    loading.value = true;
    await fetchDevices({
      perPage: itemsPerPage.value,
      page: page.value,
      status: props.status,
      sortField: sortField.value,
      sortOrder: sortOrder.value,
    });
  } catch (error: unknown) {
    handleError(error);
  }
  loading.value = false;
};

const redirectToDevice = (deviceId: string) => {
  router.push({ name: "DeviceDetails", params: { identifier: deviceId } });
};

const getSshid = (item: IDevice) => `${item.namespace}.${item.name}@${window.location.hostname}`;

const openTerminalHelper = (item: IDevice) => {
  selectedSshid.value = getSshid(item);
  showTerminalHelper.value = true;
};

const shouldOpenTerminalHelper = () => {
  try {
    const dispensedUsers = JSON.parse(localStorage.getItem("dispenseTerminalHelper") || "[]");
    return !dispensedUsers.includes(userId);
  } catch {
    return true;
  }
};

const handleSshidClick = (item: IDevice, copyFn: (text: string) => void) => {
  if (shouldOpenTerminalHelper()) {
    openTerminalHelper(item);
    return;
  }
  copyFn(getSshid(item));
};

const canUpdateDeviceTag = hasPermission("tag:update");

const canRemoveDevice = hasPermission("device:remove");

const getSortOrder = () => {
  const currentOrder = sortOrder.value;
  if (currentOrder === "asc") return "desc";
  return "asc";
};

const sortByItem = async (field: string) => {
  sortField.value = field;
  sortOrder.value = getSortOrder();
  await getDevices();
};

watch([page, itemsPerPage], async () => {
  await getDevices();
});

const refreshDevices = async () => {
  await getDevices();
};

onMounted(async () => {
  await getDevices();
});

defineExpose({ page, showTerminalHelper, openTerminalHelper });
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
