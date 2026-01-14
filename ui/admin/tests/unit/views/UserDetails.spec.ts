import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import { formatFullDateTime } from "@/utils/date";
import useUsersStore from "@admin/store/modules/users";
import UserDetails from "@admin/views/UserDetails.vue";
import { mockUser } from "../mocks";

vi.mock("@admin/store/api/users");

describe("UserDetails", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserDetails>>;

  const mountWrapper = async (mockError?: Error) => {
    const router = createCleanAdminRouter();
    await router.push({ name: "userDetails", params: { id: mockUser.id } });
    await router.isReady();

    wrapper = mountComponent(UserDetails, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: { adminUsers: mockError ? {} : { user: mockUser } },
        stubActions: !mockError,
      },
    });

    const usersStore = useUsersStore();
    if (mockError) vi.mocked(usersStore.fetchUserById).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when user loads successfully", () => {
    beforeEach(() => mountWrapper());

    it("displays the title", () => {
      expect(wrapper.find("h1").text()).toBe("User Details");
    });

    it("displays the username in the card title", () => {
      expect(wrapper.find(".text-h6").text()).toContain(mockUser.username);
    });

    it("displays admin chip when user is admin", () => {
      const chip = wrapper.find('[data-test="user-admin-chip"]');
      expect(chip.exists()).toBe(true);
      expect(chip.text()).toContain("Admin");
    });

    it("displays user uid", () => {
      const uidField = wrapper.find('[data-test="user-uid-field"]');
      expect(uidField.text()).toContain("UID:");
      expect(uidField.text()).toContain(mockUser.id);
    });

    it("displays user name", () => {
      const nameField = wrapper.find('[data-test="user-name-field"]');
      expect(nameField.text()).toContain("Name:");
      expect(nameField.text()).toContain(mockUser.name);
    });

    it("displays username", () => {
      const usernameField = wrapper.find('[data-test="user-username-field"]');
      expect(usernameField.text()).toContain("Username:");
      expect(usernameField.text()).toContain(mockUser.username);
    });

    it("displays email", () => {
      const emailField = wrapper.find('[data-test="user-email-field"]');
      expect(emailField.text()).toContain("Email:");
      expect(emailField.text()).toContain(mockUser.email);
    });

    it("displays status", () => {
      const statusField = wrapper.find('[data-test="user-status-field"]');
      expect(statusField.text()).toContain("Status:");
    });

    it("displays created at date", () => {
      const createdField = wrapper.find('[data-test="user-created-field"]');
      expect(createdField.text()).toContain("Created:");
      expect(createdField.text()).toContain(formatFullDateTime(mockUser.created_at));
    });

    it("displays last login field", () => {
      const lastLoginField = wrapper.find('[data-test="user-last-login-field"]');
      expect(lastLoginField.text()).toContain("Last Login:");
    });

    it("displays mfa status", () => {
      const row = wrapper.find('[data-test="user-mfa-marketing-row"]');
      expect(row.exists()).toBe(true);
      expect(row.text()).toContain("MFA:");
    });

    it("displays auth methods", () => {
      const authField = wrapper.find('[data-test="user-auth-methods-field"]');
      expect(authField.text()).toContain("Auth Methods:");
      expect(authField.text()).toContain("local");
    });

    it("displays namespace counters", () => {
      const row = wrapper.find('[data-test="user-max-namespace-row"]');
      expect(row.text()).toContain("Max Namespaces:");
      expect(row.text()).toContain(String(mockUser.max_namespaces));
      expect(row.text()).toContain("Namespaces Owned:");
      expect(row.text()).toContain(String(mockUser.namespacesOwned));
    });
  });

  describe("when user fails to load", () => {
    it("shows error snackbar", async () => {
      await mountWrapper(createAxiosError(404, "Not Found"));

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to get user details.");
    });
  });
});
