import Vue from 'vue';
import { login, info } from '@/store/api/auth';

export default {
  namespaced: true,

  state: {
    status: '',
    token: localStorage.getItem('token') || '',
    user: localStorage.getItem('user') || '',
    name: localStorage.getItem('name') || '',
    tenant: localStorage.getItem('tenant') || '',
  },

  getters: {
    isLoggedIn: (state) => !!state.token,
    authStatus: (state) => state.status,
    currentUser: (state) => state.user,
    currentName: (state) => state.name,
    tenant: (state) => state.tenant,
  },

  mutations: {
    authRequest(state) {
      Vue.set(state, 'status', 'loading');
    },

    authSuccess(state, data) {
      Vue.set(state, 'status', 'success');
      Vue.set(state, 'token', data.token);
      Vue.set(state, 'user', data.user);
      Vue.set(state, 'name', data.name);
      Vue.set(state, 'tenant', data.tenant);
    },

    authError(state) {
      Vue.set(state, 'status', 'error');
    },

    logout(state) {
      Vue.set(state, 'status', '');
      Vue.set(state, 'token', '');
      Vue.set(state, 'name', '');
      Vue.set(state, 'user', '');
      Vue.set(state, 'tenant', '');
    },
  },

  actions: {
    async login(context, user) {
      context.commit('authRequest');

      try {
        const resp = await login(user);

        localStorage.setItem('token', resp.data.token);
        localStorage.setItem('user', resp.data.user);
        localStorage.setItem('name', resp.data.name);
        localStorage.setItem('tenant', resp.data.tenant);

        context.commit('authSuccess', resp.data);
      } catch (err) {
        context.commit('authError');
      }
    },

    async loginToken(context, token) {
      context.commit('authRequest');

      localStorage.setItem('token', token);

      try {
        const resp = await info();

        localStorage.setItem('user', resp.data.user);
        localStorage.setItem('name', resp.data.name);
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
    },
  },
};
