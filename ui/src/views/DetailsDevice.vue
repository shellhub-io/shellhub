<template>
  <div class="d-flex pa-0 align-center">
    <h1>Device Details</h1>
  </div>
  <v-card class="mt-2 bg-v-theme-surface" v-if="!deviceIsEmpty">
    <v-card-title class="pa-4 d-flex align-center justify-space-between">
      <div class="d-flex align-center">
        <TerminalDialog
          v-if="device.status === 'accepted'"
          :enable-connect-button="true"
          :uid="device.uid"
          :online="device.online"
          data-test="terminalDialog-component"
        />
        <span class="ml-2">{{ device.name }}</span>
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
              @newHostname="receiveName"
              data-test="deviceRename-component"
            />

            <TunnelCreate
              v-if="envVariables.hasTunnels && envVariables.isEnterprise"
              :uid="device.uid"
              @update="getTunnels"
            />

            <TagFormUpdate
              :device-uid="device.uid"
              :tagsList="device.tags"
              @update="refreshUsers"
              data-test="tagFormUpdate-component"
            />

            <DeviceDelete
              variant="device"
              :uid="device.uid"
              @update="refreshUsers"
              data-test="deviceDelete-component"
            />
          </v-list>
        </v-menu>
      </div>
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <div>
        <div class="text-overline mt-3">uid:</div>
        <div data-test="deviceUid-field">
          <p>{{ device.uid }}</p>
        </div>
      </div>

      <div v-if="device.identity">
        <div class="text-overline mt-3">mac:</div>
        <div data-test="deviceMac-field">
          <code>
            {{ device.identity.mac }}
          </code>
        </div>
      </div>

      <div v-if="device.info">
        <div class="text-overline mt-3">Operating System:</div>
        <div data-test="devicePrettyName-field">
          <DeviceIcon :icon="device.info.id" class="mr-2" />
          <span>{{ device.info.pretty_name }}</span>
        </div>
      </div>

      <div v-if="device.info">
        <div class="text-overline mt-3">Agent Version:</div>
        <div data-test="deviceVersion-field">
          <p>{{ device.info.version }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">Tags:</div>
        <div v-if="device.tags" data-test="deviceTags-field">
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
        <div data-test="deviceConvertDate-field">
          <p>{{ formatFullDateTime(device.last_seen) }}</p>
        </div>
      </div>
      <div v-if="envVariables.hasTunnels && envVariables.isEnterprise">
        <div class="text-overline mt-3" data-test="tunnel-list">Tunnel List:</div>
        <TunnelList />
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
import TagFormUpdate from "../components/Tags/TagFormUpdate.vue";
import TunnelList from "../components/Tunnels/TunnelList.vue";
import DeviceDelete from "../components/Devices/DeviceDelete.vue";
import DeviceRename from "../components/Devices/DeviceRename.vue";
import { INotificationsError } from "../interfaces/INotifications";
import TerminalDialog from "../components/Terminal/TerminalDialog.vue";
import { formatFullDateTime } from "@/utils/date";
import handleError from "@/utils/handleError";
import { envVariables } from "@/envVariables";
import TunnelCreate from "@/components/Tunnels/TunnelCreate.vue";

const store = useStore();
const route = useRoute();
const deviceId = computed(() => route.params.id);
const device = computed(() => store.getters["devices/get"]);

onMounted(async () => {
  try {
    await store.dispatch("devices/get", deviceId.value);
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.deviceDetails,
    );
    handleError(error);
  }
});

const deviceIsEmpty = computed(
  () => store.getters["devices/get"]
        && Object.keys(store.getters["devices/get"]).length === 0,
);

const getTunnels = async () => {
  await store.dispatch("tunnels/get", deviceId.value);
};

const refreshUsers = async () => {
  try {
    await store.dispatch("devices/get", deviceId.value);
    if (envVariables.isEnterprise) {
      await store.dispatch("tunnels/get", deviceId.value);
    }
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.deviceDetails,
    );
    handleError(error);
  }
};

const receiveName = (params: string) => {
  device.value.name = params;
};

</script>
