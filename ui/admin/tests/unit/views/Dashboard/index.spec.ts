import { createVuetify } from "vuetify";
import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useStatsStore from "@admin/store/modules/stats";
import routes from "@admin/router";
import Dashboard from "@admin/views/Dashboard.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

const stats = {
  registered_users: 0,
  registered_devices: 0,
  online_devices: 0,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

const numberOfCards = 7;

describe("Dashboard", () => {
  const pinia = createPinia();
  setActivePinia(pinia);
  const statsStore = useStatsStore();
  statsStore.getStats = vi.fn().mockResolvedValue(stats);
  const vuetify = createVuetify();

  const wrapper = mount(Dashboard, {
    global: {
      plugins: [pinia, vuetify, routes, SnackbarPlugin],
    },
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", async () => {
    expect(wrapper.vm.stats).toEqual(stats);
    expect(wrapper.vm.hasStatus).toBe(false);
  });

  it("Must show all the card in the view", () => {
    // Main card and 6 stat cards
    expect(wrapper.findAll("div.v-card").length).toBe(numberOfCards);
  });
});
