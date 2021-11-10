<template>
  <v-container>
    <v-row
      align="center"
      justify="center"
      class="mt-4"
    >
      <v-col
        sm="8"
      >
        <SettingOwnerInfo
          :is-owner="hasAuthorization"
          data-test="settingOwnerInfo-component"
        />

        <div
          v-if="hasAuthorization"
          data-test="content-div"
        >
          <v-row>
            <v-col>
              <h3>
                Subscription info
              </h3>
            </v-col>

            <v-spacer />

            <v-col
              v-if="state === 'inactive'"
              md="auto"
              class="ml-auto"
            >
              <PaymentMethod
                type-operation="subscription"
                data-test="subscriptionPaymentMethod-component"
              />
            </v-col>
          </v-row>

          <div class="mt-6 pl-4 pr-4">
            <div
              v-if="state === 'pending' && !retrialExceeded"
              data-test="pendingRetrial-div"
            >
              <p class="ma-4">
                You have a pending request, please wait a while ...
              </p>
            </div>

            <div
              v-else-if="state === 'pending' && retrialExceeded"
              data-test="pendingExceeded-div"
            >
              <p class="ma-4">
                Couldn't proccess your last request, please try again later.
              </p>
            </div>

            <div v-else-if="state === 'inactive'">
              <div data-test="freePlan-div">
                <p>
                  Plan: <b> Free </b>
                </p>

                <p>
                  Description: You can add up to 3 devices while using the 'Free' plan.
                </p>
              </div>
            </div>

            <div v-else-if="active && state === 'processed'">
              <div data-test="premiumPlan-div">
                <p>
                  Plan: <b> Premium usage </b>
                </p>

                <p>
                  Description: In this plan, the amount is charged according to the number of
                  devices used.
                </p>
              </div>
            </div>
          </div>
        </div>

        <div
          v-if="hasAuthorization && active && renderData && state==='processed'"
          class="mt-4 mb-4"
          data-test="subscriptionActive-div"
        >
          <v-divider />
          <v-divider />

          <div class="mt-6">
            <v-row>
              <v-col>
                <h3>
                  Next bill
                </h3>
              </v-col>
            </v-row>

            <div class="mt-6 pl-4 pr-4">
              <p>
                Date: <b> {{ infoBillingData.periodEnd | formatDateWithoutDayAndHours }} </b>
              </p>

              <p>
                Estimated total:
                <b v-if="renderData">
                  {{ infoBillingData.nextPaymentDue | formatCurrency }}
                </b>
              </p>
            </div>
          </div>

          <v-divider />
          <v-divider />

          <div class="mt-6">
            <v-row>
              <v-col>
                <h3>
                  Payment method details
                </h3>
              </v-col>

              <v-spacer />

              <v-col
                md="auto"
                class="ml-auto"
              >
                <PaymentMethod
                  type-operation="update"
                  data-test="updatePaymentMethod-component"
                  @update="getSubscriptionInfo()"
                />
              </v-col>
            </v-row>

            <div class="mt-5 pl-4 pr-4">
              <p>
                <BillingIcon
                  v-if="renderData"
                  :icon-name="cardBillingData.brand"
                  data-test="billingIcon-component"
                />
                {{ cardBillingData.expMonth }}/{{ cardBillingData.expYear }} -
                {{ cardBillingData.last4 }}
              </p>
            </div>
          </div>

          <v-divider />
          <v-divider />

          <div
            data-test="cancel-div"
            class="mt-6"
          >
            <v-row>
              <v-col>
                <h3>
                  Cancel Subscription
                </h3>
              </v-col>
            </v-row>

            <div class="mt-2 pl-4">
              <v-row>
                <v-col>
                  <p>
                    When canceling subscription, you may lose access to devices.
                  </p>
                </v-col>

                <v-spacer />

                <v-col
                  md="auto"
                  class="ml-auto"
                >
                  <BillingCancel
                    v-if="renderData"
                    :next-payment-due="infoBillingData.nextPaymentDue"
                    @cancel="cancel()"
                  />
                </v-col>
              </v-row>
            </div>
          </div>
        </div>

        <div v-else-if="state==='processed' && active">
          <div
            data-test="activeLoading-div"
          >
            <v-divider />
            <v-divider />
            <div class="mt-6 mb-2">
              <p>
                Loading data
              </p>
            </div>
          </div>
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>

