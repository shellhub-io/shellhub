<template>
  <fragment>
    <DeviceChooser
      v-if="isBillingEnabled() && hasWarning"
      data-test="deviceChooser-component"
    />

    <Welcome
      :show.sync="show"
      data-test="welcome-component"
    />

    <NamespaceInstructions
      :show.sync="showInstructions"
      data-test="namespaceInstructions-component"
    />

    <BillingWarning
      v-if="isBillingEnabled()"
      data-test="billingWarning-component"
    />
  </fragment>
</template>

<script>

import DeviceChooser from '@/components/device/DeviceChooser';
import Welcome from '@/components/welcome/Welcome';
import NamespaceInstructions from '@/components/namespace/NamespaceInstructions';
import BillingWarning from '@/components/billing/BillingWarning';

export default {
  name: 'AppLayoutComponent',

  components: {
    DeviceChooser,
    Welcome,
    NamespaceInstructions,
    BillingWarning,
  },

  data() {
    return {
      show: false,
      showInstructions: false,
    };
  },

  computed: {
    hasNamespaces() {
      return this.$store.getters['namespaces/getNumberNamespaces'] !== 0;
    },

    hasSpinner() {
      return this.$store.getters['spinner/getStatus'];
    },

    hasWarning() {
      return this.$store.getters['devices/getDeviceChooserStatus'];
    },

    stats() {
      return this.$store.getters['stats/stats'];
    },
  },

  created() {
    this.showDialogs();
  },

  methods: {
    async statusWarning() {
      const bill = this.$store.getters['namespaces/get'].billing;

      if (bill === undefined) {
        await this.$store.dispatch('namespaces/get', localStorage.getItem('tenant'));
      }

      return this.$store.getters['stats/stats'].registered_devices > 3
        && !this.$store.getters['billing/active'];
    },

    async showDialogs() {
      try {
        await this.$store.dispatch('namespaces/fetch');

        if (this.hasNamespaces) {
          await this.$store.dispatch('stats/get');

          this.showScreenWelcome();
          if (this.isBillingEnabled()) {
            await this.billingWarning();
          }
        } else {
          // This shows the namespace instructions when the user has no namespace
          this.showInstructions = true;
        }
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.namespaceList);
      }
    },

    isBillingEnabled() {
      return this.$env.billingEnable;
    },

    async billingWarning() {
      const status = await this.statusWarning();
      await this.$store.dispatch('devices/setDeviceChooserStatus', status);
    },

    namespaceHasBeenShown(tenant) {
      return JSON.parse(localStorage.getItem('namespacesWelcome'))[tenant] !== undefined;
    },

    hasDevices() {
      return this.stats.registered_devices !== 0
        || this.stats.pending_devices !== 0
        || this.stats.rejected_devices !== 0;
    },

    async showScreenWelcome() {
      let status = false;

      const tenantID = await this.$store.getters['namespaces/get'].tenant_id;

      if (!this.namespaceHasBeenShown(tenantID) && !this.hasDevices()) {
        this.$store.dispatch('auth/setShowWelcomeScreen', tenantID);
        status = true;
      }

      this.show = status;
    },
  },
};

</script>
