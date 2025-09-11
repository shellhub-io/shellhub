<template>
  <div>
    <v-tooltip location="bottom" :disabled="disableTooltip">
      <template v-slot:activator="{ props }">
        <div v-bind="props">
          <slot :loading :disabled :openDialog />
        </div>
      </template>
      <span>{{ tooltipMessage }}</span>
    </v-tooltip>

    <BaseDialog
      :transition="false"
      :forceFullscreen="true"
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
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  ref,
} from "vue";
import hasPermission from "@/utils/permission";
import { envVariables } from "@/envVariables";
import handleError from "@/utils/handleError";
import Player from "./Player.vue";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useSessionsStore from "@/store/modules/sessions";
import useUsersStore from "@/store/modules/users";

const props = defineProps<{
  uid: string;
  recorded: boolean;
  authenticated: boolean;
}>();

const showDialog = ref(false);
const sessionsStore = useSessionsStore();
const usersStore = useUsersStore();
const snackbar = useSnackbar();
const loading = ref(false);
const logs = ref<string | null>(null);
const { isCommunity } = envVariables;
const disabled = computed(() => !isCommunity && (!props.recorded || !props.authenticated));
const tooltipMessage = computed(() => props.recorded
  ? "You don't have permission to play this session."
  : "This session was not recorded.");

const canPlaySession = hasPermission("session:play");

const disableTooltip = computed(() => isCommunity || (canPlaySession && props.recorded));

const getSessionLogs = async () => {
  if (!props.recorded) return false;
  logs.value = await sessionsStore.getSessionLogs(props.uid);
  return typeof logs.value === "string";
};

const displayDialog = async () => {
  try {
    loading.value = true;
    const hasLogs = await getSessionLogs();
    if (hasLogs) showDialog.value = true;
    else snackbar.showError("The session logs were deleted or not recorded.");
  } catch (error: unknown) {
    snackbar.showError("Failed to play the session.");
    handleError(error);
  }

  loading.value = false;
};

const openDialog = async () => {
  if (isCommunity) {
    usersStore.showPaywall = true;
    return;
  }
  await displayDialog();
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
