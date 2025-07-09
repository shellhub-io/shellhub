import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useStatsStore from "@admin/store/modules/stats";
import routes from "@admin/router";
import Dashboard from "@admin/views/Dashboard.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DashboardWrapper = VueWrapper<InstanceType<typeof Dashboard>>;

const stats = {
  registered_users: 0,
  registered_devices: 0,
  online_devices: 0,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

const cardsContent = [
  {
    title: "Registered Users",
    content: "Registered users",
    icon: "mdi-account-group",
    buttonLabel: "View all Users",
    path: "users",
    stat: stats.registered_users,
  },
  {
    title: "Registered Devices",
    content: "Registered devices",
    icon: "mdi-devices",
    buttonLabel: "View all Devices",
    path: "devices",
    stat: stats.registered_devices,
  },
  {
    title: "Online Devices",
    content: "Devices are online and ready for connecting",
    icon: "mdi-devices",
    buttonLabel: "View all Devices",
    path: "devices",
    stat: stats.online_devices,
  },
  {
    title: "Active Sessions",
    content: "Active SSH Sessions opened by users",
    icon: "mdi-devices",
    buttonLabel: "View all Sessions",
    path: "sessions",
    stat: stats.active_sessions,
  },
  {
    title: "Pending Devices",
    content: "Pending devices",
    icon: "mdi-devices",
    buttonLabel: "View all Devices",
    path: "devices",
    stat: stats.pending_devices,
  },
  {
    title: "Rejected Devices",
    content: "Rejected devices",
    icon: "mdi-devices",
    buttonLabel: "View all Devices",
    path: "devices",
    stat: stats.rejected_devices,
  },
];

const numberOfCards = 6;

describe("Dashboard", () => {
  let wrapper: DashboardWrapper;

  beforeEach(async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const statsStore = useStatsStore();
    statsStore.get = vi.fn().mockResolvedValue(undefined);
    statsStore.stats = stats;

    const vuetify = createVuetify();

    wrapper = mount(Dashboard, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
      },
    });

    await statsStore.get();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", async () => {
    expect(wrapper.vm.items).toEqual(cardsContent);
    expect(wrapper.vm.itemsStats).toEqual(stats);
    expect(wrapper.vm.hasStatus).toBe(false);
  });

  it("Must show all the card in the view", () => {
    expect(wrapper.findAll("div.v-card").length).toBe(numberOfCards);
  });
});
