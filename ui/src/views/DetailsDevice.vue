<template>
  <div class="d-flex pa-0 align-center">
    <h1>Device Details</h1>
  </div>
  <v-card class="mt-2 bg-v-theme-surface" v-if="!deviceIsEmpty">
    <v-card-title class="pa-4 d-flex align-center justify-space-between">
      <div class="d-flex align-center">
        <TerminalConnectButton
          v-if="device.status === 'accepted'"
          :deviceUid="device.uid"
          :online="device.online"
          :sshid="sshidAddress(device)"
          data-test="connect-btn"
        />
        <span class="ml-4">{{ device.name }}</span>
      </div>

      <div>
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
            <DeviceRename
              :uid="device.uid"
              @newHostname="receiveName"
              data-test="device-rename-component"
            />

            <TunnelCreate
              v-if="envVariables.hasTunnels && envVariables.isEnterprise"
              :uid="device.uid"
              @update="getTunnels"
            />

            <TagFormUpdate
              :device-uid="device.uid"
              :tags-list="device.tags"
              :has-authorization="hasAuthorizationFormUpdate()"
              @update="refreshDevices"
              data-test="tag-form-update-component"
            />

            <DeviceDelete
              variant="device"
              :has-authorization="true"
              :redirect="true"
              :uid="device.uid"
              @update="refreshDevices"
              data-test="device-delete-component"
            />
          </v-list>
        </v-menu>
      </div>
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <div>
        <div class="text-overline mt-3">uid:</div>
        <div data-test="device-uid-field">
          <p>{{ device.uid }}</p>
        </div>
      </div>

      <div v-if="device.identity">
        <div class="text-overline mt-3">mac:</div>
        <div data-test="device-mac-field">
          <code>
            {{ device.identity.mac }}
          </code>
        </div>
      </div>

      <div v-if="device.info">
        <div class="text-overline mt-3">Operating System:</div>
        <div data-test="device-pretty-name-field">
          <DeviceIcon :icon="device.info.id" class="mr-2" />
          <span>{{ device.info.pretty_name }}</span>
        </div>
      </div>

      <div v-if="device.info">
        <div class="text-overline mt-3">Agent Version:</div>
        <div data-test="device-version-field">
          <p>{{ device.info.version }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">Tags:</div>
        <div v-if="device.tags" data-test="device-tags-field">
          <v-tooltip
            v-for="(tag, index) in device.tags"
            :key="index"
            bottom
            :disabled="!showTag(tag)"
          >
            <template #activator="{ props }">
              <v-chip size="small" v-bind="props" class="mr-2">
                {{ displayOnlyTenCharacters(tag) }}
              </v-chip>
            </template>

            <span v-if="showTag(tag)">
              {{ tag }}
            </span>
          </v-tooltip>
        </div>
      </div>
      <div>
        <div class="text-overline mt-3">Last Seen:</div>
        <div data-test="device-last-seen-field">
          <p>{{ formatFullDateTime(device.last_seen) }}</p>
        </div>
      </div>
      <div v-if="envVariables.hasTunnels && envVariables.isEnterprise">
        <div class="text-overline mt-3" data-test="tunnel-list">Tunnel List:</div>
        <TunnelList :deviceUid />
      </div>

    </v-card-text>
  </v-card>
  <v-card class="mt-2 pa-4 bg-v-theme-surface" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import { useStore } from "../store";
import { displayOnlyTenCharacters } from "../utils/string";
import showTag from "../utils/tag";
import DeviceIcon from "../components/Devices/DeviceIcon.vue";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import TagFormUpdate from "../components/Tags/TagFormUpdate.vue";
import TunnelList from "../components/Tunnels/TunnelList.vue";
import DeviceDelete from "../components/Devices/DeviceDelete.vue";
import DeviceRename from "../components/Devices/DeviceRename.vue";
import TerminalConnectButton from "../components/Terminal/TerminalConnectButton.vue";
import { formatFullDateTime } from "@/utils/date";
import handleError from "@/utils/handleError";
import { envVariables } from "@/envVariables";
import TunnelCreate from "@/components/Tunnels/TunnelCreate.vue";
import useSnackbar from "@/helpers/snackbar";

type DeviceResolver = "uid" | "hostname";

const store = useStore();
const route = useRoute();
const snackbar = useSnackbar();
const { identifier } = route.params;
const resolver = route.query.resolver as DeviceResolver || "uid";
const device = computed(() => store.getters["devices/get"]);
const deviceUid = computed(() => device.value.uid);

const sshidAddress = (item) => `${item.namespace}.${item.name}@${window.location.hostname}`;

onMounted(async () => {
  try {
    await store.dispatch("devices/get", { [resolver]: identifier });
  } catch (error: unknown) {
    snackbar.showError("There was an error loading the device details.");
    handleError(error);
  }
});

const deviceIsEmpty = computed(
  () => store.getters["devices/get"]
        && Object.keys(store.getters["devices/get"]).length === 0,
);

const getTunnels = async () => {
  await store.dispatch("tunnels/get", deviceUid.value);
};

const refreshDevices = async () => {
  try {
    await store.dispatch("devices/get", deviceUid.value);
    if (envVariables.isEnterprise) {
      await store.dispatch("tunnels/get", deviceUid.value);
    }
  } catch (error: unknown) {
    snackbar.showError("There was an error loading the device details.");
    handleError(error);
  }
};

const hasAuthorizationFormUpdate = () => {
  const role = store.getters["auth/role"];
  return !!role && hasPermission(authorizer.role[role], actions.tag.deviceUpdate);
};

const receiveName = (params: string) => {
  device.value.name = params;
};
</script>
