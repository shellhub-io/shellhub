import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import AppLayout from '@/layouts/AppLayout';

describe('AppLayout', () => {
  const localVue = createLocalVue();
  const router = new Router();
  localVue.use(Vuex);
  localVue.use(Router);
  const vuetify = new Vuetify();

  let wrapper;

  const tenant = 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx';
  const isLoggedIn = true;
  const isMobile = false;
  const numberNamespacesEqualZero = 0;
  const numberNamespacesGreaterThanZero = 1;
  const hasSpinner = false;
  const isEnterprise = false;

  const admins = [
    ['Management', 'people_outline'],
    ['Settings', 'settings'],
  ];

  const items = [
    {
      icon: 'dashboard',
      title: 'Dashboard',
      path: '/',
    },
    {
      icon: 'devices',
      title: 'Devices',
      path: '/devices',
    },
    {
      icon: 'history',
      title: 'Sessions',
      path: '/sessions',
    },
    {
      icon: 'security',
      title: 'Firewall Rules',
      path: '/firewall/rules',
      hidden: !isEnterprise,
    },
    {
      icon: 'vpn_key',
      title: 'Public Keys',
      path: '/sshkeys/public-keys',
    },
  ];

  const itemsIsNotEnterprise = JSON.parse(JSON.stringify(items));
  itemsIsNotEnterprise.splice(3, 1);

  const itemsIsEnterprise = JSON.parse(JSON.stringify(items));
  itemsIsEnterprise[3].hidden = isEnterprise;

  const storeWithoutNamespace = new Vuex.Store({
    namespaced: true,
    state: {
      tenant,
      isLoggedIn,
      numberNamespaces: numberNamespacesEqualZero,
      isMobile,
      hasSpinner,
    },
    getters: {
      'auth/tenant': (state) => state.tenant,
      'auth/isLoggedIn': (state) => state.isLoggedIn,
      'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
      'mobile/isMobile': (state) => state.isMobile,
      'spinner/getStatus': (state) => state.hasSpinner,
    },
    actions: {
      'auth/logout': () => {},
      'privatekeys/fetch': () => {},
      'mobile/setIsMobileStatus': () => {},
    },
  });

  const storeWithNamespace = new Vuex.Store({
    namespaced: true,
    state: {
      tenant,
      isLoggedIn,
      numberNamespaces: numberNamespacesGreaterThanZero,
      isMobile,
      hasSpinner,
    },
    getters: {
      'auth/tenant': (state) => state.tenant,
      'auth/isLoggedIn': (state) => state.isLoggedIn,
      'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
      'mobile/isMobile': (state) => state.isMobile,
      'spinner/getStatus': (state) => state.hasSpinner,
    },
    actions: {
      'auth/logout': () => {},
      'privatekeys/fetch': () => {},
      'mobile/setIsMobileStatus': () => {},
    },
  });

  ///////
  // In this case, when shellhub is not enterprice and user has
  // not namespace.
  ///////

  describe('Is not enterprise and not has namespace', () => {
    beforeEach(() => {
      wrapper = shallowMount(AppLayout, {
        store: storeWithoutNamespace,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $env: {
            isEnterprise,
          },
        },
        router,
        vuetify,
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
      expect(wrapper.vm.drawer).toEqual(false);
      expect(wrapper.vm.clipped).toEqual(false);
      expect(wrapper.vm.items).toEqual(items);
      expect(wrapper.vm.admins).toEqual(admins);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isLoggedIn).toEqual(true);
      expect(wrapper.vm.visibleItems).toEqual(itemsIsNotEnterprise);
      expect(wrapper.vm.hasNamespaces).toEqual(false);
      expect(wrapper.vm.hasSpinner).toEqual(false);
      expect(wrapper.vm.showNavigationDrawer).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', async () => {
      expect(wrapper.find('[data-test="deviceWarning-component"]').exists()).toEqual(true);
    });
    it('Renders the template with data', async () => {
      items.forEach(async (item) => {
        expect(wrapper.find(`[data-test="${item.icon}-listItem"]`).exists()).toEqual(false);
      });
    });
  });

  ///////
  // In this case, when shellhub is not enterprice and user has
  // namespace.
  ///////

  describe('Is not enterprise and has namespace', () => {
    beforeEach(() => {
      wrapper = shallowMount(AppLayout, {
        store: storeWithNamespace,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $env: {
            isEnterprise,
          },
        },
        router,
        vuetify,
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
      expect(wrapper.vm.drawer).toEqual(false);
      expect(wrapper.vm.clipped).toEqual(false);
      expect(wrapper.vm.items).toEqual(items);
      expect(wrapper.vm.admins).toEqual(admins);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isLoggedIn).toEqual(true);
      expect(wrapper.vm.visibleItems).toEqual(itemsIsNotEnterprise);
      expect(wrapper.vm.hasNamespaces).toEqual(true);
      expect(wrapper.vm.hasSpinner).toEqual(false);
      expect(wrapper.vm.showNavigationDrawer).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', async () => {
      expect(wrapper.find('[data-test="deviceWarning-component"]').exists()).toEqual(true);
    });
    it('Renders the template with data', async () => {
      itemsIsNotEnterprise.forEach(async (item) => {
        expect(wrapper.find(`[data-test="${item.icon}-listItem"]`).exists()).toEqual(true);
      });

      expect(wrapper.find('[data-test="dashboard-security"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when shellhub is enterprice and user has
  // namespace.
  ///////

  describe('Is enterprise and has namespace', () => {
    beforeEach(() => {
      wrapper = shallowMount(AppLayout, {
        store: storeWithNamespace,
        localVue,
        stubs: ['fragment'],
        mocks: {
          $env: {
            isEnterprise: true,
          },
        },
        router,
        vuetify,
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
      expect(wrapper.vm.drawer).toEqual(false);
      expect(wrapper.vm.clipped).toEqual(false);
      expect(wrapper.vm.items).toEqual(itemsIsEnterprise);
      expect(wrapper.vm.admins).toEqual(admins);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.isLoggedIn).toEqual(true);
      expect(wrapper.vm.visibleItems).toEqual(itemsIsEnterprise);
      expect(wrapper.vm.hasNamespaces).toEqual(true);
      expect(wrapper.vm.hasSpinner).toEqual(false);
      expect(wrapper.vm.showNavigationDrawer).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', async () => {
      expect(wrapper.find('[data-test="deviceWarning-component"]').exists()).toEqual(true);
    });
    it('Renders the template with data', async () => {
      items.forEach(async (item) => {
        expect(wrapper.find(`[data-test="${item.icon}-listItem"]`).exists()).toEqual(true);
      });
    });
  });
});
