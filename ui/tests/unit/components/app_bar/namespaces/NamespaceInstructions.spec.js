import Vuex from 'vuex';
import { shallowMount, createLocalVue, config } from '@vue/test-utils';
import NamespaceInstructions from '@/components/app_bar/namespace/NamespaceInstructions';

config.mocks = {
  $env: {
    isHosted: true,
  },
};

describe('NamespaceInstructions', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const show = true;

  beforeEach(() => {
    wrapper = shallowMount(NamespaceInstructions, {
      localVue,
      stubs: ['fragment'],
      propsData: { show },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receives data in props', () => {
    expect(wrapper.vm.show).toEqual(show);
  });
});
