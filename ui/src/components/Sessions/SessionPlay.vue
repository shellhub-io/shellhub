<template>
  <div>
    <v-tooltip location="bottom" :disabled="disableTooltip">
      <template v-slot:activator="{ props }">
        <div v-bind="props">
          <v-btn
            color="primary"
            prepend-icon="mdi-play"
            variant="outlined"
            density="comfortable"
            data-test="connect-btn"
            @click="openDialog"
            :disabled="!isCommunity && disabled"
          >
            Play
          </v-btn>
        </div>
      </template>
      <span>{{ tooltipMessage }}</span>
    </v-tooltip>

    <v-dialog
      :transition="false"
      :fullscreen="true"
      v-model="showPlayer"
    >
      <v-card class="bg-v-theme-surface position-relative">
        <v-btn
          class="position-absolute top-0 right-0 ma-2 close-btn"
          variant="text"
          data-test="close-btn"
          icon="mdi-close"
          @click="closePlayerDialog"
        />

        <Player :logs @close="closePlayerDialog" />
      </v-card>
    </v-dialog>

    <SessionDownload
      v-model="showDownloadDialog"
      :sessionBlob
      @play="handleManualUpload"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import { envVariables } from "@/envVariables";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import Player from "./Player.vue";
import useSnackbar from "@/helpers/snackbar";
import SessionDownload from "./SessionDownload.vue";

const props = defineProps<{
  uid: string;
  recorded: boolean;
  authenticated: boolean;
}>();

const showPlayer = ref(false);
const showDownloadDialog = ref(false);
const store = useStore();
const snackbar = useSnackbar();
const disabled = computed(() => !props.recorded || !props.authenticated);
const logs = ref<string | null>(null);
const sessionBlob = ref<Blob | null>(null);
const maxBlobSize = 300 * 1000 * 1000; // 300 MB, change this to test with smaller blobs
const isCommunity = computed(() => envVariables.isCommunity);
const tooltipMessage = computed(() => props.recorded
  ? "You don't have permission to play this session."
  : "This session was not recorded.");

const hasAuthorizationToPlay = () => {
  const role = store.getters["auth/role"];
  return !!role && hasPermission(authorizer.role[role], actions.session.play);
};

const disableTooltip = computed(() => isCommunity.value || (hasAuthorizationToPlay() && props.recorded));

const getSessionLogs = async () => {
  if (props.recorded) {
    await store.dispatch("sessions/getSessionLogs", props.uid);
    const blob = store.getters["sessions/getLogs"];

    // Check if the browser supports showSaveFilePicker (Firefox does not support it yet)
    const hasShowSaveFilePicker = "showSaveFilePicker" in window;

    if (blob.size > maxBlobSize && hasShowSaveFilePicker) {
      sessionBlob.value = blob;
      showDownloadDialog.value = true;
    } else logs.value = await blob.text();
  }
};

const displayDialog = async () => {
  try {
    await getSessionLogs();
    if (logs.value) showPlayer.value = true;
  } catch (error: unknown) {
    snackbar.showError("Failed to play the session.");
    handleError(error);
  }
};

const handleManualUpload = (text: string) => {
  logs.value = text;
  showPlayer.value = true;
};

const openDialog = () => {
  if (envVariables.isCommunity) {
    store.commit("users/setShowPaywall", true);
    return;
  }
  displayDialog();
};

const closePlayerDialog = () => {
  showPlayer.value = false;
};
</script>

<style lang="scss" scoped>
.close-btn {
  z-index: 999;
}
</style>
