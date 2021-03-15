import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import SignUp from '@/views/SignUp';
import Vuetify from 'vuetify';
import flushPromises from 'flush-promises';
import router from '@/router/index';
import { ValidationProvider, ValidationObserver, extend } from 'vee-validate';
import { required, email } from 'vee-validate/dist/rules';
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

describe('SignUp', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  let wrapper;

  const store = new Vuex.Store({
    state: {
    },
    getters: {
    },
    actions: {
      'users/signUp': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
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

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="name-text"]').element.value).toEqual(newUser.name);
    expect(wrapper.find('[data-test="username-text"]').element.value).toEqual(newUser.username);
    expect(wrapper.find('[data-test="email-text"]').element.value).toEqual(newUser.email);
    expect(wrapper.find('[data-test="password-text"]').element.value).toEqual(newUser.password);
    expect(wrapper.find('[data-test="confirmPassword-text"]').element.value).toEqual(newUser.confirmPassword);
  });
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
});
