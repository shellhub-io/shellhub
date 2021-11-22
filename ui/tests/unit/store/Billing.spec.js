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

    const state = 'pending';

    store.commit('billing/setSubscription', data);

    ['status', 'customer', 'id'].map((v) => (
      Reflect.deleteProperty(data, v)
    ));

    expect(store.getters['billing/get']).toEqual({
      ...data,
      active,
      customer_id: customerId,
      subscription_id: subsId,
      state,
    });
    expect(store.getters['billing/active']).toEqual(active);
    expect(store.getters['billing/status']).toEqual(state);
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
      state: 'processed',
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

    const state = 'pending';

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
      state,
    });
    expect(store.getters['billing/active']).toEqual(!active);
    expect(store.getters['billing/status']).toEqual(state);
  });

  ///////
  // In this case, the setUpdatePaymentMethod is checked.
  ///////

  it('Verify initial state change for setUpdatePaymentMethod mutation', () => {
    const infoData = {
      info: {
        periodEnd: '2021-12-24T18:16:21Z',
        description: 'Shellhub',
        latestPaymentDue: 0,
        latestPaymentPaid: 0,
        nextPaymentDue: 0,
        nextPaymenPaid: 0,
      },
      defaultCard: {
        brand: 'visa',
        expYear: 2024,
        default: true,
        expMonth: 4,
        last4: '4042',
        id: 'pm_1JzQ80KJsksFHO6pREJA5TrK',
      },
      cards: [
        {
          brand: 'visa',
          expYear: 2024,
          default: true,
          expMonth: 4,
          last4: '4042',
          id: 'pm_1JzQ80KJsksFHO6pREJA5TrK',
        },
        {
          brand: 'visa',
          expYear: 2028,
          default: false,
          expMonth: 4,
          last4: '4042',
          id: 'pm_1JzQ80KJsksFHO6pREJA5TrG',
        },
        {
          brand: 'visa',
          expYear: 2029,
          default: false,
          expMonth: 4,
          last4: '4042',
          id: 'pm_1JzQ80KJsksFHO6pREJA5TrF',
        },
      ],
    };

    store.commit('billing/setGetSubscription', infoData);
    const updateChoiceId = 'pm_1JzQ80KJsksFHO6pREJA5TrF';
    const chosenCard = { ...infoData.cards[2], default: true };
    const expectedCards = [
      { ...infoData.cards[0], default: false },
      infoData.cards[1],
      chosenCard,
    ];
    store.commit('billing/setUpdatePaymentMethod', updateChoiceId);

    expect(store.getters['billing/getBillInfoData'].defaultCard).toEqual(chosenCard);
    expect(store.getters['billing/getBillInfoData'].cards).toEqual(expectedCards);
  });

  ///////
  // In this case, the setDeletePaymentMethod is checked.
  ///////

  it('Verify initial state change for setDeletePaymentMethod mutation', () => {
    const infoData = {
      info: {
        periodEnd: '2021-12-24T18:16:21Z',
        description: 'Shellhub',
        latestPaymentDue: 0,
        latestPaymentPaid: 0,
        nextPaymentDue: 0,
        nextPaymenPaid: 0,
      },
      defaultCard: {
        brand: 'visa',
        expYear: 2029,
        default: false,
        expMonth: 4,
        last4: '4042',
        id: 'pm_1JzQ80KJsksFHO6pREJA5TrF',
      },
      cards: [
        {
          brand: 'visa',
          expYear: 2029,
          default: false,
          expMonth: 4,
          last4: '4042',
          id: 'pm_1JzQ80KJsksFHO6pREJA5TrF',
        },
        {
          brand: 'visa',
          expYear: 2028,
          default: false,
          expMonth: 4,
          last4: '4042',
          id: 'pm_1JzQ80KJsksFHO6pREJA5TrG',
        },
        {
          brand: 'visa',
          expYear: 2028,
          default: false,
          expMonth: 4,
          last4: '4042',
          id: 'pm_1JzQ80KJsksFHO6pREJA5TrP',
        },
      ],
    };

    store.commit('billing/setGetSubscription', infoData);
    const updateChoiceId = 'pm_1JzQ80KJsksFHO6pREJA5TrF';
    const expectedCards = infoData.cards.slice(1);
    store.commit('billing/setDeletePaymentMethod', updateChoiceId);
    expect(store.getters['billing/getBillInfoData'].cards).toEqual(expectedCards);
    expect(store.getters['billing/getBillInfoData'].cards.length).toEqual(infoData.cards.length - 1);
    expect(store.getters['billing/getBillInfoData'].defaultCard).toEqual(infoData.defaultCard);
    expect(store.getters['billing/getBillInfoData'].info).toEqual(infoData.info);
  });
});
