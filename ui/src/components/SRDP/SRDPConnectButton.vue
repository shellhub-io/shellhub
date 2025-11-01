<template>
  <v-tooltip text="Connect to Remote Desktop">
    <template v-slot:activator="{ props }">
      <v-btn
        :disabled="!online"
        :color="online ? 'primary' : 'normal'"
        icon="mdi-monitor-screenshot"
        variant="outlined"
        density="compact"
        data-test="SRDP-connect-btn"
        v-bind="props"
        @click="openSRDP"
      />
    </template>
  </v-tooltip>
  
  <SRDPDialog
    v-model="showSRDP"
    :device-uid
    :device-name
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import SRDPDialog from "./SRDPDialog.vue";

defineOptions({
  inheritAttrs: false,
});

defineProps<{
  online: boolean;
  deviceUid: string;
  deviceName: string;
}>();

const showSRDP = ref(false);

const openSRDP = () => {
  showSRDP.value = true;
};

defineExpose({ showSRDP });
</script>
