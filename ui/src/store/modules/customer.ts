import { defineStore } from "pinia";
import { ref } from "vue";
import axios from "axios";
import * as billingApi from "../api/billing";
import type { ICustomer } from "@/interfaces/ICustomer";
import handleError from "@/utils/handleError";

const useCustomerStore = defineStore("customer", () => {
  const customer = ref<ICustomer>({} as ICustomer);

  const fetchCustomer = async () => {
    try {
      const { data } = await billingApi.getCustomer();
      customer.value = data as ICustomer;
    } catch (error) {
      handleError(error);
    }
  };

  const createCustomer = async () => {
    try {
      await billingApi.createCustomer();
    } catch (error) {
      handleError(error);
    }
  };

  const attachPaymentMethod = async (id: string) => {
    try {
      await billingApi.attachPaymentMethod(id);
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw error.response?.data;
      } else {
        handleError(error);
      }
    }
  };

  const detachPaymentMethod = async (id: string) => {
    try {
      await billingApi.detachPaymentMethod(id);
    } catch (error) {
      handleError(error);
    }
  };

  const createSubscription = async () => {
    try {
      await billingApi.createSubscription();
    } catch (error) {
      if (axios.isAxiosError(error)) throw error;
      handleError(error);
    }
  };

  const setDefaultPaymentMethod = async (id: string) => {
    try {
      await billingApi.setDefaultPaymentMethod(id);
    } catch (error) {
      handleError(error);
    }
  };

  return {
    customer,
    fetchCustomer,
    createCustomer,
    attachPaymentMethod,
    detachPaymentMethod,
    createSubscription,
    setDefaultPaymentMethod,
  };
});

export default useCustomerStore;
