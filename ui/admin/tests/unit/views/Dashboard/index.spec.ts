import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import Dashboard from "../../../../src/views/Dashboard.vue";

type DashboardWrapper = VueWrapper<InstanceType<typeof Dashboard>>;

const stats = {
  registeredUsers: 0,
  registeredDevices: 0,
  onlineDevices: 0,
  activeSessions: 0,
  pendingDevices: 0,
  rejectedDevices: 0,
};

const numberOfCards = 6;

const itemsTestDashboard = [
  "viewUsers-btn",
  "viewRegisteredDevices-btn",
  "viewOnlineDevices-btn",
  "viewActiveSession-btn",
  "viewPendingDevices-btn",
  "viewRejectedDevices-btn",
];

const store = createStore({
  state: {
    stats,
  },
  getters: {
    "stats/stats": (state) => state.stats,
  },
  actions: {
    "stats/get": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

const cardsContent = [
  {
    id: 0,
    title: "Registered Users",
    fieldObject: "registered_users",
    content: "Registered users",
    icon: "mdi-account-group",
    stats: stats.registeredUsers,
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
    stats: stats.registeredDevices,
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
    stats: stats.onlineDevices,
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
    stats: stats.activeSessions,
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
    stats: stats.pendingDevices,
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
    stats: stats.rejectedDevices,
    buttonName: "View all Devices",
    pathName: "devices",
    nameUseTest: "viewRejectedDevices-btn",
  },
];
describe("Dashboard", () => {
  let wrapper: DashboardWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(Dashboard, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", async () => {
    const statsItems = await wrapper.vm.items;
    const statsContent = await wrapper.vm.itemsStats;
    expect(statsItems).toEqual(cardsContent);
    expect(statsContent).toEqual(stats);
    expect(wrapper.vm.hasStatus).toEqual(false);
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
