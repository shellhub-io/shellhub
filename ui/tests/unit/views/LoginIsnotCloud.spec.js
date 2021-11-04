import Vuex from 'vuex';
import { mount, createLocalVue, config } from '@vue/test-utils';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import Login from '@/views/Login';

config.mocks = {
  $env: {
    isCloud: false,
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
      'layout/setLayout': () => {},
      'snackbar/showSnackbarErrorIncorrect': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  describe('Login screen', () => {
    beforeEach(() => {
      wrapper = mount(Login, {
        store,
        localVue,
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
      expect(wrapper.vm.showPassword).toEqual(false);
      expect(wrapper.vm.showMessage).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="accountCreated-component"]').exists()).toEqual(false);
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="username-text"]').element.value).toEqual('');
      expect(wrapper.find('[data-test="password-text"]').element.value).toEqual('');
      expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="unknownReason-card"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="forgotPassword-card"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(false);
    });
  });

  describe('Login screen unknown reason', () => {
    beforeEach(() => {
      wrapper = mount(Login, {
        store,
        localVue,
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
      expect(wrapper.vm.showPassword).toEqual(false);
      expect(wrapper.vm.showMessage).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="accountCreated-component"]').exists()).toEqual(false);
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="username-text"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="password-text"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="unknownReason-card"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="forgotPassword-card"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(false);
    });
  });
});
