import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import BillingWarning from '@/components/billing/BillingWarning';
import router from '@/router/index';

describe('BillingWarning', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;
  const isOwner = true;
  let stateBilling = false;

  const stats = {
    registered_devices: 0,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 1,
    rejected_devices: 0,
  };

  const storeHasZeroDevices = new Vuex.Store({
    namespaced: true,
    state: {
      stateBilling,
      stats,
      isOwner,
    },
    getters: {
      'billing/active': (state) => state.stateBilling,
      'stats/stats': (state) => state.stats,
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
    },
  });

  const storeHasThreeDevices = new Vuex.Store({
    namespaced: true,
    state: {
      stateBilling,
      isOwner,
      stats: { ...stats, registered_devices: 3 },
    },
    getters: {
      'billing/active': (state) => state.stateBilling,
      'stats/stats': (state) => state.stats,
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
    },
  });

  const storeHasThreeDevicesWithSubscription = new Vuex.Store({
    namespaced: true,
    state: {
      stateBilling: true,
      isOwner,
      stats: { ...stats, registered_devices: 3 },
    },
    getters: {
      'billing/active': (state) => state.stateBilling,
      'stats/stats': (state) => state.stats,
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
    },
  });

  const storeNotOwner = new Vuex.Store({
    namespaced: true,
    state: {
      stateBilling: true,
      isOwner,
      stats: { ...stats, registered_devices: 3 },
    },
    getters: {
      'billing/active': (state) => state.stateBilling,
      'stats/stats': (state) => state.stats,
      'namespaces/owner': (state) => !state.isOwner,
    },
    actions: {
    },
  });

  ///////
  // In this case, the test dialog is closes when the user has less
  // than 3 devices and has no subscription.
  ///////

  describe('Dialog is closes when has less than 3 devices', () => {
    beforeEach(() => {
      wrapper = mount(BillingWarning, {
        localVue,
        store: storeHasZeroDevices,
        stubs: ['fragment'],
        router,
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
    ///////

    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="billingWarning-dialog"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="goToBilling-btn"]').exists()).toBe(false);
    });
  });

  ///////
  // In this case, the test dialog is closes when the user has 3
  // devices and has subscription.
  ///////

  describe('Dialog is closes when user has subscription', () => {
    beforeEach(() => {
      wrapper = mount(BillingWarning, {
        localVue,
        store: storeHasThreeDevicesWithSubscription,
        stubs: ['fragment'],
        router,
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
    ///////

    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="billingWarning-dialog"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="goToBilling-btn"]').exists()).toBe(false);
    });
  });

  ///////
  // In this case, the test dialog is opened when the user has 3
  // devices and has no subscription.
  ///////

  describe('Dialog is opens', () => {
    stateBilling = false;

    beforeEach(() => {
      wrapper = mount(BillingWarning, {
        localVue,
        store: storeHasThreeDevices,
        stubs: ['fragment'],
        router,
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
    ///////

    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="billingWarning-dialog"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="goToBilling-btn"]').exists()).toBe(true);
    });
  });

  ///////
  // In this case, the test dialog does not open for user not owner
  ///////

  describe('Avoid opening for user not owner', () => {
    stateBilling = false;

    beforeEach(() => {
      wrapper = mount(BillingWarning, {
        localVue,
        store: storeNotOwner,
        stubs: ['fragment'],
        router,
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
    ///////

    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      expect(wrapper.find('[data-test="billingWarning-dialog"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="goToBilling-btn"]').exists()).toBe(false);
    });
  });
});
