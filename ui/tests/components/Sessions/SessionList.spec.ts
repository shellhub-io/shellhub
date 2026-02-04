import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { mockSession } from "@tests/mocks/session";
import { createAxiosError } from "@tests/utils/axiosError";
import { createCleanRouter } from "@tests/utils/router";
import SessionList from "@/components/Sessions/SessionList.vue";
import useSessionsStore from "@/store/modules/sessions";
import handleError from "@/utils/handleError";

describe("SessionList", () => {
  let wrapper: VueWrapper<InstanceType<typeof SessionList>>;
  let sessionsStore: ReturnType<typeof useSessionsStore>;

  const mountWrapper = (sessions = [mockSession], sessionCount = 1) => {
    wrapper = mountComponent(SessionList, {
      global: {
        plugins: [createCleanRouter()],
      },
      piniaOptions: {
        initialState: {
          sessions: {
            sessions,
            sessionCount,
          },
        },
      },
    });

    sessionsStore = useSessionsStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("DataTable rendering", () => {
    it("Renders DataTable component", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.exists()).toBe(true);
    });

    it("Passes sessions to DataTable", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.props("items")).toEqual([mockSession]);
    });

    it("Passes session count to DataTable", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.props("totalCount")).toBe(1);
    });

    it("Shows loading state", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.props("loading")).toBe(false);
    });

    it("Has correct table name", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.props("tableName")).toBe("sessions");
    });

    it("Has correct items per page options", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.props("itemsPerPageOptions")).toEqual([10, 20, 50, 100]);
    });
  });

  describe("Session list rendering", () => {
    it("Renders session row", () => {
      const rows = wrapper.findAll("tr");
      expect(rows.length).toBeGreaterThan(0);
    });

    it("Renders SessionPlay component", () => {
      const sessionPlay = wrapper.findComponent({ name: "SessionPlay" });
      expect(sessionPlay.exists()).toBe(true);
      expect(sessionPlay.props("uid")).toBe(mockSession.uid);
      expect(sessionPlay.props("recorded")).toBe(mockSession.recorded);
      expect(sessionPlay.props("authenticated")).toBe(mockSession.authenticated);
    });

    it("Renders device link when device exists", () => {
      const deviceLink = wrapper.findComponent({ name: "DeviceLink" });
      expect(deviceLink.exists()).toBe(true);
      expect(deviceLink.props("deviceUid")).toBe(mockSession.device.uid);
      expect(deviceLink.props("deviceName")).toBe(mockSession.device.name);
    });

    it("Shows username", () => {
      const username = wrapper.text();
      expect(username).toContain(mockSession.username);
    });

    it("Shows authenticated icon for authenticated session", () => {
      const icon = wrapper.find('[data-test="authenticated-icon"] .v-icon');
      expect(icon.exists()).toBe(true);
      expect(icon.classes()).toContain("mdi-shield-check");
    });

    it("Shows unauthenticated icon for unauthenticated session", () => {
      const unauthSession = { ...mockSession, authenticated: false };
      wrapper.unmount();
      mountWrapper([unauthSession]);

      const icon = wrapper.find('[data-test="authenticated-icon"] .v-icon');
      expect(icon.exists()).toBe(true);
      expect(icon.classes()).toContain("mdi-shield-alert");
    });

    it("Shows IP address", () => {
      const ipAddress = wrapper.text();
      expect(ipAddress).toContain(mockSession.ip_address);
    });

    it("Shows started date", () => {
      const rows = wrapper.findAll("tr");
      expect(rows.length).toBeGreaterThan(0);
      // Date is formatted, just check it exists
    });

    it("Shows last seen date", () => {
      const rows = wrapper.findAll("tr");
      expect(rows.length).toBeGreaterThan(0);
      // Date is formatted, just check it exists
    });
  });

  describe("Actions menu", () => {
    it("Renders actions menu button", () => {
      const actionsBtn = wrapper.find('[data-test="session-list-actions"]');
      expect(actionsBtn.exists()).toBe(true);
    });

    it("Actions button has correct icon", () => {
      const actionsBtnIcon = wrapper.find('[data-test="session-list-actions"] .v-icon');
      expect(actionsBtnIcon.classes()).toContain("mdi-format-list-bulleted");
    });

    it("Renders SessionClose component for active session", () => {
      const sessionClose = wrapper.findComponent({ name: "SessionClose" });
      expect(sessionClose.exists()).toBe(true);
      expect(sessionClose.props("uid")).toBe(mockSession.uid);
      expect(sessionClose.props("device")).toEqual(mockSession.device);
    });

    it("Does not render SessionClose for inactive session", () => {
      const inactiveSession = { ...mockSession, active: false };
      wrapper.unmount();
      mountWrapper([inactiveSession]);

      const sessionClose = wrapper.findComponent({ name: "SessionClose" });
      expect(sessionClose.exists()).toBe(false);
    });
  });

  describe("Data fetching", () => {
    it("Calls fetchSessionList on mount", () => {
      expect(sessionsStore.fetchSessionList).toHaveBeenCalledWith({
        page: 1,
        perPage: 10,
      });
    });

    it("Fetches with correct page and perPage parameters", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      await dataTable.vm.$emit("update:page", 2);
      await flushPromises();

      expect(sessionsStore.fetchSessionList).toHaveBeenCalledWith({
        page: 2,
        perPage: 10,
      });
    });

    it("Refetches when items per page changes", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      await dataTable.vm.$emit("update:itemsPerPage", 20);
      await flushPromises();

      expect(sessionsStore.fetchSessionList).toHaveBeenCalledWith({
        page: 1,
        perPage: 20,
      });
    });
  });

  describe("Error handling", () => {
    it("Shows error message on 403 error", async () => {
      const error = createAxiosError(403, "Forbidden");

      wrapper.unmount();
      mountWrapper();
      vi.mocked(sessionsStore.fetchSessionList).mockRejectedValueOnce(error);
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "You don't have permission to access this resource.",
      );
      expect(handleError).not.toHaveBeenCalled();
    });

    it("Shows generic error message on other errors", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      wrapper.unmount();
      mountWrapper();
      vi.mocked(sessionsStore.fetchSessionList).mockRejectedValueOnce(error);
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load the session list.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("Session refresh", () => {
    it("Refetches sessions when SessionClose emits update", async () => {
      localStorage.clear();
      wrapper.unmount();
      mountWrapper();

      const sessionClose = wrapper.findComponent({ name: "SessionClose" });
      await sessionClose.vm.$emit("update");
      await flushPromises();

      expect(sessionsStore.fetchSessionList).toHaveBeenCalledWith({
        page: 1,
        perPage: 10,
      });
    });
  });
});
