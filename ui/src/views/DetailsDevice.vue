<template>
  <div class="d-flex pa-0 align-center">
    <h1>Device Details</h1>
  </div>
  <v-card class="mt-2 border rounded bg-background" elevation="0" v-if="device.uid">
    <v-card-title class="pa-4 d-flex align-center justify-space-between bg-v-theme-surface">
      <div class="d-flex align-center ml-2">
        <TerminalConnectButton
          v-if="device.status === 'accepted'"
          :deviceUid="device.uid"
          :online="device.online"
          :sshid="sshidAddress(device)"
          data-test="connect-btn"
        />
        <span class="ml-6">{{ device.name }}</span>
      </div>

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
              :disabled="!canCreateWebEndpoint"
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
            :has-authorization="canUpdateDeviceTag"
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
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <v-row class="py-3">
        <v-col cols="12" md="6" class="my-0 py-0">
          <div data-test="device-uid-field">
            <div class="item-title">UID:</div>
            <p class="text-truncate">{{ device.uid }}</p>
          </div>

          <div v-if="device.identity" data-test="device-mac-field">
            <div class="item-title">MAC:</div>
            <code>{{ device.identity.mac }}</code>
          </div>

          <div v-if="device.info" data-test="device-pretty-name-field">
            <div class="item-title">Operating System:</div>
            <div>
              <DeviceIcon :icon="device.info.id" class="mr-2" />
              <span>{{ device.info.pretty_name }}</span>
            </div>
          </div>
        </v-col>

        <v-col cols="12" md="6" class="my-0 py-0">
          <div v-if="device.info" data-test="device-version-field">
            <div class="item-title">Agent Version:</div>
            <p>{{ device.info.version }}</p>
          </div>

          <div v-if="device.tags?.length" data-test="device-tags-field">
            <div class="item-title">Tags:</div>
            <div v-if="device.tags">
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

          <div data-test="device-last-seen-field">
            <div class="item-title">Last Seen:</div>
            <p>{{ formatFullDateTime(device.last_seen) }}</p>
          </div>
        </v-col>
      </v-row>
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
import handleError from "@/utils/handleError";
import { envVariables } from "@/envVariables";
import useSnackbar from "@/helpers/snackbar";
import WebEndpointCreate from "@/components/WebEndpoints/WebEndpointCreate.vue";
import useDevicesStore from "@/store/modules/devices";

type DeviceResolver = "uid" | "hostname";

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

const canCreateWebEndpoint = hasPermission("webEndpoint:create");

const canUpdateDeviceTag = hasPermission("tag:update");

const refreshDevices = async () => {
  try {
    await devicesStore.fetchDevice({ uid: deviceUid.value });
  } catch (error: unknown) {
    snackbar.showError("There was an error loading the device details.");
    handleError(error);
  }
};

</script>

<style lang="scss" scoped>
.item-title {
  margin-top: 0.75rem;
  // Vuetify's text-overline styles
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.1666666667em;
  line-height: 2.667;
}
</style>
