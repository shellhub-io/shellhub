import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import VueRouter from 'vue-router';
import Settings from '@/views/Settings';

describe('Settings', () => {
  const localVue = createLocalVue();
  localVue.use(VueRouter);
  localVue.use(Vuex);

  let wrapper;

  let numberNamespaces = 1;

  const items = [
    {
      title: 'Profile',
      path: '/settings',
    },
    {
      title: 'Namespace',
      path: '/settings/namespace-manager',
    },
    {
      title: 'Private Keys',
      path: '/settings/private-keys',
    },
    {
      title: 'Billing',
      path: '/settings/billing',
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      numberNamespaces,
    },
    getters: {
      'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
    },
    actions: {
    },
  });

  ///////
  // In this case, the billing tab can only be rendered in the cloud.
  // Checks if this tab has been rendered when user has namespace.
  ///////

  describe('Cloud is true', () => {
    beforeEach(() => {
      wrapper = shallowMount(Settings, {
        localVue,
        store,
        stubs: ['fragment'],
        mocks: {
          $env: {
            isCloud: true,
            billingEnable: true,
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

    it('Compare data with default value', () => {
      expect(wrapper.vm.drawer).toEqual(true);
      expect(wrapper.vm.clipped).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      Object.keys(items).forEach((item) => {
        expect(wrapper.find(`[data-test="${items[item].title}-tab"]`).text()).toEqual(items[item].title);
      });
    });
  });

  ///////
  // In this case, the billing tab can only be rendered in the cloud.
  // Checks if this tab is not rendered when user has no namespace.
  ///////

  describe('Cloud is false', () => {
    numberNamespaces = 0;

    beforeEach(() => {
      wrapper = shallowMount(Settings, {
        localVue,
        stubs: ['fragment'],
        mocks: {
          $env: {
            isCloud: false,
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

    it('Compare data with default value', () => {
      expect(wrapper.vm.drawer).toEqual(true);
      expect(wrapper.vm.clipped).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      Object.keys(items.slice(0, -1)).forEach((item) => {
        expect(wrapper.find(`[data-test="${items[item].title}-tab"]`).text()).toEqual(items[item].title);
      });
    });
  });

  ///////
  // In this case, the billing tab can only be rendered in the cloud.
  // Checks if this tab is not rendered.
  ///////

  describe('Cloud is false', () => {
    beforeEach(() => {
      wrapper = shallowMount(Settings, {
        localVue,
        stubs: ['fragment'],
        mocks: {
          $env: {
            isCloud: false,
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

    it('Compare data with default value', () => {
      expect(wrapper.vm.drawer).toEqual(true);
      expect(wrapper.vm.clipped).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', async () => {
      Object.keys(items.slice(0, -1)).forEach((item) => {
        expect(wrapper.find(`[data-test="${items[item].title}-tab"]`).text()).toEqual(items[item].title);
      });
    });
  });
});
