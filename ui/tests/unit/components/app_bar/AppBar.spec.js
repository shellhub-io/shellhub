import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import AppBar from '@/components/app_bar/AppBar';
import router from '@/router/index';

describe('AppBar', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const tenant = 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx';
  const isLoggedIn = true;
  const numberNamespaces = 1;
  const drawer = true;
  const isMobile = false;

  const menu = [
    {
      title: 'Settings',
      type: 'path',
      path: '/settings',
      icon: 'mdi-cog',
      items: [{ title: 'Profile', path: '/settings/profile' }],
    },
    {
      title: 'Logout',
      type: 'method',
      icon: 'mdi-logout',
      method: 'logout',
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isLoggedIn,
      numberNamespaces,
      tenant,
      isMobile,
    },
    getters: {
      'auth/isLoggedIn': (state) => state.isLoggedIn,
      'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
      'auth/tenant': (state) => state.tenant,
      'mobile/isMobile': (state) => state.isMobile,
    },
    actions: {
      'auth/logout': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(AppBar, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { drawer },
      router,
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

  it('Receive data in props', () => {
    expect(wrapper.vm.drawer).toEqual(drawer);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.clipped).toEqual(false);
    expect(wrapper.vm.chatOpen).toEqual(false);
    expect(wrapper.vm.defaultSize).toEqual(24);
    expect(wrapper.vm.menu).toEqual(menu);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.isLoggedIn).toEqual(isLoggedIn);
    expect(wrapper.vm.hasNamespaces).toEqual(numberNamespaces !== 0);
    expect(wrapper.vm.isMobile).toEqual(isMobile);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with components', () => {
    expect(wrapper.find('[data-test="namespaceMenu-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="notification-component"]').exists()).toBe(true);
  });
  it('Renders the template with data', async () => {
    expect(wrapper.find('[data-test="Settings"]').exists()).toEqual(true);
    expect(wrapper.find('[data-test="Logout"]').exists()).toEqual(true);
  });
});
