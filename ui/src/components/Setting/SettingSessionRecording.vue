<template>
  <v-switch
    hide-details
    inset
    v-model="isSessionRecordingEnabled"
    :disabled="!hasAuthorization"
    color="primary"
    data-test="session-recording-switch"
  />
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import useSessionRecordingStore from "@/store/modules/session_recording";

const { tenantId } = defineProps<{ tenantId: string }>();

const authStore = useAuthStore();
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
  set: async (isEnabled: boolean) => {
    await updateSessionRecordingStatus(isEnabled);
    return isEnabled;
  },
});

const hasAuthorization = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.namespace.enableSessionRecord);
});

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
