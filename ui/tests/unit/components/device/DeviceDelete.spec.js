import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceDelete from '@/components/device/DeviceDelete.vue';

describe('DeviceDelete', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  let uid = 'a582b47a42d';
  let redirect = true;

  const store = new Vuex.Store({
    namespaced: true,
    actions: {
      'devices/remove': () => {
      }
    }
  });

  beforeEach(() => {

    wrapper = shallowMount(DeviceDelete, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, redirect }
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper.isVueInstance()).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
