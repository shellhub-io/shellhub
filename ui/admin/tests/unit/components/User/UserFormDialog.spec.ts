import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useUsersStore from "@admin/store/modules/users";
import UserFormDialog from "@admin/components/User/UserFormDialog.vue";
import { mockUser, mockInvitedUser, mockNotConfirmedUser } from "../../mocks";

describe("UserFormDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserFormDialog>>;
  let usersStore: ReturnType<typeof useUsersStore>;

  const getDialog = () => new DOMWrapper(document.body).find('[role="dialog"]');

  const openDialog = async () => {
    const buttonSelector = wrapper.props("createUser") ? '[data-test="user-add-btn"]' : '[data-test="user-edit-btn"]';
    const openBtn = wrapper.find(buttonSelector);
    await openBtn.trigger("click");
    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("create user mode", () => {
    const mountWrapper = () => {
      wrapper = mountComponent(UserFormDialog, {
        props: { createUser: true },
        attachTo: document.body,
      });

      usersStore = useUsersStore();
    };

    describe("rendering", () => {
      beforeEach(() => mountWrapper());

      it("renders the add user button", () => {
        const addBtn = wrapper.find('[data-test="user-add-btn"]');
        expect(addBtn.exists()).toBe(true);
        expect(addBtn.text()).toContain("Add User");
      });

      it("does not show dialog initially", () => {
        expect(getDialog().exists()).toBe(false);
      });
    });

    describe("opening dialog", () => {
      beforeEach(async () => {
        mountWrapper();
        await openDialog();
      });

      it("shows dialog when clicking add button", () => {
        const dialog = getDialog();
        expect(dialog.exists()).toBe(true);
        expect(dialog.text()).toContain("Add new user");
      });

      it("shows all form fields", () => {
        const dialog = getDialog();
        expect(dialog.find('[data-test="name-field"]').exists()).toBe(true);
        expect(dialog.find('[data-test="username-field"]').exists()).toBe(true);
        expect(dialog.find('[data-test="email-field"]').exists()).toBe(true);
        expect(dialog.find('[data-test="password-field"]').exists()).toBe(true);
        expect(dialog.find('[data-test="is-admin-checkbox"]').exists()).toBe(true);
      });

      it("does not show user confirmed checkbox in create mode", () => {
        const dialog = getDialog();
        expect(dialog.find('[data-test="is-confirmed-checkbox"]').exists()).toBe(false);
      });

      it("shows create button", () => {
        const dialog = getDialog();
        const confirmBtn = dialog.find('[data-test="confirm-btn"]');
        expect(confirmBtn.exists()).toBe(true);
        expect(confirmBtn.text()).toContain("Create");
      });
    });

    describe("form validation", () => {
      beforeEach(async () => {
        mountWrapper();
        await openDialog();
      });

      it("shows error when name is empty", async () => {
        const dialog = getDialog();
        const nameInput = dialog.find('[data-test="name-field"] input');
        await nameInput.setValue("");
        await nameInput.trigger("blur");
        await flushPromises();

        expect(dialog.text()).toContain("this is a required field");
      });

      it("shows error for invalid email", async () => {
        const dialog = getDialog();
        const emailInput = dialog.find('[data-test="email-field"] input');
        await emailInput.setValue("invalid-email");
        await emailInput.trigger("blur");
        await flushPromises();

        expect(dialog.text()).toContain("this must be a valid email");
      });

      it("accepts valid email", async () => {
        const dialog = getDialog();
        const emailInput = dialog.find('[data-test="email-field"] input');
        await emailInput.setValue("valid@example.com");
        await emailInput.trigger("blur");
        await flushPromises();

        expect(dialog.text()).not.toContain("this must be a valid email");
      });
    });

    describe("password visibility", () => {
      beforeEach(async () => {
        mountWrapper();
        await openDialog();
      });

      it("hides password by default", () => {
        const dialog = getDialog();
        const passwordInput = dialog.find('[data-test="password-field"] input');
        expect(passwordInput.attributes("type")).toBe("password");
      });

      it("toggles password visibility when clicking the eye icon", async () => {
        const dialog = getDialog();
        let passwordInput = dialog.find('[data-test="password-field"] input');
        expect(passwordInput.attributes("type")).toBe("password");

        const eyeIcon = dialog.find('[data-test="password-field"] .mdi-eye-off');
        await eyeIcon.trigger("click");
        await flushPromises();

        passwordInput = dialog.find('[data-test="password-field"] input');
        expect(passwordInput.attributes("type")).toBe("text");
      });
    });

    describe("namespace limit options", () => {
      beforeEach(async () => {
        mountWrapper();
        await openDialog();
      });

      it("shows namespace limit field when checkbox is checked", async () => {
        const dialog = getDialog();
        const checkbox = dialog.find('[data-test="change-namespace-limit-checkbox"] input');
        await checkbox.setValue(true);
        await flushPromises();

        expect(dialog.find('[data-test="max-namespaces-input"]').exists()).toBe(true);
      });

      it("shows disable namespace creation checkbox when change limit is checked", async () => {
        const dialog = getDialog();
        const checkbox = dialog.find('[data-test="change-namespace-limit-checkbox"] input');
        await checkbox.setValue(true);
        await flushPromises();

        expect(dialog.find('[data-test="disable-namespace-creation-checkbox"]').exists()).toBe(true);
      });

      it("disables number input when disable namespace creation is checked", async () => {
        const dialog = getDialog();
        const changeLimitCheckbox = dialog.find('[data-test="change-namespace-limit-checkbox"] input');
        await changeLimitCheckbox.setValue(true);
        await flushPromises();

        const disableCheckbox = dialog.find('[data-test="disable-namespace-creation-checkbox"] input');
        await disableCheckbox.setValue(true);
        await flushPromises();

        const numberInput = dialog.find('[data-test="max-namespaces-input"] input');
        expect(numberInput.attributes("disabled")).toBeDefined();
      });
    });

    describe("creating user", () => {
      beforeEach(async () => {
        mountWrapper();
        await openDialog();
      });

      it("calls store action and shows success message on submit", async () => {
        const dialog = getDialog();
        await dialog.find('[data-test="name-field"] input').setValue("New User");
        await dialog.find('[data-test="username-field"] input').setValue("newuser");
        await dialog.find('[data-test="email-field"] input').setValue("newuser@example.com");
        await dialog.find('[data-test="password-field"] input').setValue("password123");
        await flushPromises();

        const confirmBtn = dialog.find('[data-test="confirm-btn"]');
        await confirmBtn.trigger("click");
        await flushPromises();

        expect(usersStore.addUser).toHaveBeenCalledWith(
          expect.objectContaining({
            name: "New User",
            username: "newuser",
            email: "newuser@example.com",
            password: "password123",
          }),
        );
        expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("User added successfully.");
      });

      it("fetches user list after successful creation", async () => {
        const dialog = getDialog();
        await dialog.find('[data-test="name-field"] input').setValue("New User");
        await dialog.find('[data-test="username-field"] input').setValue("newuser");
        await dialog.find('[data-test="email-field"] input').setValue("newuser@example.com");
        await dialog.find('[data-test="password-field"] input').setValue("password123");
        await flushPromises();

        const confirmBtn = dialog.find('[data-test="confirm-btn"]');
        await confirmBtn.trigger("click");
        await flushPromises();

        expect(usersStore.fetchUsersList).toHaveBeenCalled();
      });

      it("includes admin flag when admin checkbox is checked", async () => {
        const dialog = getDialog();
        await dialog.find('[data-test="name-field"] input').setValue("Admin User");
        await dialog.find('[data-test="username-field"] input').setValue("adminuser");
        await dialog.find('[data-test="email-field"] input').setValue("admin@example.com");
        await dialog.find('[data-test="password-field"] input').setValue("password123");
        await dialog.find('[data-test="is-admin-checkbox"] input').setValue(true);
        await flushPromises();

        const confirmBtn = dialog.find('[data-test="confirm-btn"]');
        await confirmBtn.trigger("click");
        await flushPromises();

        expect(usersStore.addUser).toHaveBeenCalledWith(
          expect.objectContaining({
            admin: true,
          }),
        );
      });

      it("includes namespace limit when specified", async () => {
        const dialog = getDialog();
        await dialog.find('[data-test="name-field"] input').setValue("Limited User");
        await dialog.find('[data-test="username-field"] input').setValue("limited");
        await dialog.find('[data-test="email-field"] input').setValue("limited@example.com");
        await dialog.find('[data-test="password-field"] input').setValue("password123");

        const changeLimitCheckbox = dialog.find('[data-test="change-namespace-limit-checkbox"] input');
        await changeLimitCheckbox.setValue(true);
        await flushPromises();

        const numberInput = dialog.find('[data-test="max-namespaces-input"] input');
        await numberInput.setValue("5");
        await flushPromises();

        const confirmBtn = dialog.find('[data-test="confirm-btn"]');
        await confirmBtn.trigger("click");
        await flushPromises();

        expect(usersStore.addUser).toHaveBeenCalledWith(
          expect.objectContaining({
            max_namespaces: 5,
          }),
        );
      });
    });

    describe("error handling", () => {
      beforeEach(async () => {
        mountWrapper();
        await openDialog();
      });

      it("shows error message when creation fails", async () => {
        vi.mocked(usersStore.addUser).mockRejectedValueOnce(
          createAxiosError(400, "Bad Request", ["name"]),
        );

        const dialog = getDialog();
        await dialog.find('[data-test="name-field"] input').setValue("New User");
        await dialog.find('[data-test="username-field"] input').setValue("newuser");
        await dialog.find('[data-test="email-field"] input').setValue("newuser@example.com");
        await dialog.find('[data-test="password-field"] input').setValue("password123");
        await flushPromises();

        const confirmBtn = dialog.find('[data-test="confirm-btn"]');
        await confirmBtn.trigger("click");
        await flushPromises();

        expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to submit the user data.");
      });

      it("shows field-specific error for duplicate username", async () => {
        vi.mocked(usersStore.addUser).mockRejectedValueOnce(createAxiosError(400, "Bad Request", ["username"]));

        const dialog = getDialog();
        await dialog.find('[data-test="name-field"] input').setValue("New User");
        await dialog.find('[data-test="username-field"] input').setValue("duplicate");
        await dialog.find('[data-test="email-field"] input').setValue("newuser@example.com");
        await dialog.find('[data-test="password-field"] input').setValue("password123");
        await flushPromises();

        const confirmBtn = dialog.find('[data-test="confirm-btn"]');
        await confirmBtn.trigger("click");
        await flushPromises();

        expect(dialog.text()).toContain("This username is invalid!");
      });
    });

    describe("closing dialog", () => {
      it("closes dialog and resets form when cancel is clicked", async () => {
        mountWrapper();
        await openDialog();

        const dialog = getDialog();
        await dialog.find('[data-test="name-field"] input').setValue("Test User");
        await flushPromises();

        const cancelBtn = dialog.find('[data-test="cancel-btn"]');
        await cancelBtn.trigger("click");
        await flushPromises();

        const dialogContent = getDialog().find(".v-overlay__content");
        expect(dialogContent.attributes("style")).toContain("display: none;");
      });
    });
  });

  describe("edit user mode", () => {
    const mountWrapper = (user = mockUser) => {
      wrapper = mountComponent(UserFormDialog, {
        props: {
          createUser: false,
          user,
        },
        piniaOptions: { initialState: { adminAuth: { currentUser: "testuser" } } },
        attachTo: document.body,
      });

      usersStore = useUsersStore();
    };

    describe("rendering", () => {
      beforeEach(() => mountWrapper());

      it("renders the edit button", () => {
        const editBtn = wrapper.find('[data-test="user-edit-btn"]');
        expect(editBtn.exists()).toBe(true);
      });

      it("shows dialog with Edit title when clicking edit button", async () => {
        await openDialog();

        const dialog = getDialog();
        expect(dialog.exists()).toBe(true);
        expect(dialog.text()).toContain("Edit user");
      });

      it("displays current user values in form", async () => {
        await openDialog();

        const dialog = getDialog();
        expect((dialog.find('[data-test="name-field"] input').element as HTMLInputElement).value).toBe(mockUser.name);
        expect((dialog.find('[data-test="username-field"] input').element as HTMLInputElement).value).toBe(mockUser.username);
        expect((dialog.find('[data-test="email-field"] input').element as HTMLInputElement).value).toBe(mockUser.email);
      });

      it("shows update button in edit mode", async () => {
        await openDialog();

        const dialog = getDialog();
        const confirmBtn = dialog.find('[data-test="confirm-btn"]');
        expect(confirmBtn.text()).toContain("Update");
      });

      it("shows user confirmed checkbox in edit mode", async () => {
        wrapper.unmount();
        mountWrapper(mockNotConfirmedUser);

        await openDialog();

        const dialog = getDialog();
        expect(dialog.find('[data-test="is-confirmed-checkbox"]').exists()).toBe(true);
      });
    });

    describe("user confirmation status", () => {
      it("enables confirmed checkbox for not-confirmed users", async () => {
        mountWrapper(mockNotConfirmedUser);

        await openDialog();

        const dialog = getDialog();
        const confirmedCheckbox = dialog.find('[data-test="is-confirmed-checkbox"] input');
        expect(confirmedCheckbox.attributes("disabled")).toBeUndefined();
      });

      it("disables confirmed checkbox for already confirmed users", async () => {
        mountWrapper(mockUser);

        await openDialog();
        await flushPromises();

        const dialog = getDialog();
        const confirmedCheckbox = dialog.find('[data-test="is-confirmed-checkbox"] input');
        expect(confirmedCheckbox.attributes("disabled")).toBeDefined();
      });

      it("disables confirmed checkbox for invited users", async () => {
        mountWrapper(mockInvitedUser);

        await openDialog();

        const dialog = getDialog();
        const confirmedCheckbox = dialog.find('[data-test="is-confirmed-checkbox"] input');
        expect(confirmedCheckbox.attributes("disabled")).toBeDefined();
      });
    });

    describe("admin privileges", () => {
      it("allows changing admin status for other users", async () => {
        const otherUser = { ...mockUser, username: "otheruser", admin: false };
        mountWrapper(otherUser);

        await openDialog();

        const dialog = getDialog();
        const adminCheckbox = dialog.find('[data-test="is-admin-checkbox"] input');
        expect(adminCheckbox.attributes("disabled")).toBeUndefined();
      });

      it("prevents current user from removing their own admin privileges", async () => {
        mountWrapper(mockUser);

        await openDialog();

        const dialog = getDialog();
        const adminCheckbox = dialog.find('[data-test="is-admin-checkbox"] input');
        expect(adminCheckbox.attributes("disabled")).toBeDefined();
      });

      describe("namespace limit options", () => {
        it("starts with namespace creation disabled if user has max_namespaces set to 0", async () => {
          mountWrapper({ ...mockUser, max_namespaces: 0 });
          await openDialog();

          const dialog = getDialog();
          const disableCheckbox = dialog.find('[data-test="disable-namespace-creation-checkbox"] input');
          expect((disableCheckbox.element as HTMLInputElement).checked).toBe(true);
        });
      });
    });

    describe("updating user", () => {
      beforeEach(async () => {
        mountWrapper();
        await openDialog();
      });

      it("calls store action with updated data", async () => {
        const dialog = getDialog();
        await dialog.find('[data-test="name-field"] input').setValue("Updated Name");
        await flushPromises();

        const confirmBtn = dialog.find('[data-test="confirm-btn"]');
        await confirmBtn.trigger("click");
        await flushPromises();

        expect(usersStore.updateUser).toHaveBeenCalledWith(
          expect.objectContaining({
            name: "Updated Name",
            id: mockUser.id,
          }),
        );
        expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("User updated successfully.");
      });

      it("shows error when update fails", async () => {
        vi.mocked(usersStore.updateUser).mockRejectedValueOnce(
          createAxiosError(500, "Internal Server Error", []),
        );

        const dialog = getDialog();
        const confirmBtn = dialog.find('[data-test="confirm-btn"]');
        await confirmBtn.trigger("click");
        await flushPromises();

        expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to submit the user data.");
      });
    });
  });
});
