import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingBilling from '@/components/setting/SettingBilling';

describe('SettingBilling', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  const stripeData = {
    latest_invoice: { amount_due: 0, amount_paid: 0 },
    upcoming_invoice: { amount_due: 0, amount_paid: 0 },
    product_description: 'Premium usage',
    card: {
      brand: 'visa', exp_year: 2024, exp_month: 4, last4: '4242',
    },
  };

  describe('Renders component according to billing instance', () => {
    const tests = [
      {
        description: 'User not owner',
        templates: {
          'settingOwnerInfo-component': true,
          'content-div': false,
        },
        computed: {
          active: false,
          isOwner: false,
        },
        dataAndProps: {
          renderData: false,
        },
        owner: false,
        instance: {
          active: false,
          state: 'inactive',
          current_period_end: 0,
          customer_id: '',
          subscription_id: '',
          payment_method_id: '',
        },
      },
      {
        description: 'Create subscription',
        templates: {
          'subscriptionPaymentMethod-component': true,
          'freePlan-div': true,
          'premiumPlan-div': false,
          'subscriptionActive-div': false,
          'updatePaymentMethod-component': false,
          'billingIcon-component': false,
          'cancel-div': false,
        },
        computed: {
          active: false,
          isOwner: true,
          state: 'inactive',
        },
        dataAndProps: {
          renderData: false,
        },
        owner: true,
        instance: {
          active: false,
          state: 'inactive',
          current_period_end: 0,
          customer_id: '',
          subscription_id: '',
          payment_method_id: '',
        },
      },
      {
        description: 'Pending request',
        templates: {
          'subscriptionPaymentMethod-component': false,
          'pendingRetrial-div': true,
          'freePlan-div': false,
          'premiumPlan-div': false,
          'subscriptionActive-div': false,
          'updatePaymentMethod-component': false,
          'billingIcon-component': false,
          'cancel-div': false,
          'activeLoading-div': false,
        },
        owner: true,
        computed: {
          active: true,
          isOwner: true,
          state: 'pending',
        },
        dataAndProps: {
          renderData: true,
        },
        instance: {
          active: true,
          state: 'pending',
          current_period_end: 0,
          customer_id: 'cus_123',
          subscription_id: 'sub_123',
          payment_method_id: 'pm_123',
        },
      },
      {
        description: 'Premium usage',
        templates: {
          'subscriptionPaymentMethod-component': false,
          'freePlan-div': false,
          'premiumPlan-div': true,
          'subscriptionActive-div': true,
          'updatePaymentMethod-component': true,
          'billingIcon-component': true,
          'cancel-div': true,
          'activeLoading-div': false,
        },
        computed: {
          active: true,
          isOwner: true,
          state: 'processed',
        },
        dataAndProps: {
          renderData: true,
        },
        owner: true,
        instance: {
          active: true,
          state: 'processed',
          current_period_end: 0,
          customer_id: 'cus_123',
          subscription_id: 'sub_123',
          payment_method_id: 'pm_123',
        },
      },
    ];

    const storeVuex = (billing, owner) => new Vuex.Store({
      namespaced: true,
      state: {
        billing,
        owner,
      },
      getters: {
        'billing/active': (state) => state.billing.active || false,
        'billing/status': (state) => state.billing.state || 'inactive',
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

    const WrapperArray = tests.map((el) => shallowMount(SettingBilling, {
      store: storeVuex(el.instance, el.owner),
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
    }));

    WrapperArray.forEach((el, i) => {
      ///////
      // Component Rendering
      //////

      it(`Is a Vue instance - ${tests[i].description}`, () => {
        expect(el).toBeTruthy();
      });
      it(`Renders the component - ${tests[i].description}`, () => {
        expect(el.html()).toMatchSnapshot();
      });

      const { templates, computed, dataAndProps } = tests[i];

      //////
      // HTML validation
      //////

      it(`Renders template - ${tests[i].description}`, () => {
        Reflect.ownKeys(templates).forEach((k) => {
          expect(el.find(`[data-test="${k}"]`).exists()).toBe(templates[k]);
        });
      });

      ///////
      // Computed properties checking
      //////

      it(`Process data in the computed - ${tests[i].description}`, () => {
        Reflect.ownKeys(computed).forEach((k) => {
          expect(el.vm[k]).toBe(computed[k]);
        });
      });

      ///////
      // Data and Props checking
      //////

      it(`Compare data with the dafault value - ${tests[i].description}`, () => {
        Reflect.ownKeys(dataAndProps).forEach((k) => {
          expect(el.vm[k]).toBe(dataAndProps[k]);
        });
      });
    });
  });
});
