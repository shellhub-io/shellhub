import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import UserWarning from "../../../src/components/User/UserWarning.vue";
import { envVariables } from "./../../../src/envVariables";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const numberNamespaces = 0;
const statusSpinner = false;
const activeBilling = true;
const DeviceChooserStatus = false;

const namespace = {
  name: "namespace",
  owner: "user",
  members: [{ name: "user" }, { name: "user2" }],
  tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484712",
};

const statsWithoutDevices = {
  registered_devices: 0,
  online_devices: 0,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

const statsWithDevices = {
  registered_devices: 4,
  online_devices: 0,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

const announcements = [
  {
    uuid: "52088548-2b99-4f38-ac09-3a8f8988476f",
    title: "This is a announcement",
    content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
    date: "2022-12-15T19:45:45.618Z",
  },
  {
    uuid: "52188548-2b99-4f38-ac09-3a8f8988476f",
    title: "This is a new announcement",
    content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
    date: "2022-12-15T19:46:45.618Z",
  },
];

const announcement = {
  uuid: "52088548-2b99-4f38-ac09-3a8f8988476f",
  title: "This is a announcement",
  content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
  date: "2022-12-15T19:45:45.618Z",
};

interface IStore {
  numberNamespaces: number;
  statusSpinner: boolean;
  stats: typeof statsWithoutDevices;
  activeBilling: boolean;
  namespace: typeof namespace;
  DeviceChooserStatus: boolean;
  announcements: typeof announcements;
  announcement: typeof announcement;
}

const getters = {
  "auth/isLoggedIn": () => true,
  "namespaces/getNumberNamespaces": (state: IStore) => state.numberNamespaces,
  "spinner/getStatus": (state: IStore) => state.statusSpinner,
  "stats/stats": (state: IStore) => state.stats,
  "billing/active": (state: IStore) => !state.activeBilling,
  "namespaces/get": (state: IStore) => state.namespace,
  "devices/getDeviceChooserStatus": (state: IStore) => state.DeviceChooserStatus,
  "announcement/list": (state: IStore) => state.announcements,
  "announcement/get": (state: IStore) => state.announcement,
};

const actions = {
  "stats/get": vi.fn(),
  "devices/setDeviceChooserStatus": vi.fn(),
  "auth/setShowWelcomeScreen": vi.fn(),
  "namespaces/fetch": vi.fn(),
  "namespaces/get": vi.fn(),
  "snackbar/showSnackbarErrorAssociation": vi.fn(),
  "snackbar/showSnackbarErrorLoading": vi.fn(),
  "announcement/getListAnnouncements": vi.fn(),
  "announcement/getAnnouncement": vi.fn(),
};

const storeWithoutDevices = createStore({
  state: {
    DeviceChooserStatus,
    numberNamespaces,
    statusSpinner,
    stats: statsWithoutDevices,
    activeBilling,
    namespace,
    announcements,
    announcement,
  },
  getters,
  actions,
});

const storeWithDevicesInactive = createStore({
  state: {
    numberNamespaces: 3,
    statusSpinner,
    stats: statsWithDevices,
    activeBilling: false,
    namespace,
    DeviceChooserStatus,
    announcements,
    announcement,
  },
  getters: {
    ...getters,
    "billing/active": (state) => state.activeBilling,
    "devices/getDeviceChooserStatus": (state) => !state.DeviceChooserStatus,
  },
  actions,
});

const storeWithDevicesActive = createStore({
  state: {
    numberNamespaces: 3,
    statusSpinner,
    stats: statsWithDevices,
    activeBilling: true,
    namespace,
    DeviceChooserStatus,
    announcements,
    announcement,
  },
  getters: {
    ...getters,
    "billing/active": (state) => state.activeBilling,
  },
  actions,
});

describe("Without devices and billing disabled", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserWarning>>;

  beforeEach(() => {
    const vuetify = createVuetify();

    envVariables.billingEnable = false;

    wrapper = mount(UserWarning, {
      global: {
        plugins: [[storeWithoutDevices, key], vuetify, routes],
      },
      shallow: true,
    });
    localStorage.setItem("namespacesWelcome", JSON.stringify({}));
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  ///////
  // Data and Props checking
  //////

  it("Compare data with the default value", () => {
    expect(wrapper.vm.show).toEqual(false);
    expect(wrapper.vm.showInstructions).toEqual(true);
  });
  it("Process data in the computed", () => {
    expect(wrapper.vm.hasNamespaces).toEqual(numberNamespaces !== 0);
    expect(wrapper.vm.hasSpinner).toEqual(statusSpinner);
    expect(wrapper.vm.stats).toEqual(statsWithoutDevices);
  });
  it("Process data in methods", () => {
    expect(wrapper.vm.hasDevices).toEqual(false);
  });

  //////
  // HTML validation
  //////

  it("Renders the template with components", () => {
    expect(wrapper.find('[data-test="deviceChooser-component"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="welcome-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="namespaceInstructions-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billingWarning-component"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="announcementsModal-component"]').exists()).toBe(true);
  });
  it("Renders the template with data", async () => {
    await wrapper.vm.showScreenWelcome();
    expect(wrapper.vm.show).toBe(true);

    localStorage.setItem(
      "namespacesWelcome",
      JSON.stringify({ ...{ [namespace.tenant_id]: true } }),
    );

    await wrapper.vm.showScreenWelcome();
    expect(wrapper.vm.show).toBe(false);
  });
});

describe("Without devices", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserWarning>>;

  beforeEach(() => {
    const vuetify = createVuetify();

    envVariables.billingEnable = true;

    wrapper = mount(UserWarning, {
      global: {
        plugins: [[storeWithoutDevices, key], vuetify, routes],
      },
      shallow: true,
    });
    localStorage.setItem("namespacesWelcome", JSON.stringify({}));
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  ///////
  // Data and Props checking
  //////

  it("Compare data with the default value", () => {
    expect(wrapper.vm.show).toEqual(false);
    expect(wrapper.vm.showInstructions).toEqual(true);
  });
  it("Process data in the computed", () => {
    expect(wrapper.vm.hasNamespaces).toEqual(numberNamespaces !== 0);
    expect(wrapper.vm.hasSpinner).toEqual(statusSpinner);
    expect(wrapper.vm.stats).toEqual(statsWithoutDevices);
  });
  it("Process data in methods", () => {
    expect(wrapper.vm.hasDevices).toEqual(false);
  });

  //////
  // HTML validation
  //////

  it("Renders the template with components", () => {
    expect(wrapper.find('[data-test="deviceChooser-component"]').exists()).toBe(
      false,
    );
    expect(wrapper.find('[data-test="welcome-component"]').exists()).toBe(true);
    expect(
      wrapper.find('[data-test="namespaceInstructions-component"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-test="billingWarning-component"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-test="announcementsModal-component"]').exists(),
    ).toBe(true);
  });
  it("Renders the template with data", async () => {
    await wrapper.vm.showScreenWelcome();
    expect(wrapper.vm.show).toBe(true);

    localStorage.setItem(
      "namespacesWelcome",
      JSON.stringify({ ...{ [namespace.tenant_id]: true } }),
    );

    await wrapper.vm.showScreenWelcome();
    expect(wrapper.vm.show).toBe(false);
  });
});

describe("With devices and inactive billing", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserWarning>>;

  beforeEach(() => {
    const vuetify = createVuetify();

    envVariables.billingEnable = true;

    wrapper = mount(UserWarning, {
      global: {
        plugins: [[storeWithDevicesInactive, key], vuetify, routes],
      },
      shallow: true,
    });
    localStorage.setItem("namespacesWelcome", JSON.stringify({}));
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  //////
  // Call actions
  //////
  it("Dispatches store on mount", async () => {
    expect(actions["namespaces/get"]).toHaveBeenCalled();
    await wrapper.vm.$nextTick();
    expect(actions["devices/setDeviceChooserStatus"]).toHaveBeenCalled();
  });

  ///////
  // Data and Props checking
  //////

  it("Compare data with the default value", () => {
    expect(wrapper.vm.show).toEqual(false);
    expect(wrapper.vm.showInstructions).toEqual(false);
  });
  it("Process data in the computed", () => {
    expect(wrapper.vm.hasNamespaces).toEqual(true);
    expect(wrapper.vm.hasSpinner).toEqual(statusSpinner);
    expect(wrapper.vm.stats).toEqual(statsWithDevices);
  });
  it("Process data in methods", () => {
    expect(wrapper.vm.hasDevices).toEqual(true);
  });
  //////
  // HTML validation
  //////

  it("Renders the template with components", () => {
    expect(wrapper.find('[data-test="deviceChooser-component"]').exists()).toBe(
      true,
    );
    expect(wrapper.find('[data-test="welcome-component"]').exists()).toBe(true);
    expect(
      wrapper.find('[data-test="namespaceInstructions-component"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-test="billingWarning-component"]').exists(),
    ).toBe(true);
  });

  it("Renders the template with data", async () => {
    expect(wrapper.vm.namespaceHasBeenShown(namespace.tenant_id)).toBe(false);

    localStorage.setItem(
      "namespacesWelcome",
      JSON.stringify({
        ...JSON.parse(localStorage.getItem("namespacesWelcome") || "{}"),
        ...{ [namespace.tenant_id]: true },
      }),
    );

    expect(wrapper.vm.namespaceHasBeenShown(namespace.tenant_id)).toBe(true);

    await wrapper.vm.showScreenWelcome();
    expect(
      Object.keys(JSON.parse(localStorage.getItem("namespacesWelcome") || "{}")),
    ).toHaveLength(1);
  });
});

///////
// In this case, The welcome screen loads with the expected
// behavior with devices and with billing environment enabled
// and active subscription
///////

describe("With devices and active billing", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserWarning>>;

  beforeEach(() => {
    const vuetify = createVuetify();

    envVariables.billingEnable = true;

    wrapper = mount(UserWarning, {
      global: {
        plugins: [[storeWithDevicesActive, key], vuetify, routes],
      },
      shallow: true,
    });
    localStorage.setItem("namespacesWelcome", JSON.stringify({}));
  });

  ///////
  // Component Rendering
  //////

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  //////
  // Call actions
  //////
  it("Dispatches on mount", async () => {
    await wrapper.vm.$nextTick();
    expect(actions["devices/setDeviceChooserStatus"]).toHaveBeenCalled();
  });

  ///////
  // Data and Props checking
  //////

  it("Compare data with the default value", () => {
    expect(wrapper.vm.show).toEqual(false);
    expect(wrapper.vm.showInstructions).toEqual(false);
  });
  it("Process data in the computed", () => {
    expect(wrapper.vm.hasNamespaces).toEqual(true);
    expect(wrapper.vm.hasSpinner).toEqual(statusSpinner);
    expect(wrapper.vm.stats).toEqual(statsWithDevices);
  });
  it("Process data in methods", () => {
    expect(wrapper.vm.hasDevices).toEqual(true);
  });

  //////
  // HTML validation
  //////

  it("Renders the template with components", () => {
    expect(wrapper.find('[data-test="deviceChooser-component"]').exists()).toBe(
      false,
    );
    expect(wrapper.find('[data-test="welcome-component"]').exists()).toBe(true);
    expect(
      wrapper.find('[data-test="namespaceInstructions-component"]').exists(),
    ).toBe(true);
    expect(
      wrapper.find('[data-test="billingWarning-component"]').exists(),
    ).toBe(true);
  });

  it("Renders the template with data", async () => {
    expect(wrapper.vm.namespaceHasBeenShown(namespace.tenant_id)).toBe(false);

    localStorage.setItem(
      "namespacesWelcome",
      JSON.stringify({
        ...JSON.parse(localStorage.getItem("namespacesWelcome") || "{}"),
        ...{ [namespace.tenant_id]: true },
      }),
    );

    expect(wrapper.vm.namespaceHasBeenShown(namespace.tenant_id)).toBe(true);

    await wrapper.vm.showScreenWelcome();
    expect(
      Object.keys(JSON.parse(localStorage.getItem("namespacesWelcome") || "{}")),
    ).toHaveLength(1);
  });
});
