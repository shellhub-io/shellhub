<template>
  <div class="d-flex pa-0 align-center">
    <h1>Connector Details</h1>
  </div>
  <v-card
    v-if="connector.uid"
    class="mt-2 bg-v-theme-surface"
  >
    <v-card-title class="pa-4 d-flex align-center justify-space-between">
      <div>
        <v-row>
          <v-col class="pr-0">
            <v-switch
              v-model="connector.enable"
              inset
              hide-details
              :color="connector.enable ? 'primary' : 'grey-darken-2'"
              @click="toggleConnectorState"
            />
          </v-col>
          <v-col class="mt-3">
            <span>Docker Host: </span><code>
              <v-chip
                data-test="sshid-chip text-overline"
                :color="connector.secure ? 'success' : 'warning'"
                :prepend-icon="connector.secure ? 'mdi-lock-check' : 'mdi-lock-open-alert'"
                variant="outlined"
              >
                <v-tooltip location="bottom">
                  <template #activator="{ props }">
                    <span
                      v-bind="props"
                      class="hover-text"
                    > {{ connector.address + ":" + connector.port }}
                    </span>
                  </template>
                  <span>{{ connector.secure ? 'Secure' : 'Insecure' }} Connection</span>
                </v-tooltip>
              </v-chip>
            </code>
          </v-col>
        </v-row>
      </div>
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
            :disabled="canEditConnector"
          >
            <template #activator="{ props }">
              <div v-bind="props">
                <ConnectorEdit
                  :ip-address="connector.address"
                  :secure="connector.secure"
                  :port-address="connector.port"
                  :uid="connector.uid"
                  :has-authorization="canEditConnector"
                  @update="getConnector"
                />
              </div>
            </template>
            <span> You don't have this kind of authorization. </span>
          </v-tooltip>

          <v-tooltip
            location="bottom"
            class="text-center"
            :disabled="canRemoveConnector"
          >
            <template #activator="{ props }">
              <div v-bind="props">
                <ConnectorDelete
                  :uid="connector.uid"
                  :has-authorization="canRemoveConnector"
                  @update="redirectContainers"
                />
              </div>
            </template>
            <span data-test="no-api-key-validate"> You don't have this kind of authorization. </span>
          </v-tooltip>
        </v-list>
      </v-menu>
    </v-card-title>

    <v-divider />

    <div class="pa-4 pt-0">
      <div
        v-if="connector.status?.message"
        class="text-overline mt-3"
      >
        <v-alert
          variant="tonal"
          type="error"
          :text="connector.status.message"
        />
      </div>
      <v-alert
        v-if="!connector.secure && !connector.status?.message"
        class="text-subtitle-2 mt-4"
        type="warning"
        variant="tonal"
        title="Your connector is vulnerable to security risks"
      >
        <template #text>
          <div>
            It is highly recommended to use TLS certificates to secure your Docker Connector.
            Without TLS, your connection is vulnerable to various security risks.
            <span>
              Please checkout the
              <a
                href="https://docs.docker.com/engine/security/protect-access/#use-tls-https-to-protect-the-docker-daemon-socket"
                target="_blank"
                rel="noopener"
              >Docker Documentation.</a>
            </span>
          </div>
        </template>
      </v-alert>
    </div>
    <v-card-text v-if="connector.status?.state === 'connected'">
      <h2>Docker Engine Info</h2>
      <v-table class="bg-v-theme-surface border mt-3">
        <thead>
          <tr>
            <th class="text-left">
              Name
            </th>
            <th class="text-left">
              Value
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(item, i) in filteredItems"
            :key="i"
          >
            <td>{{ item.name }}</td>
            <td>
              <v-chip
                v-if="connectorInfo[item.value] === ''"
                density="compact"
              >
                unknown
              </v-chip>{{ connectorInfo[item.value] }}
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card-text>
  </v-card>
  <v-card
    v-else
    class="mt-2 pa-4 bg-v-theme-surface"
  >
    <p class="text-center">
      Something went wrong, try again!
    </p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { storeToRefs } from "pinia";
import ConnectorDelete from "../components/Connector/ConnectorDelete.vue";
import ConnectorEdit from "../components/Connector/ConnectorEdit.vue";
import hasPermission from "../utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useConnectorStore from "@/store/modules/connectors";
import { IConnectorPayload } from "@/interfaces/IConnector";

const connectorStore = useConnectorStore();
const router = useRouter();
const route = useRoute();
const snackbar = useSnackbar();
const connectorUid = computed(() => route.params.id as string);
const { connector, connectorInfo } = storeToRefs(connectorStore);

const filteredItems = [
  {
    name: "Hostname",
    value: "Name",
  },
  {
    name: "Kernel Version",
    value: "KernelVersion",
  },
  {
    name: "Architecture",
    value: "Architecture",
  },
  {
    name: "Operating System",
    value: "OperatingSystem",
  },
  {
    name: "Operating System Version",
    value: "OSVersion",
  },
  {
    name: "Docker Host Version",
    value: "ServerVersion",
  },
  {
    name: "Containers Count",
    value: "Containers",
  },
  {
    name: "Containers Running",
    value: "ContainersRunning",
  },
  {
    name: "Containers Paused",
    value: "ContainersPaused",
  },
  {
    name: "Containers Stopped",
    value: "ContainersStopped",
  },
];

const redirectContainers = async () => {
  await router.push({ name: "containers" });
};

const canEditConnector = hasPermission("connector:edit");

const canRemoveConnector = hasPermission("connector:remove");

const getConnectorInfo = async () => {
  try {
    await connectorStore.getConnectorInfo(connectorUid.value);
  } catch (error: unknown) {
    snackbar.showError("Failed to load the connector info.");
    handleError(error);
  }
};

const getConnector = async () => {
  try {
    await connectorStore.fetchConnectorById(connectorUid.value);
    await getConnectorInfo();
  } catch (error: unknown) {
    snackbar.showError("Error loading the connector.");
    handleError(error);
  }
};

const toggleConnectorState = async () => {
  try {
    const payload = {
      ...connector.value,
      enable: !connector.value?.enable,
    };
    await connectorStore.updateConnector(payload as IConnectorPayload);
    snackbar.showSuccess("The connector has been updated.");
    await getConnector();
  } catch (error) {
    snackbar.showError("Failed to update the connector.");
    handleError(error);
  }
};

onMounted(async () => { await getConnector(); });
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
