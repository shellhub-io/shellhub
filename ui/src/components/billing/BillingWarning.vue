<template>
  <fragment>
    <v-dialog
      v-model="dialog"
      max-width="510"
    >
      <v-card data-test="billingWarning-dialog">
        <v-card-title class="headline grey lighten-2 text-center">
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
            @click="dialog=!dialog"
          >
            Close
          </v-btn>

          <v-btn
            to="/settings/billing"
            text
            data-test="goToBilling-btn"
            @click="dialog=!dialog"
          >
            Go to Billing
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'BillingWarningComponent',

  data() {
    return {
      dialog: true,
    };
  },

  computed: {
    active() {
      return this.$store.getters['billing/active'];
    },

    stats() {
      return this.$store.getters['stats/stats'];
    },
  },

  created() {
    this.dialog = this.stats.registered_devices === 3 && !this.active;
  },
};

</script>
