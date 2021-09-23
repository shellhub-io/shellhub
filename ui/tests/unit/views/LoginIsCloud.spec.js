import Vuex from 'vuex';
import { mount, createLocalVue, config } from '@vue/test-utils';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import Login from '@/views/Login';

config.mocks = {
  $env: {
    isCloud: true,
  },
};

const router = new Router({
  routes: [
    {
      path: '/sign-up',
      name: 'signUp',
      component: () => import('@/views/SignUp'),
    },
    {
      path: '/forgot-password',
      name: 'forgotPassword',
      component: () => import('@/views/ForgotPassword'),
    },
  ],
});

describe('Login', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.use(Router);

  let wrapper;

  const auth = {
    status: true,
    token: 'xxxxxxxx',
    user: 'user',
    tenant: 'xxxxxxxx',
  };

  const store = new Vuex.Store({
    state: {
      auth,
    },
    getters: {
    },
    actions: {
      'auth/logout': () => {},
      'auth/login': () => {},
      'auth/loginToken': () => {},
      'notifications/fetch': () => {},
      'snackbar/showSnackbarErrorIncorrect': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  ///////
  // In this case, the login screen appears to enter the data.
  //////

  describe('Account has been activated', () => {
    beforeEach(() => {
      wrapper = mount(Login, {
        store,
        localVue,
        stubs: ['fragment'],
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
      expect(wrapper.vm.username).toEqual('');
      expect(wrapper.vm.password).toEqual('');
      expect(wrapper.vm.error).toEqual(false);
      expect(wrapper.vm.showMessage).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="accountCreated-component"]').exists()).toEqual(true);
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="username-text"]').element.value).toEqual('');
      expect(wrapper.find('[data-test="password-text"]').element.value).toEqual('');
      expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="forgotPassword-card"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(true);
    });
  });

  ///////
  // In this case, user tries to login but your account has not been activated.
  //////

  describe('Account has not been activated', () => {
    beforeEach(() => {
      wrapper = mount(Login, {
        store,
        localVue,
        stubs: ['fragment'],
        router,
        vuetify,
      });

      wrapper.setData({ showMessage: true });
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
      expect(wrapper.vm.username).toEqual('');
      expect(wrapper.vm.password).toEqual('');
      expect(wrapper.vm.error).toEqual(false);
      expect(wrapper.vm.showMessage).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="accountCreated-component"]').exists()).toEqual(true);
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="username-text"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="password-text"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="forgotPassword-card"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(false);
    });
  });
});
