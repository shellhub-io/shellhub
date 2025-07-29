<template>
  <DeviceChooser
    v-if="hasWarning"
    data-test="device-chooser-component"
  />

  <Welcome
    v-model="showWelcome"
    data-test="welcome-component"
  />

  <NamespaceInstructions
    v-model="showInstructions"
    data-test="namespace-instructions-component"
  />

  <BillingWarning
    data-test="billing-warning-component"
  />

  <AnnouncementsModal
    v-model="showAnnouncements"
    :announcement="currentAnnouncement"
    data-test="announcements-modal-component"
  />

  <DeviceAcceptWarning
    v-model:show="showDeviceWarning"
    @update="showDeviceWarning = false"
    data-test="device-accept-warning-component"
  />

  <RecoveryHelper
    v-model="showRecoverHelper"
    data-test="recovery-helper-component"
  />

  <MfaForceRecoveryMail
    v-model="showForceRecoveryMail"
    data-test="mfa-force-recovery-mail-component"
  />

  <PaywallDialog
    v-model="showPaywall"
    data-test="paywall-dialog-component"
  />
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import Welcome from "../Welcome/Welcome.vue";
import NamespaceInstructions from "../Namespace/NamespaceInstructions.vue";
import { useStore } from "@/store";
import { envVariables } from "@/envVariables";
import BillingWarning from "../Billing/BillingWarning.vue";
import DeviceChooser from "../Devices/DeviceChooser.vue";
import AnnouncementsModal from "../Announcements/AnnouncementsModal.vue";
import handleError from "@/utils/handleError";
import DeviceAcceptWarning from "../Devices/DeviceAcceptWarning.vue";
import RecoveryHelper from "../AuthMFA/RecoveryHelper.vue";
import MfaForceRecoveryMail from "../AuthMFA/MfaForceRecoveryMail.vue";
import PaywallDialog from "./PaywallDialog.vue";
import useSnackbar from "@/helpers/snackbar";
import useAnnouncementStore from "@/store/modules/announcement";

defineOptions({
  inheritAttrs: false,
});

const snackbar = useSnackbar();
const store = useStore();
const announcementStore = useAnnouncementStore();
const router = useRouter();
const showInstructions = ref(false);
const showWelcome = ref<boolean>(false);
const showAnnouncements = ref<boolean>(false);
const showDeviceWarning = computed(() => store.getters["users/deviceDuplicationError"]);
const showRecoverHelper = computed(() => store.getters["auth/showRecoveryModal"]);
const showForceRecoveryMail = computed(() => store.getters["auth/showForceRecoveryMail"]);
const showPaywall = computed(() => store.getters["users/showPaywall"]);
const stats = computed(() => store.getters["stats/stats"]);
const currentAnnouncement = computed(() => announcementStore.currentAnnouncement);
const hasNamespaces = computed(
  () => store.getters["namespaces/getNumberNamespaces"] !== 0,
);
const hasWarning = computed(
  () => store.getters["devices/getDeviceChooserStatus"],
);
const statusWarning = async () => {
  const bill = store.getters["namespaces/get"].billing;

  if (bill === undefined) {
    await store.dispatch("namespaces/get", localStorage.getItem("tenant"));
  }

  return (
    store.getters["stats/stats"].registered_devices > 3
        && !store.getters["billing/active"]
  );
};

const billingWarning = async () => {
  const status = await statusWarning();
  await store.dispatch("devices/setDeviceChooserStatus", status);
};

const namespaceHasBeenShown = (tenant: string) => (
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  JSON.parse(localStorage.getItem("namespacesWelcome"))[tenant]
        !== undefined
);

const hasDevices = computed(() => (
  stats.value.registered_devices !== 0
        || stats.value.pending_devices !== 0
        || stats.value.rejected_devices !== 0
));

const showScreenWelcome = async () => {
  let status = false;

  const tenantID = await store.getters["namespaces/get"].tenant_id;
  if (!namespaceHasBeenShown(tenantID) && !hasDevices.value) {
    store.dispatch("auth/setShowWelcomeScreen", tenantID);
    status = true;
  }

  showWelcome.value = status;
};

const checkForNewAnnouncements = async () => {
  if (!envVariables.announcementsEnable) return;

  try {
    const announcements = await announcementStore.fetchAnnouncements();

    if (announcements.length > 0) {
      const latestAnnouncement = announcements[0];
      await announcementStore.fetchById(latestAnnouncement.uuid);

      const storedAnnouncementHash = localStorage.getItem("announcement");
      const currentAnnouncementHash = btoa(JSON.stringify(currentAnnouncement));

      if (storedAnnouncementHash !== currentAnnouncementHash) {
        showAnnouncements.value = true;
      }
    }
  } catch (error: unknown) {
    handleError(error);
  }
};

const showDialogs = async () => {
  try {
    if (!store.getters["auth/isLoggedIn"]) return;

    await store.dispatch("namespaces/fetch", {
      page: 1,
      perPage: 30,
    });

    if (hasNamespaces.value) {
      await store.dispatch("stats/get");

      showScreenWelcome();

      if (envVariables.isCloud && !store.getters["billing/active"]) await billingWarning();
    } else showInstructions.value = true;
  } catch (error: unknown) {
    snackbar.showError("An error occurred while fetching the namespaces.");
    handleError(error);
  }
};

onMounted(async () => {
  await showDialogs();
  await checkForNewAnnouncements();

  if (showRecoverHelper.value === true) router.push("/settings");
});
</script>
