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
          <td class="text-center">
            <div :class="(item.status === 'connected' ? 'enabled' : 'disabled') + ' text-center'" />
          </td>
          <td class="text-center">
            <v-row>
              <v-col class="d-flex justify-center">
                <v-switch
                  class=""
                  @click="switchConnector(item.uid, item.enable)"
                  v-model="item.enable"
                  inset
                  hide-details
                  :color="item.enable ? 'primary' : 'grey-darken-2'"
                />
              </v-col>
            </v-row>
          </td>
          <td class="text-center">
            <v-chip data-test="sshid-chip">
              <v-tooltip location="bottom">
                <template v-slot:activator="{ props }">
                  <span
                    v-bind="props"
                    @click="copyText(item.address)"
                    @keypress="copyText(item.address)"
                    class="hover-text"
                  >
                    {{ item.address + ":" + item.port }}
                  </span>
                </template>
                <span>Copy IP</span>
              </v-tooltip>
            </v-chip>
          </td>
          <td class="text-center">
            <v-icon
              class="mr-1"
              size="26"
              :color="item.secure ? 'primary' : 'grey-darken-2'"
              :icon="item.secure ? 'mdi-shield-check' : 'mdi-shield-off'"
            />
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
                  :disabled="hasAuthorizationEdit()"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <ConnectorEdit
                        :ipAddress="item.address"
                        :secure="item.secure"
                        :portAddress="item.port"
                        :uid="item.uid"
                        :notHasAuthorization="!hasAuthorizationEdit()"
                        @update="refresh()"
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
                      <ConnectorDelete
                        :uid="item.uid"
                        :notHasAuthorization="!hasAuthorizationRemove()"
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
import { useStore } from "../../store";
import DataTable from "../DataTable.vue";
import ConnectorDelete from "../Connector/ConnectorDelete.vue";
import ConnectorEdit from "../Connector/ConnectorEdit.vue";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import { INotificationsCopy, INotificationsError, INotificationsSuccess } from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const headers = [
  {
    text: "Status",
    value: "status",
  },
  {
    text: "Enable",
    value: "enable",
  },
  {
    text: "Address",
    value: "address",
  },
  {
    text: "Secure",
    value: "secure",
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
  () => store.getters["connectors/getNumberConnectors"],
);
const keyList = computed(() => store.getters["connectors/list"]);
const hasAuthorizationEdit = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.connector.edit,
    );
  }
  return false;
};
const hasAuthorizationRemove = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.connector.remove,
    );
  }
  return false;
};
const getConnectors = async (perPagaeValue: number, pageValue: number) => {
  try {
    loading.value = true;
    await store.dispatch("connectors/fetch", {
      page: pageValue,
      perPage: perPagaeValue,
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
  getConnectors(itemsPerPage.value, page.value);
});

const refresh = () => {
  getConnectors(itemsPerPage.value, page.value);
};

const next = async () => {
  await getConnectors(itemsPerPage.value, ++page.value);
};

const prev = async () => {
  try {
    if (page.value > 1) await getConnectors(itemsPerPage.value, --page.value);
  } catch (error) {
    store.dispatch("snackbar/setSnackbarErrorDefault");
  }
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, async (newItemsPerPage) => {
  await getConnectors(newItemsPerPage, page.value);
});

const sortByItem = async (field: string) => {
  let sortStatusString = store.getters["apiKeys/getSortStatusString"];
  const sortStatusField = store.getters["apiKeys/getSortStatusField"];

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
  await store.dispatch("apiKeys/setSortStatus", {
    sortStatusField: field,
    sortStatusString,
  });
  await getConnectors(itemsPerPage.value, page.value);
};

const copyText = (value: string | undefined) => {
  if (value) {
    navigator.clipboard.writeText(value);
    store.dispatch("snackbar/showSnackbarCopy", INotificationsCopy.connector);
  }
};

const switchConnector = async (uid: string, enable: boolean) => {
  try {
    const payload = {
      uid,
      enable: !enable,
    };
    await store.dispatch("connectors/edit", payload);
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.connectorEdit,
    );
    refresh();
  } catch (error) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.connectorEdit,
    );
    handleError(error);
  }
};

defineExpose({ refresh, getConnectors, itemsPerPage });
</script>

<style scoped>
.enabled {
height: 20px;
width: 20px;
background-color: #4CAF50;
filter: blur(2px);
border-radius: 50%;
display: inline-block;
-webkit-box-shadow: 0px 0px 10px 1px rgba(76,175,79,0.75);
-moz-box-shadow: 0px 0px 10px 1px rgba(76,175,79,0.75);
box-shadow: 0px 0px 10px 1px rgba(76,175,79,0.75);
}

.disabled {
height: 20px;
width: 20px;
background-color: #F44336;
filter: blur(2px);
border-radius: 50%;
display: inline-block;
-webkit-box-shadow: 0px 0px 10px 1px rgba(244, 67, 54,0.75);
-moz-box-shadow: 0px 0px 10px 1px rgba(244, 67, 54,0.75);
box-shadow: 0px 0px 10px 1px rgba(244, 67, 54,0.75);
}
</style>
