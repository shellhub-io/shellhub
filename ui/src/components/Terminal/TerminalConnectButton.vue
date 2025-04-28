<template>
  <v-btn
    :disabled="!online"
    :color="online ? 'success' : 'normal'"
    variant="outlined"
    density="comfortable"
    data-test="connect-btn"
    @click="showDialog = true"
  >
    {{ online ? "Connect" : "Offline" }}
  </v-btn>

  <TerminalDialog
    v-model="showDialog"
    :deviceUid
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import TerminalDialog from "./TerminalDialog.vue";

defineOptions({
  inheritAttrs: false,
});

defineProps<{
  online: boolean;
  deviceUid: string;
}>();

const showDialog = ref(false);

defineExpose({ showDialog });
</script>

<style scoped lang="scss">
button {
  width: min(100%, 120px);
}
</style>
