import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import AppLayout from '@/layouts/AppLayout';

const router = new Router({
  routes: [
    {
      path: '',
      name: 'dashboard',
      component: () => import(/* webpackChunkName: "dashboard" */ '@/views/Dashboard'),
    },
  ],
});

describe('AppLayout', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  localVue.use(Router);
  const vuetify = new Vuetify();

  let wrapper;

  const tenant = 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx';
  const isLoggedIn = true;
  const isMobile = false;
  const numberNamespaces = 0;
  const hasSpinner = false;
  const statusNavigationDrawer = false;
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
    {
      icon: 'mdi-cog',
      title: 'Settings',
      path: '/settings/namespace-manager',
    },
  ];

  const itemsIsNotEnterprise = JSON.parse(JSON.stringify(items));
  itemsIsNotEnterprise.splice(3, 1);

  const itemsIsEnterprise = JSON.parse(JSON.stringify(items));
  itemsIsEnterprise[3].hidden = isEnterprise;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      tenant,
      isLoggedIn,
      numberNamespaces,
      isMobile,
      hasSpinner,
      statusNavigationDrawer,
    },
    getters: {
      'auth/isLoggedIn': (state) => state.isLoggedIn,
      'namespaces/getNumberNamespaces': (state) => state.numberNamespaces,
      'mobile/isMobile': (state) => state.isMobile,
      'spinner/getStatus': (state) => state.hasSpinner,
      'layout/getStatusNavigationDrawer': (state) => state.statusNavigationDrawer,
    },
    actions: {
      'privatekeys/fetch': () => {},
      'mobile/setIsMobileStatus': () => {},
      'layout/setStatusNavigationDrawer': () => {},
    },
  });

  ///////
  // In this case, when shellhub is not enterprice
  ///////

  describe('Is not enterprise', () => {
    beforeEach(() => {
      wrapper = shallowMount(AppLayout, {
        store,
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
      expect(wrapper.vm.items).toEqual(items);
      expect(wrapper.vm.admins).toEqual(admins);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.visibleItems).toEqual(itemsIsNotEnterprise);
      expect(wrapper.vm.hasNamespaces).toEqual(false);
      expect(wrapper.vm.hasSpinner).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', async () => {
      expect(wrapper.find('[data-test="namespace-component"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="userWarning-component"]').exists()).toEqual(true);
    });
    items.forEach(async (item) => {
      it(`Renders the template with data - icon ${item.icon}`, async () => {
        if (item.icon !== 'security') {
          expect(wrapper.find(`[data-test="${item.icon}-listItem"]`).exists()).toEqual(true);
        } else {
          expect(wrapper.find(`[data-test="${item.icon}-listItem"]`).exists()).toEqual(false);
        }
      });
    });
  });

  ///////
  // In this case, when shellhub is enterprice and user has
  // namespace.
  ///////

  describe('Is enterprise and has namespace', () => {
    beforeEach(() => {
      wrapper = shallowMount(AppLayout, {
        store,
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
      expect(wrapper.vm.items).toEqual(itemsIsEnterprise);
      expect(wrapper.vm.admins).toEqual(admins);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.visibleItems).toEqual(itemsIsEnterprise);
      expect(wrapper.vm.hasNamespaces).toEqual(false);
      expect(wrapper.vm.hasSpinner).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', async () => {
      expect(wrapper.find('[data-test="namespace-component"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="userWarning-component"]').exists()).toEqual(true);
    });
    items.forEach(async (item) => {
      it(`Renders the template with data - icon ${item.icon}`, async () => {
        expect(wrapper.find(`[data-test="${item.icon}-listItem"]`).exists()).toEqual(true);
      });
    });
  });
});
