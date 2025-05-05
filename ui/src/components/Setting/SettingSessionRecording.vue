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
import { computed, onMounted, ref, watch } from "vue";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import { useStore } from "@/store";
import { INotificationsSuccess } from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";

const props = defineProps({
  hasTenant: {
    type: Boolean,
    default: false,
  },
});
const store = useStore();

const sessionRecordingStatus = ref(store.getters["sessionRecording/getStatus"]);

watch(sessionRecordingStatus, async (value: boolean) => {
  const data = {
    id: localStorage.getItem("tenant"),
    status: value,
  };
  try {
    await store.dispatch("sessionRecording/setStatus", data);
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.namespaceEdit,
    );
  } catch (error: unknown) {
    store.dispatch("snackbar/showSnackbarErrorDefault");
    handleError(error);
  }
});

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.namespace.enableSessionRecord,
    );
  }
  return false;
});

onMounted(async () => {
  try {
    if (props.hasTenant) {
      await store.dispatch("sessionRecording/getStatus");
    }
  } catch (error: unknown) {
    store.dispatch("snackbar/showSnackbarErrorDefault");
    handleError(error);
  }
});
defineExpose({ sessionRecordingStatus });
</script>
