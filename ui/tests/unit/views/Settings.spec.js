import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import VueRouter from 'vue-router';
import Settings from '@/views/Settings';

describe('Settings', () => {
  const localVue = createLocalVue();
  localVue.use(VueRouter);
  localVue.use(Vuex);

  let wrapper;

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
  ];

  beforeEach(() => {
    wrapper = shallowMount(Settings, {
      localVue,
      stubs: ['fragment'],
      mocks: {
        $env: (isEnterprise) => isEnterprise,
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
