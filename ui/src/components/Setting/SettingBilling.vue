<template>
  <v-container>
    <v-dialog
      v-model="statusCheckout"
      max-width="600px"
      min-width="50vw"
      persistent
    >
      <v-card class="bg-v-theme-surface">
        <v-card-title class="bg-primary">
          Payment pending
        </v-card-title>
        <v-card-text>
          <p>
            You have a payment pending! Access the Billing Portal to regularize your payment status to create a new subscription.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" data-test="close-btn" @click="statusCheckout = !statusCheckout">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

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

          <v-row class="mt-4 justify-center align-center">
            <v-col>
              <h3>Subscription info</h3>
            </v-col>

            <v-spacer />

            <v-col v-if="state === 'inactive' || state === 'pending'" md="auto" class="ml-auto">
              <v-col align="left">
                <v-btn
                  color="primary"
                  class="text-none text-uppercase"
                  :disabled="state === ''"
                  @click="checkout"
                  data-test="checkout-button"
                >
                  <v-icon class="mr-2">mdi-credit-card</v-icon>
                  Subscribe
                </v-btn>
              </v-col>
            </v-col>
          </v-row>

          <div class="mt-6 pl-4 pr-4">
            <div v-if="state === 'inactive' || state === 'pending' ">
              <div data-test="freePlan-div">
                <p>Plan: <b> Free </b></p>

                <p>
                  Description: You can add up to 3 devices while using the
                  'Free' plan.
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

            <div class="mt-4 mb-4">
              <h4 class="">Billing Portal</h4>
              <p>ShellHub patterns with Stripe payment and invoicing. To update your payment method, or downloading previous invoices
                click on the button bellow.</p>
              <v-btn
                color="primary"
                class="mt-2 text-none text-uppercase"
                @click="portal"
                data-test="portal-button"
              >
                <v-icon class="mr-2">mdi-account</v-icon>
                Open Billing Portal
              </v-btn>
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
                    {{
                      formatDateWithoutDayAndHours(infoBillingData.periodEnd)
                    }}
                  </b>
                </p>

                <p>
                  Estimated total:
                  <b v-if="renderData">
                    {{
                      formatCurrency(
                        infoBillingData.nextPaymentDue,
                        infoBillingData.currency,
                      )
                    }}
                  </b>
                </p>
              </div>
            </div>

            <v-divider />
            <v-divider />

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
                      When canceling subscription, you may lose access to
                      devices.
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
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
/* eslint-disable */
import {
  defineComponent,
  ref,
  computed,
  onMounted,
  watch, onUnmounted,
} from "vue";
import { Stripe, StripeCardElement, StripeElements, loadStripe } from "@stripe/stripe-js";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import SettingOwnerInfo from "./SettingOwnerInfo.vue";
import BillingPaymentMethod from "../Billing/BillingPaymentMethod.vue";
import formatCurrency from "@/utils/currency";
import { formatDateWithoutDayAndHours } from "../../utils/formateDate";
import BillingCancel from "../Billing/BillingCancel.vue";
import { INotificationsError } from "../../interfaces/INotifications";
import { envVariables } from "@/envVariables";
import handleError from "@/utils/handleError";
import axios, {AxiosError} from "axios";
import AnnouncementsModal from "@/components/Announcements/AnnouncementsModal.vue";

export default defineComponent({
  setup() {
    const store = useStore();
    const card = ref<StripeCardElement>({} as StripeCardElement);
    const pollMax = ref(4);
    const retrials = ref(0);
    const elements = ref<StripeElements>({} as StripeElements);
    const renderData = ref(false);
    const statusCheckout = ref(false);
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
    const stripeKey = computed<string>(() => envVariables.stripeKey);
    let polling;
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
      // if (val === "inactive") {
      //   startPolling();
      // } else {
      //   clearInterval(polling);
      //   if (state.value === "processed") {
      //     getSubscriptionInfo();
      //   }
      //   retrials.value = 0;
      // }
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
      console.log("Mounted")
      if (hasAuthorization.value) {
        stripeData();
      }
      startPolling();
    });
    onUnmounted(() => {
      console.log("Unmounted")
      clearInterval(polling);
    });
    const startPolling = () => {
      polling = setInterval(() => {
        console.log("polling")
        updateNamespace();
      }, 3000);
    };
    const stripeData = () => {
      mountStripeElements();
      if (active.value) {
        getSubscriptionInfo();
      }
    };

    const mountStripeElements = async () => {
      const stripe = await loadStripe(stripeKey.value || "") as Stripe;
      elements.value = stripe.elements();
      card.value = elements.value.create("card");
    };

    const getSubscriptionInfo = async () => {
      if (active.value && hasAuthorization.value) {
        try {
          await store.dispatch("billing/getSubscription");
          renderData.value = true;
        } catch (error: unknown) {
          renderData.value = false;
          store.dispatch("snackbar/showSnackbarErrorDefault");
          handleError(error);
        }
      }
    };

    const updateNamespace = async () => {
      try {
        await store.dispatch("namespaces/get", localStorage.getItem("tenant"));
      } catch (error: unknown) {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.namespaceLoad,
        );
        handleError(error);
      }
    };
    const checkout = async () => {
      try {
        const res = await axios.post("/api/billing/checkout", {}, {
          headers: { Authorization: `Bearer ${localStorage.getItem("token")}` },
        });

        const { url } = res.data;

        window.open(url, "_self");
      } catch (error: unknown) {
        if (axios.isAxiosError(error)) {
          const axiosError = error as AxiosError;
          switch (true) {
            case axiosError.response?.status === 403: {
              statusCheckout.value = true;
              break;
            }
            default:
              handleError(error);
          }
        }

      }
    };

    const portal = async () => {
      try {
        const res = await axios.post("/api/billing/portal", {}, {
          headers: { Authorization: `Bearer ${localStorage.getItem("token")}` },
        });

        const { url } = res.data;

        window.open(url, "_self");
      } catch (error: unknown) {
        handleError(error);
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
      checkout,
      portal,
      statusCheckout
    };
  },
  components: {
    AnnouncementsModal,
    SettingOwnerInfo,
    BillingPaymentMethod,
    BillingCancel,
  },
});
</script>