import SettingOwnerInfo from '@/components/setting/SettingOwnerInfo';
import BillingCancel from '@/components/billing/BillingCancel';
import PaymentMethod from '@/components/billing/BillingDialogPaymentMethod';
import BillingIcon from '@/components/billing/BillingIcon';

import { formatDateWithoutDayAndHours } from '@/components/filter/date';
import formatCurrency from '@/components/filter/currency';
import hasPermission from '@/components/filter/permission';

export default {
  name: 'SettingBillingComponent',

  components: {
    SettingOwnerInfo,
    BillingCancel,
    PaymentMethod,
    BillingIcon,
  },

  filters: {
    formatDateWithoutDayAndHours,
    formatCurrency,
    hasPermission,
  },

  data() {
    return {
      card: null,
      pollMax: 4,
      retrials: 0,
      elements: null,
      billingData: { info: Object, card: Object },
      renderData: false,
    };
  },

  computed: {
    retrialExceeded() {
      return this.retrials >= this.pollMax;
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

    state() {
      return this.$store.getters['billing/status'];
    },

    infoBillingData() {
      return this.billingData.info;
    },

    cardBillingData() {
      return this.billingData.card;
    },

    hasAuthorization() {
      const accessType = this.$store.getters['auth/accessType'];
      if (accessType !== '') {
        return hasPermission(
          this.$authorizer.accessType[accessType],
          this.$actions.billing.subscribe,
        );
      }

      return false;
    },
  },

  watch: {
    state(val) {
      if (val === 'pending') {
        this.startPolling();
      } else {
        clearInterval(this.polling);
        if (this.state === 'processed') {
          this.getSubscriptionInfo();
        }
        this.retrials = 0;
      }
    },

    isOwner(status) {
      if (status) {
        this.stripeData();
      }
    },
  },

  created() {
    if (this.state === 'pending') {
      this.startPolling();
    }
    this.updateNamespace();
  },

  destroyed() {
    if (this.polling !== null) {
      clearInterval(this.polling);
    }
  },

  mounted() {
    if (this.isOwner) {
      this.stripeData();
    }
  },

  methods: {
    startPolling() {
      this.polling = setInterval(() => {
        if (this.retrialExceeded) {
          clearInterval(this.polling);
        } else {
          this.updateNamespace();
          this.retrials += 1;
        }
      }, 3000);
    },

    stripeData() {
      this.mountStripeElements();

      if (this.active) {
        this.getSubscriptionInfo();
      }
    },

    mountStripeElements() {
      this.elements = this.$stripe.elements();
      this.card = this.elements.create('card');
    },

    formatSubscriptionData(data) {
      const latestInvoice = data.latest_invoice;
      const upcomingInvoice = data.upcoming_invoice;
      const productDescription = data.product_description;
      const { card } = data;

      this.billingData.info = {
        periodEnd: this.billing.current_period_end,
        description: productDescription,
        latestPaymentDue: latestInvoice.amount_due,
        latestPaymentPaid: latestInvoice.amount_paid,
        nextPaymentDue: upcomingInvoice.amount_due,
        nextPaymentPaid: upcomingInvoice.amount_paid,
      };

      this.billingData.card = {
        brand: card.brand,
        expYear: card.exp_year,
        expMonth: card.exp_month,
        last4: card.last4,
      };
    },

    async getSubscriptionInfo() {
      if (this.active) {
        try {
          const data = await this.$store.dispatch('billing/getSubscription');
          await this.formatSubscriptionData(data);

          this.renderData = true;
        } catch {
          this.renderData = false;
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      }
    },

    async updateNamespace() {
      try {
        await this.$store.dispatch('namespaces/get', localStorage.getItem('tenant'));
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.namespaceLoad);
      }
    },
  },
};

</script>
