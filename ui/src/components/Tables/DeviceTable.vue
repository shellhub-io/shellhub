<template>
  <div>
    <DataTable
      v-model:page="page"
      v-model:items-per-page="itemsPerPage"
      :headers="computedHeaders"
      :items
      :total-count="deviceCount"
      :loading
      :items-per-page-options="[10, 20, 50, 100]"
      data-test="items-list"
      @update:sort="sortByItem"
    >
      <template
        v-if="status === 'accepted'"
        #rows
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
          <td class="text-center">
            {{ item.name }}
          </td>
          <td class="text-center">
            <DeviceIcon
              :icon="item.info.id"
              class="mr-2"
              data-test="deviceIcon-component"
            />
            <span>{{ item.info.pretty_name }}</span>
          </td>
          <td class="text-center">
            <CopyWarning :copied-item="'Device SSHID'">
              <template #default="{ copyText }">
                <v-chip data-test="sshid-chip">
                  <v-tooltip location="bottom">
                    <template #activator="{ props }">
                      <span
                        v-bind="props"
                        class="hover-text text-mono"
                        @click="handleSshidClick(item, copyText)"
                        @keypress.enter="handleSshidClick(item, copyText)"
                      >
                        {{ getSshid(item) }}
                      </span>
                    </template>
                    <span>Copy SSHID</span>
                  </v-tooltip>

                  <template #append>
                    <v-tooltip location="bottom">
                      <template #activator="{ props }">
                        <v-icon
                          v-bind="props"
                          icon="mdi-help-circle-outline"
                          size="small"
                          color="primary"
                          class="ml-2"
                          data-test="sshid-help-btn"
                          @click.stop="forceOpenTerminalHelper(item)"
                        />
                      </template>
                      <span>What is an SSHID?</span>
                    </v-tooltip>
                  </template>
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
              <v-chip size="small">
                No tags
              </v-chip>
            </div>
          </td>

          <td class="text-center">
            <v-menu
              location="bottom"
              scrim
              eager
              data-test="v-menu"
            >
              <template #activator="{ props }">
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
                  data-test="mdi-information-list-item"
                  @click="redirectToDevice(item.uid)"
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
                  <template #activator="{ props }">
                    <div v-bind="props">
                      <TagFormUpdate
                        :device-uid="item.uid"
                        :tags-list="item.tags"
                        :has-authorization="canUpdateDeviceTag"
                        @update="getDevices"
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
                  <template #activator="{ props }">
                    <div v-bind="props">
                      <DeviceDelete
                        :variant
                        :uid="item.uid"
                        :has-authorization="canRemoveDevice"
                        @update="getDevices"
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
        #rows
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
              <template #activator="{ props }">
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
                  :is-in-notification="false"
                  action="accept"
                  :show="showDeviceAcceptButton"
                  data-test="DeviceActionButtonAccept-component"
                  @update="getDevices"
                />
                <DeviceActionButton
                  :uid="item.uid"
                  :variant
                  :action="status === 'pending' ? 'reject' : 'remove'"
                  :is-in-notification="false"
                  :show="showDeviceRejectButton"
                  data-test="deviceActionButtonReject-component"
                  @update="getDevices"
                />
              </v-list>
            </v-menu>
          </td>
        </tr>
      </template>
    </DataTable>
    <SSHIDHelper
      v-if="showTerminalHelper"
      v-model="showTerminalHelper"
      :sshid="selectedSshid"
      data-test="sshid-helper-component"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from "vue";
import { useRouter } from "vue-router";
import DataTable from "./DataTable.vue";
import DeviceIcon from "../Devices/DeviceIcon.vue";
import DeviceActionButton from "../Devices/DeviceActionButton.vue";
import DeviceDelete from "../Devices/DeviceDelete.vue";
import TagFormUpdate from "../Tags/TagFormUpdate.vue";
import TerminalConnectButton from "../Terminal/TerminalConnectButton.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import SSHIDHelper from "../Terminal/SSHIDHelper.vue";
import { IDevice, IDeviceMethods, DeviceStatus } from "@/interfaces/IDevice";
import hasPermission from "@/utils/permission";
import showTag from "@/utils/tag";
import { displayOnlyTenCharacters } from "@/utils/string";
import handleError from "@/utils/handleError";
import { formatFullDateTime } from "@/utils/date";
import { IContainerMethods } from "@/interfaces/IContainer";

const props = defineProps<{
  storeMethods: IDeviceMethods | IContainerMethods;
  status: DeviceStatus;
  header: "primary" | "secondary";
  variant: "device" | "container";
}>();

const { fetchDevices, getList, getCount, getFilter } = props.storeMethods;
const router = useRouter();
const loading = ref(false);
const items = computed(() => getList());
const deviceCount = computed(() => getCount());
const showDeviceAcceptButton = ref(false);
const showDeviceRejectButton = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const filter = computed(() => getFilter());
const sortField = ref<string>();
const sortOrder = ref<"asc" | "desc">();
const showTerminalHelper = ref(false);
const selectedSshid = ref("");

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

const computedHeaders = computed(() =>
  props.header === "primary" ? headers : headersSecondary,
);

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

const redirectToDevice = async (deviceId: string) => {
  await router.push({
    name: "DeviceDetails",
    params: { identifier: deviceId },
  });
};

const getSshid = (item: IDevice) =>
  `${item.namespace}.${item.name}@${window.location.hostname}`;

const openTerminalHelper = (item: IDevice) => {
  selectedSshid.value = getSshid(item);
  showTerminalHelper.value = true;
};

const handleSshidClick = (item: IDevice, copyFn: (text: string) => void) => {
  copyFn(getSshid(item));
};

const forceOpenTerminalHelper = (item: IDevice) => {
  openTerminalHelper(item);
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

watch(filter, async () => {
  page.value = 1;
  await getDevices();
});

watch([page, itemsPerPage], async () => {
  await getDevices();
});

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
