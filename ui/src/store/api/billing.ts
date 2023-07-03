import { ChoiceDevicesRequest } from "@/api/client";
import { billingApi } from "../../api/http";

export const getCustomer = async () => billingApi.getCustomer();

export const attachPaymentMethod = async (id: string) => billingApi.attachPaymentMethod({ id });

export const detachPaymentMethod = async (id: string) => billingApi.detachPaymentMethod({ id });

export const createSubscription = async () => billingApi.createSubscription();

export const setDefaultPaymentMethod = async (id: string) => billingApi.setDefaultPaymentMethod({ id });

export const createCustomer = async () => billingApi.createCustomer();

export const postDevicesChooser = async (data: ChoiceDevicesRequest) => billingApi.choiceDevices(data);

export const getSubscriptionInfo = async () => billingApi.getSubscription();

export const getDevicesMostUsed = async () => billingApi.getDevicesMostUsed();
