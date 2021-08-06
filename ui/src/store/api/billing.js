import http from '@/store/helpers/http';

export const subscritionPaymentMethod = async (data) => http().post('/billing/subscription', data);

export const getSubscriptionInfo = async () => http().get('/billing/subscription');

export const updatePaymentMethod = async (data) => http().patch('/billing/payment-method', data);

export const cancelSubscription = async () => http().delete('/billing/subscription');
