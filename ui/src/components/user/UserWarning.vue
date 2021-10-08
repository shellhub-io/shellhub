<template>
  <fragment>
    <DeviceWarning data-test="deviceWarning-component" />

    <Welcome
      :show.sync="show"
      data-test="welcome-component"
    />

    <NamespaceInstructions
      :show.sync="showInstructions"
      data-test="namespaceInstructions-component"
    />

    <BillingWarning data-test="billingWarning-component" />
  </fragment>
</template>

<script>

import DeviceWarning from '@/components/device/DeviceWarning';
import Welcome from '@/components/welcome/Welcome';
import NamespaceInstructions from '@/components/app_bar/namespace/NamespaceInstructions';
import BillingWarning from '@/components/billing/BillingWarning';

export default {
  name: 'AppLayout',

  components: {
    DeviceWarning,
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

    stats() {
      return this.$store.getters['stats/stats'];
    },
  },

  created() {
    this.showDialogs();
  },

  methods: {
    async showDialogs() {
      try {
        await this.getNamespaces();

        if (this.hasNamespaces) {
          await this.$store.dispatch('stats/get');

          this.showScreenWelcome();

          this.$store.dispatch('devices/setDeviceWarning',
            this.$store.getters['stats/stats'].registered_devices > 3
            && !this.$store.getters['billing/active']);
        } else {
          // This shows the namespace instructions when the user has no namespace
          this.showInstructions = true;
        }
      } catch (error) {
        switch (true) {
        case (error.response.status === 403): {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
          break;
        }
        default: {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.dashboard);
        }
        }
      }
    },

    async getNamespaces() {
      try {
        await this.$store.dispatch('namespaces/fetch');
      } catch (error) {
        switch (true) {
        case (!this.inANamespace && error.response.status === 403): { // dialog pops
          break;
        }
        case (error.response.status === 403): {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
          break;
        }
        default: {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.namespaceList);
        }
        }
      }
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
