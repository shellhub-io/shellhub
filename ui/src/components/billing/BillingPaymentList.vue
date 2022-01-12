<template>
  <fragment>
    <div class="mt-5">
      <v-data-table
        class="elevation-0"
        :headers="headers"
        :items="paymentList"
        hide-default-footer
        data-test="dataTable-field"
      >
        <template #[`item.brand`]="{ item }">
          <BillingIcon :icon-name="item.brand" />
        </template>

        <template #[`item.last4`]="{ item }">
          {{ item.last4 }}
        </template>

        <template #[`item.expdate`]="{ item }">
          {{ item.expMonth }} / {{ item.expYear }}
        </template>

        <template #[`item.actions`]="{ item }">
          <v-menu
            v-if="!item.default"
            offset-y
          >
            <template #activator="{ on, attrs }">
              <v-icon
                small
                class="icons"
                v-bind="attrs"
                v-on="on"
              >
                mdi-dots-horizontal
              </v-icon>
            </template>

            <v-list>
              <v-list-item @click.stop="updatePaymentMethod(item.id)">
                <v-icon class="mr-2">
                  mdi-pencil
                </v-icon>

                <v-list-item-title>
                  Make default
                </v-list-item-title>
              </v-list-item>

              <v-list-item @click.stop="deletePaymentMethod(item.id)">
                <v-icon class="mr-2">
                  mdi-delete
                </v-icon>

                <v-list-item-title>
                  Remove
                </v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>

          <div v-else>
            Default
          </div>
        </template>
      </v-data-table>
    </div>
  </fragment>
</template>

<script>

import BillingIcon from '@/components/billing/BillingIcon';

export default {
  name: 'BillingPaymentList',

  components: {
    BillingIcon,
  },

  props: {
    cards: {
      type: Array,
      required: true,
    },
  },

  data() {
    return {
      headers: [
        {
          text: 'Brand',
          value: 'brand',
          align: 'center',
          sortable: false,
        },
        {
          text: 'Exp. Date',
          value: 'expdate',
          align: 'center',
          sortable: false,
        },
        {
          text: 'Ends with',
          value: 'last4',
          align: 'center',
          sortable: false,
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'center',
          sortable: false,
        },
      ],
    };
  },

  computed: {
    paymentList() {
      return this.$props.cards;
    },
  },

  methods: {
    async updatePaymentMethod(paymentMethodId) {
      try {
        await this.$store.dispatch('billing/updatePaymentMethod', paymentMethodId);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.updateSubscription);
        this.$emit('update');
      } catch (error) {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.updateSubscription);
      }
    },

    async deletePaymentMethod(paymentMethodId) {
      try {
        await this.$store.dispatch('billing/removePaymentMethod', paymentMethodId);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.updateSubscription);
        this.$emit('update');
      } catch (error) {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deletePaymentMethod);
      }
    },
  },
};

</script>
