import { VueWrapper, flushPromises } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import Home from "@/views/Home.vue";
import { mockNamespace } from "../mocks";

describe("Home View", () => {
  let wrapper: VueWrapper<InstanceType<typeof Home>>;

  const mountWrapper = (hasNamespace = true, deviceStats = { total: 5, online: 3, pending: 1 }) => {
    const initialState = hasNamespace
      ? {
        namespaces: {
          namespaceList: [mockNamespace],
          currentNamespace: mockNamespace,
        },
        devices: {
          totalDevicesCount: deviceStats.total,
          onlineDevicesCount: deviceStats.online,
          pendingDevicesCount: deviceStats.pending,
        },
      }
      : {
        namespaces: {
          namespaceList: [],
          currentNamespace: {},
        },
        devices: {
          totalDevicesCount: 0,
          onlineDevicesCount: 0,
          pendingDevicesCount: 0,
        },
      };

    wrapper = mountComponent(Home, { piniaOptions: { initialState } });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("when namespace exists", () => {
    beforeEach(() => { mountWrapper(); });

    it("displays namespace name and description", () => {
      expect(wrapper.text()).toContain(mockNamespace.name);
      expect(wrapper.text()).toContain("This is your active namespace");
    });

    it("displays settings button for the namespace", () => {
      const settingsBtn = wrapper.find('[data-test="namespace-settings-btn"]');
      expect(settingsBtn.text()).toContain("Settings");
    });

    it("displays device statistics cards", () => {
      const acceptedCard = wrapper.find('[data-test="accepted-devices-card"]');
      expect(acceptedCard.text()).toContain("Accepted Devices");
      expect(acceptedCard.text()).toContain("5");
      expect(acceptedCard.text()).toContain("View all devices");

      const onlineCard = wrapper.find('[data-test="online-devices-card"]');
      expect(onlineCard.text()).toContain("Online Devices");
      expect(onlineCard.text()).toContain("3");
      expect(onlineCard.text()).toContain("View Online Devices");

      const pendingCard = wrapper.find('[data-test="pending-devices-card"]');
      expect(pendingCard.text()).toContain("Pending Devices");
      expect(pendingCard.text()).toContain("1");
      expect(pendingCard.text()).toContain("Approve Devices");
    });

    it("updates displayed statistics when different values are provided", () => {
      wrapper.unmount();
      mountWrapper(true, { total: 10, online: 8, pending: 2 });

      const acceptedCard = wrapper.find('[data-test="accepted-devices-card"]');
      expect(acceptedCard.text()).toContain("10");

      const onlineCard = wrapper.find('[data-test="online-devices-card"]');
      expect(onlineCard.text()).toContain("8");

      const pendingCard = wrapper.find('[data-test="pending-devices-card"]');
      expect(pendingCard.text()).toContain("2");
    });
  });

  describe("when no namespace exists", () => {
    beforeEach(() => { mountWrapper(false); });

    it("displays no active namespace message", () => {
      expect(wrapper.text()).toContain("No Active Namespace");
      expect(wrapper.text()).toContain(
        "A namespace is a logical grouping that isolates your devices, sessions, and configurations from others.",
      );
    });

    it("displays create namespace button with instructions", () => {
      const text = wrapper.text();
      expect(text).toContain("Create your first namespace");
      expect(text).toContain("Create Namespace");
      expect(text).toContain("You need to create or join a namespace");
    });

    it("opens create namespace dialog when button is clicked", async () => {
      const createBtn = wrapper.find('[data-test="create-namespace-home-btn"]');
      await createBtn.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "NamespaceAdd" });
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("does not display device statistics cards", () => {
      expect(wrapper.text()).not.toContain("Accepted Devices");
      expect(wrapper.text()).not.toContain("Online Devices");
      expect(wrapper.text()).not.toContain("Pending Devices");
    });

    it("does not display settings button", () => {
      expect(wrapper.find('[data-test="namespace-settings-btn"]').exists()).toBe(false);
    });
  });
});
