import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceDetails from '@/components/device/DeviceDetails';

describe('DeviceDetails', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const deviceOnline = {
    uid: 'a582b47a',
    name: '39-5e-2a',
    identity: {
      mac: '00:00:00',
    },
    info: {
      id: 'arch',
      pretty_name: 'Linux',
      version: '',
    },
    public_key: 'xxxxxxxx',
    tenant_id: '00000000',
    last_seen: '2020-05-20T18:58:53.276Z',
    online: true,
    namespace: 'user',
    tags: ['device1', 'device2'],
  };

  const deviceOffline = { ...deviceOnline, online: false };

  const storeDeviceOnline = new Vuex.Store({
    namespaced: true,
    state: {
      device: deviceOnline,
    },
    getters: {
      'devices/get': (state) => state.device,
    },
    actions: {
      'devices/get': () => {},
      'devices/updateTag': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  const storeDeviceOffline = new Vuex.Store({
    namespaced: true,
    state: {
      device: deviceOffline,
    },
    getters: {
      'devices/get': (state) => state.device,
    },
    actions: {
      'devices/get': () => {},
      'devices/updateTag': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  ///////
  // In this case, it is checking the rendering of components and
  // the device is online.
  ///////

  describe('Online', () => {
    beforeEach(() => {
      wrapper = shallowMount(DeviceDetails, {
        store: storeDeviceOnline,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $route: {
            params: {
              id: deviceOnline.uid,
            },
          },
        },
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

    it('Compare data with default value', () => {
      expect(wrapper.vm.uid).toEqual(deviceOnline.uid);
      expect(wrapper.vm.hostname).toEqual('localhost');
      expect(wrapper.vm.hide).toEqual(true);
      expect(wrapper.vm.device).toEqual(deviceOnline);
      expect(wrapper.vm.dialogDelete).toEqual(false);
      expect(wrapper.vm.dialogError).toEqual(false);
      expect(wrapper.vm.list).toEqual(deviceOnline.tags);
    });
    it('Process data in methods', () => {
      wrapper.vm.receiveName('ShellHub');
      expect(wrapper.vm.device.name).toEqual('ShellHub');
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="deviceRename-component"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="terminalDialog-component"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="deviceDelete-component"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="deviceTag-field"]').exists()).toEqual(true);
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="deviceUid-field"]').text()).toEqual(deviceOnline.uid);
      expect(wrapper.find('[data-test="deviceMac-field"]').text()).toEqual(deviceOnline.identity.mac);
      expect(wrapper.find('[data-test="devicePrettyName-field"]').text()).toEqual(deviceOnline.info.pretty_name);
      expect(wrapper.find('[data-test="deviceConvertDate-field"]').text()).toEqual('Wednesday, May 20th 2020, 6:58:53 pm');
    });
  });

  ///////
  // In this case, it is checking the rendering of components and
  // the device is offline.
  ///////

  describe('Offline', () => {
    beforeEach(() => {
      wrapper = shallowMount(DeviceDetails, {
        store: storeDeviceOffline,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $route: {
            params: {
              id: deviceOffline.uid,
            },
          },
        },
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

    it('Compare data with default value', () => {
      expect(wrapper.vm.uid).toEqual(deviceOffline.uid);
      expect(wrapper.vm.hostname).toEqual('localhost');
      expect(wrapper.vm.hide).toEqual(true);
      expect(wrapper.vm.device).toEqual(deviceOffline);
      expect(wrapper.vm.dialogDelete).toEqual(false);
      expect(wrapper.vm.dialogError).toEqual(false);
    });
    it('Process data in methods', () => {
      wrapper.vm.receiveName('ShellHub');
      expect(wrapper.vm.device.name).toEqual('ShellHub');
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="deviceRename-component"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="terminalDialog-component"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="deviceDelete-component"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="deviceTag-field"]').exists()).toEqual(true);
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="deviceUid-field"]').text()).toEqual(deviceOffline.uid);
      expect(wrapper.find('[data-test="deviceMac-field"]').text()).toEqual(deviceOffline.identity.mac);
      expect(wrapper.find('[data-test="devicePrettyName-field"]').text()).toEqual(deviceOffline.info.pretty_name);
      expect(wrapper.find('[data-test="deviceConvertDate-field"]').text()).toEqual('Wednesday, May 20th 2020, 6:58:53 pm');
    });
  });
});
