import { VueWrapper, flushPromises } from "@vue/test-utils";
import { Router } from "vue-router";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import DetailsSessions from "@/views/DetailsSessions.vue";
import { ISession } from "@/interfaces/ISession";
import { formatFullDateTime } from "@/utils/date";
import { mockDetailedSession } from "../mocks";
import useSessionsStore from "@/store/modules/sessions";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/store/api/sessions");

describe("Details Sessions View", () => {
  let wrapper: VueWrapper<InstanceType<typeof DetailsSessions>>;
  let router: Router;

  const mockSession: ISession = { ...mockDetailedSession, active: false };

  const mountWrapper = async ({
    sessionId = "1",
    initialSession = mockSession,
    mockError,
  }: {
    sessionId?: string;
    initialSession?: Partial<ISession>;
    mockError?: Error;
  } = {}) => {
    localStorage.setItem("tenant", "fake-tenant-data");

    router = createCleanRouter();
    await router.push({ name: "SessionDetails", params: { id: sessionId } });
    await router.isReady();

    wrapper = mountComponent(DetailsSessions, {
      global: { plugins: [router] },
      piniaOptions: {
        ...(mockError ? {} : { initialState: { sessions: { session: initialSession } } }),
        stubActions: !mockError,
      },
    });

    const sessionsStore = useSessionsStore();
    if (mockError) vi.mocked(sessionsStore.getSession).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when session loads successfully", () => {
    beforeEach(() => mountWrapper());

    it("renders all session detail fields with correct values", () => {
      const uidField = wrapper.find('[data-test="session-uid-field"]');
      const userField = wrapper.find('[data-test="session-user-field"]');
      const authenticatedField = wrapper.find('[data-test="session-authenticated-field"]');
      const activeField = wrapper.find('[data-test="session-active-field"]');
      const ipAddressField = wrapper.find('[data-test="session-ip-address-field"]');
      const startedAtField = wrapper.find('[data-test="session-started-at-field"]');
      const lastSeenField = wrapper.find('[data-test="session-last-seen-field"]');

      expect(uidField.text()).toContain(mockSession.uid);
      expect(userField.text()).toContain(mockSession.username);
      expect(authenticatedField.find("i").classes()).toContain("mdi-shield-check");
      expect(activeField.find("i").classes()).toContain("mdi-alert-circle");
      expect(ipAddressField.text()).toContain(mockSession.ip_address);
      expect(startedAtField.text()).toContain(formatFullDateTime(mockSession.started_at));
      expect(lastSeenField.text()).toContain(formatFullDateTime(mockSession.last_seen));
    });

    it.each([
      [true, true],
      [false, false],
    ] as const)(
      "has correct disable attribute for play button for recorded = %s -> %s",
      async (recorded, shouldShowButton) => {
        await mountWrapper({ initialSession: { ...mockSession, recorded } });
        // If enabled, attribute is undefined, otherwise it's an empty string
        expect(wrapper.find('[data-test="session-details-play-btn"]').attributes("disabled")).toBe(shouldShowButton ? undefined : "");
      },
    );
  });

  describe("when session fails to load", () => {
    beforeEach(() => mountWrapper({ sessionId: "inexistent-session", mockError: createAxiosError(404, "Not Found") }));

    it("shows error message when session does not load", () => {
      expect(wrapper.text()).toContain("Something is wrong, try again !");
    });

    it("does not render session detail fields", () => {
      expect(wrapper.find('[data-test="session-uid-field"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="session-details-play-btn"]').exists()).toBe(false);
    });

    it("displays error snackbar notification", () => {
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load session details.");
    });
  });
});
