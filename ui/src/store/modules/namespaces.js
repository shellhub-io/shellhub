import Vue from 'vue';
import * as apiNamespace from '@/store/api/namespaces';

export default {
  namespaced: true,

  state: {
    namespace: {},
    billInfoData: {},
    billing: {},
    namespaces: [],
    numberNamespaces: 0,
    owner: false,
  },

  getters: {
    list: (state) => state.namespaces,
    get: (state) => state.namespace,
    getNumberNamespaces: (state) => state.numberNamespaces,
    owner: (state) => state.owner,
    billing: (state) => state.billing,
  },

  mutations: {
    setNamespaces: (state, res) => {
      Vue.set(state, 'namespaces', res.data);
      Vue.set(state, 'numberNamespaces', parseInt(res.headers['x-total-count'], 10));
    },

    setNamespace: (state, res) => {
      Vue.set(state, 'namespace', res.data);
    },

    setBilling: (state, data) => {
      Vue.set(state, 'billing', data);
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

      const { billing } = res.data;
      if (billing !== null) {
        context.commit('setBilling', billing);
      }
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

    editUser: async (context, data) => {
      await apiNamespace.editUserToNamespace(data);
    },

    removeUser: async (context, data) => {
      await apiNamespace.removeUserFromNamespace(data);
    },

    clearNamespaceList: (context) => {
      context.commit('clearNamespaceList');
    },

    switchNamespace: async (context, data) => {
      localStorage.removeItem('accessType');

      const res = await apiNamespace.tenantSwitch(data);
      if (res.status === 200) {
        localStorage.setItem('token', res.data.token);
        localStorage.setItem('tenant', data.tenant_id);
        localStorage.setItem('accessType', res.data.type);
      }
    },

    setOwnerStatus: async (context, status) => {
      context.commit('setOwnerStatus', status);
    },
  },
};
