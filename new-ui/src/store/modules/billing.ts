/* eslint-disable */
import { Module } from "vuex";
import { State } from "./../index";
import * as apiBilling from "../api/billing";
import { namespaces, NamespacesState } from "./namespaces";
import infoExtract from "../../utils/billInfoExtract";
import { IBIllingDataCard } from "@/interfaces/IBilling";

export const billing: Module<NamespacesState, State> = {
  namespaced: true,
  state: namespaces.state,
  getters: {
    get: (state) => state.billing,
    active: (state) => state.billing.active || false,
    status: (state) => state.billing.state || "inactive",
    getBillInfoData: (state) => state.billInfoData,
    getInvoices: (state) => state.invoices,
    getInvoicesLength: (state) => state.invoicesLength,
    getPerPage: (state) => state.defaultPerPage,
  },

  mutations: {
    setSubscription: (state, data) => {
      state.billing = {
        ...state.billing,
        active: data.status === "active",
        current_period_end: data.current_period_end,
        customer_id: data.customer.id,
        payment_method_id: data.payment_method_id,
        subscription_id: data.id,
        state: "pending",
      };
    },

    setGetSubscription: (state, data) => {
      const perPage = state.defaultPerPage;

      state.billInfoData = data;
      state.invoices = data.invoices.slice(0, perPage);
      state.invoicesLength = data.invoices.length;
    },

    setPaymentMethod: (state, data) => {
      state.billing = {
        ...state.billing,
        state: "processed",
        payment_method_id: data.pm,
      };
    },

    setDeletePaymentMethod: (state, id) => {
      const { cards } = state.billInfoData;
      const newCards = cards.filter((c: any) => c.id !== id);

      state.billInfoData = {
        ...state.billInfoData,
        cards: newCards,
      };
    },

    setUpdatePaymentMethod: (state, id) => {
      const { defaultCard, cards } = state.billInfoData;

      const index = cards.findIndex((c: IBIllingDataCard) => c.id === id);
      const prevDefault = cards.find((c: IBIllingDataCard) => c.id === defaultCard.id);
      cards[index].default = true;
      if (prevDefault) {
        prevDefault.default = false;
      }
      state.billInfoData = {
        ...state.billInfoData,
        cards,
        defaultCard: cards[index],
      };
    },

    setPagination: (state, data) => {
      const { perPage, page } = data;
      const { invoices } = state.billInfoData;

      state.invoices = invoices.slice((page - 1) * perPage, page * perPage);
    },

    deactivateSubscription: (state) => {
      state.billing = {
        ...state.billing,
        state: "pending",
        active: false,
      };
    },
  },

  actions: {
    subscritionPaymentMethod: async (context, data) => {
      const { payment_method_id } = data;
      const res = await apiBilling.subscritionPaymentMethod(payment_method_id);
      if (res.status === 200) {
        context.commit("setSubscription", {
          ...res.data,
          ...data,
        });
      }
    },

    getSubscription: async (context) => {
      try {
        const res = await apiBilling.getSubscriptionInfo();
        if (res.status === 200) {
          const { billing } = context.state;
          const data = infoExtract(res.data, billing.current_period_end); // TODO
          context.commit("setGetSubscription", data);
        }
      } catch (error) {
        throw error;
      }
    },

    getPagination: (context, data) => {
      context.commit("setPagination", data);
    },

    updatePaymentMethod: async (context, id) => {
      try {
        const res = await apiBilling.updatePaymentMethod(id);
        if (res.status === 200) {
          context.commit("setUpdatePaymentMethod", id);
        }
      } catch (error) {
        throw error;
      }
    },

    addPaymentMethod: async (context, data) => {
      try {
        const res = await apiBilling.addPaymentMethod(data);
        if (res.status === 200) {
          context.commit("setPaymentMethod", data);
        }
      } catch (error) {
        throw error;
      }
    },

    removePaymentMethod: async (context, id) => {
      try {
        const res = await apiBilling.removePaymentMethod(id);
        if (res.status === 200) {
          context.commit("setDeletePaymentMethod", id);
        }
      } catch (error) {
        throw error;
      }
    },

    cancelSubscription: async (context) => {
      try {
        const res = await apiBilling.cancelSubscription();
        if (res.status === 200) {
          context.commit("deactivateSubscription");
        }
      } catch (error) {
        throw error;
      }
    },
  },
};
