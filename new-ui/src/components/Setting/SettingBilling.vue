<template>
  <v-container>
    <v-row align="center" justify="center" class="mt-4 mb-4">
      <v-col sm="8">
        <SettingOwnerInfo
          :is-owner="hasAuthorization"
          data-test="settingOwnerInfo-component"
        />

        <div v-if="hasAuthorization" data-test="content-div">
          <div
            v-if="active && state === 'processed' && notifyWarning"
            data-test="warning-div"
          >
            <v-alert
              class="mt-4 pl-4 pr-4 d-flex justify-center align-center"
              variant="outlined"
              type="info"
            >
              <div>
                <p class="font-weight-bold">
                  {{ warningTitle }}
                </p>
              </div>

              <div>
                {{ warningMessage }}
              </div>
            </v-alert>
          </div>
        </div>

        <v-row class="mt-4">
          <v-col>
            <h3>Subscription info</h3>
          </v-col>

          <v-spacer />

          <v-col v-if="state === 'inactive'" md="auto" class="ml-auto">
            <BillingPaymentMethod
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
              <p>Plan: <b> Free </b></p>

              <p>
                Description: You can add up to 3 devices while using the 'Free'
                plan.
              </p>
            </div>
          </div>

          <div v-else-if="active && state === 'processed'">
            <div data-test="premiumPlan-div">
              <p>Plan: <b> Premium usage </b></p>

              <p>
                Description: In this plan, the amount is charged according to
                the number of devices used.
              </p>
            </div>
          </div>
        </div>

        <div
          v-if="
            hasAuthorization && active && renderData && state === 'processed'
          "
          class="mt-4 mb-4"
          data-test="subscriptionActive-div"
        >
          <v-divider />
          <v-divider />

          <div class="mt-6 mb-6">
            <v-row>
              <v-col>
                <h3>Next bill</h3>
              </v-col>
            </v-row>

            <div class="mt-6 pl-4 pr-4">
              <p>
                Date:
                <b>
                  {{ formatDateWithoutDayAndHours(infoBillingData.periodEnd) }}
                </b>
              </p>

              <p>
                Estimated total:
                <b v-if="renderData">
                  {{ formatCurrency(infoBillingData.nextPaymentDue, infoBillingData.currency) }}
                </b>
              </p>
            </div>
          </div>

          <v-divider />
          <v-divider />

          <div class="mt-6 mb-6">
            <v-row>
              <v-col>
                <h3>Latest invoices</h3>
              </v-col>
            </v-row>
          </div>

          <BillingInvoiceList data-test="invoiceList-component" />

          <v-divider />
          <v-divider />
          <div class="mt-6 mb-2">
            <v-row>
              <v-col>
                <h3>Payment methods</h3>
              </v-col>

              <v-spacer />

              <v-col md="auto" class="ml-auto">
                <BillingPaymentMethod
                  type-operation="update"
                  data-test="updatePaymentMethod-component"
                  @update="getSubscriptionInfo()"
                />
              </v-col>
            </v-row>
          </div>

          <div>
            <BillingPaymentList
              data-test="paymentMethods-component"
              v-model:cards="cardBillingData"
            />

            <v-divider />
            <v-divider />
          </div>

          <div data-test="cancel-div" class="mt-6">
            <v-row>
              <v-col>
                <h3>Cancel Subscription</h3>
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

                <v-col md="auto" class="ml-auto">
                  <BillingCancel
                    v-if="renderData"
                    :nextPaymentDue="infoBillingData.nextPaymentDue"
                    :currency="infoBillingData.currency"
                  />
                </v-col>
              </v-row>
            </div>
          </div>
        </div>

        <div v-else-if="hasAuthorization && state === 'processed' && active">
          <div data-test="activeLoading-div">
            <v-divider />
            <v-divider />
            <div class="mt-6 mb-2">
              <p>Loading data</p>
            </div>
          </div>
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
/* eslint-disable */
import { defineComponent, ref, computed, onMounted, watch, onBeforeMount, onUpdated } from "vue";
import { loadStripe } from "@stripe/stripe-js";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import SettingOwnerInfo from "./SettingOwnerInfo.vue";
import BillingPaymentMethod from "../Billing/BillingPaymentMethod.vue";
import formatCurrency from "@/utils/currency";
import { formatDateWithoutDayAndHours } from "../../utils/formateDate";
import BillingInvoiceList from "../Billing/BillingInvoiceList.vue";
import BillingCancel from "../Billing/BillingCancel.vue";
import { INotificationsError } from "../../interfaces/INotifications";
import BillingPaymentList from "../Billing/BillingPaymentList.vue";
import { envVariables } from "@/envVariables";

