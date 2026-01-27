import { describe, expect, it, afterEach, beforeEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import Device from "@/components/Devices/Device.vue";
import useDevicesStore from "@/store/modules/devices";

describe("Device", () => {
  let wrapper: VueWrapper<InstanceType<typeof Device>>;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  beforeEach(() => {
    wrapper = mountComponent(Device, { global: { plugins: [createCleanRouter()] } });
    devicesStore = useDevicesStore();
  });

  afterEach(() => { wrapper?.unmount(); });

  describe("navigation tabs", () => {
    it("displays Accepted tab", () => {
      expect(wrapper.text()).toContain("Accepted");
    });

    it("displays Pending tab", () => {
      expect(wrapper.text()).toContain("Pending");
    });

    it("displays Rejected tab", () => {
      expect(wrapper.text()).toContain("Rejected");
    });
  });

  describe("search functionality", () => {
    it("renders search input field", () => {
      const searchField = wrapper.find('[data-test="search-text"]');
      expect(searchField.exists()).toBe(true);
    });

    it("updates device filter when search input changes", async () => {
      const searchField = wrapper.find('[data-test="search-text"] input');
      await searchField.setValue("test-device");
      await flushPromises();

      expect(devicesStore.deviceListFilter).toBeDefined();
    });

    it("clears filter when search input is empty", async () => {
      const searchField = wrapper.find('[data-test="search-text"] input');
      await searchField.setValue("test");
      await flushPromises();

      await searchField.setValue("");
      await flushPromises();

      expect(devicesStore.deviceListFilter).toBeUndefined();
    });
  });

  describe("router view", () => {
    it("renders router-view for nested routes", () => {
      expect(wrapper.findComponent({ name: "RouterView" }).exists()).toBe(true);
    });
  });
});
