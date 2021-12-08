import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import NamespaceDelete from '@/components/app_bar/namespace/NamespaceDelete';
import { actions, authorizer } from '../../../../../src/authorizer';

describe('NamespaceDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

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

  document.body.setAttribute('data-app', true);

  let wrapper;

  const nsTenant = 'xxxxxx';

  const accessType = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

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

  const namespaceObject = {
    name: 'namespace3',
    owner: 'user1',
    member_names: ['user1', 'user7', 'user8'],
    tenant_id: 'xxxxxxxx',
  };

  const getter = {
    billingActive: [
      inactiveBilling,
      activeBilling,
    ],
  };

  const tests = [
    {
      description: 'Button',
      props: {
        nsTenant,
      },
      data: {
        name: namespaceObject.name,
        dialog: false,
        action: 'remove',
      },
      computed: {
        tenant: nsTenant,
        active: getter.billingActive[0].active,
        billing: getter.billingActive[0],
      },
      namespace: namespaceObject,
      env: {
        billingEnable: false,
      },
      template: {
        'delete-btn': true,
        'namespaceDelete-dialog': false,
        'contentSubscription-p': false,
        'close-btn': false,
        'remove-btn': false,
      },
    },
    {
      description: 'Dialog without subscription',
      props: {
        nsTenant,
      },
      data: {
        name: namespaceObject.name,
        dialog: true,
        action: 'remove',
      },
      computed: {
        tenant: nsTenant,
        active: getter.billingActive[0].active,
        billing: getter.billingActive[0],
      },
      namespace: namespaceObject,
      env: {
        billingEnable: false,
      },
      template: {
        'delete-btn': true,
        'namespaceDelete-dialog': true,
        'contentSubscription-p': false,
        'close-btn': true,
        'remove-btn': true,
      },
    },
    {
      description: 'Dialog with subscription',
      props: {
        nsTenant,
      },
      data: {
        name: namespaceObject.name,
        dialog: true,
        action: 'remove',
      },
      computed: {
        tenant: nsTenant,
        active: getter.billingActive[1].active,
        billing: getter.billingActive[1],
      },
      namespace: namespaceObject,
      env: {
        billingEnable: true,
      },
      template: {
        'delete-btn': true,
        'namespaceDelete-dialog': true,
        'contentSubscription-p': true,
        'close-btn': true,
        'remove-btn': true,
      },
    },
  ];

  const storeVuex = (active, billing, currentAccessType, namespace) => new Vuex.Store({
    namespaced: true,
    state: {
      active,
      billing,
      currentAccessType,
      namespace,
      info: infoData,
    },
    getters: {
      'billing/active': (state) => state.active,
      'billing/get': (state) => state.billing,
      'auth/accessType': (state) => state.currentAccessType,
      'namespaces/get': (state) => state.namespace,
      'billing/getBillInfoData': (state) => state.info,
    },
    actions: {
      'namespaces/remove': () => {},
      'auth/logout': () => {},
      'billing/getSubscription': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  tests.forEach((test) => {
    accessType.forEach((currentAccessType) => {
      describe(`${test.description} ${currentAccessType}`, () => {
        beforeEach(() => {
          wrapper = mount(NamespaceDelete, {
            store: storeVuex(
              test.computed.active,
              test.computed.billing,
              currentAccessType,
              test.namespace,
            ),
            localVue,
            stubs: ['fragment'],
            propsData: { nsTenant: test.props.nsTenant },
            vuetify,
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
              $env: {
                billingEnable: test.env.billingEnable,
              },
              $stripe: {
                elements: () => ({
                  create: () => ({
                    mount: () => null,
                  }),
                }),
              },
            },
          });

          wrapper.setData({ dialog: test.data.dialog });

          if (test.env.billingEnable) wrapper.setData({ amountDue: 100 });
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

        it('Receive data in props', () => {
          Object.keys(test.props).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.props[item]);
          });
        });
        it('Compare data with default value', () => {
          Object.keys(test.data).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.data[item]);
          });
        });
        it('Process data in the computed', () => {
          Object.keys(test.computed).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.computed[item]);
          });
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentAccessType]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', async () => {
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        });
      });
    });
  });
});
