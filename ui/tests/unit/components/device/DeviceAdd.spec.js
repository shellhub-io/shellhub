import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceAdd from '@/components/device/DeviceAdd';

describe('DeviceAdd', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      addDevice: false,
      tenant: '00000000',
    },
    getters: {
      'modals/addDevice': (state) => state.addDevice,
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'modals/showAddDevice': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(DeviceAdd, {
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
  it('Has a command field', () => {
    expect(wrapper.find('[data-test="command-field"]').exists()).toBe(true);
  });
});
