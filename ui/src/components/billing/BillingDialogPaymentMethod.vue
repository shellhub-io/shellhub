<template>
  <fragment>
    <v-btn
      outlined
      data-test="show-btn"
      @click="show"
    >
      {{ typeOperation }}
    </v-btn>

    <v-dialog
      v-model="dialog"
      max-width="600"
      @click:outside="dialog=!dialog"
    >
      <v-card data-test="BillingDialogPaymentMethod-dialog">
        <v-card-title
          class="headline grey lighten-2 text-center"
          data-test="text-cardTitle"
        >
          {{ typeOperation | capitalizeFirstLetter }} payment method
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          <div v-if="typeOperation === 'subscription'">
            Add payment method information below
          </div>

          <v-card class="mt-8">
            <v-col>
              <div ref="card" />
            </v-col>
          </v-card>

          <div
            ref="card-element-errors"
            class="card-errors mt-4"
            role="alert"
          />

          <v-spacer />

          <v-row class="mt-2">
            <v-spacer />
            <v-col
              md="auto"
              class="ml-auto"
            />
          </v-row>
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            data-test="cancel-btn"
            @click="dialog=!dialog"
          >
            Close
          </v-btn>

          <v-btn
            text
            data-test="confirm-btn"
            :disabled="lockButton"
            @click="doAction()"
          >
            confirm
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import capitalizeFirstLetter from '@/components/filter/string';

export default {
  name: 'BillingDialogPaymentMethod',

  filters: { capitalizeFirstLetter },

  props: {
    typeOperation: {
      type: String,
      default: 'subscription',
      validator: (value) => ['subscription', 'update'].includes(value),
    },
  },

  data() {
    return {
      dialog: false,
      card: null,
      elementError: null,
      elements: null,
      lockButton: false,
    };
  },

  computed: {
    requestWaiting() {
      return this.$store.getters['spinner/getStatus'];
    },
  },

  methods: {
    show() {
      this.dialog = !this.dialog;
      this.$nextTick(async () => {
        await this.mountStripeElements();
      });
    },

    displayError(e) {
      if (e.error) {
        this.elementError.textContent = e.error.message;
      } else {
        this.elementError.textContent = '';
      }
    },

    doAction() {
      this.lockButton = true;

      switch (this.typeOperation) {
      case 'subscription':
        this.subscriptionPaymentMethod();
        break;
      case 'update':
        this.updatePaymentMethod();
        break;
      default:
        this.lockButton = false;
      }
    },

    async mountStripeElements() {
      this.elements = this.$stripe.elements();
      this.card = this.elements.create('card');
      this.card.mount(this.$refs.card);
      this.elementError = this.$refs['card-element-errors'];
      this.card.on('change', (ev) => {
        this.displayError(ev);
      });
    },

    async subscriptionPaymentMethod() {
      const result = await this.$stripe.createPaymentMethod({
        type: 'card',
        card: this.card,
      });

      if (result.error) {
        this.displayError(result.error);
      } else {
        try {
          await this.$store.dispatch('billing/subscritionPaymentMethod', {
            payment_method_id: result.paymentMethod.id,
          });
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.subscription);
          this.$emit('update');
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.subscription);
        }
      }

      this.lockButton = false;
    },

    async updatePaymentMethod() {
      const result = await this.$stripe.createPaymentMethod({
        type: 'card',
        card: this.card,
      });

      if (result.error) {
        this.displayError(result.error);
      } else {
        try {
          await this.$store.dispatch('billing/updatePaymentMethod', {
            payment_id: result.paymentMethod.id,
          });
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.updateSubscription);
          this.$emit('update');
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.updateSubscription);
        }
      }

      this.lockButton = false;
    },
  },
};

</script>
