<template>
  <fragment>
    <v-btn
      class="primary"
      data-test="show-btn"
      @click="show"
    >
      {{ actionButton(typeOperation) }}
    </v-btn>

    <v-dialog
      v-model="dialog"
      max-width="600"
      @click:outside="dialog=!dialog"
    >
      <v-card data-test="BillingDialogPaymentMethod-dialog">
        <v-card-title
          class="headline primary text-center"
          data-test="text-cardTitle"
        >
          {{ typeTitle(typeOperation) }}
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          <div v-if="typeOperation === 'subscription'">
            <div data-test="subscription-description">
              <h4>Subscribe to premium plan</h4>
              <p data-test="subscription-message">
                The subscription is charged monthly, based on the number of devices
                you have in your namespace.
              </p>
              <div>
                <b>
                  Estimated cost:
                </b>
                {{ currentQuantity }} devices : {{ priceEstimator(currentQuantity) }} / month
              </div>
            </div>
          </div>

          <v-card class="mt-6">
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
import formatCurrency from '@/components/filter/currency';

export default {
  name: 'BillingDialogPaymentMethodComponent',

  filters: { capitalizeFirstLetter, formatCurrency },

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
    currentQuantity() {
      return this.$store.getters['stats/stats'].registered_devices;
    },

    requestWaiting() {
      return this.$store.getters['spinner/getStatus'];
    },
  },

  watch: {
    dialog(status) {
      if (!status) {
        this.elementError.textContent = '';
      }
    },
  },

  methods: {
    priceEstimator(n) {
      let sumPrice = 0;

      const ranges = [3, 10, 25, 40, 55, 70, 85, 100,
        115, 130, 145, 160, 175, 190, Infinity];
      const ks = [3, 2.91, 2.82, 2.74, 2.66, 2.58, 2.50,
        2.42, 2.35, 2.28, 2.21, 2.15, 2.08, 2.02, 2.00];

      const tiers = Array.from({ length: ks.length - 1 },
        (_, i) => ({ begin: ranges[i], upTo: ranges[i + 1], k: ks[i] }));

      tiers.forEach((t) => {
        if (n > t.begin) {
          sumPrice += ((n <= t.upTo) ? (n - t.begin) : (t.upTo - t.begin)) * t.k;
        }
      });

      return formatCurrency(String(sumPrice * 100));
    },

    typeTitle(type) {
      switch (type) {
      case 'subscription':
        return 'Create subscription';
      case 'update':
        return 'Add payment method';
      default:
        return 'Operation not found';
      }
    },

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

    showError(e) {
      this.elementError.textContent = e.response.data;
    },

    actionButton(type) {
      if (type === 'subscription') {
        return 'subscribe';
      }

      if (type === 'update') {
        return 'add card';
      }

      return type;
    },

    async doAction() {
      this.lockButton = true;

      switch (this.typeOperation) {
      case 'subscription':
        await this.subscriptionPaymentMethod();
        break;
      case 'update':
        await this.updatePaymentMethod();
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
        } catch (error) {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.subscription);

          const { status } = error.response;
          if (status === 400 || status === 423) {
            this.showError(error);
          }
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
          await this.$store.dispatch('billing/addPaymentMethod', result.paymentMethod.id);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.updateSubscription);
          this.$emit('update');
          this.dialog = false;
        } catch (error) {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.updatePaymentMethod);

          const { status } = error.response;
          if (status === 400 || status === 423) {
            this.showError(error);
          }
        }
      }

      this.lockButton = false;
    },
  },
};

</script>
