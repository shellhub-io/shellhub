import { Module } from "vuex";
import axios, { AxiosError } from "axios";
import { State } from "..";
import * as apiBilling from "../api/billing";
import { ICustomer } from "@/interfaces/ICustomer";
import handleError from "@/utils/handleError";

export interface CustomerState {
  customer: ICustomer;
}

type errorResponseData = {
  message: string;
};

export function isAxiosError<ResponseType>(error: unknown): error is AxiosError<ResponseType> {
  return axios.isAxiosError(error);
}

export const customer: Module<CustomerState, State> = {
  namespaced: true,
  state: {
    customer: {} as ICustomer,
  },
  getters: {
    getCustomer: (state) => state.customer,
    hasPaymentMethods: (state) => state.customer.payment_methods || false,
  },
  mutations: {
    setCustomer: (state, customer: ICustomer) => {
      state.customer = customer;
    },
  },
  actions: {
    fetchCustomer: async ({ commit }) => {
      try {
        const customer = await apiBilling.getCustomer();
        commit("setCustomer", customer);
      } catch (error) {
        handleError(error);
      }
    },
    createCustomer: async ({ commit }) => {
      try {
        const customer = await apiBilling.createCustomer();
        commit("setCustomer", customer);
      } catch (error) {
        handleError(error);
      }
    },
    attachPaymentMethod: async ({ commit }, id: string) => {
      try {
        await apiBilling.attachPaymentMethod(id);
        commit("setCustomer", customer);
      } catch (error: any) {
        if (isAxiosError<errorResponseData>(error)) {
          throw error.response?.data;
        } else {
          handleError(error);
        }
      }
    },
    detachPaymentMethod: async ({ commit }, id: string) => {
      try {
        await apiBilling.detachPaymentMethod(id);
        commit("setCustomer", customer);
      } catch (error) {
        handleError(error);
      }
    },
    createSubscription: async ({ commit }) => {
      try {
        await apiBilling.createSubscription();
        commit("setCustomer", customer);
      } catch (error) {
        if (isAxiosError(error)) {
          throw error.response?.status;
        } else {
          handleError(error);
        }
      }
    },
    setDefaultPaymentMethod: async ({ commit }, id: string) => {
      try {
        await apiBilling.setDefaultPaymentMethod(id);
        commit("setCustomer", customer);
      } catch (error) {
        handleError(error);
      }
    },
  },
};
