<template>
  <h1>Device Details</h1>
  <v-card class="mt-2 pa-4 bg-background border">
    <v-card-text v-if="!isDeviceEmpty">
      <div>
        <h3 class="text-overline">UID:</h3>
        <p :data-test="device.uid">{{ device.uid }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Name:</h3>
        <p :data-test="device.name">{{ device.name }}</p>
      </div>

      <div v-if="device.identity">
        <h3 class="text-overline mt-3">MAC Address:</h3>
        <p :data-test="device.identity.mac">{{ device.identity.mac }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Info:</h3>
        <ul v-for="(value, name, index) in device.info" :key="index" :data-test="device.info.id">
          <li class="ml-8">
            <span class="font-weight-bold mr-1">{{ name }}:</span>
            <span>{{ value }}</span>
          </li>
        </ul>
      </div>

      <div>
        <h3 class="text-overline mt-3">Public Key:</h3>
        <p :data-test="device.public_key">{{ device.public_key }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Tenant ID:</h3>
        <p :data-test="device.tenant_id">{{ device.tenant_id }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Online:</h3>
        <p :data-test="device.online">{{ device.online }}</p>
      </div>

      <div v-if="device.tags?.length">
        <h3 class="text-overline mt-3">Tags:</h3>
        <div data-test="device-tags">
          <v-tooltip
            v-for="(tag, index) in device.tags"
            :key="index"
            bottom
            :disabled="!showTag(tag.name)"
          >
            <template #activator="{ props }">
              <v-chip size="small" v-bind="props">
                {{ displayOnlyTenCharacters(tag.name) }}
              </v-chip>
            </template>

            <span>
              {{ tag.name }}
            </span>
          </v-tooltip>
        </div>
      </div>

      <div>
        <h3 class="text-overline mt-3">Namespace:</h3>
        <p :data-test="device.namespace">{{ device.namespace }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Status:</h3>
        <p :data-test="device.status" class="text-capitalize">{{ device.status }}</p>
      </div>
    </v-card-text>
    <p v-else class="text-center">Something is wrong, try again!</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import useDevicesStore from "@admin/store/modules/devices";
import useSnackbar from "@/helpers/snackbar";
import { IAdminDevice } from "../interfaces/IDevice";
import { displayOnlyTenCharacters } from "@/utils/string";
import showTag from "@/utils/tag";

const route = useRoute();
const snackbar = useSnackbar();
const devicesStore = useDevicesStore();
const deviceId = computed(() => route.params.id);
const device = ref({} as IAdminDevice);
const isDeviceEmpty = computed(() => !device.value || Object.keys(device.value).length === 0);

onMounted(async () => {
  try {
    device.value = await devicesStore.fetchDeviceById(deviceId.value as string);
  } catch { snackbar.showError("Failed to get device details."); }
});

defineExpose({ device });
</script>
