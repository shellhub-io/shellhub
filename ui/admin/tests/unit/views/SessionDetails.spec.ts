import { describe, expect, it, beforeEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useSessionsStore from "@admin/store/modules/sessions";
import { mockSession } from "../mocks";
import SessionDetails from "@admin/views/SessionDetails.vue";
import { formatFullDateTime } from "@/utils/date";
import { afterEach } from "vitest";

vi.mock("@admin/store/api/sessions");

describe("SessionDetails", () => {
  let wrapper: VueWrapper<InstanceType<typeof SessionDetails>>;

  const mountWrapper = async (session = mockSession, mockError?: Error) => {
    const router = createCleanAdminRouter();
    await router.push({ name: "sessionDetails", params: { id: session.uid } });
    await router.isReady();

    wrapper = mountComponent(SessionDetails, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: { adminSessions: mockError ? {} : { session } },
        stubActions: !mockError,
      },
    });

    const sessionsStore = useSessionsStore();
    if (mockError) vi.mocked(sessionsStore.fetchSessionById).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when session loads successfully", () => {
    beforeEach(() => mountWrapper());

    it("displays the session uid in the card title", () => {
      expect(wrapper.find(".text-h6").text()).toBe(mockSession.uid);
    });

    it("displays session uid field", () => {
      const uidField = wrapper.find('[data-test="session-uid-field"]');
      expect(uidField.text()).toContain("UID:");
      expect(uidField.text()).toContain(mockSession.uid);
    });

    it("displays device field with link", () => {
      const deviceField = wrapper.find('[data-test="session-device-field"]');
      expect(deviceField.text()).toContain("Device:");
      const link = deviceField.find("a");
      expect(link.exists()).toBe(true);
      expect(link.text()).toBe(mockSession.device.name);
    });

    it("displays username field", () => {
      const usernameField = wrapper.find('[data-test="session-username-field"]');
      expect(usernameField.text()).toContain("Username:");
      expect(usernameField.text()).toContain(mockSession.username);
    });

    it("displays ip address field", () => {
      const ipField = wrapper.find('[data-test="session-ip-field"]');
      expect(ipField.text()).toContain("IP Address:");
      expect(ipField.text()).toContain(mockSession.ip_address);
    });

    it("displays session type field", () => {
      const typeField = wrapper.find('[data-test="session-type-field"]');
      expect(typeField.text()).toContain("Type:");
      expect(typeField.text()).toContain(mockSession.type);
    });

    it("displays started at field", () => {
      const startedField = wrapper.find('[data-test="session-started-field"]');
      expect(startedField.text()).toContain("Started At:");
      expect(startedField.text()).toContain(formatFullDateTime(mockSession.started_at));
    });

    it("displays last seen field", () => {
      const lastSeenField = wrapper.find('[data-test="session-last-seen-field"]');
      expect(lastSeenField.text()).toContain("Last Seen:");
      expect(lastSeenField.text()).toContain(formatFullDateTime(mockSession.last_seen));
    });

    it("displays authenticated status field", () => {
      const authenticatedField = wrapper.find('[data-test="session-authenticated-field"]');
      expect(authenticatedField.text()).toContain("Authenticated:");
      expect(authenticatedField.text()).toContain(mockSession.authenticated ? "Yes" : "No");
    });

    it("displays active status icon", () => {
      const activeIcon = wrapper.find('[data-test="active-icon"]');
      expect(activeIcon.exists()).toBe(true);
      expect(activeIcon.classes()).toContain("text-success");
    });
  });

  describe("when session fails to load", () => {
    it("shows error snackbar", async () => {
      await mountWrapper(mockSession, createAxiosError(404, "Not Found"));

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to get session details.");
    });
  });
});
