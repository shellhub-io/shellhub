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
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('has a username field', () => {
    expect(wrapper.find('[data-cy="username-text"]').exists()).toBe(true);
  });
  it('has a password field', () => {
    expect(wrapper.find('[data-cy="password-text"]').exists()).toBe(true);
  });
  it('has a button', () => {
    expect(wrapper.find('[data-cy="login-btn"]').exists()).toBe(true);
  });
});
