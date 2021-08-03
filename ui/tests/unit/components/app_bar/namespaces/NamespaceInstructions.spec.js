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

  ///////
  // In this case, check owner fields rendering in enterprise version of
  // the template.
  ///////

  describe('', () => {
    beforeEach(() => {
      wrapper = shallowMount(NamespaceInstructions, {
        localVue,
        stubs: ['fragment'],
        propsData: { show },
        mocks: ['$env'],
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receives data in props', () => {
      expect(wrapper.vm.show).toEqual(show);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialogAdd).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="namespaceAdd-component"]').exists()).toEqual(true);
    });

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="namespaceInstructions-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="openContentFirst-text"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="openContentSecond-text"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="add-btn"]').exists()).toEqual(true);
    });
  });

  ///////
  // In this case, check owner fields rendering in open version of
  // the template.
  ///////

  describe('', () => {
    beforeEach(() => {
      wrapper = shallowMount(NamespaceInstructions, {
        localVue,
        stubs: ['fragment'],
        propsData: { show },
        mocks: {
          $env: {
            isEnterprise: false,
          },
        },
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receives data in props', () => {
      expect(wrapper.vm.show).toEqual(show);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialogAdd).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="namespaceAdd-component"]').exists()).toEqual(false);
    });

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="namespaceInstructions-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="openContentFirst-text"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="openContentSecond-text"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="add-btn"]').exists()).toEqual(false);
    });
  });
});
