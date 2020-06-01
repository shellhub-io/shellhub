import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceAdd from '@/components/device/DeviceAdd.vue';

describe('DeviceAdd', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      addDevice: false,
      tenant: '',
    },
    getters: {
      'modals/addDevice': (state) => state.addDevice,
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'modals/showAddDevice': () => {
      }
    }
  });

  beforeEach(() => {

    wrapper = shallowMount(DeviceAdd, {
      store,
      localVue,
      stubs: ['fragment']
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper.isVueInstance()).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
