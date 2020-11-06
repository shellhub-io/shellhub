import Vue from 'vue';
import {
  postNamespace,
  fetchNamespaces,
  getNamespace,
  removeNamespace,
  putNamespace,
  addUserToNamespace,
  removeUserFromNamespace,
  tenantSwitch,
} from '@/store/api/namespaces';

export default {
  namespaced: true,

  state: {
    namespace: {},
    namespaces: [],
    numberNamespaces: 0,
  },

  getters: {
    list: (state) => state.namespaces,
    get: (state) => state.namespace,
    getNumberNamespaces: (state) => state.numberNamespaces,
  },

  mutations: {
    setNamespaces: (state, res) => {
      Vue.set(state, 'namespaces', res.data);
      Vue.set(state, 'numberNamespaces', parseInt(res.headers['x-total-count'], 10));
    },

    setNamespace: (state, res) => {
      Vue.set(state, 'namespace', res.data);
    },

    removeNamespace: (state, id) => {
      state.namespaces.splice(state.namespaces.findIndex((d) => d.tenant_id === id), 1);
    },

    removeMember: (state, usr) => {
      state.namespace.member_names.splice(state.namespace.member_names.findIndex(
        (m) => m === usr,
      ), 1);
    },
  },

  actions: {
    post: async (context, data) => {
      await postNamespace(data);
    },

    fetch: async (context) => {
      const res = await fetchNamespaces();
      context.commit('setNamespaces', res);
    },

    get: async (context, id) => {
      const res = await getNamespace(id);
      context.commit('setNamespace', res);
    },

    put: async (context, data) => {
      const res = await putNamespace(data);
      context.commit('setNamespace', res);
    },

    remove: async (context, id) => {
      await removeNamespace(id);
      context.commit('removeNamespace', id);
    },

    addUser: async (context, data) => {
      const res = await addUserToNamespace(data);
      if (res.status === 200) {
        context.commit('setNamespace', res);
      }
    },

    removeUser: async (context, data) => {
      const res = await removeUserFromNamespace(data);
      if (res.status === 200) {
        context.commit('removeMember', data.username);
      }
    },

    switchNamespace: async (context, data) => {
      await tenantSwitch(data);
    },
  },
};
