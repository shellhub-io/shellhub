<template>
  <MessageDialog
    v-if="canSubscribeToBilling"
    v-model="showDialog"
    @close="close"
    @cancel="close"
    title="You already have a device using the same name"
    :description="`${duplicatedDeviceName} name is already taken by another accepted device, please choose another name.`"
    icon="mdi-alert"
    icon-color="warning"
    cancel-text="Close"
    cancel-data-test="close-btn"
    :show-footer="true"
    data-test="device-accept-warning-dialog"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import hasPermission from "@/utils/permission";
import MessageDialog from "../MessageDialog.vue";
import useDevicesStore from "@/store/modules/devices";

const devicesStore = useDevicesStore();
const duplicatedDeviceName = computed(() => devicesStore.duplicatedDeviceName);
const showDialog = computed(() => !!duplicatedDeviceName.value);

const canSubscribeToBilling = hasPermission("billing:subscribe");

const close = () => { devicesStore.duplicatedDeviceName = ""; };
</script>
