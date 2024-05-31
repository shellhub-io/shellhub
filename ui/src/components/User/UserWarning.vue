<template>
  <DeviceChooser
    v-if="isBillingEnabled && hasWarning"
    data-test="deviceChooser-component"
  />

  <Welcome
    v-model:show="show"
    @update="show = false"
    data-test="welcome-component"
  />

  <NamespaceInstructions
    v-model:show="showInstructions"
    @update="showInstructions = false"
    data-test="namespaceInstructions-component"
  />

  <BillingWarning
    v-if="isBillingEnabled"
    data-test="billingWarning-component"
  />

  <AnnouncementsModal
    :show="showAnnouncements"
    :announcement="announcement"
    @update="showAnnouncements = false"
    data-test="announcementsModal-component"
  />

  <DeviceAcceptWarning
    v-model:show="showDeviceWarning"
    @update="showDeviceWarning = false"
    data-test="DeviceAcceptWarning-component"

  />

  <RecoveryHelper
    v-model="showRecoverHelper"
    data-test="RecoveryHelper-component"
  />

  <MfaForceRecoveryMail
    v-model="showForceRecoveryMail"
    data-test="MfaForceRecoveryMail-component"
  />
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import Welcome from "../Welcome/Welcome.vue";
import NamespaceInstructions from "../Namespace/NamespaceInstructions.vue";
import { INotificationsError } from "../../interfaces/INotifications";
import { useStore } from "../../store";
import { envVariables } from "../../envVariables";
import BillingWarning from "../Billing/BillingWarning.vue";
import DeviceChooser from "../Devices/DeviceChooser.vue";
import AnnouncementsModal from "../Announcements/AnnouncementsModal.vue";
import handleError from "@/utils/handleError";
import DeviceAcceptWarning from "../Devices/DeviceAcceptWarning.vue";
import RecoveryHelper from "../AuthMFA/RecoveryHelper.vue";
import MfaForceRecoveryMail from "../AuthMFA/MfaForceRecoveryMail.vue";

const store = useStore();
const router = useRouter();
const showInstructions = ref(false);
const show = ref<boolean>(false);
const showAnnouncements = ref<boolean>(false);
const showDeviceWarning = computed(() => store.getters["users/deviceDuplicationError"]);
const showRecoverHelper = computed(() => store.getters["auth/showRecoveryModal"]);
const showForceRecoveryMail = computed(() => store.getters["auth/showForceRecoveryMail"]);
const hasNamespaces = computed(
  () => store.getters["namespaces/getNumberNamespaces"] !== 0,
);
const hasWarning = computed(
  () => store.getters["devices/getDeviceChooserStatus"],
);
const stats = computed(() => store.getters["stats/stats"]);
const announcements = computed(() => store.getters["announcement/list"]);
const announcement = computed(() => store.getters["announcement/get"]);

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

const ShowAnnouncementsCheck = async () => {
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
    store.dispatch(
      "snackbar/showSnackbarErrorLoading",
      INotificationsError.namespaceList,
    );
    handleError(error);
  }
};

onMounted(() => {
  showDialogs();
  ShowAnnouncementsCheck();

  if (showRecoverHelper.value === true) {
    router.push("/settings");
  }
});
</script>
