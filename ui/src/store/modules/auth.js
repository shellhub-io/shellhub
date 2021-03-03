import Vue from 'vue';
import * as apiAuth from '@/store/api/auth';

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
    },

    changeData(state, data) {
      Vue.set(state, 'user', data.username);
      Vue.set(state, 'email', data.email);
    },
  },

  actions: {
    async login(context, user) {
      context.commit('authRequest');

      try {
        const resp = await apiAuth.login(user);

        localStorage.setItem('token', resp.data.token);
        localStorage.setItem('user', resp.data.user);
        localStorage.setItem('name', resp.data.name);
        localStorage.setItem('tenant', resp.data.tenant);
        localStorage.setItem('email', resp.data.email);
        localStorage.setItem('id', resp.data.id);
        localStorage.setItem('namespacesWelcome', JSON.stringify({}));

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
        const resp = await apiAuth.info();

        localStorage.setItem('user', resp.data.user);
        localStorage.setItem('name', resp.data.name);
        localStorage.setItem('tenant', resp.data.tenant);
        localStorage.setItem('id', resp.data.id);
        localStorage.setItem('email', resp.data.email);
        localStorage.setItem('namespacesWelcome', JSON.stringify({}));

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
      localStorage.removeItem('namespacesWelcome');
      localStorage.removeItem('noNamespace');
      localStorage.removeItem('email');
      localStorage.removeItem('id');
      localStorage.removeItem('name');
    },

    changeUserData(context, data) {
      localStorage.setItem('user', data.username);
      localStorage.setItem('email', data.email);
      context.commit('changeData', data);
    },

    setShowWelcomeScreen(context, tenantID) {
      localStorage.setItem('namespacesWelcome', JSON.stringify(
        Object.assign(
          JSON.parse(localStorage.getItem('namespacesWelcome')) || {},
          { ...{ [tenantID]: true } },
        ),
      ));
    },
  },
};
