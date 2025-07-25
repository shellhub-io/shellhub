<template>
  <BaseDialog
    v-if="hasAuthorization"
    v-model="showDialog"
    transition="dialog-bottom-transition"
    data-test="device-accept-warning-dialog"
  >
    <v-card class="bg-v-theme-surface" data-test="card-dialog">
      <v-card-title class="pa-3 bg-primary" data-test="card-title">
        You already have a device using the same name
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1">
        <p class="mb-2" data-test="card-text">
          <strong>{{ device }} </strong> name is already taken by another accepted device,
          please choose another name.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="close()"> Close </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { actions, authorizer } from "@/authorizer";
import hasPermission from "@/utils/permission";
import { useStore } from "@/store";
import BaseDialog from "../BaseDialog.vue";

const store = useStore();
const device = computed(() => store.getters["devices/getDeviceToBeRenamed"]);
const showDialog = computed(() => store.getters["users/deviceDuplicationError"]);

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  return !!role && hasPermission(authorizer.role[role], actions.billing.subscribe);
});

const close = () => { store.dispatch("users/setDeviceDuplicationOnAcceptance", false); };
</script>
