import { VueWrapper } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import createCleanRouter from "@tests/utils/router";
import NamespaceInviteCard from "@/views/NamespaceInviteCard.vue";
import { routes } from "@/router";

const mockRoutes = [
  ...routes,
  { name: "AcceptInvite", path: "/accept-invite", component: NamespaceInviteCard },
];

describe("Namespace Invite Card View", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceInviteCard>>;

  const mountWrapper = async (setLocalStorage = true) => {
    const userId = "507f1f77bcf86cd799439011";
    const tenantId = "fake-tenant";

    if (setLocalStorage) {
      localStorage.setItem("tenant", tenantId);
      localStorage.setItem("id", userId);
    }

    const router = createCleanRouter(mockRoutes);
    await router.push({ name: "AcceptInvite", query: { "user-id": userId, "tenant-id": tenantId } });
    await router.isReady();

    wrapper = mountComponent(NamespaceInviteCard, { global: { plugins: [router] } });
  };

  afterEach(() => {
    wrapper?.unmount();
    localStorage.clear();
  });

  describe("when user is logged in with matching account", () => {
    beforeEach(() => mountWrapper());

    it("renders the invitation dialog with correct title", () => {
      expect(wrapper.find('[data-test="title"]').text()).toBe("Namespace Invitation");
    });

    it("displays the invitation message", () => {
      const message = wrapper.find('[data-test="message"]');
      expect(message.exists()).toBe(true);
    });

    it("displays accept and decline buttons", () => {
      expect(wrapper.find('[data-test="accept-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="decline-btn"]').exists()).toBe(true);
    });

    it("renders all dialog action elements", () => {
      expect(wrapper.find('[data-test="actions"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="spacer"]').exists()).toBe(true);
    });
  });

  describe("when user is not logged in with matching account", () => {
    beforeEach(() => mountWrapper(false));

    it("displays error alert for invalid user", () => {
      const errorAlert = wrapper.find('[data-test="error-alert"]');
      expect(errorAlert.exists()).toBe(true);
      expect(errorAlert.text()).toContain("You aren't logged in the account meant for this invitation.");
    });

    it("disables the accept button", () => {
      expect(wrapper.find('[data-test="accept-btn"]').attributes("disabled")).toBeDefined();
    });

    it("changes decline button text to redirect home", () => {
      expect(wrapper.find('[data-test="decline-btn"]').text()).toContain("Back to Home Page");
    });
  });
});
