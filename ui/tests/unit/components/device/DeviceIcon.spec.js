import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceIcon from '@/components/device/DeviceIcon.vue';

describe('DeviceIcon', () => {
  let wrapper;
  let iconName = 'fl-alpine';

  const localVue = createLocalVue();
  localVue.use(Vuex);

  beforeEach(() => {
    const localVue = createLocalVue();
    localVue.use(Vuex);

    wrapper = shallowMount(DeviceIcon, {
      localVue,
      stubs: ['fragment'],
      propsData: { iconName }
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper.isVueInstance()).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
