import Vue from 'vue';
import * as apiBilling from '@/store/api/billing';
import ns from '@/store/modules/namespaces';

export default {
  namespaced: true,

  state: ns.state,

  getters: {
    get: (state) => state.billing,
    active: (state) => state.billing.active || false,
    status: (state) => state.billing.state || 'inactive',
  },

  mutations: {
    setSubscription: (state, data) => {
      Vue.set(state, 'billing', {
        active: data.status === 'active',
        current_period_end: data.current_period_end,
        customer_id: data.customer.id,
        payment_method_id: data.payment_method_id,
        subscription_id: data.id,
        state: 'pending',
      });
    },

    setPaymentMethod: (state, data) => {
      Vue.set(state, 'billing', {
        ...state.billing,
        state: 'processed',
        payment_method_id: data.pm,
      });
    },

    deactivateSubscription: (state) => {
      Vue.set(state, 'billing', {
        ...state.billing,
        state: 'pending',
        active: false,
      });
    },
  },

  actions: {
    subscritionPaymentMethod: async (context, data) => {
      const res = await apiBilling.subscritionPaymentMethod(data);
      if (res.status === 200) {
        context.commit('setSubscription', {
          ...res.data,
          ...data,
        });
      }
    },

    getSubscription: async () => {
      const res = await apiBilling.getSubscriptionInfo();
      if (res.status === 200) {
        return res.data;
      }
      return new Error('failed to get subscrition');
    },

    updatePaymentMethod: async (context, data) => {
      const res = await apiBilling.updatePaymentMethod(data);
      if (res.status === 200) {
        context.commit('setPaymentMethod', data);
      }
    },

    cancelSubscription: async (context) => {
      const res = await apiBilling.cancelSubscription();
      if (res.status === 200) {
        context.commit('deactivateSubscription');
      }
    },
  },
};
