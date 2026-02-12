<template>
  <DataTable
    v-model:page="page"
    v-model:items-per-page="itemsPerPage"
    :headers="headers"
    :items="webEndpoints"
    :total-count="totalCount"
    :loading="loading"
    :items-per-page-options="[10, 20, 50]"
    table-name="webEndpoints"
    data-test="web-endpoints-table"
    @update:sort="sortByItem"
  >
    <template #rows>
      <tr
        v-for="endpoint in webEndpoints"
        :key="endpoint.address"
        data-test="web-endpoint-row"
        :class="isExpired(endpoint.expires_in) ? 'text-warning' : ''"
      >
        <td class="d-flex align-center justify-center text-center">
          <DeviceIcon
            :icon="endpoint.device?.info?.id"
            class="mr-2"
          />
          <div class="d-flex flex-column align-center">
            <p
              class="link text-truncate"
              @click="redirectDevice(endpoint.device_uid)"
              @keyup="redirectDevice(endpoint.device_uid)"
            >
              {{ endpoint.device?.name }}
            </p>
            <small class="text-medium-emphasis">{{ endpoint.device?.info?.pretty_name }}</small>
          </div>
        </td>

        <td data-test="web-endpoint-url">
          <a
            :href="`${protocol}//${endpoint.full_address}`"
            target="_blank"
            rel="noopener noreferrer"
            @click="handleClick"
          >
            {{ `${protocol}//${endpoint.full_address}` }}
          </a>
        </td>

        <td class="text-center">
          {{ endpoint.host }}
        </td>
        <td class="text-center">
          {{ endpoint.port }}
        </td>

        <td
          class="text-center"
          data-test="web-endpoint-tls"
        >
          <v-chip
            v-if="endpoint.tls?.enabled"
            size="small"
          >
            {{ endpoint.tls?.domain }}
          </v-chip>

          <v-chip
            v-else
            size="small"
            color="error"
          >
            Disabled
          </v-chip>
        </td>

        <td class="text-center">
          {{ formatDate(endpoint.expires_in) }}
        </td>
        <td class="text-center">
          <WebEndpointDelete
            :uid="endpoint.device_uid"
            :address="endpoint.address"
            @update="refresh"
          />
        </td>
      </tr>
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted } from "vue";
import moment from "moment";
import { useRouter } from "vue-router";
import DataTable from "@/components/Tables/DataTable.vue";
import WebEndpointDelete from "@/components/WebEndpoints/WebEndpointDelete.vue";
import DeviceIcon from "@/components/Devices/DeviceIcon.vue";
import { IWebEndpoint } from "@/interfaces/IWebEndpoints";
import useWebEndpointsStore from "@/store/modules/web_endpoints";
import handleError from "@/utils/handleError";

type SortField = "created_at" | "updated_at" | "address" | "uid";

const webEndpointsStore = useWebEndpointsStore();
const router = useRouter();

const webEndpoints = computed<IWebEndpoint[]>(() => webEndpointsStore.webEndpoints);
const totalCount = computed(() => webEndpointsStore.webEndpointCount);

const page = ref(1);
const itemsPerPage = ref(10);
const loading = ref(false);
const sortField = ref<SortField>();
const sortOrder = ref<"asc" | "desc">();
const { protocol } = window.location;

const headers = [
  { text: "Device", value: "device", sortable: false },
  { text: "Address", value: "address", sortable: true },
  { text: "Host", value: "host", sortable: true },
  { text: "Port", value: "port", sortable: true },
  { text: "Domain", value: "tls", sortable: false },
  { text: "Expiration Date", value: "expires_in", sortable: true },
  { text: "Actions", value: "actions", sortable: false },
];

const fetchWebEndpoints = async () => {
  loading.value = true;
  try {
    await webEndpointsStore.fetchWebEndpointsList({
      page: page.value,
      perPage: itemsPerPage.value,
      sortField: sortField.value,
      sortOrder: sortOrder.value,
    });
  } catch (error) {
    handleError(error);
  }

  loading.value = false;
};

const getSortOrder = () => sortOrder.value === "asc" ? "desc" : "asc";

const sortByItem = async (field: SortField) => {
  sortField.value = field;
  sortOrder.value = getSortOrder();
  await fetchWebEndpoints();
};

watch([page, itemsPerPage], async () => {
  await fetchWebEndpoints();
});

const refresh = async () => {
  await fetchWebEndpoints();
};

const isExpired = (date: string) => date !== "0001-01-01T00:00:00Z" && moment().utc().isAfter(moment(date));

const formatDate = (expiresIn: string) => {
  if (expiresIn === "0001-01-01T00:00:00Z") return "Never Expires";
  const expirationDate = moment(expiresIn);
  const format = "MMM D YYYY, h:mm:ss a";
  return isExpired(expiresIn)
    ? `Expired on ${expirationDate.format(format)}`
    : `Expires on ${expirationDate.format(format)}`;
};

const handleClick = () => {
  setTimeout(() => void fetchWebEndpoints(), 30000);
};

const redirectDevice = async (deviceUid: string) => {
  await router.push({ name: "DeviceDetails", params: { identifier: deviceUid } });
};

onMounted(fetchWebEndpoints);
</script>

<style scoped>

.link {
  text-decoration: underline;
  cursor: pointer;
}
</style>
