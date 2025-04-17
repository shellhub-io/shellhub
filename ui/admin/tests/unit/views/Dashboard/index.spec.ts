import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useStatsStore from "@admin/store/modules/stats";
import useSnackbarStore from "@admin/store/modules/snackbar";
import routes from "../../../../src/router";
import Dashboard from "../../../../src/views/Dashboard.vue";

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
    id: 0,
    title: "Registered Users",
    fieldObject: "registered_users",
    content: "Registered users",
    icon: "mdi-account-group",
    stats: stats.registered_users,
    buttonName: "View all Users",
    pathName: "users",
    nameUseTest: "viewUsers-btn",
  },
  {
    id: 1,
    title: "Registered Devices",
    fieldObject: "registered_devices",
    content: "Registered devices",
    icon: "mdi-devices",
    stats: stats.registered_devices,
    buttonName: "View all Devices",
    pathName: "devices",
    nameUseTest: "viewRegisteredDevices-btn",
  },
  {
    id: 2,
    title: "Online Devices",
    fieldObject: "online_devices",
    content: "Devices are online and ready for connecting",
    icon: "mdi-devices",
    stats: stats.online_devices,
    buttonName: "View all Devices",
    pathName: "devices",
    nameUseTest: "viewOnlineDevices-btn",
  },
  {
    id: 3,
    title: "Active Sessions",
    fieldObject: "active_sessions",
    content: "Active SSH Sessions opened by users",
    icon: "mdi-devices",
    stats: stats.active_sessions,
    buttonName: "View all Sessions",
    pathName: "sessions",
    nameUseTest: "viewActiveSession-btn",
  },
  {
    id: 4,
    title: "Pending Devices",
    fieldObject: "pending_devices",
    content: "Pending devices",
    icon: "mdi-devices",
    stats: stats.pending_devices,
    buttonName: "View all Devices",
    pathName: "devices",
    nameUseTest: "viewPendingDevices-btn",
  },
  {
    id: 5,
    title: "Rejected Devices",
    fieldObject: "rejected_devices",
    content: "Rejected devices",
    icon: "mdi-devices",
    stats: stats.rejected_devices,
    buttonName: "View all Devices",
    pathName: "devices",
    nameUseTest: "viewRejectedDevices-btn",
  },
];

const numberOfCards = 6;

const itemsTestDashboard = [
  "viewUsers-btn",
  "viewRegisteredDevices-btn",
  "viewOnlineDevices-btn",
  "viewActiveSession-btn",
  "viewPendingDevices-btn",
  "viewRejectedDevices-btn",
];

describe("Dashboard", () => {
  let wrapper: DashboardWrapper;

  beforeEach(async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const statsStore = useStatsStore();
    statsStore.get = vi.fn().mockResolvedValue(undefined);
    statsStore.stats = stats;

    const snackbarStore = useSnackbarStore();
    snackbarStore.showSnackbarErrorAction = vi.fn();

    const vuetify = createVuetify();

    wrapper = mount(Dashboard, {
      global: {
        plugins: [pinia, vuetify, routes],
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

  itemsTestDashboard.forEach((testItem) => {
    it(`Renders the Card ${testItem} with data`, () => {
      expect(wrapper.find(`[data-test=${testItem}]`).exists()).toBe(true);
    });
  });
});
