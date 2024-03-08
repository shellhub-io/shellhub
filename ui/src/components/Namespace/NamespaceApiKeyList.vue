<template>
  <div>
    <DataTable
      :headers="headers"
      :items="keyList"
      :itemsPerPage="itemsPerPage"
      :nextPage="next"
      :previousPage="prev"
      :loading="loading"
      :actualPage="page"
      :totalCount="numberKeys"
      :comboboxOptions="[10, 20, 50, 100]"
      @changeItemsPerPage="changeItemsPerPage"
      @clickNextPage="next"
      @clickPreviousPage="prev"
      @clickSortableIcon="sortByItem"
      data-test="api-key-list"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in keyList" :key="i">
          <td :class="formatKey(item.expires_in) ? 'text-warning' : ''">
            <v-icon class="mr-1" :icon="formatKey(item.expires_in) ? 'mdi-clock-alert-outline' : 'mdi-key-outline'" data-test="key-icon" />
            {{ item.name }}
          </td>
          <td :class="formatKey(item.expires_in) ? 'text-warning text-center' : 'text-center'" data-test="key-name">
            {{ formatDate(item.expires_in) }}
          </td>
          <td class="text-center" data-test="menu-key-component">
            <v-menu location="bottom" scrim eager>
              <template v-slot:activator="{ props }">
                <v-chip v-bind="props" density="comfortable" size="small">
                  <v-icon>mdi-dots-horizontal</v-icon>
                </v-chip>
              </template>
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationRemoveKey()"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <NamespaceEditApiKey
                        :key-name="item.name"
                        :key-id="item.id"
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
                      <NamespaceDeleteApiKey
                        :keyId="item.id"
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
import { useStore } from "../../store";
import DataTable from "../DataTable.vue";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import { INotificationsError } from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";
import NamespaceDeleteApiKey from "./NamespaceDeleteApiKey.vue";
import NamespaceEditApiKey from "./NamespaceEditApiKey.vue";

const headers = [
  {
    text: "Key Name",
    value: "name",
    sortable: true,
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
const numberKeys = computed<number>(
  () => store.getters["auth/getNumberApiKeys"],
);
const keyList = computed(() => store.getters["auth/apiKeyList"]);
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

const tenant = computed(() => localStorage.getItem("tenant"));

const getKey = async (perPagaeValue: number, pageValue: number) => {
  try {
    loading.value = true;
    await store.dispatch("auth/getApiKey", {
      tenant: tenant.value,
      perPage: perPagaeValue,
      page: pageValue,
      sortStatusField: store.getters["auth/getSortStatusField"],
      sortStatusString: store.getters["auth/getSortStatusString"],
    });
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        store.dispatch("snackbar/showSnackbarErrorAssociation");
        handleError(error);
      }
    } else {
      store.dispatch(
        "snackbar/showSnackbarErrorAction",
        INotificationsError.namespaceLoad,
      );
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

const next = async () => {
  await getKey(itemsPerPage.value, ++page.value);
};

const prev = async () => {
  try {
    if (page.value > 1) await getKey(itemsPerPage.value, --page.value);
  } catch (error) {
    store.dispatch("snackbar/setSnackbarErrorDefault");
  }
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, async (newItemsPerPage) => {
  await getKey(newItemsPerPage, page.value);
});

const sortByItem = async (field: string) => {
  let sortStatusString = store.getters["auth/getSortStatusString"];
  const sortStatusField = store.getters["auth/getSortStatusField"];

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
  await store.dispatch("auth/setSortStatus", {
    sortStatusField: field,
    sortStatusString,
  });
  await getKey(itemsPerPage.value, page.value);
};

defineExpose({ refresh, formatKey, formatDate, itemsPerPage });
</script>
