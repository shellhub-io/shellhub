import { defineStore } from "pinia";
import premiumContent from "../api/global";

type PremiumItem = {
  title: string;
  features: string[];
  button: {
    label: string;
    link: string;
  };
};

const useGlobalStore = defineStore("global", {
  state: () => ({
    premiumContent: [] as PremiumItem[],
  }),
  actions: {
    async getPaywallPremiumContent() {
      try {
        const res = await premiumContent();
        this.premiumContent = res;
      } catch (error) {
        console.error(error);
        throw error;
      }
    },
  },
});

export default useGlobalStore;
