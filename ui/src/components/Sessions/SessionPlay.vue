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
      <v-card class="bg-v-theme-surface">
        <v-card-title
          class="text-h5 pa-3 bg-primary d-flex justify-space-between ga-4 align-center"
        >
          <v-btn
            variant="text"
            data-test="close-btn"
            icon="mdi-close"
            @click="showDialog = false"
          />
        </v-card-title>
        <div class="ma-0 pa-0 w-100 fill-height position-relative bg-v-theme-terminal">
          <Player :logs />
        </div>
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
import { INotificationsError } from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";
import Player from "./Player.vue";

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
const logs = ref<string | null>(null);
const isCommunity = computed(() => envVariables.isCommunity);

const openPlay = async () => {
  if (props.recorded) {
    await store.dispatch("sessions/getLogSession", props.uid);
    logs.value = store.getters["sessions/get"];
  }
};

const displayDialog = async () => {
  try {
    await openPlay();
    showDialog.value = true;
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorLoading",
      INotificationsError.sessionPlay,
    );
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
</script>

<style lang="scss" scoped>
.terminal {
  position: absolute;
  top: 0px;
  bottom: 0px;
  left: 0;
  right:0;
  margin-right: 0px;
}
</style>
