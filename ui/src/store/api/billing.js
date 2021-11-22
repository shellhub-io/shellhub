import http from '@/store/helpers/http';

export const subscritionPaymentMethod = async (data) => http().post('/billing/subscription', data);

export const postDevicesChooser = async (data) => http().post('/billing/devices-choice', data);

export const getSubscriptionInfo = async () => http().get('/billing/subscription');

export const getDevicesMostUsed = async () => http().get('/billing/devices-most-used');

export const updatePaymentMethod = async (id) => http().patch(`/billing/${id}/payment-method`);

export const addPaymentMethod = async (id) => http().post(`/billing/${id}/payment-method`);

export const removePaymentMethod = async (id) => http().delete(`/billing/${id}/payment-method`);

export const cancelSubscription = async () => http().delete('/billing/subscription');
