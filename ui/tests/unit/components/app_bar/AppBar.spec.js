import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Router from 'vue-router';
import AppBar from '@/components/app_bar/AppBar';

describe('AppBar', () => {
  const localVue = createLocalVue();
  const router = new Router();
  localVue.use(Vuex);
  localVue.use(Router);

  let wrapper;

  const tenant = 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx';
  const isLoggedIn = true;
  const numberNamespaces = 1;
  const isMobile = false;
  const statusDarkMode = true;

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
      statusDarkMode,
    },
    getters: {
      'auth/isLoggedIn': (state) => state.isLoggedIn,
      'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
      'auth/tenant': (state) => state.tenant,
      'mobile/isMobile': (state) => state.isMobile,
      'layout/getStatusDarkMode': (state) => state.statusDarkMode,
    },
    actions: {
      'auth/logout': () => {},
      'layout/setLayout': () => {},
      'namespaces/clearNamespaceList': () => {},
      'snackbar/showSnackbarErrorNotRequest': () => {},
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(AppBar, {
      store,
      localVue,
      stubs: ['fragment'],
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
    expect(wrapper.vm.getStatusDarkMode).toEqual(true);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with components', () => {
    expect(wrapper.find('[data-test="notification-component"]').exists()).toBe(true);
  });
  it('Renders the template with data', async () => {
    expect(wrapper.find('[data-test="Settings"]').exists()).toEqual(true);
    expect(wrapper.find('[data-test="Logout"]').exists()).toEqual(true);
  });
});
