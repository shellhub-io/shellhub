<template>
  <DataTable
    v-model:page="page"
    v-model:itemsPerPage="itemsPerPage"
    :headers="headers"
    :items="items"
    :totalCount="totalCount"
    :loading="loading"
    :itemsPerPageOptions="[10, 20, 50]"
    @update:sort="sortByItem"
    data-test="web-endpoints-table"
  >
    <template #rows>
      <tr
        v-for="(endpoint) in items"
        :key="endpoint.address"
        :class="isExpired(endpoint.expires_in) ? 'text-warning' : ''"
      >
        <td data-test="web-endpoint-url">
          <a
            :href="`${urlProtocol}//${endpoint.full_address}`"
            target="_blank"
            rel="noopener noreferrer"
            @click="handleClick"
          >
            {{ `${urlProtocol}//${endpoint.full_address}` }}
          </a>
        </td>
        <td class="text-center">{{ endpoint.host }}</td>
        <td class="text-center">{{ endpoint.port }}</td>
        <td class="text-center">{{ formatDate(endpoint.expires_in) }}</td>
        <td class="text-center">
          <WebEndpointDelete
            :uid="endpoint.device"
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
import { useStore } from "@/store";
import DataTable from "@/components/DataTable.vue";
import WebEndpointDelete from "@/components/WebEndpoints/WebEndpointDelete.vue";
import { IWebEndpoints } from "@/interfaces/IWebEndpoints";

type SortField = "created_at" | "updated_at" | "address" | "uid";

const store = useStore();

const items = computed<IWebEndpoints[]>(() => store.getters["webEndpoints/listWebEndpoints"]);
const totalCount = computed(() => store.getters["webEndpoints/getTotalCount"]);

const page = ref(store.getters["webEndpoints/getPage"]);
const itemsPerPage = ref(store.getters["webEndpoints/getPerPage"]);
const loading = ref(false);

const sortBy = ref<SortField>(store.getters["webEndpoints/getSortBy"]);
const sortDesc = ref<boolean>(store.getters["webEndpoints/getOrderBy"] === "desc");

const headers = [
  { text: "Address", value: "address", sortable: true },
  { text: "Host", value: "host", sortable: true },
  { text: "Port", value: "port", sortable: true },
  { text: "Expiration Date", value: "expires_in", sortable: true },
  { text: "Actions", value: "actions", sortable: false },
];

const fetchWebEndpoints = async () => {
  loading.value = true;
  try {
    await store.dispatch("webEndpoints/get", {
      page: page.value,
      perPage: itemsPerPage.value,
      filter: store.getters["webEndpoints/getFilter"],
      sortBy: sortBy.value,
      orderBy: sortDesc.value ? "desc" : "asc",
    });

    store.commit("webEndpoints/setPagePerPage", {
      page: page.value,
      perPage: itemsPerPage.value,
      filter: store.getters["webEndpoints/getFilter"],
      sortBy: sortBy.value,
      orderBy: sortDesc.value ? "desc" : "asc",
    });
  } finally {
    loading.value = false;
  }
};

const sortByItem = (field: string) => {
  const validFields: SortField[] = ["created_at", "updated_at", "address", "uid"];
  if (!validFields.includes(field as SortField)) return;

  if (sortBy.value === field) {
    sortDesc.value = !sortDesc.value;
  } else {
    sortBy.value = field as SortField;
    sortDesc.value = false;
  }

  fetchWebEndpoints();
};

watch([page, itemsPerPage], () => {
  fetchWebEndpoints();
});

const refresh = () => {
  fetchWebEndpoints();
};

const urlProtocol = ref(window.location.protocol);

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
  setTimeout(() => fetchWebEndpoints(), 30000);
};

onMounted(fetchWebEndpoints);
</script>
