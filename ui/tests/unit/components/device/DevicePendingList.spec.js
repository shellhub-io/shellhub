import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DevicePendingList from '@/components/device/DevicePendingList';

describe('DevicePendingList', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      devices: [],
      numberDevices: 0,
    },
    getters: {
      'devices/list': (state) => state.devices,
      'devices/getNumberDevices': (state) => state.numberDevices,
    },
    actions: {
      'modals/showAddDevice': () => {
      },
      'devices/fetch': () => {
      },
      'devices/rename': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(DevicePendingList, {
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
});
