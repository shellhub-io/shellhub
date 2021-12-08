import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import BillingCancel from '@/components/billing/BillingCancel';
import { actions, authorizer } from '../../../../src/authorizer';

describe('BillingCancel', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const accessType = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

  const tests = [
    {
      description: 'Button',
      variables: {
        dialog: false,
      },
      props: {
        nextPaymentDue: 1234,
      },
      data: {
        dialog: false,
        action: 'unsubscribe',
      },
      template: {
        'cancel-btn': true,
        'billingCancel-dialog': false,
        'close-btn': false,
        'cancelDialog-btn': false,
      },
      templateText: {
        'cancel-btn': 'Cancel',
      },
    },
    {
      description: 'Dialog',
      variables: {
        dialog: true,
      },
      props: {
        nextPaymentDue: 1234,
      },
      data: {
        dialog: true,
        action: 'unsubscribe',
      },
      template: {
        'cancel-btn': true,
        'billingCancel-dialog': true,
        'close-btn': true,
        'cancelDialog-btn': true,
      },
      templateText: {
        'cancel-btn': 'Cancel',
      },
    },
  ];

  const storeVuex = (currentAccessType) => new Vuex.Store({
    namespaced: true,
    state: {
      currentAccessType,
    },
    getters: {
      'auth/accessType': (state) => state.currentAccessType,
    },
    actions: {
      'billing/cancelSubscription': () => {},
      'devices/setDeviceChooserStatus': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    accessType.forEach((currentAccessType) => {
      describe(`${test.description} ${currentAccessType}`, () => {
        beforeEach(() => {
          wrapper = mount(BillingCancel, {
            store: storeVuex(currentAccessType),
            localVue,
            stubs: ['fragment', 'router-link'],
            propsData: { nextPaymentDue: test.props.nextPaymentDue },
            vuetify,
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
            },
          });

          wrapper.setData({ dialog: test.variables.dialog });
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
        // Data checking
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
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentAccessType]);
        });

        //////
        // HTML validation
        //////

        Object.keys(test.template).forEach((item) => {
          it(`Renders the template ${item} with data`, () => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        });
        it('Renders template with expected text', () => {
          Object.keys(test.templateText).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).text()).toContain(test.templateText[item]);
          });
        });
      });
    });
  });
});
