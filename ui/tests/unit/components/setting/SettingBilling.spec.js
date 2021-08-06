import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import flushPromises from 'flush-promises';
import SettingBilling from '@/components/setting/SettingBilling';

describe('SettingBilling', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const owner = true;

  const inactiveBilling = {
    active: false,
    current_period_end: 0,
    customer_id: '',
    subscription_id: '',
    payment_method_id: '',
  };

  const activeBilling = {
    active: true,
    current_period_end: 12121,
    customer_id: 'cus_123',
    subscription_id: 'subs_123',
    payment_method_id: 'pm_123',
  };

  const stripeData = {
    latest_invoice: { amount_due: 0, amount_paid: 0 },
    upcoming_invoice: { amount_due: 0, amount_paid: 0 },
    product_description: 'Premium usage',
    card: {
      brand: 'visa', exp_year: 2024, exp_month: 4, last4: '4242',
    },
  };

  const billingData = {
    info: {
      description: stripeData.product_description,
      latestPaymentDue: stripeData.latest_invoice.amount_due,
      latestPaymentPaid: stripeData.latest_invoice.amount_paid,
      nextPaymentDue: stripeData.upcoming_invoice.amount_due,
      nextPaymentPaid: stripeData.upcoming_invoice.amount_paid,
    },
    card: stripeData.card,
  };

  const storeOwnerWithoutSubscription = new Vuex.Store({
    namespaced: true,
    state: {
      stateBilling: inactiveBilling.active,
      billing: inactiveBilling,
      owner,
    },
    getters: {
      'billing/active': (state) => state.stateBilling,
      'billing/get': (state) => state.billing,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'billing/getSubscription': () => stripeData,
      'namespaces/get': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  const storeOwnerWithSubscription = new Vuex.Store({
    namespaced: true,
    state: {
      stateBilling: activeBilling.active,
      billing: activeBilling,
      owner,
    },
    getters: {
      'billing/active': (state) => state.stateBilling,
      'billing/get': (state) => state.billing,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'billing/getSubscription': () => stripeData,
      'billing/cancelSubscription': () => {},
      'namespaces/get': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  const storeNotOwner = new Vuex.Store({
    namespaced: true,
    state: {
      stateBilling: activeBilling.active,
      billing: activeBilling,
      owner: !owner,
    },
    getters: {
      'billing/active': (state) => state.stateBilling,
      'billing/get': (state) => state.billing,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'billing/getSubscription': () => {},
      'billing/cancelSubscription': () => {},
      'namespaces/get': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is the unrealized owner of the subscription.
  ///////

  describe('Namespace owner without subscription', () => {
    beforeEach(() => {
      wrapper = shallowMount(SettingBilling, {
        store: storeOwnerWithoutSubscription,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $stripe: {
            elements: () => ({
              create: () => ({
                mount: () => null,
              }),
            }),
          },
        },
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Compare data with default value', () => {
      expect(wrapper.vm.billingData).toEqual({ card: Object, info: Object });
      expect(wrapper.vm.renderData).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.active).toEqual(inactiveBilling.active);
      expect(wrapper.vm.billing).toEqual(inactiveBilling);
      expect(wrapper.vm.isOwner).toEqual(owner);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="settingOwnerInfo-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="subscriptionPaymentMethod-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="updatePaymentMethod-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="billingIcon-component"]').exists()).toBe(false);
    });
    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="content-div"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="subscriptionActive-div"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is subscription is already done.
  ///////

  describe('Namespace owner with subscription', () => {
    beforeEach(async () => {
      wrapper = shallowMount(SettingBilling, {
        store: storeOwnerWithSubscription,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $stripe: {
            elements: () => ({
              create: () => ({
                mount: () => null,
              }),
            }),
          },
        },
      });

      wrapper.setData({
        renderData: true,
        billingData,
      });
      await flushPromises();
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Compare data with default value', () => {
      expect(wrapper.vm.renderData).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.active).toEqual(activeBilling.active);
      expect(wrapper.vm.billing).toEqual(activeBilling);
      expect(wrapper.vm.isOwner).toEqual(owner);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="settingOwnerInfo-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="subscriptionPaymentMethod-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="updatePaymentMethod-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="billingIcon-component"]').exists()).toBe(true);
    });
    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="content-div"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="subscriptionActive-div"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(false);
    });
  });

  // ///////
  // // In this case, when the user doesn't owns the namespace.
  // ///////

  describe('Doesn\'t own the namespace', () => {
    beforeEach(async () => {
      wrapper = shallowMount(SettingBilling, {
        store: storeNotOwner,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $stripe: {
            elements: () => ({
              create: () => ({
                mount: () => null,
              }),
            }),
          },
        },
      });

      wrapper.setData({ renderData: true });
      await flushPromises();
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Compare data with default value', () => {
      expect(wrapper.vm.billingData).toEqual({ card: Object, info: Object });
      expect(wrapper.vm.renderData).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.active).toEqual(activeBilling.active);
      expect(wrapper.vm.billing).toEqual(activeBilling);
      expect(wrapper.vm.isOwner).toEqual(!owner);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="settingOwnerInfo-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="subscriptionPaymentMethod-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="updatePaymentMethod-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="billingIcon-component"]').exists()).toBe(false);
    });
    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="content-div"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="subscriptionActive-div"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(false);
    });
  });
});
