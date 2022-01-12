import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import App from '@/App';

describe('App', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const isLoggedIn = false;
  const authId = 'xxxxxxxx';
  const layouts = ['simpleLayout', 'appLayout'];
  const statusDarkMode = true;

  const storeNotLoggedIn = new Vuex.Store({
    namespaced: true,
    state: {
      layout: layouts[0],
      isLoggedIn,
      authId,
    },
    getters: {
      'layout/getLayout': (state) => state.layout,
      'auth/isLoggedIn': (state) => state.isLoggedIn,
      'auth/id': (state) => state.authId,
      'layout/getStatusDarkMode': (state) => state.statusDarkMode,
    },
    actions: {
      'layout/setLayout': () => {},
    },
  });

  const storeLoggedIn = new Vuex.Store({
    namespaced: true,
    state: {
      layout: layouts[1],
      isLoggedIn: !isLoggedIn,
      authId,
      statusDarkMode,
    },
    getters: {
      'layout/getLayout': (state) => state.layout,
      'auth/isLoggedIn': (state) => state.isLoggedIn,
      'auth/id': (state) => state.authId,
      'layout/getStatusDarkMode': (state) => state.statusDarkMode,
    },
    actions: {
      'layout/setLayout': () => {},
    },
  });

  ///////
  // In this case, is to check the rendering app layout when the user
  // is not logged in.
  ///////

  describe('Not logged in', () => {
    beforeEach(() => {
      wrapper = shallowMount(App, {
        store: storeNotLoggedIn,
        localVue,
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.layout).toEqual(layouts[0]);
      expect(wrapper.vm.isLoggedIn).toEqual(isLoggedIn);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="snackbar-component"]').exists()).toBe(true);
      expect(wrapper.find(`[data-test="${layouts[0]}-component"]`).exists()).toBe(true);
    });
  });

  ///////
  // In this case, is to check the rendering app layout when the user
  // is logged in.
  ///////

  describe('Logged in', () => {
    beforeEach(() => {
      wrapper = shallowMount(App, {
        store: storeLoggedIn,
        localVue,
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.layout).toEqual(layouts[1]);
      expect(wrapper.vm.isLoggedIn).toEqual(!isLoggedIn);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="snackbar-component"]').exists()).toBe(true);
      expect(wrapper.find(`[data-test="${layouts[1]}-component"]`).exists()).toBe(true);
    });
  });
});
