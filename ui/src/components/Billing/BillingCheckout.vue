<template>
  <v-container>
    <v-row>
      <v-col>
        <h2 data-test="title">Payment Method:</h2>
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <p data-test="sub-title">This is the payment method you have selected for your ShellHub Cloud subscription</p>
      </v-col>
    </v-row>
    <v-col>
      <v-card align="center" class="bg-v-theme-card" variant="flat" data-test="card">
        <v-list-item class="pl-0 pr-0">
          <v-row align="center" cols="12" class="pa-0 pa-2">
            <v-col cols="1">
              <BillingIcon :icon-name="paymentMethod.brand" />
            </v-col>
            <v-col cols="3">
              <b>{{ paymentMethod.number }}</b>
            </v-col>
            <v-col cols="3">
              {{ paymentMethod.exp_month + "/" + paymentMethod.exp_year }}
            </v-col>
            <v-col cols="2">
              {{ paymentMethod.cvc }}
            </v-col>
            <v-col cols="3">
              <v-chip>
                <b>default</b>
                <v-tooltip
                  activator="parent"
                  location="top"
                >This payment method will be used on your subscription</v-tooltip>
              </v-chip>
            </v-col>
          </v-row>
        </v-list-item>
      </v-card>
    </v-col>
    <v-spacer class="mb-5" />
    <v-row>
      <v-col>
        <h3 data-test="additional-information">Additional Information:</h3>
        <v-spacer class="mb-2" />
        <ul data-test="additional-information-list" class="text-justify">
          <li><p>Your selected payment method will be charged automatically on a monthly basis</p></li>
          <li><p>You can manage your payment methods, invoices and subscription details in the <strong>Billing Portal.</strong></p></li>
          <li><p>Invoices will be generated and available for download at the beginning of each billing cycle.</p></li>
          <li><p>You have the option to cancel your subscription at any time through the <strong>Billing Portal.</strong></p></li>
        </ul>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { onMounted, reactive, computed } from "vue";
import { IPaymentMethod } from "@/interfaces/ICustomer";
import BillingIcon from "@/components/Billing/BillingIcon.vue";
import useCustomerStore from "@/store/modules/customer";

const customerStore = useCustomerStore();
const paymentMethod = reactive<IPaymentMethod>({ brand: "", cvc: "", default: false, exp_year: 0, exp_month: 0, id: "", number: "" });
const customer = computed(() => customerStore.customer);

onMounted(async () => {
  await customerStore.fetchCustomer();
  const defaultPaymentMethod = customer.value.payment_methods?.filter((value: IPaymentMethod) => value.default === true)[0];
  paymentMethod.brand = defaultPaymentMethod?.brand || "";
  paymentMethod.cvc = defaultPaymentMethod?.cvc || "";
  paymentMethod.exp_year = defaultPaymentMethod?.exp_year || 0;
  paymentMethod.exp_month = defaultPaymentMethod?.exp_month || 0;
  paymentMethod.number = defaultPaymentMethod?.number || "";
  paymentMethod.default = defaultPaymentMethod?.default || false;
});
</script>
