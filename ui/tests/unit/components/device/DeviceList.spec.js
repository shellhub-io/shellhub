import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceList from '@/components/device/DeviceList.vue';

describe('DeviceList', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  let device = {};

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
    }
  });

  beforeEach(() => {

    wrapper = shallowMount(DeviceList, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { device }
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper.isVueInstance()).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
