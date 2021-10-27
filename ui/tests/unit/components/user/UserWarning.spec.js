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
  const activeBilling = true;
  const DeviceChooserStatus = false;

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
    registered_devices: 4,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 0,
    rejected_devices: 0,
  };

  const getters = {
    'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
    'spinner/getStatus': (state) => state.statusSpinner,
    'stats/stats': (state) => state.stats,
    'billing/active': (state) => !state.activeBilling,
    'namespaces/get': (state) => state.namespace,
    'devices/getDeviceChooserStatus': (state) => state.DeviceChooserStatus,
  };

  const actions = {
    'stats/get': () => {},
    'devices/setDeviceChooserStatus': () => {},
    'auth/setShowWelcomeScreen': () => {},
    'namespaces/fetch': () => {},
    'snackbar/showSnackbarErrorAssociation': () => {},
    'snackbar/showSnackbarErrorLoading': () => {},
  };

  const storeWithoutDevices = new Vuex.Store({
    namespaced: true,
    state: {
      DeviceChooserStatus,
      numberNamespaces,
      statusSpinner,
      stats: statsWithoutDevices,
      activeBilling,
      namespace,
    },
    getters,
    actions,
  });

  const storeWithDevicesInactive = new Vuex.Store({
    namespaced: true,
    state: {
      numberNamespaces: 3,
      statusSpinner,
      stats: statsWithDevices,
      active: false,
      namespace,
    },
    getters: {
      ...getters,
      'billing/active': (state) => state.active,
      'devices/getDeviceChooserStatus': (state) => !state.DeviceChooserStatus,
    },
    actions,
  });

  const storeWithDevicesActive = new Vuex.Store({
    namespaced: true,
    state: {
      numberNamespaces: 3,
      statusSpinner,
      stats: statsWithDevices,
      active: true,
      namespace,
    },
    getters: {
      ...getters,
      'billing/active': (state) => state.active,
    },
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
      expect(wrapper.find('[data-test="deviceChooser-component"]').exists()).toBe(false);
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
      expect(wrapper.find('[data-test="deviceChooser-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="welcome-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="namespaceInstructions-component"]').exists()).toBe(true);
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
  // and inactive subscription
  ///////

  describe('With devices and inactive billing', () => {
    beforeEach(() => {
      storeWithDevicesInactive.dispatch = jest.fn();

      wrapper = shallowMount(UserWarning, {
        store: storeWithDevicesInactive,
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
    //

    it('Is a Vue instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    //////
    // Call actions
    //////
    it('Dispatches on mount', () => {
      expect(storeWithDevicesInactive.dispatch).toHaveBeenCalledWith('devices/setDeviceChooserStatus', true);
    });

    ///////
    // Data and Props checking
    //////

    it('Compare data with the default value', () => {
      expect(wrapper.vm.show).toEqual(false);
      expect(wrapper.vm.showInstructions).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.hasNamespaces).toEqual(true);
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
      expect(wrapper.find('[data-test="deviceChooser-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="welcome-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="namespaceInstructions-component"]').exists()).toBe(true);
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

  ///////
  // In this case, The welcome screen loads with the expected
  // behavior with devices and with billing environment enabled
  // and active subscription
  ///////

  describe('With devices and active billing', () => {
    beforeEach(() => {
      storeWithDevicesActive.dispatch = jest.fn();

      wrapper = shallowMount(UserWarning, {
        store: storeWithDevicesActive,
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

    //////
    // Call actions
    //////
    it('Dispatches on mount', () => {
      expect(storeWithDevicesActive.dispatch).toHaveBeenCalledWith('devices/setDeviceChooserStatus', false);
    });

    ///////
    // Data and Props checking
    //////

    it('Compare data with the default value', () => {
      expect(wrapper.vm.show).toEqual(false);
      expect(wrapper.vm.showInstructions).toEqual(false);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.hasNamespaces).toEqual(true);
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
      expect(wrapper.find('[data-test="deviceChooser-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="welcome-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="namespaceInstructions-component"]').exists()).toBe(true);
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
