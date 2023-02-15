<template>
  <v-btn class="bg-primary" data-test="show-btn" @click="dialog = !dialog" v-bind="$attrs">
    {{ actionButton(typeOperation) }}
  </v-btn>

  <v-dialog v-model="dialog" max-width="600">
    <v-card
      v-model="dialog"

      class="bg-v-theme-surface"
      data-test="BillingDialogPaymentMethod-dialog"
    >
      <v-card-title class="bg-primary pa-4" data-test="text-cardTitle">
        {{ typeTitle(typeOperation) }}
      </v-card-title>

      <v-card-text class="mt-2 mb-3 pb-1">
        <div v-if="typeOperation === 'subscription'">
          <div data-test="subscription-description" class="text-high-emphasis">
            <h4 class="text-body-1 font-weight-bold">
              Subscribe to premium plan:
            </h4>
            <p data-test="subscription-message" class="text-medium-emphasis">
              The subscription is charged monthly, based on the number of
              devices you have in your namespace.
            </p>
            <div class="mt-4">
              <b> Estimated cost: </b>
              <span>
                {{ currentQuantity }} devices :
                {{ priceEstimator(currentQuantity) }} / month
              </span>
            </div>
          </div>
        </div>

        <v-card class="paymentForm mt-6 pa-3 bg-white">
          <StripeElements
            v-if="stripeLoaded"
            v-slot="{ elements }"
            ref="elms"
            :stripe-key="stripeKey"
            :instance-options="instanceOptions"
            :elements-options="elementsOptions"
          >
            <StripeElement
              ref="card"
              :elements="elements"
              :options="cardOptions"
            />
          </StripeElements>
        </v-card>

        <div ref="card-element-errors" class="card-errors mt-4" role="alert" />

        <v-spacer />

        <v-row class="mt-2">
          <v-spacer />
          <v-col md="auto" class="ml-auto" />
        </v-row>
      </v-card-text>
      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="cancel-btn" @click="dialog = !dialog">
          Close
        </v-btn>

        <v-btn
          variant="text"
          data-test="confirm-btn"
          :disabled="lockButton"
          @click="doAction()"
        >
          confirm
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, ref, computed, onBeforeMount, onMounted } from "vue";
import { StripeElements, StripeElement } from "vue-stripe-js";
import { loadStripe } from "@stripe/stripe-js";
import formatCurrency from "@/utils/currency";
import { useStore } from "../../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { envVariables } from "@/envVariables";

export default defineComponent({
  props: {
    typeOperation: {
      type: String,
      default: "subscription",
      validator: (value: string) => ["subscription", "update"].includes(value),
    },
  },
  components: {
    StripeElements,
    StripeElement,
  },
  emits: ["update"],
  setup(props, ctx) {
    const dialog = ref(false);
    const lockButton = ref(false);
    const elementError = ref("");
    const store = useStore();

    const currentQuantity = computed(
      () => store.getters["stats/stats"].registered_devices,
    );

    const stripeKey = computed(() => envVariables.stripeKey);
    const stripeLoaded = ref(false);
    const card = ref();
    const elms = ref();

    const instanceOptions = ref({
      // https://stripe.com/docs/js/initializing#init_stripe_js-options
    });
    const elementsOptions = ref({
      // https://stripe.com/docs/js/elements_object/create#stripe_elements-options
    });
    const cardOptions = ref({
      // https://stripe.com/docs/stripe.js#element-options
      value: {
        postalCode: "",
      },
    });

    onMounted(() => {
      try {
        store.dispatch("stats/get");
      } catch (error: any) {
        throw new Error(error);
      }
    });

    onBeforeMount(() => {
      const stripePromise = loadStripe(stripeKey.value || "");
      stripePromise.then(() => {
        stripeLoaded.value = true;
      });
    });

    const displayError = (e: any) => {
      if (e.error) {
        elementError.value = e.error.message;
      } else {
        elementError.value = e;
      }
    };

    const showError = (e: any) => {
      elementError.value = e.response.data;
    };

    const priceEstimator = (n: number) => {
      let sumPrice = 0;

      const ranges = [
        3,
        10,
        25,
        40,
        55,
        70,
        85,
        100,
        115,
        130,
        145,
        160,
        175,
        190,
        Infinity,
      ];
      const ks = [
        3, 2.91, 2.82, 2.74, 2.66, 2.58, 2.5, 2.42, 2.35, 2.28, 2.21, 2.15,
        2.08, 2.02, 2.0,
      ];

      const tiers = Array.from({ length: ks.length - 1 }, (_, i) => ({
        begin: ranges[i],
        upTo: ranges[i + 1],
        k: ks[i],
      }));

      tiers.forEach((t) => {
        if (n > t.begin) {
          sumPrice += (n <= t.upTo ? n - t.begin : t.upTo - t.begin) * t.k;
        }
      });
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      return formatCurrency(String(sumPrice * 100));
    };

    const actionButton = (type: string) => {
      if (type === "subscription") {
        return "subscribe";
      }

      if (type === "update") {
        return "add card";
      }

      return type;
    };

    const typeTitle = (type: string) => {
      switch (type) {
        case "subscription":
          return "Create subscription";
        case "update":
          return "Add payment method";
        default:
          return "Operation not found";
      }
    };

    const subscriptionPaymentMethod = async () => {
      const cardElement = card.value.stripeElement;
      const result = await elms.value.instance.createPaymentMethod({
        type: "card",
        card: cardElement,
      });
      if (result) {
        try {
          await store.dispatch("billing/subscritionPaymentMethod", {
            payment_method_id: result.paymentMethod.id,
          });
          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.subscription,
          );
        } catch (error: any) {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.subscription,
          );
          const { status } = error.response;
          if (status === 400 || status === 423) {
            showError(error);
          }
        }
      } else {
        displayError(result.error);
      }

      lockButton.value = false;
    };

    const updatePaymentMethod = async () => {
      const cardElement = card.value.stripeElement;
      const result = await elms.value.instance.createPaymentMethod({
        type: "card",
        card: cardElement,
      });

      if (result) {
        try {
          await store.dispatch(
            "billing/addPaymentMethod",
            result.paymentMethod.id,
          );
          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.updateSubscription,
          );
          ctx.emit("update");
          dialog.value = false;
        } catch (error: any) {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.updatePaymentMethod,
          );

          const { status } = error.response;
          if (status === 400 || status === 423) {
            showError(error);
          }
          throw new Error(error);
        }
      } else {
        displayError(result.error);
      }

      lockButton.value = false;
    };

    const doAction = async () => {
      lockButton.value = true;
      switch (props.typeOperation) {
        case "subscription":
          await subscriptionPaymentMethod();
          break;
        case "update":
          await updatePaymentMethod();
          break;
        default:
          lockButton.value = false;
      }
    };

    return {
      dialog,
      lockButton,
      currentQuantity,
      actionButton,
      typeTitle,
      priceEstimator,
      doAction,
      stripeLoaded,
      stripeKey,
      instanceOptions,
      elementsOptions,
      cardOptions,
      elms,
      card,
      elementError,
    };
  },
});
</script>
