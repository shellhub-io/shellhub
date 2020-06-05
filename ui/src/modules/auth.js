import { login } from '@/api/auth';

export default {
  namespaced: true,

  state: {
    status: '',
    token: localStorage.getItem('token') || '',
    user: localStorage.getItem('user') || '',
    tenant: localStorage.getItem('tenant') || ''
  },

  getters: {
    isLoggedIn: (state) => !!state.token,
    authStatus: (state) => state.status,
    currentUser: (state) => state.user,
    tenant: (state) => state.tenant,
  },

  mutations: {
    authRequest(state) {
      state.status = 'loading';
    },

    authSuccess(state, data) {
      state.status = 'success';
      state.token = data.token;
      state.user = data.user;
      state.tenant = data.tenant;
    },

    authError(state) {
      state.status = 'error';
    },

    logout(state) {
      state.status = '';
      state.token = '';
      state.user = '';
      state.tenant = '';
    },
  },

  actions: {
    async login(context, user) {
      context.commit('authRequest');

      try {
        const resp = await login(user);

        localStorage.setItem('token', resp.data.token);
        localStorage.setItem('user', resp.data.user);
        localStorage.setItem('tenant', resp.data.tenant);

        context.commit('authSuccess', resp.data);
      } catch (err) {
        context.commit('authError');
      }
    },

    logout(context) {
      context.commit('logout');
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('tenant');
      localStorage.removeItem('onceWelcome');
    }
  }
};
