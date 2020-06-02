import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceRename from '@/components/device/DeviceRename.vue';

describe('DeviceRename', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  let device = {};

  const store = new Vuex.Store({
    namespaced: true,
    actions: {
      'devices/rename': () => {
      }
    }
  });

  beforeEach(() => {

    wrapper = shallowMount(DeviceRename, {
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
