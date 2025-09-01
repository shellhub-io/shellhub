<template>
  <BaseDialog
    v-if="hasAuthorization"
    v-model="showDialog"
    @close="close"
    transition="dialog-bottom-transition"
    data-test="device-accept-warning-dialog"
  >
    <v-card class="bg-v-theme-surface" data-test="card-dialog">
      <v-card-title class="pa-3 bg-primary" data-test="card-title">
        You already have a device using the same name
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1">
        <p class="mb-2" data-test="card-text">
          <strong>{{ duplicatedDeviceName }} </strong> name is already taken by another accepted device,
          please choose another name.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="close"> Close </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { actions, authorizer } from "@/authorizer";
import hasPermission from "@/utils/permission";
import BaseDialog from "../BaseDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useDevicesStore from "@/store/modules/devices";

const authStore = useAuthStore();
const devicesStore = useDevicesStore();
const duplicatedDeviceName = computed(() => devicesStore.duplicatedDeviceName);
const showDialog = computed(() => !!duplicatedDeviceName.value);

const hasAuthorization = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.billing.subscribe);
});

const close = () => { devicesStore.duplicatedDeviceName = ""; };
</script>
