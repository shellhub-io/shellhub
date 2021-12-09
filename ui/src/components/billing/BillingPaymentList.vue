<template>
  <fragment>
    <div class="mt-5">
      <v-row class="text-center mb-2">
        <v-col
          v-for="(item, index) in heading"
          :key="item.id"
          :data-test="item.id+'-div'"
          :cols="(index === 3) ? 4 : null"
        >
          <b>
            {{ item.title }}
          </b>
        </v-col>
      </v-row>

      <v-list class="mb-2">
        <v-list-item
          v-for="item in paymentList"
          :key="item.id"
          class="mb-1"
        >
          <v-row class="text-center">
            <v-col>
              <BillingIcon
                :icon-name="item.brand"
                :data-test="'icon-'+item.id+'-component'"
              />
            </v-col>
            <v-col
              :data-test="'exp-date-'+item.id+'-col'"
              class="pm-data"
            >
              <p>
                {{ item.expMonth }}/{{ item.expYear }}
              </p>
            </v-col>
            <v-col
              :data-test="'last4-'+item.id+'-col'"
              class="pm-data"
            >
              <p>
                {{ item.last4 }}
              </p>
            </v-col>
            <v-col
              cols="4"
              class="actions-column"
            >
              <div
                v-if="item.default"
                :data-test="'default-'+item.id+'-div'"
                class="ml-4 pm-text"
              >
                <p>
                  Default
                </p>
              </div>
              <div
                v-else
                class="ml-2"
                :data-test="'actions-'+item.id+'-div'"
              >
                <v-btn
                  outlined
                  class="mr-2"
                  @click="updatePaymentMethod(item.id)"
                >
                  <div>
                    <v-tooltip bottom>
                      <template #activator="{ on }">
                        <span v-on="on">
                          <v-icon v-on="on">
                            mdi-pencil
                          </v-icon>
                        </span>
                      </template>
                      <span>
                        Make default
                      </span>
                    </v-tooltip>
                  </div>
                </v-btn>
                <v-btn
                  outlined
                  class="mr-2"
                  @click="deletePaymentMethod(item.id)"
                >
                  <div>
                    <v-tooltip bottom>
                      <template #activator="{ on }">
                        <span v-on="on">
                          <v-icon
                            v-on="on"
                          >
                            delete
                          </v-icon>
                        </span>
                      </template>
                      <span>
                        Remove
                      </span>
                    </v-tooltip>
                  </div>
                </v-btn>
              </div>
            </v-col>
          </v-row>
        </v-list-item>
      </v-list>
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
      heading: [
        {
          id: 'brand',
          title: 'Brand',
        },
        {
          id: 'expdate',
          title: 'Exp. Date',
        },
        {
          id: 'last4',
          title: 'Ends with',
        },
        {
          id: 'actions',
          title: 'Actions',
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