export default defineComponent({
  setup() {
    const store = useStore();
    const card = ref(null);
    const cards = ref(null);
    const pollMax = ref(4);
    const retrials = ref(0);
    const elements = ref<any>(null);
    const renderData = ref(false);
    const warningTitle = ref("Payment failed");
    const warningMessage = ref(`Please update your payment method
        by adding a new card, or attempt to pay failed latest
        invoices through url`);
    const billingData = computed(() => store.getters["billing/getBillInfoData"]);
    const retrialExceeded = computed(() => retrials.value >= pollMax.value);
    const active = computed(() => store.getters["billing/active"]);
    const notifyWarning = computed(() => billingData.value.warning);
    const billing = computed(() => store.getters["billing/get"]);
    const state = computed(() => store.getters["billing/status"]);
    const infoBillingData = computed(() => billingData.value.info);
    const cardBillingData = computed(() => billingData.value.cards);
    const cardBillingDefault = computed(() => billingData.value.defaultCard);
    const stripeKey = computed(() => envVariables.stripeKey);
    let polling: any;
    const hasAuthorization = computed(() => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.billing["subscribe"]
        );
      }
      return false;
    });
    watch(state, (val) => {
      if (val === "pending") {
        startPolling();
      } else {
        clearInterval(polling);
        if (state.value === "processed") {
          getSubscriptionInfo();
        }
        retrials.value = 0;
      }
    });
    watch(hasPermission, (status: any) => {
      if (status) {
        stripeData();
      }
    });
    watch(active, (val) => {
      if (val) {
        getSubscriptionInfo();
      }
    });
    onMounted(() => {
      if (hasAuthorization.value) {
        stripeData();
      }
    });
    const startPolling = () => {
      polling = setInterval(() => {
        if (retrialExceeded.value) {
          clearInterval(polling);
        } else {
          updateNamespace();
          retrials.value += 1;
        }
      }, 3000);
    };
    const stripeData = () => {
      mountStripeElements();
      if (active.value) {
        getSubscriptionInfo();
      }
    };

    const mountStripeElements = async () => {
      const stripe = await loadStripe(stripeKey.value || "");
      elements.value = stripe?.elements();
      card.value = elements.value.create("card");
    };

    const getSubscriptionInfo = async () => {
      if (active.value) {
        try {
          await store.dispatch("billing/getSubscription");
          console.log("billing data", billing.value)
          console.log("billing billingData", billingData.value)
          renderData.value = true;
        } catch (error: any) {
          renderData.value = false;
          store.dispatch("snackbar/showSnackbarErrorDefault");
          throw new Error(error);
        }
      }
    };

    const updateNamespace = async () => {
      try {
        await store.dispatch("namespaces/get", localStorage.getItem("tenant"));
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.namespaceLoad,
        );
        throw new Error(error);
      }
    };

    return {
      warningTitle,
      warningMessage,
      active,
      state,
      notifyWarning,
      hasAuthorization,
      retrialExceeded,
      infoBillingData,
      formatCurrency,
      formatDateWithoutDayAndHours,
      getSubscriptionInfo,
      renderData,
      cardBillingData,
    };
  },
  components: {
    SettingOwnerInfo,
    BillingPaymentMethod,
    BillingInvoiceList,
    BillingCancel,
    BillingPaymentList,
  },
});
</script>
