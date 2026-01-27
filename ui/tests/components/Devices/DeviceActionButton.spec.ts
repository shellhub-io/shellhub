import { describe, expect, it, afterEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import DeviceActionButton from "@/components/Devices/DeviceActionButton.vue";
import useDevicesStore from "@/store/modules/devices";
import useBillingStore from "@/store/modules/billing";
import useStatsStore from "@/store/modules/stats";

vi.mock("@/utils/permission", () => ({
  default: vi.fn(() => true),
}));

describe("DeviceActionButton", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceActionButton>>;
  let dialog: DOMWrapper<Element>;
  let devicesStore: ReturnType<typeof useDevicesStore>;
  let billingStore: ReturnType<typeof useBillingStore>;
  let statsStore: ReturnType<typeof useStatsStore>;

  const openActionDialog = async () => {
    await wrapper.find('[data-test="open-action-dialog"]').trigger("click");
    await flushPromises();
  };

  const triggerAction = async () => {
    await dialog.find('[data-test="action-btn"]').trigger("click");
    await flushPromises();
  };

  const mountWrapper = (
    { action = "accept", isInDevicesDropdown = false, isBillingActive = false }:
    { action?: "accept" | "reject" | "remove"; isInDevicesDropdown?: boolean; isBillingActive?: boolean } = {},
  ) => {
    wrapper = mountComponent(DeviceActionButton, {
      props: {
        uid: "test-device-uid",
        variant: "device",
        action,
        isInDevicesDropdown,
        name: "test-device",
      },
      piniaOptions: {
        initialState: { billing: { billing: { active: isBillingActive } } },
      },
    });

    devicesStore = useDevicesStore();
    billingStore = useBillingStore();
    statsStore = useStatsStore();
    dialog = new DOMWrapper(document.body);
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("list item rendering", () => {
    it("renders list item when not in devices dropdown", () => {
      mountWrapper();

      expect(wrapper.find('[data-test="open-action-dialog"]').exists()).toBe(true);
    });

    it("displays accept action text", () => {
      mountWrapper();

      expect(wrapper.find('[data-test="action-item"]').text()).toContain("Accept");
    });

    it("displays reject action text", () => {
      mountWrapper({ action: "reject" });

      expect(wrapper.find('[data-test="action-item"]').text()).toContain("Reject");
    });

    it("displays remove action text", () => {
      mountWrapper({ action: "remove" });

      expect(wrapper.find('[data-test="action-item"]').text()).toContain("Remove");
    });

    it("renders button when in devices dropdown", () => {
      mountWrapper({ isInDevicesDropdown: true });

      expect(wrapper.find("button").exists()).toBe(true);
    });
  });

  describe("action dialog", () => {
    it("opens dialog when action item is clicked", async () => {
      mountWrapper();

      await openActionDialog();

      expect(dialog.find('[data-test="device-action-dialog"]').exists()).toBe(true);
    });

    it("shows dialog with correct title for accept", async () => {
      mountWrapper();

      await openActionDialog();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toContain("Accept");
    });

    it("shows dialog with correct icon color for accept", async () => {
      mountWrapper();

      await openActionDialog();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("iconColor")).toBe("primary");
    });

    it("shows dialog with correct icon color for reject", async () => {
      mountWrapper({ action: "reject" });

      await openActionDialog();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("iconColor")).toBe("warning");
    });

    it("shows dialog with correct icon color for remove", async () => {
      mountWrapper({ action: "remove" });

      await openActionDialog();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("iconColor")).toBe("error");
    });

    it("displays billing alert when billing is active for accept action", async () => {
      mountWrapper({ isBillingActive: true });

      await openActionDialog();

      expect(dialog.text()).toContain("Accepted devices");
      expect(dialog.text()).toContain("billed");
    });

    it("closes dialog when cancel is clicked", async () => {
      mountWrapper();

      await openActionDialog();

      await dialog.find('[data-test="close-btn"]').trigger("click");
      await flushPromises();

      expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
    });
  });

  describe("device actions", () => {
    it("calls acceptDevice when accept action is confirmed", async () => {
      mountWrapper();

      await openActionDialog();

      await triggerAction();

      expect(devicesStore.acceptDevice).toHaveBeenCalledWith("test-device-uid");
      expect(statsStore.fetchStats).toHaveBeenCalled();
    });

    it("calls rejectDevice when reject action is confirmed", async () => {
      mountWrapper({ action: "reject" });

      await openActionDialog();

      await triggerAction();

      expect(devicesStore.rejectDevice).toHaveBeenCalledWith("test-device-uid");
      expect(statsStore.fetchStats).toHaveBeenCalled();
    });

    it("calls removeDevice when remove action is confirmed", async () => {
      mountWrapper({ action: "remove" });

      await openActionDialog();

      await triggerAction();

      expect(devicesStore.removeDevice).toHaveBeenCalledWith("test-device-uid");
      expect(statsStore.fetchStats).toHaveBeenCalled();
    });

    it("handles billing error (402) on accept", async () => {
      mountWrapper();
      vi.spyOn(devicesStore, "acceptDevice").mockRejectedValue(createAxiosError(402, "Payment Required"));

      await openActionDialog();

      await triggerAction();

      expect(billingStore.showBillingWarning).toBe(true);
    });

    it("handles device limit error (403) on accept", async () => {
      mountWrapper();
      vi.spyOn(devicesStore, "acceptDevice").mockRejectedValue(createAxiosError(403, "Forbidden"));

      await openActionDialog();

      await triggerAction();

      expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
    });

    it("handles duplicate name error (409) on accept", async () => {
      mountWrapper();
      vi.spyOn(devicesStore, "acceptDevice").mockRejectedValue(createAxiosError(409, "Conflict"));

      await openActionDialog();

      await triggerAction();

      expect(devicesStore.duplicatedDeviceName).toBe("test-device");
    });

    it("emits update event after successful action", async () => {
      mountWrapper();

      await openActionDialog();

      await triggerAction();

      expect(wrapper.emitted("update")).toBeTruthy();
    });
  });
});
