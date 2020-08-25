import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Login from '@/views/Login';

const localVue = createLocalVue();
localVue.use(Vuex);

const store = new Vuex.Store({
  state: {
    auth: {
      status: true,
      token: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      user: 'user',
      tenant: 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoibGV',
    },
  },
  getters: {
  },
});

describe('Login', () => {
  const wrapper = shallowMount(Login, {
    store,
    localVue,
    mocks: {
      $route: {
        query: {},
      },
    },
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Has a username field', () => {
    expect(wrapper.find('[data-test="username-text"]').exists()).toBe(true);
  });
  it('Has a password field', () => {
    expect(wrapper.find('[data-test="password-text"]').exists()).toBe(true);
  });
  it('Has a button', () => {
    expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
  });
  it('Verify the variables and fields with data', () => {
    wrapper.setData({ username: 'ShellHub' });
    wrapper.setData({ password: 'ShellHub' });
    expect(wrapper.vm.username).toEqual('ShellHub');
    expect(wrapper.vm.password).toEqual('ShellHub');

    const textInputUsername = wrapper.find('[data-test="username-text"]');
    const textInputPassword = wrapper.find('[data-test="password-text"]');

    textInputUsername.element.value = 'ShellHub';
    textInputPassword.element.value = 'ShellHub';

    expect(textInputUsername.element.value).toEqual('ShellHub');
    expect(textInputPassword.element.value).toEqual('ShellHub');
  });
});
