<template>
  <v-dialog
    v-if="hasAuthorization"
    v-model="showMessage"
    transition="dialog-bottom-transition"
    width="650"
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
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { actions, authorizer } from "@/authorizer";
import hasPermission from "@/utils/permission";
import { useStore } from "@/store";

const store = useStore();
const device = computed(() => store.getters["devices/getDeviceToBeRenamed"]);

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.billing.subscribe,
    );
  }

  return false;
});

const close = () => {
  if (store.getters["users/deviceDuplicationError"]) {
    store.dispatch("users/setDeviceDuplicationOnAcceptance", false);
  }
};

const showMessage = computed({
  get() {
    return (
      (store.getters["users/deviceDuplicationError"])
    );
  },
  set() {
    close();
  },
});

defineExpose({ showMessage });
</script>
