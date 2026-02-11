import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper } from "@vue/test-utils";
import WelcomeThirdScreen from "@/components/Welcome/WelcomeThirdScreen.vue";
import { mountComponent } from "@tests/utils/mount";
import useDevicesStore from "@/store/modules/devices";
import { mockDevicePending } from "@tests/mocks/device";

describe("WelcomeThirdScreen", () => {
  let wrapper: VueWrapper<InstanceType<typeof WelcomeThirdScreen>>;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  const firstPendingDevice = mockDevicePending;

  const mountWrapper = async () => {
    wrapper = mountComponent(WelcomeThirdScreen, {
      props: { modelValue: firstPendingDevice },
    });

    devicesStore = useDevicesStore();

    await flushPromises();
    await wrapper.setProps({ firstPendingDevice });
  };

  beforeEach(async () => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Component initialization", () => {
    it("fetches first pending device on mount", () => {
      expect(devicesStore.getFirstPendingDevice).toHaveBeenCalled();
    });

    it("shows device detected screen when device is available", async () => {
      await flushPromises();

      expect(wrapper.text()).toContain("Device Detected!");
      expect(wrapper.text()).toContain("Confirm this device to add it to your account");
    });
  });

  describe("Device display", () => {
    it("renders success icon", () => {
      const avatar = wrapper.find(".v-avatar");
      expect(avatar.exists()).toBe(true);
      expect(wrapper.html()).toContain("mdi-check-circle");
    });

    it("renders connection successful alert", () => {
      const alert = wrapper.find('[data-test="welcome-third-screen-name"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("Connection Successful");
    });

    it("renders device name", () => {
      const deviceField = wrapper.find('[data-test="device-field"]');
      expect(deviceField.exists()).toBe(true);
      expect(deviceField.text()).toBe("device-three");
    });

    it("renders device pretty name", () => {
      const prettyNameField = wrapper.find('[data-test="device-pretty-name-field"]');
      expect(prettyNameField.exists()).toBe(true);
      expect(prettyNameField.text()).toBe("Debian 12");
    });

    it("renders device icon", () => {
      const deviceIcon = wrapper.findComponent({ name: "DeviceIcon" });
      expect(deviceIcon.exists()).toBe(true);
      expect(deviceIcon.props("icon")).toBe("debian");
    });

    it("renders device expansion panel with details", async () => {
      const expansionPanel = wrapper.find(".v-expansion-panel");
      expect(expansionPanel.exists()).toBe(true);

      await expansionPanel.find(".v-expansion-panel-title").trigger("click");
      await flushPromises();

      expect(wrapper.text()).toContain("Device UID:");
      expect(wrapper.text()).toContain("device-three");
      expect(wrapper.text()).toContain("MAC Address:");
      expect(wrapper.text()).toContain("00:00:00:00:00:03");
    });
  });

  describe("No device fallback", () => {
    beforeEach(() => wrapper.setProps({ firstPendingDevice: undefined }));

    it("shows no device message when device is not available", () => {
      expect(wrapper.find('[data-test="no-device-heading"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="no-device-heading"]').text()).toBe("No Device Detected Yet");
      expect(wrapper.find('[data-test="no-device-text"]').text()).toContain("Please run the installation command");
    });

    it("renders warning icon in no device state", () => {
      expect(wrapper.html()).toContain("mdi-clock-alert");
    });
  });
});
