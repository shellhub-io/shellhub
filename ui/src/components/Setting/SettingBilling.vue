<template>
  <v-container>
    <v-container>
      <v-row align="center" justify="center" class="mt-4 mb-4">
        <v-col sm="8">
          <SettingOwnerInfo :is-owner="hasAuthorization" data-test="settingOwnerInfo-component" />
          <v-alert
            v-if="message"
            :text="message"
            :type="messageType"
            variant="tonal"
            data-test="message-alert"
          />
          <div v-if="hasAuthorization" data-test="content-div">
            <v-row class="mt-4 justify-center align-center">
              <v-col>
                <h3>Subscription info</h3>
              </v-col>
              <v-spacer />
              <v-col v-if="!active" md="auto" class="ml-auto">
                <v-col align="left">
                  <v-btn
                    color="primary"
                    class="text-none text-uppercase"
                    :disabled="status === ''"
                    @click="checkout"
                    data-test="subscribe-button"
                  >
                    <v-icon class="mr-2">mdi-credit-card</v-icon>
                    Subscribe
                  </v-btn>
                </v-col>
              </v-col>
            </v-row>

            <div class="mt-6 pl-4 pr-4">
              <div v-if="!active" data-test="freePlan-div">
                <p>Plan: <b>Free</b></p>
                <p>
                  Description: You can add up to 3 devices while using the 'Free' plan.
                </p>
              </div>

              <div v-else data-test="premiumPlan-div">
                <p>Plan: <b>Premium usage</b></p>
                <p>
                  Description: In this plan, the amount is charged according to the number of devices used.
                </p>
              </div>

              <div class="mt-4 mb-4" data-test="billing-portal-text">
                <h4>Billing Portal</h4>
                <p>
                  ShellHub patterns with Stripe payment and invoicing. To update your payment method or download previous invoices,
                  click on the button below.
                </p>
                <v-btn
                  :disabled="noCustomer.value"
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

            <div v-if="hasAuthorization && active" class="mt-4 mb-4" data-test="subscriptionActive-div">
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
                      {{ formattedDate }}
                    </b>
                  </p>
                  <p>
                    Estimated total:
                    <b>
                      {{ formattedCurrency }}
                    </b>
                  </p>
                </div>
              </div>
              <v-divider />
              <v-divider />
            </div>
          </div>
        </v-col>
      </v-row>
    </v-container>
    <v-dialog v-model="dialogCheckout" persistent width="600" transition="dialog-bottom-transition" data-test="dialog-checkout">
      <v-window v-model="el">
        <v-window-item :value="1">
          <v-card class="bg-v-theme-surface content" data-test="card-first-page">
            <v-container>
              <v-card-subtitle class="mb-1" style="font-size: 12px;"><b>Welcome</b> > Payment Details > Checkout</v-card-subtitle>
              <BillingLetter />
              <v-row>
                <v-col>
                  <v-card-actions>
                    <v-btn color="primary" @click="close()" data-test="payment-letter-close-button">Close</v-btn>
                  </v-card-actions>
                </v-col>
                <v-col>
                  <v-card-actions class="justify-end">
                    <v-btn
                      color="primary"
                      @click="goToNextStep"
                      data-test="payment-letter-next-button"
                    >
                      Next
                    </v-btn>
                  </v-card-actions>
                </v-col>
              </v-row>
            </v-container>
          </v-card>
        </v-window-item>
        <v-window-item :value="2">
          <v-card class="bg-v-theme-surface content" data-test="card-second-page">
            <v-container>
              <v-card-subtitle class="mb-1" style="font-size: 12px;">Welcome > <b>Payment Details</b> > Checkout</v-card-subtitle>
              <v-card-title align="center" class="mb-1" data-test="billing-payment-details">Payment Details</v-card-title>
              <BillingPayment
                @no-payment-methods="existingDefaultCard = false"
                @has-default-payment="existingDefaultCard = true"
                @customer-id-created="noCustomer.value = false"
              />
              <v-row>
                <v-col>
                  <v-card-actions>
                    <v-btn color="primary" @click="goToPreviousStep" data-test="payment-details-back-button">Back</v-btn>
                  </v-card-actions>
                </v-col>
                <v-col class="d-flex flex-column align-end">
                  <v-card-actions>
                    <v-btn
                      :disabled="!existingDefaultCard"
                      color="primary"
                      @click="goToNextStep"
                      data-test="payment-details-next-button"
                    >
                      Next
                    </v-btn>
                  </v-card-actions>
                </v-col>
              </v-row>
            </v-container>
          </v-card>
        </v-window-item>
        <v-window-item :value="3">
          <v-card class="bg-v-theme-surface content" data-test="card-third-page">
            <v-container>
              <v-card-subtitle class="mb-1" style="font-size: 12px;">Welcome > Payment Details > <b>Checkout</b></v-card-subtitle>
              <BillingCheckout :key="componentKey" />
              <v-row>
                <v-col>
                  <v-alert v-if="alertRender" icon="$error" :text="errorMessage" type="error" data-test="checkout-error-alert" />
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <v-card-actions>
                    <v-btn color="primary" @click="goToPreviousStep" data-test="checkout-back-button">Back</v-btn>
                  </v-card-actions>
                </v-col>
                <v-col class="d-flex flex-column align-end">
                  <v-btn @click="subscribe()" color="primary" data-test="checkout-button">Subscribe now</v-btn>
                </v-col>
              </v-row>
            </v-container>
          </v-card>
        </v-window-item>
        <v-window-item :value="4">
          <v-card class="bg-v-theme-surface content" data-test="card-fourth-page">
            <v-container>
              <BillingSuccesful />
              <v-row>
                <v-col>
                  <v-card-actions>
                    <v-btn @click="reload()" data-test="successful-close-button">Close</v-btn>
                  </v-card-actions>
                </v-col>
              </v-row>
            </v-container>
          </v-card>
        </v-window-item>
      </v-window>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import {
  ref,
  computed,
  watch,
  onMounted,
  reactive,
} from "vue";
import axios from "axios";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import SettingOwnerInfo from "./SettingOwnerInfo.vue";
import BillingPayment from "@/components/Billing/BillingPayment.vue";
import formatCurrency from "@/utils/currency";
import { formatDateWithoutDayAndHours } from "../../utils/formateDate";
import handleError from "@/utils/handleError";
import BillingCheckout from "../Billing/BillingCheckout.vue";
import BillingSuccesful from "../Billing/BillingSuccessful.vue";
import BillingLetter from "../Billing/BillingLetter.vue";

