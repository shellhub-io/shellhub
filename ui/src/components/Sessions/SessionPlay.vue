<template>
  <div>
    <v-btn
      color="primary"
      prepend-icon="mdi-play"
      variant="outlined"
      density="comfortable"
      data-test="connect-btn"
      @click="openDialog"
      :disabled="!isCommunity && props.disabled"
    >
      Play
    </v-btn>

    <v-dialog
      :transition="false"
      :fullscreen="true"
      v-model="showDialog"
    >
      <v-card class="bg-v-theme-surface position-relative">
        <v-btn
          class="position-absolute top-0 right-0 ma-2 close-btn"
          variant="text"
          data-test="close-btn"
          icon="mdi-close"
          @click="closeDialog"
        />

        <Player :logs @close="closeDialog" />
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  ref,
} from "vue";
import { envVariables } from "@/envVariables";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import Player from "./Player.vue";
import useSnackbar from "@/helpers/snackbar";

const props = defineProps({
  uid: {
    type: String,
    required: true,
  },
  recorded: {
    type: Boolean,
    required: true,
  },
  notHasAuthorization: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
});

const showDialog = ref(false);
const store = useStore();
const snackbar = useSnackbar();
const logs = ref<string | null>(null);
const isCommunity = computed(() => envVariables.isCommunity);

const getSessionLogs = async () => {
  if (props.recorded) {
    await store.dispatch("sessions/getSessionLogs", props.uid);
    logs.value = store.getters["sessions/getLogs"];
  }
};

const displayDialog = async () => {
  try {
    await getSessionLogs();
    showDialog.value = true;
  } catch (error: unknown) {
    snackbar.showError("Failed to play the session.");
    handleError(error);
  }
};

const openDialog = () => {
  if (envVariables.isCommunity) {
    store.commit("users/setShowPaywall", true);
    return;
  }
  displayDialog();
};

const closeDialog = () => {
  showDialog.value = false;
};
</script>

<style lang="scss" scoped>
.close-btn {
  z-index: 999;
}
</style>
