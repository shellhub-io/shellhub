<template>
  <DeviceChooser
    v-if="showDeviceChooser"
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
    v-if="showRecoverHelper"
    v-model:showDialog="showRecoverHelper"
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
import useAuthStore from "@/store/modules/auth";
import useBillingStore from "@/store/modules/billing";
import useDevicesStore from "@/store/modules/devices";
import useNamespacesStore from "@/store/modules/namespaces";
import useStatsStore from "@/store/modules/stats";

defineOptions({
  inheritAttrs: false,
});

const snackbar = useSnackbar();
const store = useStore();
const announcementStore = useAnnouncementStore();
const authStore = useAuthStore();
const billingStore = useBillingStore();
const devicesStore = useDevicesStore();
const namespacesStore = useNamespacesStore();
const statsStore = useStatsStore();
const router = useRouter();
const showInstructions = ref(false);
const showWelcome = ref<boolean>(false);
const showAnnouncements = ref<boolean>(false);
const showDeviceWarning = computed(() => store.getters["users/deviceDuplicationError"]);
const showRecoverHelper = computed(() => authStore.showRecoveryModal);
const showForceRecoveryMail = computed(() => authStore.showForceRecoveryMail);
const showPaywall = computed(() => store.getters["users/showPaywall"]);
const stats = computed(() => statsStore.stats);
const currentAnnouncement = computed(() => announcementStore.currentAnnouncement);
const hasNamespaces = computed(() => namespacesStore.namespaceList.length !== 0);
const showDeviceChooser = computed(() => devicesStore.showDeviceChooser);

const billingWarning = async () => {
  await billingStore.getSubscriptionInfo();
  const showDeviceChooser = stats.value.registered_devices > 3 && !billingStore.isActive;
  devicesStore.showDeviceChooser = showDeviceChooser;
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

  const tenantID = namespacesStore.currentNamespace.tenant_id;
  if (!namespaceHasBeenShown(tenantID) && !hasDevices.value) {
    authStore.setShowWelcomeScreen(tenantID);
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
    if (!authStore.isLoggedIn) return;

    await namespacesStore.fetchNamespaceList({ perPage: 30 });

    if (hasNamespaces.value) {
      await statsStore.fetchStats();

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
