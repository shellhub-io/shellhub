<template>
  <v-switch
    hide-details
    inset
    v-model="sessionRecordingStatus"
    :disabled="!hasAuthorization"
    color="primary"
    data-test="session-recording-switch"
  />
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

const props = defineProps({
  hasTenant: {
    type: Boolean,
    default: false,
  },
});

const store = useStore();
const snackbar = useSnackbar();

const updateSessionRecordingStatus = async (value: boolean) => {
  const data = {
    id: localStorage.getItem("tenant"),
    status: value,
  };
  try {
    await store.dispatch("sessionRecording/setStatus", data);
    snackbar.showSuccess(`Session recording was successfully ${value ? "enabled" : "disabled"}.`);
  } catch (error: unknown) {
    snackbar.showError("Failed to update session recording status.");
    handleError(error);
  }
};

const sessionRecordingStatus = computed({
  get: () => store.getters["sessionRecording/isEnabled"],
  set: async (value: boolean) => {
    await updateSessionRecordingStatus(value);
  },
});

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role === "") return false;

  return hasPermission(
    authorizer.role[role],
    actions.namespace.enableSessionRecord,
  );
});

onMounted(async () => {
  try {
    if (props.hasTenant) await store.dispatch("sessionRecording/getStatus");
  } catch (error: unknown) {
    snackbar.showError("Failed to fetch session recording status.");
    handleError(error);
  }
});

defineExpose({ sessionRecordingStatus });
</script>
