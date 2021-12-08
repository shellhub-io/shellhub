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

            <div
              data-test="paymentMethods-div"
              class="mt-5 pl-4 pr-4 pb-5"
            >
              <v-row class="text-row mb-2">
                <v-col>
                  <b>Brand</b>
                </v-col>
                <v-col>
                  <b>Expiration date</b>
                </v-col>
                <v-col>
                  <b>Ends with</b>
                </v-col>
                <v-col
                  class="pm-actionsText"
                  cols="3"
                >
                  <b>Actions</b>
                </v-col>
              </v-row>
              <v-row
                v-for="item in cardBillingData"
                :key="item.id"
                class="pm-rowItem"
              >
                <v-col>
                  <BillingIcon
                    v-if="renderData"
                    :icon-name="item.brand"
                    data-test="billingIcon-component"
                  />
                </v-col>
                <v-col class="pm-data">
                  <p>
                    {{ item.expMonth }}/{{ item.expYear }}
                  </p>
                </v-col>
                <v-col class="pm-data">
                  <p>
                    {{ item.last4 }}
                  </p>
                </v-col>
                <v-col
                  class="actions-column"
                  cols="4"
                >
                  <div
                    v-if="item.default"
                    class="pm-text"
                  >
                    <p>
                      Default
                    </p>
                  </div>
                  <div
                    v-else
                    class="pm-actions"
                  >
                    <v-btn
                      class="ml-4"
                      outlined
                      @click="updatePaymentMethod(item.id)"
                    >
                      <div>
                        <v-tooltip bottom>
                          <template #activator="{ on }">
                            <span v-on="on">
                              <v-icon
                                v-on="on"
                              >
                                mdi-pencil
                              </v-icon>
                            </span>
                          </template>
                          <span>
                            Make default
                          </span>
                        </v-tooltip>
                      </div>
                    </v-btn>
                    <v-btn
                      class="ml-4"
                      outlined
                      @click="deletePaymentMethod(item.id)"
                    >
                      <div>
                        <v-tooltip bottom>
                          <template #activator="{ on }">
                            <span v-on="on">
                              <v-icon
                                v-on="on"
                              >
                                delete
                              </v-icon>
                            </span>
                          </template>
                          <span>
                            Remove
                          </span>
                        </v-tooltip>
                      </div>
                    </v-btn>
                  </div>
                </v-col>
              </v-row>
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
      const accessType = this.$store.getters['auth/accessType'];
      if (accessType !== '') {
        return hasPermission(
          this.$authorizer.accessType[accessType],
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

    async updatePaymentMethod(paymentMethodId) {
      try {
        await this.$store.dispatch('billing/updatePaymentMethod', paymentMethodId);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.updateSubscription);
        this.$emit('update');
      } catch (error) {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.updateSubscription);

        const { status } = error.response;
        if (status === 400 || status === 423) {
          this.showError(error);
        }
      }
    },

    async deletePaymentMethod(paymentMethodId) {
      try {
        await this.$store.dispatch('billing/removePaymentMethod', paymentMethodId);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.updateSubscription);
        this.$emit('update');
      } catch (error) {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deletePaymentMethod);

        const { status } = error.response;
        if (status === 400 || status === 423) {
          this.showError(error);
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

<style>

.pm-data {
  text-align: center;
  margin-right: 2rem;
}

.pm-actionsText {
  text-align: center;
}

.pm-text {
  text-align: center;
}

</style>
