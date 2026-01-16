<template>
  <div>
    <DataTable
      v-model:items-per-page="itemsPerPage"
      v-model:page="page"
      :headers
      :items="devices"
      :loading
      :total-count="devicesCount"
      :items-per-page-options="[10, 20, 50, 100]"
      table-name="adminDevices"
      data-test="devices-list"
      @update:sort="sortByItem"
    >
      <template #rows>
        <tr
          v-for="(item, i) in devices"
          :key="i"
        >
          <td>
            <v-icon
              v-if="item.online"
              color="success"
              data-test="success-icon"
              icon="mdi-check-circle"
            />
            <v-icon
              v-else
              color="#E53935"
              data-test="error-icon"
              icon="mdi-close-circle"
            />
          </td>
          <td>{{ item.name }}</td>
          <td>
            <span class="d-inline-flex align-center ga-2">
              <DeviceIcon :icon="item.info.id" />
              {{ item.info.pretty_name }}
            </span>
          </td>
          <td>
            <router-link
              :to="{ name: 'namespaceDetails', params: { id: item.tenant_id } }"
              class="hyper-link"
              data-test="namespace-link"
            >
              {{ item.namespace }}
            </router-link>
          </td>
          <td>
            <div v-if="item.tags[0]">
              <v-tooltip
                v-for="(tag, index) in item.tags"
                :key="index"
                bottom
                :disabled="!showTag(tag.name)"
              >
                <template #activator="{ props }">
                  <v-chip
                    size="small"
                    v-bind="props"
                  >
                    {{ displayOnlyTenCharacters(tag.name) }}
                  </v-chip>
                </template>

                <span>
                  {{ tag.name }}
                </span>
              </v-tooltip>
            </div>
          </td>
          <td>
            {{ formatFullDateTime(item.last_seen) }}
          </td>
          <td>
            <v-chip
              size="small"
              class="text-capitalize"
              :text="item.status"
            />
          </td>
          <td>
            <v-tooltip
              bottom
              anchor="bottom"
            >
              <template #activator="{ props }">
                <v-icon
                  tag="a"
                  dark
                  v-bind="props"
                  tabindex="0"
                  icon="mdi-information"
                  data-test="info-button"
                  @click="redirectToDevice(item.uid)"
                  @keypress.enter="redirectToDevice(item.uid)"
                />
              </template>
              <span>Info</span>
            </v-tooltip>
          </td>
        </tr>
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from "vue";
import { useRouter } from "vue-router";
import useDevicesStore from "@admin/store/modules/devices";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "@/components/Tables/DataTable.vue";
import DeviceIcon from "@/components/Devices/DeviceIcon.vue";
import { formatFullDateTime } from "@/utils/date";
import { displayOnlyTenCharacters } from "@/utils/string";
import showTag from "@/utils/tag";
import handleError from "@/utils/handleError";

const router = useRouter();
const snackbar = useSnackbar();
const devicesStore = useDevicesStore();
const page = ref(1);
const itemsPerPage = ref(10);
const loading = ref(false);
const devices = computed(() => devicesStore.devices);
const devicesCount = computed(() => devicesStore.deviceCount);
const sortField = ref<string>();
const sortOrder = ref<"asc" | "desc" | undefined>(undefined);

const headers = ref([
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
    text: "Info",
    value: "info",
    sortable: true,
  },
  {
    text: "Namespace",
    value: "namespace",
    sortable: true,
  },
  {
    text: "Tags",
    value: "tags",
  },
  {
    text: "Last Seen",
    value: "last_seen",
    sortable: true,
  },
  {
    text: "Status",
    value: "status",
    sortable: true,
  },
  {
    text: "Actions",
    value: "actions",
  },
]);

const fetchDevices = async () => {
  try {
    loading.value = true;
    await devicesStore.fetchDeviceList({
      perPage: itemsPerPage.value,
      page: page.value,
      sortField: sortField.value,
      sortOrder: sortOrder.value,
    });
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to fetch devices.");
  } finally {
    loading.value = false;
  }
};

const getSortOrder = () => sortOrder.value === "asc" ? "desc" : "asc";

const sortByItem = async (field: string) => {
  sortField.value = field;
  sortOrder.value = getSortOrder();
  await fetchDevices();
};

const redirectToDevice = async (deviceId: string) => {
  await router.push({ name: "deviceDetails", params: { id: deviceId } });
};

watch([itemsPerPage, page], async () => { await fetchDevices(); });

onMounted(async () => { await fetchDevices(); });
</script>

<style scoped>
.hyper-link {
  color: inherit;
  text-decoration: underline;
}

.hyper-link:visited,
.hyper-link:hover,
.hyper-link:active {
  color: inherit;
}
</style>
