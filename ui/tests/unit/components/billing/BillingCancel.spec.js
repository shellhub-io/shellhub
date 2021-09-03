import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import BillingCancel from '@/components/billing/BillingCancel';

describe('BillingCancel', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const nextPaymentDue = 1234;

  const stats = {
    registered_devices: 0,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 1,
    rejected_devices: 0,
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      stats,
    },
    getters: {
      'stats/stats': (state) => state.stats,
    },
    actions: {
      'billing/cancelSubscription': () => {},
      'devices/setDeviceWarning': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  ///////
  // In this case, focus of the test is button rendering.
  ///////

  describe('Button', () => {
    beforeEach(() => {
      wrapper = mount(BillingCancel, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { nextPaymentDue },
        vuetify,
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

    it('Receive data in props', () => {
      expect(wrapper.vm.nextPaymentDue).toBe(nextPaymentDue);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="billingWarning-dialog"]').exists()).toBe(false);
    });
  });

  ///////
  // In this case, focus of the test is dialog rendering.
  ///////

  describe('Dialog', () => {
    beforeEach(() => {
      wrapper = mount(BillingCancel, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { nextPaymentDue },
        vuetify,
      });

      wrapper.setData({ dialog: true });
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
      expect(wrapper.vm.nextPaymentDue).toBe(nextPaymentDue);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="billingWarning-dialog"]').exists()).toBe(true);
    });
  });
});
