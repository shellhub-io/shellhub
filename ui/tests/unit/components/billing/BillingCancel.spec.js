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

  const accessType = ['owner', 'administrator', 'operator', 'observer'];

  const hasAuthorization = {
    owner: true,
    administrator: true,
    operator: false,
    observer: false,
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
      },
      template: {
        'delete-btn': true,
        'billingCancel-dialog': false,
        'close-btn': false,
        'cancel-btn': false,
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
      },
      template: {
        'delete-btn': true,
        'billingCancel-dialog': true,
        'close-btn': true,
        'cancel-btn': true,
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

        it('Renders the template with data', () => {
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        });
      });
    });
  });
});
