<template>
  <v-row>
    <v-col>
      <v-text-field
        v-model="customer.name"
        label="Name"
        disabled
        data-test="customer-name" />
    </v-col>
    <v-col>
      <v-text-field
        v-model="customer.email"
        label="E-mail"
        disabled
        data-test="customer-email" />
    </v-col>
  </v-row>
  <v-col class="ma-0 pa-0">
    <h4 data-test="credit-card-text">Your credit cards</h4>
  </v-col>
  <v-list
    v-if="!(customer.payment_methods == null)"
    data-test="payment-methods-list"
    nav
    bg-color="transparent"
    class="w-100 pa-0 pt-2 content-card">
    <v-col v-for="(item, i) in customer.payment_methods" :key="i" class="pa-0 mt-2 mb-2">
      <v-card :key="i" class="bg-v-theme-card" variant="flat">
        <v-list-item :key="i" data-test="payment-methods-item" @click="setDefaultPayment(item.id)" class="card">
          <v-row align="center" cols="12">
            <v-col cols="1">
              <BillingIcon :icon-name="item.brand" />
            </v-col>
            <v-col cols="4">
              <b>{{ item.number }}</b>
            </v-col>
            <v-col cols="2">
              {{ item.exp_month + "/" + item.exp_year }}
            </v-col>
            <v-col cols="2">
              {{ item.cvc }}
            </v-col>
            <v-col cols="3" v-if="item.default === true">
              <v-chip density="comfortable">
                <b>default</b>
                <v-tooltip
                  activator="parent"
                  location="top"
                >This payment method will be used on your subscription</v-tooltip>
              </v-chip>
            </v-col>
            <v-col cols="3" v-if="!item.default" class="d-flex flex-column align-end">
              <v-btn variant="text" icon="mdi-delete" data-test="payment-methods-delete-btn" @click.stop="deletePaymentMethod(item.id)" />
            </v-col>
          </v-row>
        </v-list-item>
      </v-card>
    </v-col>
  </v-list>
  <v-col v-else>
    <v-card-subtitle>You don't have any registered cards yet, please add one</v-card-subtitle>
  </v-col>
  <v-row class="mt-0 pt-0 pt-0" cols="12">
    <v-col cols="8" class="pr-0">
      <StripeElements
        v-if="stripeLoaded && addNewCard"
        v-slot="{ elements }"
        ref="elms"
        :stripe-key="stripeKey"
        :instance-options="instanceOptions"
        :elements-options="elementsOptions">
        <StripeElement
          type="card"
          ref="card"
          :elements="elements"
          :options="cardOptions" />
      </StripeElements>
    </v-col>
    <v-col cols="4" class="d-flex flex-lg-column align-center justify-center pt-0">
      <v-btn
        @click='addNewCard ? savePayment() : addNewCard = true'
        :text='addNewCard ? "Save card" : "Add new card "'
        prepend-icon="mdi-credit-card-plus"
        color="primary"
        data-test="add-card-btn" />
    </v-col>
  </v-row>
  <v-row>
    <v-col>
      <v-alert v-if="alertRender" icon="$error" :text="errorMessage" type="error" data-test="alert-message" />
    </v-col>
  </v-row>
</template>

<script lang='ts' setup>
import { onBeforeMount, onMounted, ref, reactive, computed } from "vue";
import { loadStripe } from "@stripe/stripe-js";
import { StripeElements, StripeElement } from "vue-stripe-js";
import type { StripeConstructorOptions, StripeElementsOptions, StripeCardElementOptions } from "@stripe/stripe-js";
import BillingIcon from "./BillingIcon.vue";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";
import { envVariables } from "../../envVariables";

const emit = defineEmits(["no-payment-methods", "has-default-payment", "customer-id-created"]);
const stripeKey = computed(() => envVariables.stripeKey);
const stripeLoaded = ref(false);
const customer: any = reactive({ name: "", email: "", payment_methods: [] });
const store = useStore();
const consumerData = computed(() => store.getters["customer/getCustomer"]);
const card = ref();
const elms = ref();
const errorMessage = ref();
const alertRender = ref(false);
const addNewCard = ref(false);
const namespace = computed(() => store.getters["namespaces/get"]);
const instanceOptions = ref<StripeConstructorOptions>({
});

