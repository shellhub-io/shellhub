<template>
  <fragment>
    <v-dialog
      v-if="hasAuthorization"
      v-model="showMessage"
      max-width="510"
    >
      <v-card data-test="billingWarning-dialog">
        <v-card-title class="headline primary text-center">
          Update account
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          <p>
            This namespace has maximum number of devices on your free account.
          </p>

          <p>
            If you create the subscription in your account settings, you can continue to take
            advantage of the features available on ShellHub by adding more devices.
          </p>
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            data-test="close-btn"
            @click="close()"
          >
            Close
          </v-btn>

          <v-btn
            to="/settings/billing"
            text
            data-test="goToBilling-btn"
            @click="close()"
          >
            Go to Billing
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import hasPermission from '@/components/filter/permission';

export default {
  name: 'BillingWarningComponent',

  filters: {
    hasPermission,
  },

  computed: {
    hasAuthorization() {
      const role = this.$store.getters['auth/role'];

      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.billing.subscribe,
        );
      }

      return false;
    },

    showMessage: {
      get() {
        return (this.$store.getters['users/statusUpdateAccountDialog']
          && this.$store.getters['stats/stats'].registered_devices === 3
          && !this.$store.getters['billing/active'])
          || this.$store.getters['users/statusUpdateAccountDialogByDeviceAction'];
      },

      set() {
        this.close();
      },
    },
  },

  methods: {
    close() {
      if (this.$store.getters['users/statusUpdateAccountDialog']) {
        this.$store.dispatch('users/setStatusUpdateAccountDialog', false);
      } else if (this.$store.getters['users/statusUpdateAccountDialogByDeviceAction']) {
        this.$store.dispatch('users/setStatusUpdateAccountDialogByDeviceAction', false);
      }
    },
  },
};

</script>
