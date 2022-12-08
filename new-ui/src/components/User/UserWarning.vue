<template>
  <DeviceChooser
    v-if="isBillingEnabled && hasWarning"
    data-test="deviceChooser-component"
  />

  <Welcome
    :show.sync="show"
    @update="show = false"
    data-test="welcome-component"
  />

  <NamespaceInstructions
    :show.sync="showInstructions"
    @update="showInstructions = false"
    data-test="namespaceInstructions-component"
  />

  <BillingWarning
    v-if="isBillingEnabled"
    data-test="billingWarning-component"
  />
</template>

<script lang="ts">
import { defineComponent, ref, computed, onMounted } from "vue";
import Welcome from "../Welcome/Welcome.vue";
import NamespaceInstructions from "../Namespace/NamespaceInstructions.vue";
import { INotificationsError } from "../../interfaces/INotifications";
import { useStore } from "../../store";
import { envVariables } from "../../envVariables";
import BillingWarning from "../Billing/BillingWarning.vue";
import DeviceChooser from "../Devices/DeviceChooser.vue";

export default defineComponent({
  inheritAttrs: false,
  setup() {
    const store = useStore();
    const showInstructions = ref(false);
    const show = ref<boolean>(false);

    const hasNamespaces = computed(
      () => store.getters["namespaces/getNumberNamespaces"] !== 0
    );
    const hasSpinner = computed(() => store.getters["spinner/getStatus"]);
    const hasWarning = computed(
      () => store.getters["devices/getDeviceChooserStatus"]
    );
    const stats = computed(() => store.getters["stats/stats"]);

    onMounted(() => {
      showDialogs();
    });

    const statusWarning = async () => {
      const bill = store.getters["namespaces/get"].billing;

      if (bill === undefined) {
        await store.dispatch("namespaces/get", localStorage.getItem("tenant"));
      }

      return (
        store.getters["stats/stats"].registered_devices > 3 &&
        !store.getters["billing/active"]
      );
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
          if (isBillingEnabled.value) {
            await billingWarning();
          }
        } else {
          // this shows the namespace instructions when the user has no namespace
          showInstructions.value = true;
        }
      } catch {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.namespaceList
        );
      }
    };

    const isBillingEnabled = computed(() => envVariables.billingEnable);

    const billingWarning = async () => {
      const status = await statusWarning();
      await store.dispatch("devices/setDeviceChooserStatus", status);
    };

    const namespaceHasBeenShown = (tenant: string) => {
      return (
        // @ts-ignore
        JSON.parse(localStorage.getItem("namespacesWelcome"))[tenant] !==
        undefined
      );
    };

    const hasDevices = computed(() => {
      return (
        stats.value.registered_devices !== 0 ||
        stats.value.pending_devices !== 0 ||
        stats.value.rejected_devices !== 0
      );
    });

    const showScreenWelcome = async () => {
      let status = false;

      const tenantID = await store.getters["namespaces/get"].tenant_id;
      if (!namespaceHasBeenShown(tenantID) && !hasDevices.value) {
        store.dispatch("auth/setShowWelcomeScreen", tenantID);
        status = true;
      }

      show.value = status;
    };

    return {
      hasNamespaces,
      hasSpinner,
      hasDevices,
      stats,
      showInstructions,
      isBillingEnabled,
      namespaceHasBeenShown,
      showScreenWelcome,
      hasWarning,
      show,
    };
  },
  components: { Welcome, NamespaceInstructions, BillingWarning, DeviceChooser },
});
</script>
