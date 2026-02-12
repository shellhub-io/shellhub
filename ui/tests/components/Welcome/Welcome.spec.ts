import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import Welcome from "@/components/Welcome/Welcome.vue";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useDevicesStore from "@/store/modules/devices";
import useStatsStore from "@/store/modules/stats";
import { mockDevice } from "@tests/mocks/device";

describe("Welcome", () => {
  let wrapper: VueWrapper<InstanceType<typeof Welcome>>;
  let dialog: DOMWrapper<Element>;
  let devicesStore: ReturnType<typeof useDevicesStore>;
  let statsStore: ReturnType<typeof useStatsStore>;

  const mockPendingDevice = {
    ...mockDevice,
    status: "pending" as const,
    uid: "pending-device-123",
  };

  const defaultStats = {
    registered_devices: 0,
    pending_devices: 0,
    rejected_devices: 0,
    online_devices: 0,
    active_sessions: 0,
  };

  const mountWrapper = async ({ tenantId = "test-tenant", stats = defaultStats } = {}) => {
    wrapper = mountComponent(Welcome, {
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          auth: { tenantId },
          stats: { stats },
        },
      },
    });

    devicesStore = useDevicesStore();
    statsStore = useStatsStore();

    await flushPromises();

    dialog = new DOMWrapper(document.body);
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    localStorage.clear();
    document.body.innerHTML = "";
  });

  describe("Dialog visibility logic", () => {
    it("does not show dialog when tenant ID is not available", async () => {
      await mountWrapper({ tenantId: "" });

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(false);
    });

    it("does not show dialog when namespace has already been shown", async () => {
      localStorage.setItem("namespacesWelcome", JSON.stringify({ "test-tenant": true }));

      await mountWrapper();

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(false);
    });

    it("does not show dialog when namespace has registered devices", async () => {
      await mountWrapper({ stats: { ...defaultStats, registered_devices: 1 } });

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(false);
    });

    it("does not show dialog when namespace has pending devices", async () => {
      await mountWrapper({ stats: { ...defaultStats, pending_devices: 1 } });

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(false);
    });

    it("does not show dialog when namespace has rejected devices", async () => {
      await mountWrapper({ stats: { ...defaultStats, rejected_devices: 1 } });

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(false);
    });

    it("shows dialog for new namespace with no devices", async () => {
      await mountWrapper({ tenantId: "new-tenant" });

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(true);
    });
  });

  describe("Dialog rendering", () => {
    beforeEach(async () => {
      await mountWrapper();
    });

    it("renders WindowDialog with correct props", () => {
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.exists()).toBe(true);
      expect(windowDialog.props("title")).toBe("Welcome to ShellHub!");
      expect(windowDialog.props("icon")).toBe("mdi-door-open");
      expect(windowDialog.props("iconColor")).toBe("primary");
    });

    it("shows correct step description", () => {
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("description")).toBe("Step 1 of 4");
    });

    it("renders all window items", () => {
      const windowItems = dialog.findAll(".v-window-item");
      expect(windowItems).toHaveLength(4);
    });
  });

  describe("Step navigation", () => {
    beforeEach(async () => {
      vi.useFakeTimers();
      await mountWrapper();
    });

    it("starts at step 1", () => {
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("description")).toBe("Step 1 of 4");
    });

    it("shows Next button on step 1", () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      expect(confirmBtn.exists()).toBe(true);
      expect(confirmBtn.text()).toBe("Next");
    });

    it("shows Close button on step 1", () => {
      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      expect(cancelBtn.exists()).toBe(true);
      expect(cancelBtn.text()).toBe("Close");
    });

    it("moves to step 2 and starts polling when Next is clicked on step 1", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("description")).toBe("Step 2 of 4");
    });

    it("shows documentation link on step 2", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const docLink = dialog.find('[data-test="second-screen-helper-link"]');
      expect(docLink.exists()).toBe(true);
      expect(docLink.text()).toContain("Check our  documentation");
    });

    it("disables Next button on step 2 when no device is detected", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const nextBtn = dialog.find('[data-test="confirm-btn"]');
      expect((nextBtn.element as HTMLButtonElement).disabled).toBe(true);
    });

    it("enables Next button on step 2 when device is detected", async () => {
      vi.mocked(devicesStore.getFirstPendingDevice).mockResolvedValue(mockPendingDevice);

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      statsStore.stats.pending_devices = 1;
      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      const nextBtn = dialog.find('[data-test="confirm-btn"]');
      expect((nextBtn.element as HTMLButtonElement).disabled).toBe(false);
    });

    it("shows Accept button on step 3", async () => {
      vi.mocked(devicesStore.getFirstPendingDevice).mockResolvedValue(mockPendingDevice);

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      statsStore.stats.pending_devices = 1;
      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      const acceptBtn = dialog.find('[data-test="confirm-btn"]');
      expect(acceptBtn.text()).toBe("Accept");
    });

    it("shows Finish button on step 4", async () => {
      vi.mocked(devicesStore.getFirstPendingDevice).mockResolvedValue(mockPendingDevice);

      const nextBtn = dialog.find('[data-test="confirm-btn"]');
      await nextBtn.trigger("click");
      await flushPromises();

      statsStore.stats.pending_devices = 1;
      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      const acceptBtn = dialog.find('[data-test="confirm-btn"]');
      await acceptBtn.trigger("click");
      await flushPromises();

      const finishBtn = dialog.find('[data-test="confirm-btn"]');
      expect(finishBtn.text()).toBe("Finish");
    });

    it("does not show Close button on step 4", async () => {
      vi.mocked(devicesStore.getFirstPendingDevice).mockResolvedValue(mockPendingDevice);

      const nextBtn = dialog.find('[data-test="confirm-btn"]');
      await nextBtn.trigger("click");
      await flushPromises();

      statsStore.stats.pending_devices = 1;
      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      const acceptBtn = dialog.find('[data-test="confirm-btn"]');
      await acceptBtn.trigger("click");
      await flushPromises();

      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      expect(cancelBtn.exists()).toBe(false);
    });
  });

  describe("Device polling", () => {
    beforeEach(async () => {
      vi.useFakeTimers();
      await mountWrapper();
    });

    it("starts polling when moving to step 2", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("description")).toBe("Step 2 of 4");

      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      expect(statsStore.fetchStats).toHaveBeenCalled();
    });

    it("detects pending device and moves to step 3", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      await vi.advanceTimersByTimeAsync(3000);

      statsStore.stats.pending_devices = 1;
      await flushPromises();

      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("description")).toBe("Step 3 of 4");
    });

    it("stops polling when device is detected", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      statsStore.stats.pending_devices = 1;
      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      const callCountBefore = vi.mocked(statsStore.fetchStats).mock.calls.length;

      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      const callCountAfter = vi.mocked(statsStore.fetchStats).mock.calls.length;
      expect(callCountAfter).toBe(callCountBefore);
    });

    it("stops polling when dialog is closed", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const callCountBefore = vi.mocked(statsStore.fetchStats).mock.calls.length;

      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      const callCountAfter = vi.mocked(statsStore.fetchStats).mock.calls.length;
      expect(callCountAfter).toBe(callCountBefore);
    });

    it("shows error when polling fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(statsStore.fetchStats).mockRejectedValueOnce(error);

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch devices.");
    });
  });

  describe("Device acceptance", () => {
    beforeEach(async () => {
      vi.useFakeTimers();

      await mountWrapper();
      vi.mocked(devicesStore.getFirstPendingDevice).mockResolvedValue(mockPendingDevice);

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      statsStore.stats.pending_devices = 1;
      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();
    });

    it("accepts device when Accept button is clicked", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(devicesStore.acceptDevice).toHaveBeenCalledWith("pending-device-123");
    });

    it("fetches stats after accepting device", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(statsStore.fetchStats).toHaveBeenCalled();
    });

    it("moves to step 4 after accepting device", async () => {
      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("description")).toBe("Step 4 of 4");
    });

    it("does not accept device when firstPendingDevice is not available", async () => {
      // @ts-expect-error Simulate device being removed before acceptance
      vi.mocked(devicesStore.getFirstPendingDevice).mockResolvedValueOnce(undefined);
      wrapper.unmount();

      vi.useFakeTimers();
      vi.mocked(devicesStore.acceptDevice).mockResolvedValue();
      await mountWrapper();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      statsStore.stats.pending_devices = 1;
      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      const acceptBtn = dialog.find('[data-test="confirm-btn"]');
      await acceptBtn.trigger("click");
      await flushPromises();

      expect(devicesStore.acceptDevice).not.toHaveBeenCalled();
    });

    it("shows error when accepting device fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(devicesStore.acceptDevice).mockRejectedValueOnce(error);

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to accept device.");
    });
  });

  describe("Dialog close behavior", () => {
    beforeEach(() => mountWrapper());

    it("closes dialog when Close button is clicked", async () => {
      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(false);
    });

    it("closes dialog when Finish button is clicked on step 4", async () => {
      vi.useFakeTimers();
      wrapper.unmount();
      await mountWrapper();
      vi.mocked(devicesStore.getFirstPendingDevice).mockResolvedValue(mockPendingDevice);

      const nextBtn = dialog.find('[data-test="confirm-btn"]');
      await nextBtn.trigger("click");
      await flushPromises();

      statsStore.stats.pending_devices = 1;
      await vi.advanceTimersByTimeAsync(3000);
      await flushPromises();

      const acceptBtn = dialog.find('[data-test="confirm-btn"]');
      await acceptBtn.trigger("click");
      await flushPromises();

      const finishBtn = dialog.find('[data-test="confirm-btn"]');
      await finishBtn.trigger("click");
      await flushPromises();

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(false);
    });

    it("closes dialog when close event is emitted", async () => {
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      await windowDialog.vm.$emit("update:modelValue", false);
      await flushPromises();

      expect(wrapper.findComponent({ name: "WindowDialog" }).props("modelValue")).toBe(false);
    });
  });
});
