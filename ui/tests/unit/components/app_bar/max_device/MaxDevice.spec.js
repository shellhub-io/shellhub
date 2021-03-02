import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import MaxDevice from '@/components/app_bar/max_device/MaxDevice';

describe('MaxDevice', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const namespace = {
    name: 'namespace',
    owner: 'user1',
    member_names: ['user6', 'user7', 'user8'],
    tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484715',
    devices_count: 0,
    max_devices: 0,
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
    },
    actions: {
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(MaxDevice, {
      store,
      localVue,
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.namespace).toEqual(namespace);
  });
  it('Compare data with default value', () => {
    expect(wrapper.find('[data-test="devices-field"]').text()).toEqual(namespace.devices_count.toString());
  });
  it('Renders correct number of Devices', () => {
    const expectedNumbers = [4, 6, 7];
    const WrapperArray = [
      shallowMount(MaxDevice, {
        store: new Vuex.Store({
          namespaced: true,
          state: {
            namespace: {
              name: 'namespace1',
              owner: 'user2',
              member_names: ['user2', 'user4', 'user5'],
              tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484716',
              devices_count: 4,
              max_devices: 0,
            },
          },
          getters: {
            'namespaces/get': (state) => state.namespace,
          },
          actions: {
          },
        }),
        localVue,
        stubs: ['fragment'],
      }),
      shallowMount(MaxDevice, {
        store: new Vuex.Store({
          namespaced: true,
          state: {
            namespace: {
              name: 'namespace2',
              owner: 'user2',
              member_names: ['user2', 'user4', 'user5'],
              tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484717',
              devices_count: 6,
              max_devices: 0,
            },
          },
          getters: {
            'namespaces/get': (state) => state.namespace,
          },
          actions: {
          },
        }),
        localVue,
        stubs: ['fragment'],
      }),
      shallowMount(MaxDevice, {
        store: new Vuex.Store({
          namespaced: true,
          state: {
            namespace: {
              name: 'namespace4',
              owner: 'user3',
              member_names: ['user3', 'user4', 'user6'],
              tenant_id: 'a736a52b-5777-4f92-b0b8-e359bf484718',
              devices_count: 7,
              max_devices: 0,
            },
          },
          getters: {
            'namespaces/get': (state) => state.namespace,
          },
          actions: {
          },
        }),
        localVue,
        stubs: ['fragment'],
      }),
    ];
    expectedNumbers.forEach((n, i) => {
      expect(WrapperArray[i].find('[data-test="devices-field"]').text()).toEqual(n.toString());
    });
  });
});
