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
    email: localStorage.getItem('email') || '',
    id: localStorage.getItem('id') || '',
    namespaceUsedToShowWelcomeScreen: [],
  },

  getters: {
    isLoggedIn: (state) => !!state.token,
    authStatus: (state) => state.status,
    currentUser: (state) => state.user,
    currentName: (state) => state.name,
    tenant: (state) => state.tenant,
    email: (state) => state.email,
    id: (state) => state.id,
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
      Vue.set(state, 'email', data.email);
      Vue.set(state, 'id', data.id);
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
      Vue.set(state, 'email', '');
      Vue.set(state, 'namespaceUsedToWelcomeScreen', []);
    },

    changeData(state, data) {
      Vue.set(state, 'user', data.username);
      Vue.set(state, 'email', data.email);
    },

    setShowWelcomeScreen(state, tenantId) {
      state.namespaceUsedToShowWelcomeScreen.push(tenantId);
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
        localStorage.setItem('email', resp.data.email);
        localStorage.setItem('id', resp.data.id);

        context.commit('authSuccess', resp.data);
      } catch (err) {
        context.commit('authError');
        throw err;
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

    logout: async ({ commit, state }) => {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('tenant');
      localStorage.removeItem('noNamespace');
      localStorage.removeItem('email');
      localStorage.removeItem('id');
      localStorage.removeItem('name');

      await state.namespaceUsedToShowWelcomeScreen.forEach((item) => {
        localStorage.removeItem(item);
      });
      commit('logout');
    },

    changeUserData(context, data) {
      localStorage.setItem('user', data.username);
      localStorage.setItem('email', data.email);
      context.commit('changeData', data);
    },

    setShowWelcomeScreen(context, tenantId) {
      context.commit('setShowWelcomeScreen', tenantId);
    },
  },
};
