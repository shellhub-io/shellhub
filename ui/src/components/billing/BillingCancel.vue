<template>
  <fragment>
    <v-tooltip
      bottom
      :disabled="hasAuthorization"
    >
      <template #activator="{ on }">
        <div v-on="on">
          <v-btn
            :disabled="!hasAuthorization"
            color="red darken-1"
            outlined
            data-test="cancel-btn"
            @click="dialog = !dialog"
          >
            Cancel
          </v-btn>
        </div>
      </template>

      <span>
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="510"
    >
      <v-card data-test="billingCancel-dialog">
        <v-card-title class="headline primary text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          Canceling the subscription will generate an invoice,
          estimated <b> {{ nextPaymentDue | formatCurrency }} </b> for the time of use.
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            data-test="close-btn"
            @click="dialog=!dialog"
          >
            Close
          </v-btn>

          <v-btn
            text
            data-test="cancelDialog-btn"
            @click="cancelSubscription()"
          >
            Cancel
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import formatCurrency from '@/components/filter/currency';
import hasPermission from '@/components/filter/permission';

export default {
  name: 'BillingCancelComponent',

  filters: {
    formatCurrency,
    hasPermission,
  },

  props: {
    nextPaymentDue: {
      type: Number,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
      action: 'unsubscribe',
    };
  },

  computed: {
    hasAuthorization() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.billing[this.action],
        );
      }

      return false;
    },
  },

  methods: {
    async cancelSubscription() {
      try {
        await this.$store.dispatch('billing/cancelSubscription');
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.cancelSubscription);

        this.$emit('update');
        this.$store.dispatch('devices/setDeviceChooserStatus', this.$store.getters['stats/stats'].registered_devices > 3);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.cancelSubscription);
      }
    },
  },
};

</script>
