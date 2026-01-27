import { describe, expect, it, afterEach, beforeEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import DeviceListChooser from "@/components/Devices/DeviceListChooser.vue";
import useDevicesStore from "@/store/modules/devices";
import { mockDevice, mockDeviceForSession } from "@tests/mocks/device";
import { createCleanRouter } from "@tests/utils/router";

const mockDevices = [
  mockDevice,
  { ...mockDevice, uid: "a582b47a42e" },
  { ...mockDevice, uid: "a582b47a42f" },
  mockDeviceForSession,
];

describe("DeviceListChooser", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceListChooser>>;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  const mountWrapper = (isSelectable = false) => {
    wrapper = mountComponent(DeviceListChooser, {
      global: { plugins: [createCleanRouter()] },
      props: { isSelectable },
      piniaOptions: { initialState: { devices: { devices: mockDevices } } },
    });

    devicesStore = useDevicesStore();
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("data table rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders devices data table", () => {
      expect(wrapper.find('[data-test="devices-dataTable"]').exists()).toBe(true);
    });

    it("displays device names in table", () => {
      expect(wrapper.text()).toContain("39-5e-2a");
      expect(wrapper.text()).toContain("00-00-00-00-00-01");
    });

    it("displays operating system information", () => {
      expect(wrapper.text()).toContain("Linux Mint 19.3");
      expect(wrapper.text()).toContain("Manjaro Linux");
    });

    it("shows device icons for each device", () => {
      const deviceIcons = wrapper.findAllComponents({ name: "DeviceIcon" });
      expect(deviceIcons.length).toBe(4);
    });

    it("displays SSHID for devices", () => {
      expect(wrapper.text()).toContain("user.39-5e-2a@localhost");
      expect(wrapper.text()).toContain("dev.00-00-00-00-00-01@localhost");
    });
  });

  describe("device selection", () => {
    it("does not show checkboxes when not selectable", () => {
      mountWrapper();

      const checkboxes = wrapper.findAllComponents({ name: "VCheckbox" });
      expect(checkboxes.length).toBe(0);
    });

    it("shows checkboxes when selectable", () => {
      mountWrapper(true);

      const checkboxes = wrapper.findAllComponents({ name: "VCheckbox" });
      expect(checkboxes.length).toBe(4);
    });

    it("updates selected devices when checkbox is clicked", async () => {
      mountWrapper(true);

      const firstCheckbox = wrapper.find('[data-test="device-selection-checkbox"] input');
      await firstCheckbox.setValue(true);
      await flushPromises();

      expect(devicesStore.selectedDevices.length).toBeGreaterThan(0);
    });

    it("prevents selecting more than 3 devices", async () => {
      mountWrapper(true);
      devicesStore.selectedDevices = [
        mockDevices[0],
        mockDevices[1],
        mockDevices[2],
      ];

      const checkbox = wrapper.findAll('[data-test="device-selection-checkbox"] input')[3];
      await checkbox.setValue(true); // Attempt to select fourth device
      await flushPromises();

      expect(devicesStore.selectedDevices.length).toBeLessThanOrEqual(3);
    });
  });

  describe("device links", () => {
    beforeEach(() => mountWrapper());

    it("renders clickable device name links", () => {
      const links = wrapper.findAll("a");
      expect(links.length).toBeGreaterThan(0);
    });

    it("links point to device details page", () => {
      const firstLink = wrapper.find("a");
      expect(firstLink.attributes("href")).toContain("/a582b47a42d");
    });
  });
});
