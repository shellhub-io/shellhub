<template>
  <div class="d-flex pa-0 align-center">
    <h1>Device Details</h1>
  </div>
  <v-card class="mt-2 pa-4" v-if="!deviceIsEmpty">
    <v-card-text>
      <div>
        <div class="text-overline mt-3">
          <h3>uid:</h3>
        </div>
        <div :data-test="device.uid">
          <p>{{ device.uid }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>name:</h3>
        </div>
        <div :data-test="device.name">
          <p>{{ device.name }}</p>
        </div>
      </div>

      <div v-if="device.identity">
        <div class="text-overline mt-3">
          <h3>mac:</h3>
        </div>
        <div :data-test="device.identity.mac">
          {{ device.identity.mac }}
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>info:</h3>
        </div>
        <ul v-for="(value, name, index) in device.info" :key="index" :data-test="device.info.id">
          <li class="ml-8">
            <span class="font-weight-bold mr-1">{{ name }}:</span>
            <span>{{ value }}</span>
          </li>
        </ul>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Public Key:</h3>
        </div>
        <div :data-test="device.public_key">
          <p>{{ device.public_key }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Tenat Id:</h3>
        </div>
        <div :data-test="device.tenant_id">
          <p>{{ device.tenant_id }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Online:</h3>
        </div>
        <div :data-test="device.online">
          <p>{{ device.online }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Tags:</h3>
        </div>
        <div v-if="device.tags" :data-test="device.tags">
          <v-tooltip
            v-for="(tag, index) in device.tags"
            :key="index"
            bottom
            :disabled="!showTag(tag)"
          >
            <template #activator="{ props }">
              <v-chip size="small" v-bind="props" v-on="props">
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
        <div class="text-overline mt-3">
          <h3>Namespace:</h3>
        </div>
        <div :data-test="device.namespace">
          <p>{{ device.namespace }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Status:</h3>
        </div>
        <div :data-test="device.status">
          <p>{{ device.status }}</p>
        </div>
      </div>
    </v-card-text>
  </v-card>
  <v-card class="mt-2 pa-4" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script lang="ts">
import { computed, ref, defineComponent, onMounted } from "vue";
import { useRoute } from "vue-router";
import { IDevice } from "../interfaces/IDevice";
import { INotificationsError } from "../interfaces/INotifications";
import { useStore } from "../store";
import displayOnlyTenCharacters from "../hooks/string";
import showTag from "../hooks/tag";

export default defineComponent({
  name: "DeviceDetails",
  setup() {
    const store = useStore();
    const route = useRoute();
    const deviceId = computed(() => route.params.id);
    const device = ref({} as IDevice);

    onMounted(async () => {
      try {
        await store.dispatch("devices/get", deviceId.value);
        device.value = store.getters["devices/get"];
      } catch {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.deviceDetails);
      }
    });

    const deviceIsEmpty = computed(() => store.getters["devices/get"] && store.getters["devices/get"].lenght === 0);

    return {
      device,
      deviceIsEmpty,
      displayOnlyTenCharacters,
      showTag,
    };
  },
});
</script>
