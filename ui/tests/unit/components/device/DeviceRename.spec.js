import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceRename from '@/components/device/DeviceRename';

describe('DeviceRename', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const uid = 'a582b47a42d';
  const name = '39-5e-2a';

  const store = new Vuex.Store({
    namespaced: true,
    actions: {
      'devices/rename': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(DeviceRename, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { name, uid },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.name).toEqual(name);
    expect(wrapper.vm.uid).toEqual(uid);
  });
});
