<template>
  <v-tooltip location="bottom" :disabled="disableTooltip" v-bind="$attrs">
    <template v-slot:activator="{ props }">
      <div v-bind="props">
        <slot :loading :disabled :openDialog />
      </div>
    </template>
    <span>{{ tooltipMessage }}</span>
  </v-tooltip>

  <PlayerDialog v-model="showDialog" :logs />
</template>

<script setup lang="ts">
import {
  computed,
  ref,
} from "vue";
import hasPermission from "@/utils/permission";
import { envVariables } from "@/envVariables";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useSessionsStore from "@/store/modules/sessions";
import useUsersStore from "@/store/modules/users";
import PlayerDialog from "./PlayerDialog.vue";

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
</script>
