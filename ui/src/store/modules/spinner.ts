import { defineStore } from "pinia";
import { ref } from "vue";

const useSpinnerStore = defineStore("spinner", () => {
  const status = ref(false);

  return { status };
});

export default useSpinnerStore;
