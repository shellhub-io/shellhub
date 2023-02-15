<template>
  <v-table class="bg-v-theme-surface">
    <thead>
      <tr>
        <th
          v-for="(head, i) in headers"
          :key="i"
          :class="head.align ? `text-${head.align}` : 'text-center'"
        >
          <span> {{ head.text }}</span>
        </th>
      </tr>
    </thead>
    <tbody v-if="paymentList.length">
      <tr v-for="(item, i) in paymentList" :key="i" data-test="dataTable-field">
        <td class="text-center">
          <BillingIcon :icon-name="item.brand" />
        </td>
        <td class="text-center">
          {{ item.last4 }}
        </td>
        <td class="text-center">{{ item.expMonth }} / {{ item.expYear }}</td>
        <td class="text-center">
          <v-menu location="bottom" scrim eager>
            <template v-slot:activator="{ props }">
              <v-chip v-bind="props" density="comfortable" size="small">
                <v-icon>mdi-dots-horizontal</v-icon>
              </v-chip>
            </template>
            <v-list class="bg-v-theme-surface" lines="two" density="compact">
              <v-list-item @click.stop="updatePaymentMethod(item.id)">
                <div class="d-flex align-center">
                  <v-icon class="mr-2"> mdi-pencil </v-icon>

                  <v-list-item-title> Make default </v-list-item-title>
                </div>
              </v-list-item>

              <v-list-item @click.stop="deletePaymentMethod(item.id)">
                <div class="d-flex align-center">
                  <v-icon class="mr-2"> mdi-delete </v-icon>

                  <v-list-item-title> Remove </v-list-item-title>
                </div>
              </v-list-item>
            </v-list>
          </v-menu>
        </td>
      </tr>
    </tbody>
    <div v-else class="mt-4">
      <p>No data avaliabe</p>
    </div>
  </v-table>
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";
import {
  INotificationsError,
  INotificationsSuccess,
} from "@/interfaces/INotifications";
import { useStore } from "../../store";
import BillingIcon from "./BillingIcon.vue";

export default defineComponent({
  props: {
    cards: {
      type: Array as any,
      required: true,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const store = useStore();
    const paymentList = computed(() => props.cards);
    const updatePaymentMethod = async (paymentMethodId: string) => {
      try {
        await store.dispatch("billing/updatePaymentMethod", paymentMethodId);
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.updateSubscription,
        );
        ctx.emit("update");
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.subscription,
        );
        throw new Error(error);
      }
    };
    const deletePaymentMethod = async (paymentMethodId: string) => {
      try {
        await store.dispatch("billing/removePaymentMethod", paymentMethodId);
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.updateSubscription,
        );
        ctx.emit("update");
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deletePaymentMethod,
        );
        throw new Error(error);
      }
    };
    return {
      paymentList,
      updatePaymentMethod,
      deletePaymentMethod,
      headers: [
        {
          text: "Brand",
          value: "brand",
          align: "center",
          sortable: false,
        },
        {
          text: "Exp. Date",
          value: "expdate",
          align: "center",
          sortable: false,
        },
        {
          text: "Ends with",
          value: "last4",
          align: "center",
          sortable: false,
        },
        {
          text: "Actions",
          value: "actions",
          align: "center",
          sortable: false,
        },
      ],
    };
  },
  components: { BillingIcon },
});
</script>
