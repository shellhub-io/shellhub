// stores/layout.ts
import { defineStore } from "pinia";

export const useLayoutStore = defineStore("layout", {
  state: () => ({
    layout: "appLayout" as string,
    statusDarkMode: (localStorage.getItem("statusDarkMode") || "dark") as string,
  }),

  getters: {
    getLayout: (state) => state.layout,
    getStatusDarkMode: (state) => state.statusDarkMode,
  },

  actions: {
    setLayout(layout: string) {
      this.layout = layout;
    },

    setStatusDarkMode(status: boolean) {
      const statusDarkMode = status ? "dark" : "light";
      this.statusDarkMode = statusDarkMode;
      localStorage.setItem("statusDarkMode", statusDarkMode);
    },
  },
});

export default useLayoutStore;
