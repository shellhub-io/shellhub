import { describe, expect, it, vi, afterEach, beforeEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useSessionsStore from "@admin/store/modules/sessions";
import SessionList from "@admin/components/Sessions/SessionList.vue";
import { mockSessions } from "../../mocks";

describe("SessionList", () => {
  let wrapper: VueWrapper<InstanceType<typeof SessionList>>;
  let sessionsStore: ReturnType<typeof useSessionsStore>;
  let router: ReturnType<typeof createCleanAdminRouter>;

  const mountWrapper = async (sessionCount = mockSessions.length) => {
    router = createCleanAdminRouter();
    wrapper = mountComponent(SessionList, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: {
          adminSessions: {
            sessions: sessionCount > 0 ? mockSessions : [],
            sessionCount,
          },
        },
      },
    });

    sessionsStore = useSessionsStore();

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders session list table", () => {
      const table = wrapper.find('[data-test="session-list"]');
      expect(table.exists()).toBe(true);
    });

    it("displays all sessions", () => {
      const rows = wrapper.findAll('[data-test="tbody-has-items"] tr');
      expect(rows).toHaveLength(mockSessions.length);
    });

    it("displays session uid", () => {
      const firstRow = wrapper.findAll('[data-test="tbody-has-items"] tr')[0];
      expect(firstRow.text()).toContain("session-1");
    });

    it("displays device name", () => {
      const firstRow = wrapper.findAll('[data-test="tbody-has-items"] tr')[0];
      expect(firstRow.text()).toContain("test-device");
    });

    it("displays username", () => {
      const firstRow = wrapper.findAll('[data-test="tbody-has-items"] tr')[0];
      expect(firstRow.text()).toContain("alice");
    });

    it("displays IP address", () => {
      const firstRow = wrapper.findAll('[data-test="tbody-has-items"] tr')[0];
      expect(firstRow.text()).toContain("192.168.1.100");
    });

    it("shows active status icon for active sessions", () => {
      const firstRow = wrapper.findAll('[data-test="tbody-has-items"] tr')[0];
      const activeIcon = firstRow.find(".mdi-check-circle");
      expect(activeIcon.exists()).toBe(true);
    });

    it("shows authenticated icon for authenticated sessions", () => {
      const firstRow = wrapper.findAll('[data-test="tbody-has-items"] tr')[0];
      const authIcon = firstRow.find(".mdi-shield-check");
      expect(authIcon.exists()).toBe(true);
    });

    it("shows not authenticated icon for unauthenticated sessions", () => {
      const secondRow = wrapper.findAll('[data-test="tbody-has-items"] tr')[1];
      const notAuthIcon = secondRow.find(".mdi-shield-alert");
      expect(notAuthIcon.exists()).toBe(true);
    });

    it("shows empty state when no sessions", async () => {
      wrapper.unmount();
      await mountWrapper(0);
      await flushPromises();

      const emptyState = wrapper.find('[data-test="empty-state"]');
      expect(emptyState.exists()).toBe(true);
    });
  });

  describe("initial data loading", () => {
    it("calls fetchSessionList on mount", async () => {
      await mountWrapper();
      expect(sessionsStore.fetchSessionList).toHaveBeenCalledWith({
        perPage: 10,
        page: 1,
      });
    });
  });

  describe("pagination", () => {
    it("changes page when next button is clicked", async () => {
      await mountWrapper(11);

      const nextBtn = wrapper.find('[data-test="pager-next"]');
      await nextBtn.trigger("click");
      await flushPromises();

      expect(sessionsStore.fetchSessionList).toHaveBeenCalledWith({
        perPage: 10,
        page: 2,
      });
    });

    it("changes items per page when selecting from dropdown", async () => {
      await mountWrapper(20);

      // Select 20 items per page option (index 1)
      const ippCombo = wrapper.find('[data-test="ipp-combo"] input');
      await ippCombo.setValue(20);
      await flushPromises();

      expect(sessionsStore.fetchSessionList).toHaveBeenCalledWith({
        perPage: 20,
        page: 1,
      });
    });
  });

  describe("navigation", () => {
    beforeEach(() => mountWrapper());

    it("redirects to device details when clicking device name", async () => {
      const pushSpy = vi.spyOn(router, "push");

      const deviceLinks = wrapper.findAll('[data-test="device-link"]');
      await deviceLinks[0].trigger("click");
      await flushPromises();

      expect(pushSpy).toHaveBeenCalledWith({
        name: "deviceDetails",
        params: { id: "device-123" },
      });
    });

    it("redirects to session details when clicking info icon", async () => {
      const pushSpy = vi.spyOn(router, "push");

      const infoButtons = wrapper.findAll('[data-test="info-button"]');
      await infoButtons[0].trigger("click");
      await flushPromises();

      expect(pushSpy).toHaveBeenCalledWith({
        name: "sessionDetails",
        params: { id: "session-1" },
      });
    });
  });

  describe("error handling", () => {
    it("shows error snackbar when fetching sessions fails", async () => {
      mountComponent(SessionList, { global: { plugins: [createCleanAdminRouter()] } });

      vi.mocked(useSessionsStore().fetchSessionList).mockRejectedValueOnce(
        createAxiosError(500, "Network Error"),
      );

      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch sessions list.");
    });
  });
});
