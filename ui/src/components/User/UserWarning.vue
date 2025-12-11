<template>
  <DeviceChooser
    v-if="showDeviceChooser"
    data-test="device-chooser-component"
  />

  <Welcome data-test="welcome-component" />

  <BillingWarning
    v-model="showBillingWarning"
    data-test="billing-warning-component"
  />

  <AnnouncementsModal
    v-model="showAnnouncements"
    :announcement="currentAnnouncement"
    data-test="announcements-modal-component"
  />

  <DeviceAcceptWarning
    v-if="showDuplicationWarning"
    data-test="device-accept-warning-component"
  />

  <RecoveryHelper
    v-if="showRecoverHelper"
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
import useUsersStore from "@/store/modules/users";

defineOptions({
  inheritAttrs: false,
});

const snackbar = useSnackbar();
const announcementStore = useAnnouncementStore();
const authStore = useAuthStore();
const billingStore = useBillingStore();
const devicesStore = useDevicesStore();
const namespacesStore = useNamespacesStore();
const statsStore = useStatsStore();
const usersStore = useUsersStore();
const router = useRouter();
const showAnnouncements = ref<boolean>(false);
const showDuplicationWarning = computed(() => !!devicesStore.duplicatedDeviceName);
const showRecoverHelper = computed(() => authStore.showRecoveryModal);
const showForceRecoveryMail = computed(() => authStore.showForceRecoveryMail);
const showPaywall = computed(() => usersStore.showPaywall);
const stats = computed(() => statsStore.stats);
const currentAnnouncement = computed(() => announcementStore.currentAnnouncement);
const hasNamespaces = computed(() => namespacesStore.hasNamespaces);
const showDeviceChooser = computed(() => devicesStore.showDeviceChooser);
const showBillingWarning = computed({
  get() {
    return billingStore.showBillingWarning;
  },
  set(value: boolean) {
    billingStore.showBillingWarning = value;
  },
});

const setShowDeviceChooser = async () => {
  await billingStore.getSubscriptionInfo();
  const showDeviceChooser = stats.value.registered_devices > 3 && !billingStore.isActive;
  devicesStore.showDeviceChooser = showDeviceChooser;
};

const checkForNewAnnouncements = async () => {
  if (!envVariables.isAnnouncementsEnabled) return;

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

    if (hasNamespaces.value) {
      await statsStore.fetchStats();

      if (envVariables.isCloud && !billingStore.isActive) await setShowDeviceChooser();
    }
  } catch (error: unknown) {
    snackbar.showError("An error occurred while fetching the namespaces.");
    handleError(error);
  }
};

onMounted(async () => {
  await showDialogs();
  await checkForNewAnnouncements();

  if (showRecoverHelper.value === true) await router.push("/settings");
});
</script>
