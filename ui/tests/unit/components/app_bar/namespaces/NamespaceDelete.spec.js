import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import NamespaceDelete from '@/components/app_bar/namespace/NamespaceDelete';
import { actions, authorizer } from '../../../../../src/authorizer';

describe('NamespaceDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const nsTenant = 'xxxxxx';

  const accessType = ['owner', 'administrator', 'operator', 'observer'];

  const hasAuthorization = {
    owner: true,
    administrator: true,
    operator: false,
    observer: false,
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
    },
    getters: {
      'billing/active': (state) => state.active,
      'billing/get': (state) => state.billing,
      'auth/accessType': (state) => state.currentAccessType,
      'namespaces/get': (state) => state.namespace,
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
