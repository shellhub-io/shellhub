import { billingApi } from "../../api/http";

export const subscritionPaymentMethod = async (data: any) =>
  billingApi.createSubscription({ payment_method_id: data });

export const postDevicesChooser = async (data: any) =>
  billingApi.choiceDevices(data);

export const getSubscriptionInfo = async () => billingApi.getSubscription();

export const getDevicesMostUsed = async () => billingApi.getDevicesMostUsed();

export const updatePaymentMethod = async (id: string) =>
  billingApi.updatePaymentMethod(id);

export const addPaymentMethod = async (id: string) =>
  billingApi.addPaymentMethod(id);

export const removePaymentMethod = async (id: string) =>
  billingApi.deattachPaymentMethod(id);

export const cancelSubscription = async () => billingApi.cancelSubscription();
