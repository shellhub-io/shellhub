<template>
  <DeviceChooser
    v-if="isBillingEnabled && hasWarning"
    data-test="device-chooser-component"
  />

  <Welcome
    v-model:show="show"
    @update="show = false"
    data-test="welcome-component"
  />

  <NamespaceInstructions
    v-model="showInstructions"
    data-test="namespace-instructions-component"
  />

  <BillingWarning
    v-if="isBillingEnabled"
    data-test="billing-warning-component"
  />

  <AnnouncementsModal
    v-model="showAnnouncements"
    :announcement="announcement"
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

defineOptions({
  inheritAttrs: false,
});

const snackbar = useSnackbar();
const store = useStore();
const router = useRouter();
const showInstructions = ref(false);
const show = ref<boolean>(false);
const showAnnouncements = ref<boolean>(false);
const showDeviceWarning = computed(() => store.getters["users/deviceDuplicationError"]);
const showRecoverHelper = computed(() => store.getters["auth/showRecoveryModal"]);
const showForceRecoveryMail = computed(() => store.getters["auth/showForceRecoveryMail"]);
const showPaywall = computed(() => store.getters["users/showPaywall"]);
const stats = computed(() => store.getters["stats/stats"]);
const announcements = computed(() => store.getters["announcement/list"]);
const announcement = computed(() => store.getters["announcement/get"]);
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

  show.value = status;
};

const checkAnnouncements = async () => {
  if (!envVariables.announcementsEnable) {
    return;
  }

  try {
    await store.dispatch("announcement/getListAnnouncements", {
      page: 1,
      perPage: 1,
      orderBy: "desc",
    });

    if (announcements.value.length > 0) {
      const announcementTest = announcements.value[0];
      await store.dispatch(
        "announcement/getAnnouncement",
        announcementTest.uuid,
      );

      const announcementStorage = localStorage.getItem("announcement");
      const lastAnnouncementEncoded = btoa(JSON.stringify(announcement.value));
      if (announcementStorage !== lastAnnouncementEncoded) {
        showAnnouncements.value = true;
      }
    }
  } catch (error: unknown) {
    handleError(error);
  }
};

const isBillingEnabled = computed(() => envVariables.billingEnable);

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
      if (isBillingEnabled.value) {
        await billingWarning();
      }
    } else {
      // this shows the namespace instructions when the user has no namespace
      showInstructions.value = true;
    }
  } catch (error: unknown) {
    snackbar.showError("An error occurred while fetching the namespaces.");
    handleError(error);
  }
};

onMounted(() => {
  showDialogs();
  checkAnnouncements();

  if (showRecoverHelper.value === true) {
    router.push("/settings");
  }
});
</script>
