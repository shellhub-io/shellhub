import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import flushPromises from 'flush-promises';
import { ValidationProvider, ValidationObserver, extend } from 'vee-validate';
import { required, email } from 'vee-validate/dist/rules';
import Router from 'vue-router';
import SignUp from '@/views/SignUp';
import '@/vee-validate';

extend('required', {
  ...required,
  message: 'This field is required',
});

extend('email', {
  ...email,
  message: 'This field must be a valid email',
});

const newUser = {
  name: '',
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
};

const router = new Router({
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/Login'),
    },
  ],
});

describe('SignUp', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.use(Router);

  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  let wrapper;

  const store = new Vuex.Store({
    state: {
    },
    getters: {
    },
    actions: {
      'users/signUp': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  beforeEach(() => {
    wrapper = mount(SignUp, {
      store,
      stubs: ['fragment'],
      localVue,
      vuetify,
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
    expect(wrapper.vm.newUser.name).toEqual(newUser.name);
    expect(wrapper.vm.newUser.username).toEqual(newUser.username);
    expect(wrapper.vm.newUser.email).toEqual(newUser.email);
    expect(wrapper.vm.newUser.password).toEqual(newUser.password);
    expect(wrapper.vm.newUser.confirmPassword).toEqual(newUser.confirmPassword);
    expect(wrapper.vm.delay).toEqual(500);
    expect(wrapper.vm.overlay).toEqual(false);
  });

  //////
  // HTML validation
  //////

  it('Show empty fields required in validation', async () => {
    const validatorName = wrapper.vm.$refs.providerName;
    const validatorUsername = wrapper.vm.$refs.providerUsername;
    const validatorEmail = wrapper.vm.$refs.providerEmail;
    const validatorPassword = wrapper.vm.$refs.providerPassword;
    const validatorConfirmPassword = wrapper.vm.$refs.providerConfirmPassword;

    await validatorName.validate();
    expect(validatorName.errors[0]).toBe('This field is required');
    await validatorUsername.validate();
    expect(validatorUsername.errors[0]).toBe('This field is required');
    await validatorEmail.validate();
    expect(validatorEmail.errors[0]).toBe('This field is required');
    await validatorPassword.validate();
    expect(validatorPassword.errors[0]).toBe('This field is required');
    await validatorConfirmPassword.validate();
    expect(validatorConfirmPassword.errors[0]).toBe('This field is required');

    wrapper.setData({ newUser: { password: 's', confirmPassword: 'h' } });
    await flushPromises();

    await validatorPassword.validate();
    expect(validatorPassword.errors[0]).toBe('Your password should be 5-30 characters long');
    await validatorConfirmPassword.validate();
    expect(validatorConfirmPassword.errors[0]).toBe('The passwords do not match');
  });
  it('Renders the template with components', () => {
    expect(wrapper.find('[data-test="accountCreated-component"]').exists()).toEqual(true);
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="name-text"]').element.value).toEqual(newUser.name);
    expect(wrapper.find('[data-test="username-text"]').element.value).toEqual(newUser.username);
    expect(wrapper.find('[data-test="email-text"]').element.value).toEqual(newUser.email);
    expect(wrapper.find('[data-test="password-text"]').element.value).toEqual(newUser.password);
    expect(wrapper.find('[data-test="confirmPassword-text"]').element.value).toEqual(newUser.confirmPassword);
    expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
  });
});