const store = useStore();
const billing = computed(() => store.getters["billing/get"]);
const active = computed(() => store.getters["billing/active"]);
const status = computed(() => store.getters["billing/status"]);
const namespace = computed(() => store.getters["namespaces/get"]);
const el = ref<number>(1);
const dialogCheckout = ref(false);
const alertRender = ref(false);
const errorMessage = ref("");
const noCustomer = reactive({ value: false });
const existingDefaultCard = ref(true);
const componentKey = ref(0);
const message = ref("");
const messageType = ref();
const formattedDate = ref();
const formattedCurrency = ref();

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.billing.subscribe,
    );
  }
  return false;
});

const errorTreatment = async () => {
  switch (status.value) {
    // eslint-disable-next-line vue/camelcase, camelcase
    case "to_cancel_at_end_of_period":
      message.value = `Your subscription will be canceled at ${formattedDate.value
      }, if you want to renew your subscription to premium, please, access our billing portal.`;
      messageType.value = "warning";
      break;
    case "past_due":
      message.value = "Your subscription payment method has failed. Please, access the billing portal to keep your subscription";
      messageType.value = "warning";
      break;
    case "unpaid":
      // eslint-disable-next-line vue/max-len
      message.value = "You have unpaid invoices which made your subscription to be canceled. Please, solve this issue opening the billing portal.";
      messageType.value = "error";
      break;
    case "canceled":
      message.value = "Your subscription was canceled. To continue to use ShellHub premium benefits, please, subscribe to a new one.";
      messageType.value = "error";
      break;
    default:
      break;
  }
};

const getSubscriptionInfo = async () => {
  if (active.value && hasAuthorization.value) {
    try {
      await store.dispatch("billing/getSubscription");
      formattedDate.value = formatDateWithoutDayAndHours(billing.value?.end_at);
      formattedCurrency.value = formatCurrency(billing.value?.invoices[0].amount, billing.value?.invoices[0].currency.toUpperCase());
    } catch (error: unknown) {
      handleError(error);
    }
  }
};

watch(active, (val) => {
  if (val) {
    getSubscriptionInfo();
  }
});

onMounted(async () => {
  const tenant = computed(() => localStorage.getItem("tenant"));
  await store.dispatch("namespaces/get", tenant.value);
  if (namespace.value.billing == null || !namespace.value.billing.customer_id) {
    noCustomer.value = true;
  }
  await getSubscriptionInfo();
  await errorTreatment();
});

const subscribe = async () => {
  try {
    await store.dispatch("customer/createSubscription");
    el.value = 4;
  } catch (status) {
    switch (status) {
      case 402:
        alertRender.value = true;
        // eslint-disable-next-line vue/max-len
        errorMessage.value = "Before attempting to subscribe again, please ensure that all your invoices have been paid or closed by checking the billing portal.";
        break;
      default:
        alertRender.value = true;
        errorMessage.value = "An error occurred during the payment process. Please try again later or contact the ShellHub team";
    }
    handleError(status);
  }
};

const checkout = async () => {
  el.value = 1;
  dialogCheckout.value = true;
};

const close = () => {
  dialogCheckout.value = false;
};

const reload = () => {
  window.location.reload();
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

const goToPreviousStep = () => {
  el.value--;
};

const forceRerender = () => {
  componentKey.value += 1;
};

const goToNextStep = () => {
  el.value++;
  if (el.value === 3) {
    forceRerender();
  }
};

defineExpose({
  dialogCheckout,
  el,
});
</script>

<style scoped>
p {
  text-align: justify;
  overflow: auto;
}
</style>
