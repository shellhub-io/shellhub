<template>
  <v-row>
    <v-col>
      <v-text-field
        v-model="customer.name"
        label="Name"
        disabled
        autocomplete="name"
        data-test="customer-name"
      />
    </v-col>
    <v-col>
      <v-text-field
        v-model="customer.email"
        label="E-mail"
        disabled
        autocomplete="email"
        data-test="customer-email"
      />
    </v-col>
  </v-row>
  <v-col class="ma-0 pa-0">
    <h4 data-test="credit-card-text">
      Your credit cards
    </h4>
  </v-col>
  <v-list
    v-if="!(customer.payment_methods == null)"
    data-test="payment-methods-list"
    nav
    bg-color="transparent"
    class="w-100 pa-0 pt-2 content-card"
  >
    <v-col
      v-for="(item, i) in customer.payment_methods"
      :key="i"
      class="pa-0 mt-2 mb-2"
    >
      <v-card
        :key="i"
        class="bg-v-theme-card"
        variant="flat"
      >
        <v-list-item
          :key="i"
          data-test="payment-methods-item"
          class="card"
          @click="setDefaultPayment(item.id)"
        >
          <v-row
            align="center"
            cols="12"
          >
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
            <v-col
              v-if="item.default === true"
              cols="3"
            >
              <v-chip density="comfortable">
                <b>default</b>
                <v-tooltip
                  activator="parent"
                  location="top"
                >
                  This payment method will be used on your subscription
                </v-tooltip>
              </v-chip>
            </v-col>
            <v-col
              v-else
              cols="3"
              class="d-flex flex-column align-end"
            >
              <v-btn
                variant="text"
                icon="mdi-delete"
                data-test="payment-methods-delete-btn"
                @click.stop="deletePaymentMethod(item.id)"
              />
            </v-col>
          </v-row>
        </v-list-item>
      </v-card>
    </v-col>
  </v-list>
  <v-col v-else>
    <v-card-subtitle>You don't have any registered cards yet, please add one</v-card-subtitle>
  </v-col>
  <v-row
    class="mt-0 pt-0 pt-0"
    cols="12"
  >
    <v-col
      cols="8"
      class="pr-0"
    >
      <StripeElements
        v-if="stripeLoaded && addNewCard"
        v-slot="{ elements }"
        ref="elms"
        :stripe-key="stripeKey"
        :instance-options="instanceOptions"
        :elements-options="elementsOptions"
      >
        <StripeElement
          ref="card"
          type="card"
          :elements="elements"
          :options="cardOptions"
        />
      </StripeElements>
    </v-col>
    <v-col
      cols="4"
      class="d-flex flex-lg-column align-center justify-center pt-0"
    >
      <v-btn
        :text="addNewCard ? 'Save card' : 'Add new card'"
        prepend-icon="mdi-credit-card-plus"
        color="primary"
        data-test="add-card-btn"
        @click="addNewCard ? savePayment() : addNewCard = true"
      />
    </v-col>
  </v-row>
  <v-row>
    <v-col>
      <v-alert
        v-if="alertRender"
        icon="$error"
        :text="errorMessage"
        type="error"
        data-test="alert-message"
        role="alert"
        aria-live="assertive"
      />
    </v-col>
  </v-row>
</template>

<script lang='ts' setup>
import { onBeforeMount, onMounted, ref, computed } from "vue";
import { loadStripe } from "@stripe/stripe-js";
import { StripeElements, StripeElement } from "vue-stripe-js";
import type {
  StripeConstructorOptions,
  StripeElementsOptions,
  StripeCardElementOptions,
  StripeCardElement,
  Stripe,
} from "@stripe/stripe-js";
import BillingIcon from "./BillingIcon.vue";
import handleError from "@/utils/handleError";
import { envVariables } from "@/envVariables";
import useCustomerStore from "@/store/modules/customer";
import useNamespacesStore from "@/store/modules/namespaces";

interface CustomError {
  code: string;
  message: string;
}

const emit = defineEmits(["no-payment-methods", "has-default-payment", "customer-id-created"]);
const stripeKey = computed(() => envVariables.stripeKey);
const stripeLoaded = ref(false);
const customerStore = useCustomerStore();
const namespacesStore = useNamespacesStore();
const customer = computed(() => customerStore.customer);
const card = ref<{ stripeElement: StripeCardElement }>();
const elms = ref<{ instance: Stripe }>();
const errorMessage = ref();
const alertRender = ref(false);
const addNewCard = ref(false);
const namespace = computed(() => namespacesStore.currentNamespace);
const instanceOptions = ref<StripeConstructorOptions>({});
const errorMessages: Record<string, string> = {
  card_declined: "Your payment method was declined, check if your card is valid or has sufficient funds",
  expired_card: "Your payment was declined because the card has expired.",
  incorrect_cvc: "Your payment was declined due to an incorrect CVC.",
  processing_error: "Your payment was declined due to a processing error.",
  incorrect_number: "Your payment was declined because the card number is incorrect.",
  default: "An error occurred during payment processing.",
};

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
  try {
    await customerStore.fetchCustomer();

    if (customer.value.payment_methods?.length === 0 || customer.value.payment_methods === null) {
      emit("no-payment-methods");
    } else {
      emit("has-default-payment");
    }
  } catch (error) {
    handleError(error);
  }
};

const savePayment = async () => {
  const cardElement = card.value?.stripeElement;
  if (!cardElement) return;
  await elms.value?.instance.createPaymentMethod({
    type: "card",
    card: cardElement,
  })
    .then(async (result) => {
      if (result.error) {
        alertRender.value = true;
        errorMessage.value = errorMessages[result.error.code || "default"] || errorMessages.default;
        return;
      }

      try {
        const id: string = result.paymentMethod?.id || "";
        await customerStore.attachPaymentMethod(id);
        await fetchData();
        alertRender.value = false;
        addNewCard.value = false;
      } catch (error) {
        alertRender.value = true;
        const isMessageError = (error: unknown): error is CustomError => typeof error === "object" && error !== null && "code" in error;
        errorMessage.value = isMessageError(error) ? errorMessages[error.code] : errorMessages.default;
      }
    });
};

const setDefaultPayment = async (id: string) => {
  try {
    await customerStore.setDefaultPaymentMethod(id).then(async () => {
      await fetchData();
    });
  } catch (error) {
    handleError(error);
  }
};

const deletePaymentMethod = async (id: string) => {
  try {
    await customerStore.detachPaymentMethod(id).then(async () => {
      await fetchData();
    });
  } catch (error) {
    handleError(error);
  }
};

onMounted(async () => {
  const tenant = computed(() => localStorage.getItem("tenant") as string);
  await namespacesStore.fetchNamespace(tenant.value);
  if (!namespace.value.billing || !namespace.value.billing?.customer_id) {
    try {
      await customerStore.createCustomer();
      emit("customer-id-created");
    } catch (error) {
      handleError(error);
    }
  }
  await fetchData();
});

onBeforeMount(async () => {
  const stripePromise = loadStripe(stripeKey.value);
  await stripePromise.then(() => {
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
