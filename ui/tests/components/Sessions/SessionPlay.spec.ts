import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import SessionPlay from "@/components/Sessions/SessionPlay.vue";
import useSessionsStore from "@/store/modules/sessions";
import useUsersStore from "@/store/modules/users";
import { envVariables } from "@/envVariables";
import handleError from "@/utils/handleError";

const mockPlayer = {
  play: vi.fn(),
  pause: vi.fn(),
  seek: vi.fn(),
  dispose: vi.fn(),
  getCurrentTime: vi.fn().mockResolvedValue(0),
  getDuration: vi.fn().mockResolvedValue(100),
  addEventListener: vi.fn(),
};

vi.mock("asciinema-player", () => ({
  create: vi.fn(() => mockPlayer),
}));

describe("SessionPlay", () => {
  let wrapper: VueWrapper<InstanceType<typeof SessionPlay>>;
  let sessionsStore: ReturnType<typeof useSessionsStore>;
  let usersStore: ReturnType<typeof useUsersStore>;

  const mockLogs = '{"version":2,"width":80,"height":24}';

  const mountWrapper = ({ uid = "session-123", recorded = true, authenticated = true, isCommunity = false } = {}) => {
    vi.spyOn(envVariables, "isCommunity", "get").mockReturnValue(isCommunity);

    wrapper = mountComponent(SessionPlay, {
      props: {
        uid,
        recorded,
        authenticated,
      },
      slots: {
        default: `
          <template #default="{ loading, disabled, openDialog }">
            <button data-test="trigger-open-dialog" @click="openDialog">Open</button>
            <span data-test="loading-state">{{ loading }}</span>
            <span data-test="disabled-state">{{ disabled }}</span>
          </template>
        `,
      },
    });

    sessionsStore = useSessionsStore();
    usersStore = useUsersStore();

    vi.mocked(sessionsStore.getSessionLogs).mockResolvedValue(mockLogs);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    vi.restoreAllMocks();
  });

  describe("Tooltip rendering", () => {
    it("Renders tooltip wrapper", () => {
      const tooltip = wrapper.findComponent({ name: "v-tooltip" });
      expect(tooltip.exists()).toBe(true);
    });

    it("Shows tooltip disabled in community edition", () => {
      wrapper.unmount();
      mountWrapper({ isCommunity: true });

      const tooltip = wrapper.findComponent({ name: "v-tooltip" });
      expect(tooltip.props("disabled")).toBe(true);
    });

    it("Shows tooltip when session not recorded in cloud", () => {
      wrapper.unmount();
      mountWrapper({ recorded: false });

      const tooltip = wrapper.findComponent({ name: "v-tooltip" });
      expect(tooltip.props("disabled")).toBe(false);
    });

    it("Shows tooltip when session not authenticated in cloud", () => {
      wrapper.unmount();
      mountWrapper({ authenticated: false });

      const tooltip = wrapper.findComponent({ name: "v-tooltip" });
      expect(tooltip.props("disabled")).toBe(false);
    });
  });

  describe("Slot rendering", () => {
    it("Provides loading state to slot", () => {
      const loadingState = wrapper.find('[data-test="loading-state"]');
      expect(loadingState.exists()).toBe(true);
    });

    it("Provides disabled state to slot", () => {
      wrapper.unmount();
      mountWrapper({ recorded: false });

      const disabledState = wrapper.find('[data-test="disabled-state"]');
      expect(disabledState.exists()).toBe(true);
      expect(disabledState.text()).toBe("true");
    });

    it("Provides openDialog function to slot", () => {
      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      expect(openBtn.exists()).toBe(true);
    });
  });

  describe("Session logs fetching in cloud", () => {
    it("Calls getSessionLogs with correct uid", async () => {
      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      expect(sessionsStore.getSessionLogs).toHaveBeenCalledWith("session-123");
    });

    it("Opens dialog when logs are successfully fetched", async () => {
      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "PlayerDialog" });
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("Passes logs to PlayerDialog", async () => {
      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "PlayerDialog" });
      expect(dialog.props("logs")).toBe(mockLogs);
    });

    it("Shows error when logs are null", async () => {
      // @ts-expect-error Testing null response
      vi.mocked(sessionsStore.getSessionLogs).mockResolvedValueOnce(null);

      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "The session logs were deleted or not recorded.",
      );
    });

    it("Does not open dialog when logs are null", async () => {
      // @ts-expect-error Testing null response
      vi.mocked(sessionsStore.getSessionLogs).mockResolvedValueOnce(null);

      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "PlayerDialog" });
      expect(dialog.props("modelValue")).toBe(false);
    });

    it("Does not fetch logs when session not recorded", async () => {
      wrapper.unmount();
      mountWrapper({ recorded: false });

      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      expect(sessionsStore.getSessionLogs).not.toHaveBeenCalled();
    });
  });

  describe("Community edition behavior", () => {
    it("Shows paywall in community edition", async () => {
      wrapper.unmount();
      mountWrapper({ isCommunity: true });

      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      expect(usersStore.showPaywall).toBe(true);
    });

    it("Does not fetch logs in community edition", async () => {
      wrapper.unmount();
      mountWrapper({ isCommunity: true });

      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      expect(sessionsStore.getSessionLogs).not.toHaveBeenCalled();
    });

    it("Does not open PlayerDialog in community edition", async () => {
      wrapper.unmount();
      mountWrapper({ isCommunity: true });

      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "PlayerDialog" });
      expect(dialog.props("modelValue")).toBe(false);
    });
  });

  describe("Error handling", () => {
    it("Shows error message on fetch failure", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(sessionsStore.getSessionLogs).mockRejectedValueOnce(error);

      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to play the session.");
      expect(handleError).toHaveBeenCalledWith(error);
    });

    it("Does not open dialog on error", async () => {
      const error = createAxiosError(404, "Not Found");
      vi.mocked(sessionsStore.getSessionLogs).mockRejectedValueOnce(error);

      const openBtn = wrapper.find('[data-test="trigger-open-dialog"]');
      await openBtn.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "PlayerDialog" });
      expect(dialog.props("modelValue")).toBe(false);
    });
  });

  describe("PlayerDialog interaction", () => {
    it("Renders PlayerDialog component", () => {
      const dialog = wrapper.findComponent({ name: "PlayerDialog" });
      expect(dialog.exists()).toBe(true);
    });

    it("Dialog is initially closed", () => {
      const dialog = wrapper.findComponent({ name: "PlayerDialog" });
      expect(dialog.props("modelValue")).toBe(false);
    });

    it("Passes null logs initially", () => {
      const dialog = wrapper.findComponent({ name: "PlayerDialog" });
      expect(dialog.props("logs")).toBeNull();
    });
  });
});
