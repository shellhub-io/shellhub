<template>
  <DataTable
    v-model:page="page"
    v-model:itemsPerPage="itemsPerPage"
    :headers
    :items="connectors"
    :totalCount="connectorCount"
    :loading
    :itemsPerPageOptions="[10, 20, 50, 100]"
    data-test="connector-list"
  >
    <template v-slot:rows>
      <tr v-for="(item, i) in connectors" :key="i">
        <td class="text-center">
          <div data-test="status-connector" :class="(item.status.state === 'connected' ? 'enabled' : 'disabled') + ' text-center'" />
        </td>
        <td class="text-center">
          <div
            class="d-flex justify-center"
            data-test="switch-enable"
          >
            <v-switch
              v-model="item.enable"
              @click="toggleConnectorState(item)"
              inset
              hide-details
              :color="item.enable ? 'primary' : 'grey-darken-2'"
            />
          </div>
        </td>
        <td class="text-center">
          <CopyWarning :copied-item="'Connector host'">
            <template #default="{ copyText }">
              <v-chip data-test="ip-chip">
                <v-tooltip location="bottom">
                  <template v-slot:activator="{ props }">
                    <span
                      v-bind="props"
                      @click='copyText(`${item.address}:${item.port}`)'
                      @keypress='copyText(`${item.address}:${item.port}`)'
                      class="hover-text"
                    >
                      {{ `${item.address}:${item.port}` }}
                    </span>
                  </template>
                  <span>Copy IP</span>
                </v-tooltip>
              </v-chip>
            </template>
          </CopyWarning>
        </td>
        <td class="text-center">
          <v-icon
            data-test="secure-icon"
            class="mr-1"
            size="26"
            :color="item.secure ? 'primary' : 'grey-darken-2'"
            :icon="item.secure ? 'mdi-lock-check' : 'mdi-lock-open-alert'"
          />
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
                data-test="connector-list-actions"
              />
            </template>
            <v-list class="bg-v-theme-surface" lines="two" density="compact">
              <v-list-item @click="redirectToDetails(item.uid)">
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
                :disabled="hasAuthorizationEdit()"
              >
                <template v-slot:activator="{ props }">
                  <div v-bind="props">
                    <ConnectorEdit
                      :ipAddress="item.address"
                      :secure="item.secure"
                      :portAddress="item.port"
                      :uid="item.uid"
                      :hasAuthorization="hasAuthorizationEdit()"
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
                      :hasAuthorization="hasAuthorizationRemove()"
                      @update="refresh()"
                    />
                  </div>
                </template>
                <span data-test="no-connector-validate"> You don't have this kind of authorization. </span>
              </v-tooltip>
            </v-list>
          </v-menu>
        </td>
      </tr>
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { envVariables } from "@/envVariables";
import DataTable from "../DataTable.vue";
import ConnectorDelete from "../Connector/ConnectorDelete.vue";
import ConnectorEdit from "../Connector/ConnectorEdit.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import handleError from "@/utils/handleError";
import { router } from "@/router";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import useConnectorStore from "@/store/modules/connectors";
import { IConnector } from "@/interfaces/IConnector";

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
    text: "Host",
    value: "host",
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
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const authStore = useAuthStore();
const connectorStore = useConnectorStore();
const { connectors, connectorCount } = storeToRefs(connectorStore);

const hasAuthorizationEdit = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.connector.edit);
};

const hasAuthorizationRemove = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.connector.remove);
};

const getConnectors = async () => {
  try {
    loading.value = true;
    await connectorStore.fetchConnectorList({
      page: page.value,
      perPage: itemsPerPage.value,
    });
  } catch (error: unknown) {
    snackbar.showError("An error occurred while loading connectors");
    handleError(error);
  }

  loading.value = false;
};

onMounted(async () => {
  if (envVariables.isCommunity) {
    return;
  }
  await getConnectors();
});

const refresh = async () => {
  await getConnectors();
};

watch([page, itemsPerPage], async () => {
  await getConnectors();
});

const redirectToDetails = (uid: string) => {
  router.push({ name: "ConnectorDetails", params: { id: uid } });
};

const toggleConnectorState = async (item: IConnector) => {
  try {
    const payload = {
      ...item,
      enable: !item.enable,
    };
    await connectorStore.updateConnector(payload);
    snackbar.showSuccess("Connector updated successfully.");
    refresh();
  } catch (error) {
    snackbar.showError("An error occurred while updating the connector.");
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
