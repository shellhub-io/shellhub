import Vuex from 'vuex';
import VueRouter from 'vue-router';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Device from '@/components/device/Device';

describe('Device', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  localVue.use(VueRouter);

  let wrapper;

  beforeEach(() => {
    wrapper = shallowMount(Device, {
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
