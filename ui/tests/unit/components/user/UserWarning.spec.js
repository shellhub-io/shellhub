import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import UserWarning from '@/components/user/UserWarning';
import 'mock-local-storage';

describe('UserWarning', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const numberNamespaces = 0;
  const statusSpinner = false;
  const activeBilling = false;

  const namespace = {
    name: 'namespace',
    owner: 'user',
    members: [{ name: 'user' }, { name: 'user2' }],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484712',
  };

  const statsWithoutDevices = {
    registered_devices: 0,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 0,
    rejected_devices: 0,
  };

  const statsWithDevices = {
    registered_devices: 2,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 0,
    rejected_devices: 0,
  };

  const getters = {
    'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
    'spinner/getStatus': (state) => state.statusSpinner,
    'stats/stats': (state) => state.stats,
    'billing/active': (state) => state.activeBilling,
    'namespaces/get': (state) => state.namespace,
  };

  const actions = {
    'stats/get': () => {},
    'devices/setDeviceWarning': () => {},
    'auth/setShowWelcomeScreen': () => {},
    'namespaces/fetch': () => {},
    'snackbar/showSnackbarErrorAssociation': () => {},
    'snackbar/showSnackbarErrorLoading': () => {},
  };

  const storeWithoutDevices = new Vuex.Store({
    namespaced: true,
    state: {
      numberNamespaces,
      statusSpinner,
      stats: statsWithoutDevices,
      activeBilling,
      namespace,
    },
    getters,
    actions,
  });

  const storeWithDevices = new Vuex.Store({
    namespaced: true,
    state: {
      numberNamespaces,
      statusSpinner,
      stats: statsWithDevices,
      activeBilling,
      namespace,
    },
    getters,
    actions,
  });

  ///////
  // In this case, The welcome screen loads with the expected
  // behavior without devices and with billing environment
  // disabled
  ///////

  describe('Without devices and billing disabled', () => {
    beforeEach(() => {
      wrapper = shallowMount(UserWarning, {
        store: storeWithoutDevices,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $env: {
            billingEnable: false,
          },
        },
      });

      localStorage.setItem('namespacesWelcome', JSON.stringify({}));
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

    it('Compare data with the default value', () => {
      expect(wrapper.vm.show).toEqual(false);
      expect(wrapper.vm.showInstructions).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.hasNamespaces).toEqual(numberNamespaces !== 0);
      expect(wrapper.vm.hasSpinner).toEqual(statusSpinner);
      expect(wrapper.vm.stats).toEqual(statsWithoutDevices);
    });
    it('Process data in methods', () => {
      expect(wrapper.vm.hasDevices()).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="deviceWarning-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="welcome-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="namespaceInstructions-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="billingWarning-component"]').exists()).toBe(false);
    });
    it('Renders the template with data', async () => {
      await wrapper.vm.showScreenWelcome();
      expect(wrapper.vm.show).toBe(true);

      localStorage.setItem('namespacesWelcome', JSON.stringify({ ...{ [namespace.tenant_id]: true } }));

      await wrapper.vm.showScreenWelcome();
      expect(wrapper.vm.show).toBe(false);
    });
  });

  ///////
  // In this case, The welcome screen loads with the expected
  // behavior without devices
  ///////

  describe('Without devices', () => {
    beforeEach(() => {
      wrapper = shallowMount(UserWarning, {
        store: storeWithoutDevices,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $env: {
            billingEnable: true,
          },
        },
      });

      localStorage.setItem('namespacesWelcome', JSON.stringify({}));
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

    it('Compare data with the default value', () => {
      expect(wrapper.vm.show).toEqual(false);
      expect(wrapper.vm.showInstructions).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.hasNamespaces).toEqual(numberNamespaces !== 0);
      expect(wrapper.vm.hasSpinner).toEqual(statusSpinner);
      expect(wrapper.vm.stats).toEqual(statsWithoutDevices);
    });
    it('Process data in methods', () => {
      expect(wrapper.vm.hasDevices()).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="deviceWarning-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="welcome-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="namespaceInstructions-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="billingWarning-component"]').exists()).toBe(true);
    });
    it('Renders the template with data', async () => {
      await wrapper.vm.showScreenWelcome();
      expect(wrapper.vm.show).toBe(true);

      localStorage.setItem('namespacesWelcome', JSON.stringify({ ...{ [namespace.tenant_id]: true } }));

      await wrapper.vm.showScreenWelcome();
      expect(wrapper.vm.show).toBe(false);
    });
  });

  ///////
  // In this case, The welcome screen loads with the expected
  // behavior with devices and with billing environment enabled
  ///////

  describe('With devices', () => {
    beforeEach(() => {
      wrapper = shallowMount(UserWarning, {
        store: storeWithDevices,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $env: {
            billingEnable: true,
          },
        },
      });

      localStorage.setItem('namespacesWelcome', JSON.stringify({}));
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

    it('Compare data with the default value', () => {
      expect(wrapper.vm.show).toEqual(false);
      expect(wrapper.vm.showInstructions).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.hasNamespaces).toEqual(numberNamespaces !== 0);
      expect(wrapper.vm.hasSpinner).toEqual(statusSpinner);
      expect(wrapper.vm.stats).toEqual(statsWithDevices);
    });
    it('Process data in methods', () => {
      expect(wrapper.vm.hasDevices()).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="deviceWarning-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="welcome-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="namespaceInstructions-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="billingWarning-component"]').exists()).toBe(true);
    });
    it('Renders the template with data', async () => {
      expect(wrapper.vm.namespaceHasBeenShown(namespace.tenant_id)).toBe(false);

      localStorage.setItem('namespacesWelcome', JSON.stringify({
        ...JSON.parse(localStorage.getItem('namespacesWelcome')),
        ...{ [namespace.tenant_id]: true },
      }));

      expect(wrapper.vm.namespaceHasBeenShown(namespace.tenant_id)).toBe(true);

      await wrapper.vm.showScreenWelcome();
      expect(Object.keys(JSON.parse(localStorage.getItem('namespacesWelcome')))).toHaveLength(1);
    });
  });
});
