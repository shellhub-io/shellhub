<template>
  <WindowDialog
    v-model="showCheckoutDialog"
    transition="dialog-bottom-transition"
    data-test="checkout-dialog"
    title="Billing & Subscription"
    description="Manage your subscription and payment details"
    icon="mdi-credit-card"
    icon-color="primary"
    :show-footer="true"
    @close="resetDialog"
  >
    <v-card-text class="pa-6">
      <v-window v-model="el">
        <v-window-item :value="1">
          <BillingLetter />
        </v-window-item>
        <v-window-item :value="2">
          <v-card-title
            align="center"
            class="pt-0 pb-4"
            data-test="billing-payment-details"
          >
            Payment Details
          </v-card-title>
          <BillingPayment
            @no-payment-methods="existingDefaultCard = false"
            @has-default-payment="existingDefaultCard = true"
            @customer-id-created="noCustomer.value = false"
          />
        </v-window-item>
        <v-window-item :value="3">
          <BillingCheckout
            :key="componentKey"
            data-test="billing-checkout"
          />
          <v-alert
            v-if="alertRender"
            icon="$error"
            :text="errorMessage"
            type="error"
            data-test="checkout-error-alert"
            role="alert"
            aria-live="assertive"
          />
        </v-window-item>
        <v-window-item :value="4">
          <div
            class="content pa-4 pb-0 px-0"
            data-test="card-fourth-page"
            @click:outside="emit('reload')"
          >
            <v-container>
              <BillingSuccessful />
            </v-container>
          </div>
        </v-window-item>
      </v-window>
    </v-card-text>

    <template #footer>
      <v-spacer />
      <template v-if="el === 1">
        <v-btn
          data-test="payment-letter-close-button"
          @click="resetDialog"
        >
          Close
        </v-btn>
        <v-btn
          color="primary"
          data-test="payment-letter-next-button"
          @click="goToNextStep"
        >
          Next
        </v-btn>
      </template>

      <template v-if="el === 2">
        <v-btn
          data-test="payment-details-back-button"
          @click="goToPreviousStep"
        >
          Back
        </v-btn>
        <v-btn
          :disabled="!existingDefaultCard"
          color="primary"
          data-test="payment-details-next-button"
          @click="goToNextStep"
        >
          Next
        </v-btn>
      </template>

      <template v-if="el === 3">
        <v-btn
          data-test="checkout-back-button"
          @click="goToPreviousStep"
        >
          Back
        </v-btn>
        <v-btn
          color="primary"
          data-test="checkout-button"
          @click="subscribe"
        >
          Subscribe now
        </v-btn>
      </template>

      <template v-if="el === 4">
        <v-btn
          data-test="successful-close-button"
          @click="emit('reload')"
        >
          Close
        </v-btn>
      </template>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import { ref, reactive } from "vue";
import BillingLetter from "./BillingLetter.vue";
import BillingPayment from "./BillingPayment.vue";
import BillingCheckout from "./BillingCheckout.vue";
import BillingSuccessful from "./BillingSuccessful.vue";
import handleError from "@/utils/handleError";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
import useCustomerStore from "@/store/modules/customer";
import { AxiosError } from "axios";

const customerStore = useCustomerStore();
const showCheckoutDialog = defineModel<boolean>({ required: true });
const el = ref(1);
const existingDefaultCard = ref(true);
const alertRender = ref(false);
const errorMessage = ref("");
const noCustomer = reactive({ value: false });
const componentKey = ref(0);

const goToPreviousStep = () => {
  el.value--;
};

const resetDialog = () => {
  el.value = 1;
  showCheckoutDialog.value = false;
};

const goToNextStep = () => {
  el.value++;
  if (el.value === 3) {
    componentKey.value++;
  }
};

const subscribe = async () => {
  try {
    await customerStore.createSubscription();
    el.value = 4;
  } catch (error) {
    const status = (error as AxiosError).response?.status;
    switch (status) {
      case 402:
        alertRender.value = true;
        errorMessage.value = `Before attempting to subscribe again, 
      please ensure that all your invoices have been paid or closed by checking the billing portal.`;
        break;
      default:
        alertRender.value = true;
        errorMessage.value
          = "An error occurred during the payment process. Please try again later or contact the ShellHub team";
    }
    handleError(status);
  }
};

const emit = defineEmits(["reload"]);
</script>
