import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceDelete from '@/components/device/DeviceDelete';

describe('DeviceDelete', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const uid = 'a582b47a42d';
  const redirect = true;

  const store = new Vuex.Store({
    namespaced: true,
    actions: {
      'devices/remove': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(DeviceDelete, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, redirect },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive uid in props', () => {
    expect(wrapper.vm.uid).toEqual(uid);
  });
});
