import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import BillingWarning from '@/components/billing/BillingWarning';
import router from '@/router/index';
import { actions, authorizer } from '../../../../src/authorizer';

describe('BillingWarning', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;
  const statusUpdateAccountDialog = true;
  const statusUpdateAccountDialogByDeviceAction = false;

  const stats = {
    registered_devices: 0,
    online_devices: 0,
    bctive_sessions: 0,
    pending_devices: 1,
    rejected_devices: 0,
  };

  const storeVuex = (statsData, billingEnabled, role) => new Vuex.Store({
    namespaced: true,
    state: {
      stateBilling: billingEnabled,
      role,
      stats: statsData,
      statusUpdateAccountDialog,
      statusUpdateAccountDialogByDeviceAction,
    },
    getters: {
      'billing/active': (state) => state.stateBilling,
      'stats/stats': (state) => state.stats,
      'auth/role': (state) => state.role,
      'users/statusUpdateAccountDialog': (state) => state.statusUpdateAccountDialog,
      'users/statusUpdateAccountDialogByDeviceAction': (state) => state.statusUpdateAccountDialogByDeviceAction,
    },
    actions: {
      'users/setStatusUpdateAccountDialog': () => {},
    },
  });

  const tests = [
    {

      ///////
      // In this case, uhe test dialog is closes when the user has less
      // than 3 devices and has no subscription.
      ///////

      description: 'Less than 3 devices, no subscription',
      storeData: {
        stats,
        enable: false,
        permission: 'owner',
      },
      computed: {
        showMessage: false,
      },
      template: {
        '[data-test="billingWarning-dialog"]': false,
        '[data-test="close-btn"]': false,
        '[data-test="goToBilling-btn"]': false,
      },
    },
    {

      ///////
      // In this case, the test dialog is closes when the user has 3
      // devices and has subscription.
      ///////

      description: '3 devices, with subscription',
      storeData: {
        stats: { ...stats, registered_devices: 3 },
        enable: true,
        permission: 'owner',
      },
      computed: {
        showMessage: false,
      },
      template: {
        '[data-test="billingWarning-dialog"]': false,
        '[data-test="close-btn"]': false,
        '[data-test="goToBilling-btn"]': false,
      },
    },
    {

      ///////
      // In this case, the test dialog is closes when the user has 3
      // devices and has no subscription.
      ///////

      description: '3 devices, no subscription',
      storeData: {
        stats: { ...stats, registered_devices: 3 },
        enable: false,
        permission: 'owner',
      },
      computed: {
        showMessage: true,
        hasAuthorization: true,
      },
      template: {
        '[data-test="billingWarning-dialog"]': true,
        '[data-test="close-btn"]': true,
        '[data-test="goToBilling-btn"]': true,
      },
    },
    {

      ///////
      // In this case, the test dialog does not open for user not owner
      ///////

      description: 'Not the owner',
      storeData: {
        stats: { ...stats, registered_devices: 3 },
        enable: false,
        permission: 'operator',
      },
      computed: {
        showMessage: true,
        hasAuthorization: false,
      },
      template: {
        '[data-test="billingWarning-dialog"]': false,
        '[data-test="close-btn"]': false,
        '[data-test="goToBilling-btn"]': false,
      },
    },
  ];

  tests.forEach((test) => {
    describe(test.description, () => {
      beforeEach(() => {
        wrapper = mount(BillingWarning, {
          localVue,
          store: storeVuex(...Object.values(test.storeData)),
          stubs: ['fragment'],
          router,
          mocks: {
            $authorizer: authorizer,
            $actions: actions,
          },
          vuetify,
        });
      });

      ///////
      // Component Rendering
      ///////

      it('Is a Vue instance', () => {
        expect(wrapper).toBeTruthy();
      });
      it('Renders the component', () => {
        expect(wrapper.html()).toMatchSnapshot();
      });

      ///////
      // Data and props checking
      ///////

      it('Process data in the computed', () => {
        Reflect.ownKeys(test.computed).forEach((c) => {
          expect(wrapper.vm[c]).toEqual(test.computed[c]);
        });
      });

      ///////
      // HTML validation
      ///////
      //

      it('Renders the template with data', async () => {
        Reflect.ownKeys(test.template).forEach((t) => {
          expect(wrapper.find(t).exists()).toBe(test.template[t]);
        });
      });
    });
  });
});
