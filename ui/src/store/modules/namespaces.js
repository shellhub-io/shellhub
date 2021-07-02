import Vue from 'vue';
import * as apiNamespace from '@/store/api/namespaces';

export default {
  namespaced: true,

  state: {
    namespace: {},
    namespaces: [],
    numberNamespaces: 0,
    owner: false,
    webhookUrl: '',
    active: false,
  },

  getters: {
    list: (state) => state.namespaces,
    get: (state) => state.namespace,
    getNumberNamespaces: (state) => state.numberNamespaces,
    owner: (state) => state.owner,
    webhookUrl: (state) => state.webhookUrl,
    webhookActive: (state) => state.active,
  },

  mutations: {
    setNamespaces: (state, res) => {
      Vue.set(state, 'namespaces', res.data);
      Vue.set(state, 'numberNamespaces', parseInt(res.headers['x-total-count'], 10));
    },

    setNamespace: (state, res) => {
      Vue.set(state, 'namespace', res.data);
      if (res.data.settings.webhook !== undefined) {
        const { url, active } = res.data.settings.webhook;
        Vue.set(state, 'webhookUrl', url);
        Vue.set(state, 'active', active);
      }
    },

    removeNamespace: (state, id) => {
      state.namespaces.splice(state.namespaces.findIndex((d) => d.tenant_id === id), 1);
    },

    removeMember: (state, usr) => {
      state.namespace.members.splice(state.namespace.members.findIndex(
        (m) => m.name === usr,
      ), 1);
    },

    clearNamespaceList: (state) => {
      Vue.set(state, 'namespaces', []);
      Vue.set(state, 'numberNamespaces', 0);
    },

    clearObjectNamespace: (state) => {
      Vue.set(state, 'namespace', {});
    },

    setOwnerStatus: (state, status) => {
      Vue.set(state, 'owner', status);
    },

    setWebhook: (state, data) => {
      Vue.set(state, 'webhookUrl', data.url);
    },

    setWebhookStatus: (state, status) => {
      Vue.set(state, 'active', status);
    },
  },

  actions: {
    post: async (context, data) => {
      const res = await apiNamespace.postNamespace(data);
      return res;
    },

    fetch: async (context) => {
      const res = await apiNamespace.fetchNamespaces();
      context.commit('setNamespaces', res);
    },

    get: async (context, id) => {
      const res = await apiNamespace.getNamespace(id);
      context.commit('setNamespace', res);
    },

    put: async (context, data) => {
      await apiNamespace.putNamespace(data);
    },

    remove: async (context, id) => {
      await apiNamespace.removeNamespace(id);
      context.commit('removeNamespace', id);
      context.commit('clearObjectNamespace');
      context.commit('clearNamespaceList');
    },

    addUser: async (context, data) => {
      await apiNamespace.addUserToNamespace(data);
    },

    removeUser: async (context, data) => {
      const res = await apiNamespace.removeUserFromNamespace(data);
      if (res.status === 200) {
        context.commit('removeMember', data.username);
      }
    },

    clearNamespaceList: (context) => {
      context.commit('clearNamespaceList');
    },

    switchNamespace: async (context, data) => {
      const res = await apiNamespace.tenantSwitch(data);
      if (res.status === 200) {
        localStorage.setItem('token', res.data.token);
        localStorage.setItem('tenant', data.tenant_id);
      }
    },

    setOwnerStatus: async (context, status) => {
      context.commit('setOwnerStatus', status);
    },

    updateWebhookStatus: async (context, data) => {
      const res = await apiNamespace.webhookStatusUpdate(data);
      if (res.status === 200) {
        context.commit('setWebhookStatus', data.status);
      }
    },

    updateWebhook: async (context, data) => {
      const res = await apiNamespace.webhookUpdate(data);
      if (res.status === 200) {
        context.commit('setWebhook', data);
      }
    },
  },
};
