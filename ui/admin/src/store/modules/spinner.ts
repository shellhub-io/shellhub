import { defineStore } from "pinia";

export const useSpinnerStore = defineStore("spinner", {
  state: () => ({
    status: false,
  }),

  getters: {
    getStatus: (state) => state.status,
  },

  actions: {
    setStatus(status: boolean) {
      this.status = status;
    },
  },
});

export default useSpinnerStore;
