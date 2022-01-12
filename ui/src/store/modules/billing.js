import Vue from 'vue';
import * as apiBilling from '@/store/api/billing';
import ns from '@/store/modules/namespaces';
import infoExtract from '@/helpers/billInfoExtract';

export default {
  namespaced: true,

  state: ns.state,

  getters: {
    get: (state) => state.billing,
    active: (state) => state.billing.active || false,
    status: (state) => state.billing.state || 'inactive',
    getBillInfoData: (state) => state.billInfoData,
    getInvoices: (state) => state.invoices,
    getInvoicesLength: (state) => state.invoicesLength,
    getPerPage: (state) => state.defaultPerPage,
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

    setGetSubscription: (state, data) => {
      const perPage = state.defaultPerPage;

      Vue.set(state, 'billInfoData', data);
      Vue.set(state, 'invoices', data.invoices.slice(0, perPage));
      Vue.set(state, 'invoicesLength', data.invoices.length);
    },

    setPaymentMethod: (state, data) => {
      Vue.set(state, 'billing', {
        ...state.billing,
        state: 'processed',
        payment_method_id: data.pm,
      });
    },

    setDeletePaymentMethod: (state, id) => {
      const { cards } = state.billInfoData;
      const newCards = cards.filter((c) => c.id !== id);

      Vue.set(state, 'billInfoData', {
        ...state.billInfoData,
        cards: newCards,
      });
    },

    setUpdatePaymentMethod: (state, id) => {
      const { defaultCard, cards } = state.billInfoData;

      const index = cards.findIndex((c) => c.id === id);
      const prevDefault = cards.find((c) => c.id === defaultCard.id);
      cards[index].default = true;
      prevDefault.default = false;

      Vue.set(state, 'billInfoData', {
        ...state.billInfoData,
        cards,
        defaultCard: cards[index],
      });
    },

    setPagination: (state, data) => {
      const { perPage, page } = data;
      const { invoices } = state.billInfoData;
      Vue.set(state, 'invoices', invoices.slice((page - 1) * perPage, page * perPage));
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

    getSubscription: async (context) => {
      const res = await apiBilling.getSubscriptionInfo();
      if (res.status === 200) {
        const { billing } = context.state;
        const data = infoExtract(res.data, billing.current_period_end);
        context.commit('setGetSubscription', data);
      }
      return new Error('failed to get subscrition');
    },

    getPagination: (context, data) => {
      context.commit('setPagination', data);
    },

    updatePaymentMethod: async (context, id) => {
      const res = await apiBilling.updatePaymentMethod(id);
      if (res.status === 200) {
        context.commit('setUpdatePaymentMethod', id);
      }
    },

    addPaymentMethod: async (context, data) => {
      const res = await apiBilling.addPaymentMethod(data);
      if (res.status === 200) {
        context.commit('setPaymentMethod', data);
      }
    },

    removePaymentMethod: async (context, id) => {
      const res = await apiBilling.removePaymentMethod(id);
      if (res.status === 200) {
        context.commit('setDeletePaymentMethod', id);
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
