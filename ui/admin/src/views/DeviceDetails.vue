<template>
  <h1>Device Details</h1>
  <v-card
    v-if="device.uid"
    class="mt-2 border rounded bg-background"
  >
    <v-card-title class="pa-4 d-flex align-center justify-space-between bg-v-theme-surface">
      <div class="d-flex align-center ml-2 ga-3">
        <v-tooltip
          location="bottom"
          :text="device.online ? 'Online' : 'Offline'"
        >
          <template #activator="{ props }">
            <v-icon
              v-bind="props"
              :color="device.online ? 'success' : '#E53935'"
              data-test="online-icon"
              :icon="device.online ? 'mdi-check-circle' : 'mdi-close-circle'"
            />
          </template>
        </v-tooltip>
        <h2 class="text-h6">{{ device.name }}</h2>
        <v-chip
          size="small"
          data-test="device-status-chip"
          class="text-capitalize"
          :text="device.status"
        />
      </div>
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <v-row class="py-3">
        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div data-test="device-uid-field">
            <h3 class="item-title">UID:</h3>
            <p class="text-truncate">{{ device.uid }}</p>
          </div>

          <div
            v-if="device.identity"
            data-test="device-mac-field"
          >
            <h3 class="item-title">MAC:</h3>
            <code>{{ device.identity.mac }}</code>
          </div>

          <div
            v-if="device.info"
            data-test="device-pretty-name-field"
          >
            <h3 class="item-title">Operating System:</h3>
            <div>
              <DeviceIcon
                :icon="device.info.id"
                class="mr-2"
              /><span>{{ device.info.pretty_name }}</span>
            </div>
          </div>

          <div
            v-if="device.info"
            data-test="device-version-field"
          >
            <h3 class="item-title">Agent Version:</h3>
            <p>{{ device.info.version }}</p>
          </div>

          <div
            v-if="device.info"
            data-test="device-architecture-field"
          >
            <h3 class="item-title">Architecture:</h3>
            <p>{{ device.info.arch }}</p>
          </div>

          <div
            v-if="device.info"
            data-test="device-platform-field"
          >
            <h3 class="item-title">Platform:</h3>
            <p>{{ device.info.platform }}</p>
          </div>
        </v-col>

        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div data-test="device-namespace-field">
            <h3 class="item-title">Namespace:</h3>
            <router-link
              :to="{ name: 'namespaceDetails', params: { id: device.tenant_id } }"
              class="hyper-link"
            >
              {{ device.namespace }}
            </router-link>
          </div>

          <div data-test="device-tenant-id-field">
            <h3 class="item-title">Tenant ID:</h3>
            <p class="text-truncate">{{ device.tenant_id }}</p>
          </div>

          <div
            v-if="device.remote_addr"
            data-test="device-remote-addr-field"
          >
            <h3 class="item-title">Remote Address:</h3>
            <p>{{ device.remote_addr }}</p>
          </div>

          <div
            v-if="device.tags?.length"
            data-test="device-tags-field"
          >
            <h3 class="item-title">Tags:</h3>
            <div v-if="device.tags">
              <v-tooltip
                v-for="(tag, index) in device.tags"
                :key="index"
                bottom
                :disabled="!showTag(tag.name)"
                :text="tag.name"
              >
                <template #activator="{ props }">
                  <v-chip
                    size="small"
                    v-bind="props"
                    class="mr-2"
                  >
                    {{ displayOnlyTenCharacters(tag.name) }}
                  </v-chip>
                </template>
              </v-tooltip>
            </div>
          </div>

          <div
            v-if="device.created_at"
            data-test="device-created-at-field"
          >
            <h3 class="item-title">Created At:</h3>
            <p>{{ formatFullDateTime(device.created_at) }}</p>
          </div>

          <div data-test="device-last-seen-field">
            <h3 class="item-title">Last Seen:</h3>
            <p>{{ formatFullDateTime(device.last_seen) }}</p>
          </div>
        </v-col>
      </v-row>

      <v-divider class="mt-4" />

      <v-row class="py-3">
        <v-col cols="12">
          <div
            v-if="device.public_key"
            data-test="device-public-key-field"
          >
            <h3 class="item-title">Public Key:</h3>
            <code class="text-break">{{ device.public_key }}</code>
          </div>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
  <v-card
    v-else
    class="mt-2 pa-4 bg-v-theme-surface"
  >
    <p class="text-center">
      Something is wrong, try again!
    </p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import useDevicesStore from "@admin/store/modules/devices";
import useSnackbar from "@/helpers/snackbar";
import { displayOnlyTenCharacters } from "@/utils/string";
import showTag from "@/utils/tag";
import { formatFullDateTime } from "@/utils/date";
import DeviceIcon from "@/components/Devices/DeviceIcon.vue";

const route = useRoute();
const snackbar = useSnackbar();
const devicesStore = useDevicesStore();
const deviceId = computed(() => route.params.id);
const device = computed(() => devicesStore.device);

onMounted(async () => {
  try {
    await devicesStore.fetchDeviceById(deviceId.value as string);
  } catch {
    snackbar.showError("Failed to get device details.");
  }
});
</script>

<style scoped>
.hyper-link {
  color: inherit;
  text-decoration: underline;
}

.hyper-link:visited,
.hyper-link:hover,
.hyper-link:active {
  color: inherit;
}
</style>
