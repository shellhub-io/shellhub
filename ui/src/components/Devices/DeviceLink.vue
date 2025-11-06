<template>
  <v-btn
    class="text-none border rounded bg-v-theme-background"
    prepend-icon="mdi-developer-board"
    density="default"
    size="default"
    @click="redirectToDeviceDetails()"
  >
    {{ deviceName }}
  </v-btn>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";

const { deviceUid, deviceName } = defineProps<{
  deviceUid: string;
  deviceName: string;
}>();

const router = useRouter();

const redirectToDeviceDetails = async () => {
  await router.push({ name: "DeviceDetails", params: { identifier: deviceUid } });
};
</script>
