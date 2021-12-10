import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingBilling from '@/components/setting/SettingBilling';
import { actions, authorizer } from '../../../../src/authorizer';

describe('SettingBilling', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const role = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

  const stripeData = {
    latest_invoice: { amount_due: 0, amount_paid: 0 },
    upcoming_invoice: { amount_due: 0, amount_paid: 0 },
    product_description: 'Premium usage',
  };

  const infoData = {
    info: {
      periodEnd: '2021-12-24T18:16:21Z',
      description: 'Shellhub',
      latestPaymentDue: 0,
      latestPaymentPaid: 0,
      nextPaymentDue: 0,
      nextPaymenPaid: 0,
    },

    card: {
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

  const info2 = {
    periodEnd: '2021-12-24T18:16:21Z',
    description: 'Shellhub',
    latestPaymentDue: 0,
    latestPaymentPaid: 0,
    nextPaymentDue: 0,
    nextPaymenPaid: 0,
  };

  const card2 = {
    brand: 'visa',
    expYear: 2024,
    default: true,
    expMonth: 4,
    last4: '4042',
    id: 'pm_123',
  };

  // describe('Renders component according to billing instance', () => {
  const tests = [
    {
      description: 'Create subscription',
      computed: {
        active: false,
        state: 'inactive',
      },
      data: {
        renderData: false,
        action: 'subscribe',
      },
      instance: {
        active: false,
        state: 'inactive',
        current_period_end: 0,
        customer_id: '',
        subscription_id: '',
        payment_method_id: '',
      },
      template: {
        'subscriptionPaymentMethod-component': true,
        'freePlan-div': true,
        'premiumPlan-div': false,
        'subscriptionActive-div': false,
        'updatePaymentMethod-component': false,
        'paymentMethods-component': false,
        'cancel-div': false,
      },
    },
    {
      description: 'Pending request',
      owner: true,
      computed: {
        active: true,
        state: 'pending',
      },
      data: {
        renderData: true,
        action: 'subscribe',
      },
      instance: {
        active: true,
        state: 'pending',
        current_period_end: 0,
        customer_id: 'cus_123',
        subscription_id: 'sub_123',
        payment_method_id: 'pm_123',
      },
      template: {
        'subscriptionPaymentMethod-component': false,
        'pendingRetrial-div': true,
        'freePlan-div': false,
        'premiumPlan-div': false,
        'subscriptionActive-div': false,
        'updatePaymentMethod-component': false,
        'paymentMethods-component': false,
        'cancel-div': false,
        'activeLoading-div': false,
      },
    },
    {
      description: 'Premium usage',
      computed: {
        active: true,
        state: 'processed',
      },
      data: {
        renderData: true,
        action: 'subscribe',
      },
      instance: {
        active: true,
        state: 'processed',
        current_period_end: 0,
        customer_id: 'cus_123',
        subscription_id: 'sub_123',
        payment_method_id: 'pm_123',
        info: info2,
        card: card2,
      },
      template: {
        'subscriptionPaymentMethod-component': false,
        'freePlan-div': false,
        'premiumPlan-div': true,
        'subscriptionActive-div': true,
        'updatePaymentMethod-component': true,
        'paymentMethods-component': true,
        'cancel-div': true,
        'activeLoading-div': false,
      },
    },
  ];

  const storeVuex = (billing, currentrole) => new Vuex.Store({
    namespaced: true,
    state: {
      billing,
      currentrole,
      info: infoData,
    },
    getters: {
      'billing/active': (state) => state.billing.active || false,
      'billing/status': (state) => state.billing.state || 'inactive',
      'billing/get': (state) => state.billing,
      'auth/role': (state) => state.currentrole,
      'billing/getBillInfoData': (state) => state.info,
    },
    actions: {
      'billing/getSubscription': () => stripeData,
      'namespaces/get': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = shallowMount(SettingBilling, {
            store: storeVuex(test.instance, currentrole),
            localVue,
            stubs: ['fragment'],
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
              $stripe: {
                elements: () => ({
                  create: () => ({
                    mount: () => null,
                  }),
                }),
              },
            },
          });

          wrapper.setData({ renderData: test.data.renderData });
          wrapper.setData({ billingData: { info: { nextPaymentDue: 0 }, card: { brand: 'cc-visa' } } });
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

        //////
        // HTML validation
        //////

        if (currentrole === 'owner') {
          describe(`Template rendering - ${test.description}`, () => {
            Reflect.ownKeys(test.template).forEach((k) => {
              it(`${test.template[k] ? 'Renders' : 'Does not render'} template ${k} for `, () => {
                expect(wrapper.find(`[data-test="${k}"]`).exists()).toBe(test.template[k]);
              });
            });
          });
        }
        ///////
        // Data checking
        //////

        it('Compare data with default value', () => {
          if (hasAuthorization[currentrole]) {
            Object.keys(test.data).forEach((item) => {
              expect(wrapper.vm[item]).toEqual(test.data[item]);
            });
          }
        });
        it('Process data in the computed', () => {
          Object.keys(test.computed).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.computed[item]);
          });
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });
      });
    });
  });
});
