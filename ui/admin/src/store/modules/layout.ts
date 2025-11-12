import { defineStore } from "pinia";
import { ref } from "vue";

export type Layout = "AppLayout";
type Theme = "dark" | "light";

const useLayoutStore = defineStore("layout", () => {
  const layout = ref<Layout>("AppLayout");
  const theme = ref<Theme>(localStorage.getItem("theme") as Theme || "dark");

  const setTheme = (newTheme: Theme) => {
    theme.value = newTheme;
    localStorage.setItem("theme", newTheme);
  };

  return {
    layout,
    theme,
    setTheme,
  };
});

export default useLayoutStore;
