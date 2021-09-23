import Vuex from 'vuex';
import { shallowMount, createLocalVue, config } from '@vue/test-utils';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
import MaxDevice from '@/components/app_bar/max_device/MaxDevice';

config.mocks = {
  $env: {
    isEnterprise: true,
  },
};

describe('MaxDevice', () => {
  const localVue = createLocalVue();
  localVue.component('FontAwesomeIcon', FontAwesomeIcon);
  localVue.use(Vuex);

  let wrapper;

  const namespaceWithoutDevices = {
    name: 'namespace',
    owner: 'user1',
    member_names: ['user6', 'user7', 'user8'],
    tenant_id: 'xxxxxxxx',
    devices_count: 0,
    max_devices: 0,
  };

  const namespaceWithDevices = {
    name: 'namespace1',
    owner: 'user2',
    member_names: ['user2', 'user4', 'user5'],
    tenant_id: 'xxxxxxxx',
    devices_count: 4,
    max_devices: 10,
  };

  const storeWithoutDevices = new Vuex.Store({
    namespaced: true,
    state: {
      namespace: namespaceWithoutDevices,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
    },
    actions: {},
  });

  const storeWithDevices = new Vuex.Store({
    namespaced: true,
    state: {
      namespace: namespaceWithDevices,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
    },
    actions: {
    },
  });

  ///////
  // In this case, check owner fields rendering in enterprise version
  // and without devices of the template.
  ///////

  describe('Without devices', () => {
    beforeEach(() => {
      wrapper = shallowMount(MaxDevice, {
        store: storeWithoutDevices,
        localVue,
        stubs: ['fragment'],
        mocks: ['$env'],
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.namespace).toEqual(namespaceWithoutDevices);
    });
    it('Compare data with default value', () => {
      expect(wrapper.find('[data-test="devices-chip"]').text()).toEqual(namespaceWithoutDevices.devices_count.toString());
    });
  });

  ///////
  // In this case, check owner fields rendering in enterprise version
  // and with devices of the template.
  ///////

  describe('With devices', () => {
    beforeEach(() => {
      wrapper = shallowMount(MaxDevice, {
        store: storeWithDevices,
        localVue,
        stubs: ['fragment'],
        mocks: ['$env'],
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.namespace).toEqual(namespaceWithDevices);
    });
    it('Compare data with default value', () => {
      expect(wrapper.find('[data-test="devices-chip"]').text()).toEqual(namespaceWithDevices.devices_count.toString());
    });
  });

  ///////
  // In this case, check owner fields rendering in open version of
  // the template.
  ///////

  describe('Without devices', () => {
    beforeEach(() => {
      wrapper = shallowMount(MaxDevice, {
        store: storeWithDevices,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $env: {
            isEnterprise: false,
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.namespace).toEqual(namespaceWithDevices);
    });
    it('Compare data with default value', () => {
      expect(wrapper.find('[data-test=devices-chip]').exists()).toEqual(false);
    });
  });
});
