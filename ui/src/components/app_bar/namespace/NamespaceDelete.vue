<template>
  <fragment>
    <v-btn
      color="red darken-1"
      outlined
      data-test="delete-btn"
      @click="dialog = !dialog"
    >
      Delete namespace
    </v-btn>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card data-test="namespaceDelete-dialog">
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          <p
            v-if="active && amountDue !== null"
            data-test="contentSubscription-p"
          >
            Deleting the namespace will generate an invoice,
            estimated <b> {{ amountDue | formatCurrency }} </b> for the time of use.
          </p>

          <p data-test="content-text">
            This action cannot be undone. This will permanently delete the
            <b> {{ displayOnlyTenCharacters(name) }} </b>and its related data.
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
            color="red darken-1"
            text
            data-test="remove-btn"
            @click="remove();"
          >
            Remove
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import formatCurrency from '@/components/filter/currency';

export default {
  name: 'NamespaceDelete',

  filters: { formatCurrency },

  props: {
    nsTenant: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      name: '',
      dialog: false,
      card: null,
      elements: null,
      amountDue: null,
      cardItems: {},
    };
  },

  computed: {
    tenant() {
      return this.$props.nsTenant;
    },

    active() {
      return this.$store.getters['billing/active'];
    },

    billing() {
      return this.$store.getters['billing/get'];
    },

    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },
  },

  async created() {
    this.name = this.$store.getters['namespaces/get'].name;
  },

  mounted() {
    if (this.isOwner && this.isBillingEnabled()) {
      this.getSubscriptionInfo();
    }
  },

  methods: {
    isBillingEnabled() {
      return this.$env.billingEnable;
    },

    async remove() {
      try {
        this.dialog = !this.dialog;
        await this.$store.dispatch('namespaces/remove', this.tenant);
        await this.$store.dispatch('auth/logout');
        await this.$router.push({ name: 'login' });
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceDelete);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.namespaceDelete);
      }
    },

    displayOnlyTenCharacters(str) {
      if (str !== undefined) {
        if (str.length > 10) return `${str.substr(0, 10)}...`;
      }
      return str;
    },

    getDueAmount(data) {
      return data.upcoming_invoice.amount_due;
    },

    async getSubscriptionInfo() {
      if (this.active) {
        try {
          const data = await this.$store.dispatch('billing/getSubscription');
          this.amountDue = this.getDueAmount(data);
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      }
    },
  },
};

</script>
