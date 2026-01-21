import { VueWrapper } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import Devices from "@/views/Devices.vue";

describe("Devices View", () => {
  let wrapper: VueWrapper<InstanceType<typeof Devices>>;
  const router = createCleanRouter();

  const mountWrapper = (showDevices = true) =>
    wrapper = mountComponent(Devices, {
      global: { plugins: [router] },
      piniaOptions: { initialState: { devices: { showDevices } } },
    });

  afterEach(() => { wrapper?.unmount(); });

  describe("when devices exist", () => {
    beforeEach(() => { wrapper = mountWrapper(); });

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="devices-header"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Device Management");
      expect(pageHeader.find('[data-test="device-add-btn"]').exists()).toBe(true);
    });

    it("displays the device list component", () => {
      expect(wrapper.find('[data-test="device-table-component"]').exists()).toBe(true);
    });

    it("does not show the no items message", () => {
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
    });
  });

  describe("when no devices exist", () => {
    beforeEach(() => { wrapper = mountWrapper(false); });

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="devices-header"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Device Management");
      expect(pageHeader.find('[data-test="device-add-btn"]').exists()).toBe(true);
    });

    it("does not display the device list component", () => {
      expect(wrapper.find('[data-test="device-table-component"]').exists()).toBe(false);
    });

    it("shows the no items message", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      expect(noItemsMessage.exists()).toBe(true);
      expect(noItemsMessage.text()).toContain("In order to register a device on ShellHub");
      expect(noItemsMessage.find('[data-test="device-add-btn"]').exists()).toBe(true);
    });
  });
});
