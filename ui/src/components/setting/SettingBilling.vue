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

          <div class="mt-6 mb-2">
            <v-row>
              <v-col>
                <h3>
                  Payment methods
                </h3>
              </v-col>

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
          </div>

          <BillingPaymentList
            data-test="paymentMethods-component"
            :cards.sync="cardBillingData"
          />

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
import BillingPaymentList from '@/components/billing/BillingPaymentList';
import { formatDateWithoutDayAndHours } from '@/components/filter/date';
import formatCurrency from '@/components/filter/currency';
import hasPermission from '@/components/filter/permission';

export default {
  name: 'SettingBillingComponent',

  components: {
    SettingOwnerInfo,
    BillingCancel,
    BillingPaymentList,
    PaymentMethod,
  },

  filters: {
    formatDateWithoutDayAndHours,
    formatCurrency,
    hasPermission,
  },

  data() {
    return {
      card: null,
      cards: null,
      pollMax: 4,
      retrials: 0,
      elements: null,
      renderData: false,
      action: 'subscribe',
    };
  },

  computed: {
    retrialExceeded() {
      return this.retrials >= this.pollMax;
    },

    active() {
      return this.$store.getters['billing/active'];
    },

    billingData() {
      return this.$store.getters['billing/getBillInfoData'];
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
      return this.billingData.cards;
    },

    cardBillingDefault() {
      return this.billingData.defaultCard;
    },

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

    async getSubscriptionInfo() {
      if (this.active) {
        try {
          await this.$store.dispatch('billing/getSubscription');
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
