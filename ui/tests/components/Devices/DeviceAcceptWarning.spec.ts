import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import DeviceAcceptWarning from "@/components/Devices/DeviceAcceptWarning.vue";
import useDevicesStore from "@/store/modules/devices";

vi.mock("@/utils/permission", () => ({
  default: vi.fn(() => true),
}));

describe("DeviceAcceptWarning", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceAcceptWarning>>;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  const mountWrapper = (duplicatedDeviceName = "") => {
    wrapper = mountComponent(DeviceAcceptWarning, {
      piniaOptions: { initialState: { devices: { duplicatedDeviceName } } },
    });

    devicesStore = useDevicesStore();
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("dialog visibility", () => {
    it("does not render dialog when no duplicated device name", () => {
      mountWrapper();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("renders dialog when duplicated device name is set", () => {
      mountWrapper("test-device");

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(true);
    });
  });

  describe("dialog content", () => {
    beforeEach(() => mountWrapper("test-device"));

    it("shows warning title", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toBe("You already have a device using the same name");
    });

    it("shows description with duplicated device name", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("description")).toContain("test-device");
      expect(messageDialog.props("description")).toContain("already taken");
    });

    it("displays warning icon", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("icon")).toBe("mdi-alert");
      expect(messageDialog.props("iconColor")).toBe("warning");
    });

    it("shows Close button", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("cancelText")).toBe("Close");
    });
  });

  describe("dialog actions", () => {
    it("clears duplicated device name when cancel is emitted", async () => {
      mountWrapper("test-device");

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      await messageDialog.vm.$emit("cancel");

      expect(devicesStore.duplicatedDeviceName).toBe("");
    });

    it("clears duplicated device name when close is emitted", async () => {
      mountWrapper("test-device");

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      await messageDialog.vm.$emit("close");

      expect(devicesStore.duplicatedDeviceName).toBe("");
    });

    it("updates dialog visibility when duplicated device name changes", async () => {
      mountWrapper();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);

      devicesStore.duplicatedDeviceName = "new-device";
      await flushPromises();

      expect(messageDialog.props("modelValue")).toBe(true);
    });
  });
});
