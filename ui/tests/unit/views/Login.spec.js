import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Login from '@/views/Login';
import router from '@/router/index';

describe('Login', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const auth = {
    status: true,
    token: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
    user: 'user',
    tenant: 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoibGV',
  };

  const store = new Vuex.Store({
    state: {
      auth,
    },
    getters: {
    },
    actions: {
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(Login, {
      store,
      localVue,
      router,
      mocks: {
        $route: {
          query: {},
        },
      },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.username).toEqual(null);
    expect(wrapper.vm.password).toEqual(null);
  });
  it('Renders the template with data', () => {
    const textInputUsername = wrapper.find('[data-test="username-text"]');
    const textInputPassword = wrapper.find('[data-test="password-text"]');

    expect(textInputUsername.element.value).toEqual(undefined);
    expect(textInputPassword.element.value).toEqual(undefined);

    textInputUsername.element.value = 'ShellHub';
    textInputPassword.element.value = 'ShellHub';
    expect(textInputUsername.element.value).toEqual('ShellHub');
    expect(textInputPassword.element.value).toEqual('ShellHub');

    expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
  });
});
