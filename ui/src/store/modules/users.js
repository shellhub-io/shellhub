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
    async put(context, data) {
      await apiUser.putUser(data);
    },
  },
};
