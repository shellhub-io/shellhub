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

    async patchData(context, data) {
      await apiUser.patchUserData(data);
    },

    async patchPassword(context, data) {
      await apiUser.patchUserPassword(data);
    },
  },
};
