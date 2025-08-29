<template>
  <BaseDialog
    v-model="showCheckoutDialog"
    @close="resetDialog"
    transition="dialog-bottom-transition"
    data-test="checkout-dialog"
  >
    <v-window v-model="el">
      <v-window-item :value="1">
        <v-card class="bg-v-theme-surface content pa-4 pb-0 px-0" data-test="card-first-page">
          <v-container>
            <v-card-subtitle class="mb-1" style="font-size: 12px;">
              <b>Welcome</b> > Payment Details > Checkout
            </v-card-subtitle>
            <BillingLetter />

          </v-container>
          <v-card-actions class="justify-end">
            <v-btn color="primary" @click="resetDialog" data-test="payment-letter-close-button">Close</v-btn>
            <v-spacer />
            <v-btn
              color="primary"
              @click="goToNextStep"
              data-test="payment-letter-next-button"
            >
              Next
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-window-item>
      <v-window-item :value="2">
        <v-card class="bg-v-theme-surface content" data-test="card-second-page">
          <v-container class="pa-4">
            <v-card-subtitle class="mb-1" style="font-size: 12px;">
              Welcome > <b>Payment Details</b> > Checkout
            </v-card-subtitle>
            <v-card-title align="center" class="mb-1" data-test="billing-payment-details">Payment Details</v-card-title>
            <BillingPayment
              @no-payment-methods="existingDefaultCard = false"
              @has-default-payment="existingDefaultCard = true"
              @customer-id-created="noCustomer.value = false"
            />
          </v-container>
          <v-card-actions>
            <v-btn color="primary" @click="goToPreviousStep" data-test="payment-details-back-button">Back</v-btn>
            <v-spacer />
            <v-btn
              :disabled="!existingDefaultCard"
              color="primary"
              @click="goToNextStep"
              data-test="payment-details-next-button"
            >
              Next
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-window-item>
      <v-window-item :value="3">
        <v-card class="bg-v-theme-surface content" data-test="card-third-page">
          <v-container class="pa-6">
            <v-card-subtitle class="mb-1" style="font-size: 12px;">
              Welcome > Payment Details > <b>Checkout</b>
            </v-card-subtitle>
            <BillingCheckout :key="componentKey" />
            <v-row>
              <v-col>
                <v-alert
                  v-if="alertRender"
                  icon="$error"
                  :text="errorMessage"
                  type="error"
                  data-test="checkout-error-alert"
                />
              </v-col>
            </v-row>
          </v-container>
          <v-card-actions>
            <v-btn color="primary" @click="goToPreviousStep" data-test="checkout-back-button">Back</v-btn>
            <v-spacer />
            <v-btn @click="subscribe" color="primary" data-test="checkout-button">Subscribe now</v-btn>
          </v-card-actions>
        </v-card>
      </v-window-item>
      <v-window-item :value="4">
        <v-card class="bg-v-theme-surface content pa-4 pb-0 px-0" @click:outside="emit('reload')" data-test="card-fourth-page">
          <v-container>
            <BillingSuccesful />
            <v-card-actions>
              <v-spacer />
              <v-btn @click="emit('reload')" data-test="successful-close-button">Close</v-btn>
            </v-card-actions>
          </v-container>
        </v-card>
      </v-window-item>
    </v-window>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive } from "vue";
import BillingLetter from "./BillingLetter.vue";
import BillingPayment from "./BillingPayment.vue";
import BillingCheckout from "./BillingCheckout.vue";
import BillingSuccesful from "./BillingSuccessful.vue";
import handleError from "@/utils/handleError";
import BaseDialog from "../BaseDialog.vue";
import useCustomerStore from "@/store/modules/customer";

const customerStore = useCustomerStore();
const showCheckoutDialog = defineModel({ default: false });
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

const emit = defineEmits(["reload"]);

defineExpose({
  showCheckoutDialog,
  el,
});
</script>
