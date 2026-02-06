import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import Sessions from "@/views/Sessions.vue";
import { mockSession } from "../mocks";
import useSessionsStore from "@/store/modules/sessions";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/store/api/sessions");

describe("Sessions View", () => {
  let wrapper: VueWrapper<InstanceType<typeof Sessions>>;
  const router = createCleanRouter();

  const mountWrapper = async (hasSessions = true, mockError?: Error) => {
    const initialState = {
      sessions: {
        sessionCount: hasSessions ? 1 : 0,
        sessions: hasSessions ? [mockSession] : [],
      },
    };

    wrapper = mountComponent(Sessions, {
      global: { plugins: [router] },
      piniaOptions: { initialState, stubActions: !mockError },
    });

    const sessionsStore = useSessionsStore();
    if (mockError) vi.mocked(sessionsStore.fetchSessionList).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when sessions exist", () => {
    beforeEach(() => mountWrapper());

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="sessions-title"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Sessions");
    });

    it("displays the sessions list", () => {
      expect(wrapper.find('[data-test="sessions-list"]').exists()).toBe(true);
    });

    it("does not show the no items message", () => {
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
    });
  });

  describe("when no sessions exist", () => {
    beforeEach(() => mountWrapper(false));

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="sessions-title"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Sessions");
    });

    it("does not display the sessions list", () => {
      expect(wrapper.find('[data-test="sessions-list"]').exists()).toBe(false);
    });

    it("shows the no items message", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      expect(noItemsMessage.exists()).toBe(true);
      expect(noItemsMessage.text()).toContain("An SSH session is created when a connection is made");
    });

    it("displays link to connection guide", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      const link = noItemsMessage.find('[data-test="how-to-connect-link"]');
      expect(link.exists()).toBe(true);
      expect(link.text()).toContain("how to connect to your devices");
    });
  });

  describe("when loading sessions fails", () => {
    it("displays error snackbar notification", async () => {
      await mountWrapper(false, createAxiosError(500, "Internal Server Error"));
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load the sessions list.");
    });
  });
});
