import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import SettingSessionRecording from "@/components/Setting/SettingSessionRecording.vue";
import useSessionRecordingStore from "@/store/modules/session_recording";
import handleError from "@/utils/handleError";
import * as hasPermissionModule from "@/utils/permission";
import { getSessionRecordStatus } from "@/store/api/users";

vi.mock("@/store/api/users");

describe("SettingSessionRecording", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingSessionRecording>>;
  let sessionRecordingStore: ReturnType<typeof useSessionRecordingStore>;

  const mountWrapper = ({ tenantId = "tenant-123", hasPermission = true, stubActions = true } = {}) => {
    vi.spyOn(hasPermissionModule, "default").mockReturnValue(hasPermission);

    wrapper = mountComponent(SettingSessionRecording, {
      props: { tenantId },
      piniaOptions: { stubActions },
    });

    sessionRecordingStore = useSessionRecordingStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Switch rendering", () => {
    it("Renders switch component", () => {
      const switchComponent = wrapper.find('[data-test="session-recording-switch"]');
      expect(switchComponent.exists()).toBe(true);
    });

    it("Switch is enabled when user has permission", () => {
      const switchComponent = wrapper.findComponent({ name: "v-switch" });
      expect(switchComponent.props("disabled")).toBe(false);
    });
  });

  describe("Initial status fetch", () => {
    it("Fetches session recording status on mount", async () => {
      await flushPromises();
      expect(sessionRecordingStore.getStatus).toHaveBeenCalled();
    });

    it("Shows error when fetch fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(getSessionRecordStatus).mockRejectedValueOnce(error);

      wrapper.unmount();
      mountWrapper({ stubActions: false });
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch session recording status.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("Status update", () => {
    it("Calls setStatus when switch is toggled to enabled", async () => {
      const switchComponent = wrapper.findComponent({ name: "v-switch" });
      await switchComponent.setValue(true);
      await flushPromises();

      expect(sessionRecordingStore.setStatus).toHaveBeenCalledWith({
        id: "tenant-123",
        status: true,
      });
    });

    it("Calls setStatus when switch is toggled to disabled", async () => {
      const switchComponent = wrapper.findComponent({ name: "v-switch" });
      await switchComponent.setValue(false);
      await flushPromises();

      expect(sessionRecordingStore.setStatus).toHaveBeenCalledWith({
        id: "tenant-123",
        status: false,
      });
    });

    it("Shows success message when enabled", async () => {
      const switchComponent = wrapper.findComponent({ name: "v-switch" });
      await switchComponent.setValue(true);
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Session recording was successfully enabled.");
    });

    it("Shows success message when disabled", async () => {
      const switchComponent = wrapper.findComponent({ name: "v-switch" });
      await switchComponent.setValue(false);
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Session recording was successfully disabled.");
    });
  });

  describe("Error handling", () => {
    it("Shows error when status update fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      wrapper.unmount();
      mountWrapper();
      vi.mocked(sessionRecordingStore.setStatus).mockRejectedValueOnce(error);
      await flushPromises();

      const switchComponent = wrapper.findComponent({ name: "v-switch" });
      await switchComponent.setValue(true);
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update session recording status.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
