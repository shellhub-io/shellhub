<template>
  <div class="d-flex pa-0 align-center">
    <h1>Device Details</h1>
  </div>
  <v-card class="mt-2 bg-v-theme-surface" v-if="device.uid">
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
              :name="device.name"
              data-test="device-rename-component"
            />

            <div v-if="envVariables.hasWebEndpoints && envVariables.isEnterprise">
              <v-list-item
                v-bind="$attrs"
                @click="showWebEndpointCreate = true"
                data-test="tunnel-create-dialog-btn"
                :disabled="!hasAuthorizationCreateWebEndpoint"
              >
                <div class="d-flex align-center">
                  <div class="mr-2" data-test="create-icon">
                    <v-icon>mdi-web-plus</v-icon>
                  </div>
                  <v-list-item-title> Create Web Endpoint </v-list-item-title>
                </div>
              </v-list-item>
            </div>

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

    </v-card-text>
  </v-card>
  <v-card class="mt-2 pa-4 bg-v-theme-surface" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>

  <WebEndpointCreate
    v-model="showWebEndpointCreate"
    :uid="device.uid"
    :useDevicesList="false"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { displayOnlyTenCharacters } from "../utils/string";
import showTag from "@/utils/tag";
import DeviceIcon from "../components/Devices/DeviceIcon.vue";
import TagFormUpdate from "../components/Tags/TagFormUpdate.vue";
import DeviceDelete from "../components/Devices/DeviceDelete.vue";
import DeviceRename from "../components/Devices/DeviceRename.vue";
import TerminalConnectButton from "../components/Terminal/TerminalConnectButton.vue";
import { formatFullDateTime } from "@/utils/date";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import handleError from "@/utils/handleError";
import { envVariables } from "@/envVariables";
import useSnackbar from "@/helpers/snackbar";
import WebEndpointCreate from "@/components/WebEndpoints/WebEndpointCreate.vue";
import useAuthStore from "@/store/modules/auth";
import useDevicesStore from "@/store/modules/devices";

type DeviceResolver = "uid" | "hostname";

const authStore = useAuthStore();
const devicesStore = useDevicesStore();
const route = useRoute();
const snackbar = useSnackbar();
const { identifier } = route.params;
const resolver = route.query.resolver as DeviceResolver || "uid";
const device = computed(() => devicesStore.device);
const deviceUid = computed(() => device.value.uid);
const showWebEndpointCreate = ref(false);

const sshidAddress = (item) => `${item.namespace}.${item.name}@${window.location.hostname}`;

onMounted(async () => {
  try {
    await devicesStore.fetchDevice({ [resolver]: identifier });
  } catch (error: unknown) {
    snackbar.showError("There was an error loading the device details.");
    handleError(error);
  }
});

const hasAuthorizationCreateWebEndpoint = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.tunnel.create);
};

const refreshDevices = async () => {
  try {
    await devicesStore.fetchDevice({ uid: deviceUid.value });
  } catch (error: unknown) {
    snackbar.showError("There was an error loading the device details.");
    handleError(error);
  }
};

const hasAuthorizationFormUpdate = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.tag.deviceUpdate);
};
</script>