const elementsOptions = ref<StripeElementsOptions>({
  fonts: [{ cssSrc: "https://fonts.googleapis.com/css2?family=Lato" }],
  appearance: {
    theme: "night",
  },
  locale: "en",
  loader: "auto",
  mode: "setup",
  currency: "usd",
  payment_method_creation: "manual",
  // https://stripe.com/docs/js/elements_object/create#stripe_elements-options
});
const cardOptions = ref<StripeCardElementOptions>({
  style: {
    base: {
      iconColor: "#c4f0ff",
      color: "#fff",
      backgroundColor: "#22252B",
      lineHeight: "50px",
      padding: "12px",
      fontWeight: "400",
      fontFamily: "Roboto, Open Sans, Segoe UI, sans-serif",
      fontSize: "16px",
      fontSmoothing: "antialiased",
      ":-webkit-autofill": {
        color: "white",
      },
      "::placeholder": {
        color: "white",
      },
    },
    invalid: {
      iconColor: "#FFC7EE",
      color: "#FFC7EE",
    },
  },
  hidePostalCode: true,
});

const fetchData = async () => {
  await store.dispatch("customer/fetchCustomer");
  const customerDetails = consumerData.value.data;
  customer.name = customerDetails.name;
  customer.email = customerDetails.email;
  customer.payment_methods = customerDetails.payment_methods;
  if (customer.payment_methods == null || customer.payment_methods.length === 0) {
    emit("no-payment-methods");
  } else {
    emit("has-default-payment");
  }
};

const savePayment = async () => {
  const cardElement = card.value.stripeElement;
  elms.value.instance.createPaymentMethod({
    type: "card",
    card: cardElement,
  })
    .then(async (result: any) => {
      try {
        const id: string = result.paymentMethod.id || "";
        await store.dispatch("customer/attachPaymentMethod", id);
        await fetchData();
        alertRender.value = false;
        addNewCard.value = false;
      } catch (error: any) {
        alertRender.value = true;
        const errorMessages = {
          card_declined: "Your payment method was declined, check if your card is valid or have sufficient funds",
          expired_card: "Your payment was declined because the card has expired.",
          incorrect_cvc: "Your payment was declined due to an incorrect CVC.",
          processing_error: "Your payment was declined due to a processing error.",
          incorrect_number: "Your payment was declined because the card number is incorrect.",
          default: "An error occurred during payment processing.",
        };
        errorMessage.value = errorMessages[error.code] || errorMessages.default;
      }
    });
};

const setDefaultPayment = async (id: string) => {
  try {
    await store.dispatch("customer/setDefaultPaymentMethod", id).then(async () => {
      await fetchData();
    });
  } catch (error) {
    handleError(error);
  }
};

const deletePaymentMethod = async (id: string) => {
  try {
    await store.dispatch("customer/detachPaymentMethod", id).then(async () => {
      await fetchData();
    });
  } catch (error) {
    handleError(error);
  }
};

onMounted(async () => {
  const tenant = computed(() => localStorage.getItem("tenant"));
  await store.dispatch("namespaces/get", tenant.value);
  if (namespace.value.billing == null || namespace.value.billing.customer_id === "") {
    try {
      await store.dispatch("customer/createCustomer");
      emit("customer-id-created");
    } catch (error) {
      handleError(error);
    }
  }
  await fetchData();
});

onBeforeMount(() => {
  const stripePromise = loadStripe(stripeKey.value);
  stripePromise.then(() => {
    stripeLoaded.value = true;
  });
});
</script>

<style scoped>
.card:hover,
.card:focus {
  border-left: 5px solid #7284d0;
  border-right: 5px solid #7284d0;
  transition: ease-in-out 200ms;
}
.content-card {
  max-height: 45vh;
  overflow: auto;
}

::-webkit-scrollbar {
    width: 6px;
  }
    ::-webkit-scrollbar-track {
      background-color: rgb(255 255 255 / 10%);
      border-radius: 10px;
    }
    ::-webkit-scrollbar-thumb {
      background-color: rgb(0 0 0 / 80%);
      border-radius: 10px;
    }
</style>
