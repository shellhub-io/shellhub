<template>
  <div>
    <DataTable
      v-model:page="page"
      v-model:itemsPerPage="itemsPerPage"
      :headers
      :items="keyList"
      :totalCount="numberKeys"
      :loading
      :itemsPerPageOptions="[10, 20, 50, 100]"
      @update:sort="sortByItem"
      data-test="api-key-list"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in keyList" :key="i">
          <td :class="formatKey(item.expires_in) ? 'text-warning text-center' : 'text-center'">
            <v-icon class="mr-1" :icon="formatKey(item.expires_in) ? 'mdi-clock-alert-outline' : 'mdi-key-outline'" />
            {{ item.name }}
          </td>
          <td :class="formatKey(item.expires_in) ? 'text-warning text-center' : 'text-center'" data-test="key-name">
            {{ item.role }}
          </td>
          <td :class="formatKey(item.expires_in) ? 'text-warning text-center' : 'text-center'" data-test="key-name">
            {{ formatDate(item.expires_in) }}
          </td>
          <td class="text-center" data-test="menu-key-component">
            <v-menu location="bottom" scrim eager>
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
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationRemoveKey()"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <ApiKeyEdit
                        :key-name="item.name"
                        :key-id="item.id"
                        :key-role="item.role"
                        :has-authorization="hasAuthorizationRemoveKey()"
                        :disabled="formatKey(item.expires_in)"
                        @update="refresh()"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationRemoveKey()"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <ApiKeyDelete
                        :keyId="item.name"
                        :has-authorization="hasAuthorizationRemoveKey()"
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
import { computed, onMounted, ref, watch } from "vue";
import axios, { AxiosError } from "axios";
import moment from "moment";
import { useStore } from "@/store";
import DataTable from "@/components/DataTable.vue";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import handleError from "@/utils/handleError";
import ApiKeyDelete from "./ApiKeyDelete.vue";
import ApiKeyEdit from "./ApiKeyEdit.vue";
import useSnackbar from "@/helpers/snackbar";

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
const store = useStore();
const snackbar = useSnackbar();
const numberKeys = computed<number>(
  () => store.getters["apiKeys/getNumberApiKeys"],
);
const keyList = computed(() => store.getters["apiKeys/apiKeyList"]);
const hasAuthorizationRemoveKey = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.apiKey.delete,
    );
  }
  return false;
};

const now = moment().utc();

const formatKey = (unixTime: number): boolean => {
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

const getKey = async (perPageValue: number, pageValue: number) => {
  try {
    loading.value = true;
    await store.dispatch("apiKeys/getApiKey", {
      page: pageValue,
      perPage: perPageValue,
      sortStatusField: store.getters["apiKeys/getSortStatusField"],
      sortStatusString: store.getters["apiKeys/getSortStatusString"],
    });
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        snackbar.showError("You are not authorized to view this API key.");
        handleError(error);
      }
    } else {
      snackbar.showError("Failed to load API keys.");
      handleError(error);
    }
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  getKey(itemsPerPage.value, page.value);
});

const refresh = () => {
  getKey(itemsPerPage.value, page.value);
};

watch([page, itemsPerPage], async () => {
  await getKey(itemsPerPage.value, page.value);
});

const getSortOrder = () => {
  const currentOrder = store.getters["apiKeys/getSortStatusString"];
  if (currentOrder === "asc") return "desc";
  return "asc";
};

const sortByItem = async (field: string) => {
  await store.dispatch("apiKeys/setSortStatus", {
    sortStatusField: field,
    sortStatusString: getSortOrder(),
  });
  await getKey(itemsPerPage.value, page.value);
};

defineExpose({ refresh, formatKey, formatDate, itemsPerPage });
</script>
