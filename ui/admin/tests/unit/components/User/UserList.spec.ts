import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { Router } from "vue-router";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useUsersStore from "@admin/store/modules/users";
import useAuthStore from "@admin/store/modules/auth";
import UserList from "@admin/components/User/UserList.vue";
import { mockUsers } from "../../mocks";
import { IAdminUser } from "@admin/interfaces/IUser";

const mockSAMLUser: IAdminUser = {
  ...mockUsers[0],
  id: "user-saml",
  username: "samluser",
  preferences: {
    auth_methods: ["saml"],
  },
};

const mockLocalUser: IAdminUser = {
  ...mockUsers[0],
  id: "user-local",
  username: "localuser",
  preferences: {
    auth_methods: ["local"],
  },
};

const testUsers = [mockSAMLUser, mockLocalUser];

describe("UserList", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserList>>;
  let router: Router;
  let usersStore: ReturnType<typeof useUsersStore>;
  let authStore: ReturnType<typeof useAuthStore>;

  const mountWrapper = (mockUserCount?: number) => {
    router = createCleanAdminRouter();

    wrapper = mountComponent(UserList, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: {
          adminUsers: {
            users: testUsers,
            usersCount: mockUserCount ?? testUsers.length,
          },
        },
      },
    });

    usersStore = useUsersStore();
    authStore = useAuthStore();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the data table", () => {
      expect(wrapper.find('[data-test="users-list"]').exists()).toBe(true);
    });

    it("displays user names", () => {
      const nameCells = wrapper.findAll('[data-test="name-cell"]');
      expect(nameCells[0].text()).toBe(testUsers[0].name);
      expect(nameCells[1].text()).toBe(testUsers[1].name);
    });

    it("displays user emails", () => {
      const emailCells = wrapper.findAll('[data-test="email-cell"]');
      expect(emailCells[0].text()).toBe(testUsers[0].email);
      expect(emailCells[1].text()).toBe(testUsers[1].email);
    });

    it("displays user usernames", () => {
      const usernameCells = wrapper.findAll('[data-test="username-cell"]');
      expect(usernameCells[0].text()).toBe(testUsers[0].username);
      expect(usernameCells[1].text()).toBe(testUsers[1].username);
    });

    it("displays user status chips", () => {
      const statusChips = wrapper.findAllComponents({ name: "UserStatusChip" });
      expect(statusChips).toHaveLength(testUsers.length);
    });

    it("displays info buttons for each user", () => {
      const infoButtons = wrapper.findAll('[data-test="info-button"]');
      expect(infoButtons).toHaveLength(testUsers.length);
    });

    it("displays edit buttons for each user", () => {
      const editButtons = wrapper.findAll('[data-test="user-edit-btn"]');
      expect(editButtons).toHaveLength(testUsers.length);
    });

    it("displays login buttons for each user", () => {
      const loginButtons = wrapper.findAll('[data-test="login-button"]');
      expect(loginButtons).toHaveLength(testUsers.length);
    });

    it("displays delete buttons for each user", () => {
      const deleteComponents = wrapper.findAllComponents({ name: "UserDelete" });
      expect(deleteComponents).toHaveLength(testUsers.length);
    });

    it("displays reset password button only for SAML users", () => {
      const resetPasswordComponents = wrapper.findAllComponents({ name: "UserResetPassword" });
      // Only SAML user should have reset password button
      expect(resetPasswordComponents).toHaveLength(1);
    });
  });

  describe("fetching users", () => {
    it("fetches users on mount", () => {
      mountWrapper();

      expect(usersStore.fetchUsersList).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 10,
          page: 1,
        }),
      );
    });

    it("refetches users when page changes", async () => {
      mountWrapper(11); // Mock total count to 11 to enable pagination

      // Click next page button
      const nextPageBtn = wrapper.find('[data-test="pager-next"]');
      await nextPageBtn.trigger("click");
      await flushPromises();

      expect(usersStore.fetchUsersList).toHaveBeenCalledWith(
        expect.objectContaining({
          page: 2,
        }),
      );
    });

    it("refetches users when items per page changes", async () => {
      mountWrapper(20);

      // Change items per page via combobox
      const ippCombo = wrapper.find('[data-test="ipp-combo"] input');
      await ippCombo.setValue(20);
      await flushPromises();

      expect(usersStore.fetchUsersList).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 20,
        }),
      );
    });
  });

  describe("navigating to user details", () => {
    beforeEach(() => mountWrapper());

    it("navigates when clicking info button", async () => {
      const pushSpy = vi.spyOn(router, "push");
      const infoButton = wrapper.findAll('[data-test="info-button"]')[0];

      await infoButton.trigger("click");

      expect(pushSpy).toHaveBeenCalledWith({
        name: "userDetails",
        params: { id: testUsers[0].id },
      });
    });
  });

  describe("login with token", () => {
    const windowOpenSpy = vi.spyOn(window, "open").mockImplementation(() => null);

    beforeEach(() => {
      mountWrapper();
      vi.mocked(authStore.getLoginToken).mockResolvedValue("mock-token-123");
    });

    it("gets login token and opens new window when clicking login button", async () => {
      const loginButton = wrapper.findAll('[data-test="login-button"]')[0];
      await loginButton.trigger("click");
      await flushPromises();

      expect(authStore.getLoginToken).toHaveBeenCalledWith(testUsers[0].id);
      expect(windowOpenSpy).toHaveBeenCalledWith(
        "/login?token=mock-token-123",
        "_target",
      );
    });

    it("shows error when getting login token fails", async () => {
      vi.mocked(authStore.getLoginToken).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const loginButton = wrapper.findAll('[data-test="login-button"]')[0];
      await loginButton.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to get the login token.");
      expect(windowOpenSpy).not.toHaveBeenCalled();
    });
  });

  describe("opening edit dialog", () => {
    beforeEach(() => mountWrapper());

    it("renders UserFormDialog components for each user", () => {
      const formDialogs = wrapper.findAllComponents({ name: "UserFormDialog" });
      expect(formDialogs).toHaveLength(testUsers.length);
    });

    it("passes user data to UserFormDialog", () => {
      const formDialogs = wrapper.findAllComponents({ name: "UserFormDialog" });
      expect(formDialogs[0].props("user")).toEqual(testUsers[0]);
      expect(formDialogs[1].props("user")).toEqual(testUsers[1]);
    });
  });

  describe("user deletion", () => {
    beforeEach(() => mountWrapper());

    it("renders UserDelete component for each user", () => {
      const deleteComponents = wrapper.findAllComponents({ name: "UserDelete" });
      expect(deleteComponents).toHaveLength(testUsers.length);
    });

    it("passes user id to UserDelete component", () => {
      const deleteComponents = wrapper.findAllComponents({ name: "UserDelete" });
      expect(deleteComponents[0].props("id")).toBe(testUsers[0].id);
      expect(deleteComponents[1].props("id")).toBe(testUsers[1].id);
    });
  });

  describe("reset password for SAML users", () => {
    beforeEach(() => mountWrapper());

    it("shows reset password button only for users with SAML-only authentication", () => {
      const resetPasswordComponents = wrapper.findAllComponents({ name: "UserResetPassword" });

      // Should only show for SAML user
      expect(resetPasswordComponents).toHaveLength(1);
      expect(resetPasswordComponents[0].props("userId")).toBe(testUsers[0].id);
    });

    it("refetches users after password reset", async () => {
      const resetPasswordComponent = wrapper.findComponent({ name: "UserResetPassword" });
      const postResetFetchSpy = vi.spyOn(usersStore, "fetchUsersList");
      await resetPasswordComponent.vm.$emit("update");
      await flushPromises();

      // Should have been called once on mount and once after update
      expect(postResetFetchSpy).toHaveBeenCalled();
    });
  });

  describe("error handling", () => {
    it("shows error snackbar when fetching users fails", async () => {
      mountWrapper(11);
      vi.mocked(usersStore.fetchUsersList).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      // Trigger refetch by changing page
      const nextPageBtn = wrapper.find('[data-test="pager-next"]');
      await nextPageBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch users.");
    });
  });
});
