import store from '@/store';

describe('Billing', () => {
  const customerId = 'cus_123';
  const subsId = 'subs_123';
  const paymentMethodId = 'pm_456';
  const active = true;

  ///////
  // In this case, the default state of the variables is checked.
  ///////

  it('Return namespace default variables', () => {
    expect(store.getters['billing/get']).toEqual({});
    expect(store.getters['billing/active']).toEqual(!active);
  });

  ///////
  // In this case, the subscription is carried out.
  ///////

  it('Verify initial state change for setSubscription mutation', () => {
    const data = {
      status: 'active',
      current_period_end: 12121,
      customer: {
        id: customerId,
      },
      payment_method_id: 'pm_123',
      id: subsId,
    };

    store.commit('billing/setSubscription', data);

    ['status', 'customer', 'id'].map((v) => (
      Reflect.deleteProperty(data, v)
    ));

    expect(store.getters['billing/get']).toEqual({
      ...data,
      active,
      customer_id: customerId,
      subscription_id: subsId,
    });
    expect(store.getters['billing/active']).toEqual(active);
  });

  ///////
  // In this case, the update is carried out.
  ///////

  it('Verify initial state change for setPaymentMethod mutation', () => {
    const data = {
      status: 'active',
      current_period_end: 12121,
      customer: {
        id: customerId,
      },
      pm: paymentMethodId,
      id: subsId,
    };

    store.commit('billing/setPaymentMethod', data);

    ['status', 'customer', 'id', 'pm'].map((v) => (
      Reflect.deleteProperty(data, v)
    ));

    expect(store.getters['billing/get']).toEqual({
      ...data,
      active,
      customer_id: customerId,
      subscription_id: subsId,
      payment_method_id: paymentMethodId,
    });
    expect(store.getters['billing/active']).toEqual(active);
  });

  ///////
  // In this case, the deactivateSubscription is checked.
  ///////

  it('Verify initial state change for deactivateSubscription mutation', () => {
    const data = {
      status: 'active',
      current_period_end: 12121,
      customer: {
        id: customerId,
      },
      payment_method_id: 'pm_123',
      id: subsId,
    };

    store.commit('billing/deactivateSubscription');

    ['status', 'customer', 'id'].map((v) => (
      Reflect.deleteProperty(data, v)
    ));

    expect(store.getters['billing/get']).toEqual({
      ...data,
      active: !active,
      customer_id: customerId,
      subscription_id: subsId,
      payment_method_id: paymentMethodId,
    });
    expect(store.getters['billing/active']).toEqual(!active);
  });
});
