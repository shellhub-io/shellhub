import putUser from '@/store/api/users';

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
      await putUser(data);
    },
  },
};
