import Vuex from 'vuex';
import { shallowMount, createLocalVue, config } from '@vue/test-utils';
import NamespaceInstructions from '@/components/app_bar/namespace/NamespaceInstructions';

config.mocks = {
  $env: {
    isEnterprise: true,
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
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialogAdd).toEqual(false);
  });
  it('Renders the template with data - enterprise version', () => {
    expect(wrapper.find('[data-test="openContentFirst-text"]').exists()).toEqual(false);
    expect(wrapper.find('[data-test="openContentSecond-text"]').exists()).toEqual(false);
    expect(wrapper.find('[data-test="namespace-btn"]').exists()).toEqual(true);
  });
  it('Renders the template with data - open version', () => {
    config.mocks = {
      $env: {
        isEnterprise: false,
      },
    };

    wrapper = shallowMount(NamespaceInstructions, {
      localVue,
      stubs: ['fragment'],
      propsData: { show },
    });

    expect(wrapper.find('[data-test="openContentFirst-text"]').exists()).toEqual(true);
    expect(wrapper.find('[data-test="openContentSecond-text"]').exists()).toEqual(true);
    expect(wrapper.find('[data-test="namespace-btn"]').exists()).toEqual(false);
  });
});
