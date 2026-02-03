import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { mockDevice } from "@tests/mocks/device";
import QuickConnectionList from "@/components/QuickConnection/QuickConnectionList.vue";
import useDevicesStore from "@/store/modules/devices";
import handleError from "@/utils/handleError";
import { createCleanRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";

const onlineDevice = {
  ...mockDevice,
  online: true,
  namespace: "user",
  name: "test-device",
};

const onlineDeviceWithoutTags = {
  ...mockDevice,
  online: true,
  namespace: "admin",
  name: "no-tags-device",
  tags: [],
};

describe("QuickConnectionList", () => {
  let wrapper: VueWrapper<InstanceType<typeof QuickConnectionList>>;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  const mountWrapper = (filter = "", onlineDevices = [onlineDevice, onlineDeviceWithoutTags]) => {
    wrapper = mountComponent(QuickConnectionList, {
      global: { plugins: [createCleanRouter()] },
      props: { filter },
      piniaOptions: {
        initialState: { devices: { onlineDevices } },
      },
    });

    devicesStore = useDevicesStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Device list rendering", () => {
    it("Renders devices list", () => {
      const list = wrapper.find('[data-test="devices-list"]');
      expect(list.exists()).toBe(true);
    });

    it("Displays all online devices", () => {
      const items = wrapper.findAll('[data-test="device-list-item"]');
      expect(items).toHaveLength(2);
    });

    it("Shows device name", () => {
      const name = wrapper.findAll('[data-test="device-name"]')[0];
      expect(name.text()).toBe(onlineDevice.name);
    });

    it("Shows device info with icon and OS name", () => {
      const info = wrapper.findAll('[data-test="device-info"]')[0];
      expect(info.text()).toContain(onlineDevice.info.pretty_name);
    });

    it("Renders DeviceIcon component", () => {
      const deviceIcon = wrapper.findComponent({ name: "DeviceIcon" });
      expect(deviceIcon.exists()).toBe(true);
      expect(deviceIcon.props("icon")).toBe(onlineDevice.info.id);
    });

    it("Shows device SSHID", () => {
      const sshid = wrapper.findAll('[data-test="device-ssh-id"]')[0];
      const expectedSshid = `${onlineDevice.namespace}.${onlineDevice.name}@${window.location.hostname}`;
      expect(sshid.text()).toContain(expectedSshid);
    });

    it("Shows device tags", () => {
      const tags = wrapper.findAll('[data-test="device-tags"]')[0];
      const tagChip = tags.find('[data-test="tag-chip"]');
      expect(tagChip.exists()).toBe(true);
      expect(tagChip.text()).toBe(onlineDevice.tags[0].name);
    });

    it("Shows 'No tags' when device has no tags", () => {
      const tags = wrapper.findAll('[data-test="device-tags"]')[1];
      const noTagsChip = tags.find('[data-test="no-tags-chip"]');
      expect(noTagsChip.exists()).toBe(true);
      expect(noTagsChip.text()).toBe("No tags");
    });

    it("Shows tag tooltip for tags with full name", () => {
      const tagName = new DOMWrapper(document.body).find('[data-test="tag-name"]');
      expect(tagName.text()).toBe(onlineDevice.tags[0].name);
    });
  });

  describe("Empty state", () => {
    beforeEach(() => {
      wrapper.unmount();
      mountWrapper("", []);
    });

    it("Shows empty state when no devices are online", () => {
      const emptyState = wrapper.find('[data-test="no-online-devices"]');
      expect(emptyState.exists()).toBe(true);
    });

    it("Shows offline icon in empty state", () => {
      const icon = wrapper.find('[data-test="no-online-devices-icon"]');
      expect(icon.exists()).toBe(true);
      expect(icon.classes()).toContain("mdi-laptop-off");
    });

    it("Shows empty state message", () => {
      const message = wrapper.find('[data-test="no-online-devices-message"]');
      expect(message.text()).toBe("There are currently no devices online.");
    });
  });

  describe("Device interactions", () => {
    it("Shows copy SSHID button", () => {
      const copyBtn = wrapper.findAll('[data-test="copy-id-button"]')[0];
      expect(copyBtn.exists()).toBe(true);
    });

    it("Shows SSHID help button", () => {
      const helpBtn = wrapper.findAll('[data-test="sshid-help-btn"]')[0];
      expect(helpBtn.exists()).toBe(true);
    });

    it("Opens terminal dialog when device is clicked", async () => {
      const deviceItem = wrapper.findAll('[data-test="device-list-item"]')[0];
      await deviceItem.trigger("click");
      await flushPromises();

      const terminalDialog = wrapper.findComponent({ name: "TerminalDialog" });
      expect(terminalDialog.props("modelValue")).toBe(true);
      expect(terminalDialog.props("deviceUid")).toBe(onlineDevice.uid);
      expect(terminalDialog.props("deviceName")).toBe(onlineDevice.name);
    });

    it("Opens SSHID helper when help button is clicked", async () => {
      const helpBtn = wrapper.findAll('[data-test="sshid-help-btn"]')[0];
      await helpBtn.trigger("click");
      await flushPromises();

      const sshidHelper = wrapper.findComponent({ name: "SSHIDHelper" });
      expect(sshidHelper.props("modelValue")).toBe(true);
      expect(sshidHelper.props("sshid")).toContain(onlineDevice.name);
    });
  });

  describe("Data fetching", () => {
    it("Calls fetchOnlineDevices on mount", () => {
      expect(devicesStore.fetchOnlineDevices).toHaveBeenCalled();
    });

    it("Fetches with correct filter", () => {
      const calls = vi.mocked(devicesStore.fetchOnlineDevices).mock.calls;
      const lastCall = calls[calls.length - 1];

      // The filter is base64 encoded JSON, so we decode it
      const filterParam = lastCall[0];
      const decodedFilter = JSON.parse(Buffer.from(filterParam as string, "base64").toString("utf-8"));

      expect(decodedFilter).toEqual([
        {
          type: "property",
          params: { name: "online", operator: "eq", value: true },
        },
        {
          type: "property",
          params: { name: "name", operator: "contains", value: "" },
        },
        { type: "operator", params: { name: "and" } },
      ]);
    });

    it("Refetches when filter prop changes", async () => {
      wrapper.unmount();
      mountWrapper("test-filter");
      await flushPromises();

      expect(devicesStore.fetchOnlineDevices).toHaveBeenCalled();

      const calls = vi.mocked(devicesStore.fetchOnlineDevices).mock.calls;
      const lastCall = calls[calls.length - 1];
      const filterParam = lastCall[0];
      const decodedFilter = JSON.parse(Buffer.from(filterParam as string, "base64").toString("utf-8"));

      // Check that the filter includes the search term
      expect(decodedFilter[1].params.value).toBe("test-filter");
    });
  });

  describe("Error handling", () => {
    it("Handles fetch error", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(devicesStore.fetchOnlineDevices).mockRejectedValue(error);

      await wrapper.setProps({ filter: "trigger-refetch" });
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while loading devices.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("CopyWarning component", () => {
    it("Renders CopyWarning component for each device", () => {
      const copyWarnings = wrapper.findAllComponents({ name: "CopyWarning" });
      expect(copyWarnings.length).toBeGreaterThan(0);
    });

    it("Passes correct SSHID to CopyWarning", () => {
      const copyWarning = wrapper.findAllComponents({ name: "CopyWarning" })[0];
      const expectedSshid = `${onlineDevice.namespace}.${onlineDevice.name}@${window.location.hostname}`;
      expect(copyWarning.props("macro")).toBe(expectedSshid);
    });

    it("Passes correct copied item label", () => {
      const copyWarning = wrapper.findAllComponents({ name: "CopyWarning" })[0];
      expect(copyWarning.props("copiedItem")).toBe("Device SSHID");
    });
  });
});
