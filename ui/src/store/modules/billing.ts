/* eslint-disable */
import { Module } from "vuex";
import * as apiBilling from "../api/billing";
import { namespaces, NamespacesState } from "./namespaces";
import { State } from "..";

export const billing: Module<NamespacesState, State> = {
  namespaced: true,
  state: namespaces.state,
  getters: {
    get: (state) => state.billing,
    active: (state) => state.billing.active || false,
    status: (state) => state.billing.status || "inactive",
    invoices: (state) => state.billing.invoices || [],
  },

  mutations: {
    setSubscription: (state, data) => {
      state.billing = data;
    }, 
  },

  actions: {
    getSubscription: async (context) => {
      try {
        const res = await apiBilling.getSubscriptionInfo();
        if (res.status === 200) {
          context.commit("setSubscription", res.data);
        }
      } catch (error) {
        throw error;
      }
    },
  },
};