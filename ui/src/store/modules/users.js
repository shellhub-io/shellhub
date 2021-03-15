import * as apiUser from '@/store/api/users';

export default {
  namespaced: true,

  state: {
  },

  getters: {
  },

  mutations: {
  },

  actions: {
    async signUp(context, data) {
      await apiUser.signUp(data);
    },

    async put(context, data) {
      await apiUser.putUser(data);
    },
  },
};
