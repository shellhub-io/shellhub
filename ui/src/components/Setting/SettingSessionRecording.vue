<template>
  <v-switch
    v-model="isSessionRecordingEnabled"
    hide-details
    inset
    :disabled="!canUpdateSessionRecording"
    color="primary"
    data-test="session-recording-switch"
  />
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useSessionRecordingStore from "@/store/modules/session_recording";

const { tenantId } = defineProps<{ tenantId: string }>();

const snackbar = useSnackbar();
const sessionRecordingStore = useSessionRecordingStore();

const updateSessionRecordingStatus = async (isEnabled: boolean) => {
  const data = {
    id: tenantId,
    status: isEnabled,
  };
  try {
    await sessionRecordingStore.setStatus(data);
    snackbar.showSuccess(`Session recording was successfully ${isEnabled ? "enabled" : "disabled"}.`);
  } catch (error: unknown) {
    snackbar.showError("Failed to update session recording status.");
    handleError(error);
  }
};

const isSessionRecordingEnabled = computed({
  get: () => sessionRecordingStore.isEnabled,
  set: (isEnabled: boolean) => {
    void updateSessionRecordingStatus(isEnabled);
  },
});

const canUpdateSessionRecording = hasPermission("namespace:updateSessionRecording");

onMounted(async () => {
  try {
    if (tenantId) await sessionRecordingStore.getStatus();
  } catch (error: unknown) {
    snackbar.showError("Failed to fetch session recording status.");
    handleError(error);
  }
});

defineExpose({ isSessionRecordingEnabled });
</script>
