<template>
  <div>
    <DataTable
      v-model:page="page"
      v-model:items-per-page="itemsPerPage"
      :headers
      :items="apiKeys"
      :total-count="apiKeysCount"
      :loading
      :items-per-page-options="[10, 20, 50, 100]"
      table-name="apiKeys"
      data-test="api-key-list"
      @update:sort="sortByItem"
    >
      <template #rows>
        <tr
          v-for="item in apiKeys"
          :key="item.id"
        >
          <td
            :class="{ 'text-warning': hasKeyExpired(item.expires_in) }"
            class="text-center"
          >
            <v-icon
              class="mr-1"
              :icon="hasKeyExpired(item.expires_in) ? 'mdi-clock-alert-outline' : 'mdi-key-outline'"
            />
            {{ item.name }}
          </td>
          <td
            :class="{ 'text-warning': hasKeyExpired(item.expires_in) }"
            class="text-center text-capitalize"
            data-test="key-name"
          >
            {{ item.role }}
          </td>
          <td
            :class="{ 'text-warning': hasKeyExpired(item.expires_in) }"
            class="text-center"
            data-test="key-expiry-date"
          >
            {{ formatDate(item.expires_in) }}
          </td>
          <td
            class="text-center"
            data-test="menu-key-component"
          >
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
                />
              </template>
              <v-list
                class="bg-v-theme-surface"
                lines="two"
                density="compact"
              >
                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="canDeleteApiKey"
                >
                  <template #activator="{ props }">
                    <div v-bind="props">
                      <ApiKeyEdit
                        :key-name="item.name"
                        :key-id="item.id"
                        :key-role="item.role"
                        :has-authorization="canDeleteApiKey"
                        :disabled="hasKeyExpired(item.expires_in)"
                        @update="refresh()"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="canDeleteApiKey"
                >
                  <template #activator="{ props }">
                    <div v-bind="props">
                      <ApiKeyDelete
                        :key-id="item.name"
                        :has-authorization="canDeleteApiKey"
                        @update="refresh()"
                      />
                    </div>
                  </template>
                  <span data-test="no-api-key-validate"> You don't have this kind of authorization. </span>
                </v-tooltip>
              </v-list>
            </v-menu>
          </td>
        </tr>
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import axios from "axios";
import moment from "moment";
import DataTable from "@/components/Tables/DataTable.vue";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import ApiKeyDelete from "./ApiKeyDelete.vue";
import ApiKeyEdit from "./ApiKeyEdit.vue";
import useSnackbar from "@/helpers/snackbar";
import useApiKeysStore from "@/store/modules/api_keys";

const headers = [
  {
    text: "Key Name",
    value: "name",
    sortable: true,
  },
  {
    text: "Role",
    value: "role",
  },
  {
    text: "Expiration Date",
    value: "expires_in",
    sortable: true,
  },
  {
    text: "Actions",
    value: "actions",
  },
];
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const sortField = ref<string>("name");
const sortOrder = ref<"asc" | "desc">("asc");
const apiKeyStore = useApiKeysStore();
const snackbar = useSnackbar();
const apiKeysCount = computed(() => apiKeyStore.apiKeysCount);
const apiKeys = computed(() => apiKeyStore.apiKeys);
const canDeleteApiKey = hasPermission("apiKey:delete");

const now = moment().utc();

const hasKeyExpired = (unixTime: number): boolean => {
  if (unixTime === -1) {
    return false;
  }

  const expiryDate = moment.unix(unixTime);

  return now.isAfter(expiryDate);
};

const formatDate = (unixTime: number): string => {
  if (unixTime === -1) {
    return "Never";
  }

  const expiryDate = moment.unix(unixTime);
  const format = "MMM D YYYY";

  return now.isAfter(expiryDate)
    ? `Expired on ${expiryDate.format(format)}.`
    : `Expires on ${expiryDate.format(format)}.`;
};

const fetchApiKeys = async () => {
  try {
    loading.value = true;
    await apiKeyStore.fetchApiKeys({
      page: page.value,
      perPage: itemsPerPage.value,
      sortField: sortField.value,
      sortOrder: sortOrder.value,
    });
  } catch (error: unknown) {
    if (axios.isAxiosError(error) && error.response?.status === 403) {
      snackbar.showError("You are not authorized to view this API key.");
      return;
    }
    snackbar.showError("Failed to load API keys.");
    handleError(error);
  } finally {
    loading.value = false;
  }
};

const refresh = async () => {
  await fetchApiKeys();
};

watch([page, itemsPerPage], async () => {
  await fetchApiKeys();
});

const toggleSortOrder = () => sortOrder.value === "asc" ? "desc" : "asc";

const sortByItem = async (field: string) => {
  sortField.value = field;
  sortOrder.value = toggleSortOrder();
  await fetchApiKeys();
};

defineExpose({ refresh });
</script>
