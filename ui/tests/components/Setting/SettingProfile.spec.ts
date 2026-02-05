import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import SettingProfile from "@/components/Setting/SettingProfile.vue";
import useUsersStore from "@/store/modules/users";
import useAuthStore from "@/store/modules/auth";
import { envVariables } from "@/envVariables";
import handleError from "@/utils/handleError";
import { mockUser } from "@tests/mocks";
import { createCleanRouter } from "@tests/utils/router";
import { UserAuthMethods } from "@admin/interfaces/IUser";

describe("SettingProfile", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingProfile>>;
  let usersStore: ReturnType<typeof useUsersStore>;
  let authStore: ReturnType<typeof useAuthStore>;

  const mountWrapper = ({
    isCommunity = false,
    authMethods = ["local"] as UserAuthMethods,
    isMfaEnabled = false,
  } = {}) => {
    vi.spyOn(envVariables, "isCommunity", "get").mockReturnValue(isCommunity);
    vi.spyOn(envVariables, "isCloud", "get").mockReturnValue(!isCommunity);

    wrapper = mountComponent(SettingProfile, {
      global: { plugins: [createCleanRouter()], stubs: ["MfaSettings", "MfaDisable", "UserDelete"] },
      piniaOptions: {
        initialState: {
          auth: {
            name: mockUser.name,
            username: mockUser.username,
            email: mockUser.email,
            recoveryEmail: mockUser.recovery_email,
            isMfaEnabled,
            authMethods,
          },
        },
      },
    });

    usersStore = useUsersStore();
    authStore = useAuthStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    vi.restoreAllMocks();
  });

  describe("Page header", () => {
    it("Renders page header with correct props", () => {
      const header = wrapper.findComponent({ name: "PageHeader" });
      expect(header.exists()).toBe(true);
      expect(header.props("icon")).toBe("mdi-account-circle");
      expect(header.props("title")).toBe("Account Profile");
      expect(header.props("overline")).toBe("Settings");
    });

    it("Shows edit button when not in edit mode", () => {
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      expect(editBtn.exists()).toBe(true);
      expect(editBtn.text()).toContain("Edit Profile");
    });

    it("Shows cancel and save buttons when in edit mode", async () => {
      await wrapper.find('[data-test="edit-profile-button"]').trigger("click");

      const cancelBtn = wrapper.find('[data-test="cancel-edit-button"]');
      const saveBtn = wrapper.find('[data-test="save-changes-button"]');

      expect(cancelBtn.exists()).toBe(true);
      expect(saveBtn.exists()).toBe(true);
      expect(wrapper.find('[data-test="edit-profile-button"]').exists()).toBe(false);
    });
  });

  describe("Name field", () => {
    it("Renders name input with current user name", async () => {
      await flushPromises();
      const nameInput = wrapper.find('[data-test="name-input"] input');
      expect(nameInput.exists()).toBe(true);
    });

    it("Name input is readonly when not in edit mode", async () => {
      await flushPromises();
      const vTextField = wrapper.findComponent({ name: "v-text-field" });
      expect(vTextField.props("readonly")).toBe(true);
      expect(vTextField.props("disabled")).toBe(true);
    });

    it("Name input is editable when in edit mode", async () => {
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      await editBtn.trigger("click");

      const nameFields = wrapper.findAllComponents({ name: "v-text-field" });
      const nameField = nameFields.find((f) => f.props("modelValue") === mockUser.name);
      expect(nameField?.props("readonly")).toBe(false);
      expect(nameField?.props("disabled")).toBe(false);
    });
  });

  describe("Username field", () => {
    it("Renders username input for local auth", () => {
      const usernameInput = wrapper.find('[data-test="username-input"]');
      expect(usernameInput.exists()).toBe(true);
    });

    it("Does not render username input for non-local auth in Community", async () => {
      wrapper.unmount();
      mountWrapper({ isCommunity: true, authMethods: ["saml"] });
      await flushPromises();

      const usernameInput = wrapper.find('[data-test="username-input"]');
      expect(usernameInput.exists()).toBe(false);
    });
  });

  describe("Email field", () => {
    it("Renders email input", () => {
      const emailInput = wrapper.find('[data-test="email-input"]');
      expect(emailInput.exists()).toBe(true);
    });
  });

  describe("Recovery email field", () => {
    it("Renders recovery email input for local auth", () => {
      const recoveryEmailInput = wrapper.find('[data-test="recovery-email-input"]');
      expect(recoveryEmailInput.exists()).toBe(true);
    });

    it("Does not render recovery email input for non-local auth in Community", async () => {
      wrapper.unmount();
      mountWrapper({ isCommunity: true, authMethods: ["saml"] });
      await flushPromises();

      const recoveryEmailInput = wrapper.find('[data-test="recovery-email-input"]');
      expect(recoveryEmailInput.exists()).toBe(false);
    });
  });

  describe("MFA toggle", () => {
    it("Renders MFA switch", async () => {
      await flushPromises();
      const mfaSwitch = wrapper.find('[data-test="switch-mfa"]');
      expect(mfaSwitch.exists()).toBe(true);
    });

    it("MFA switch is disabled in community", async () => {
      wrapper.unmount();
      mountWrapper({ isCommunity: true });
      await flushPromises();

      const mfaSwitch = wrapper.findComponent({ name: "v-switch" });
      expect(mfaSwitch.props("disabled")).toBe(true);
    });

    it("Shows MFA settings dialog when toggled on", async () => {
      const mfaSwitch = wrapper.find('[data-test="switch-mfa"]');
      await mfaSwitch.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "MfaSettings" });
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("Shows MFA disable dialog when toggled off", async () => {
      wrapper.unmount();
      mountWrapper({ isMfaEnabled: true });
      await flushPromises();

      const mfaSwitch = wrapper.find('[data-test="switch-mfa"]');
      await mfaSwitch.trigger("click");
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "MfaDisable" });
      expect(dialog.props("modelValue")).toBe(true);
    });
  });

  describe("Delete account", () => {
    it("Renders delete account button", () => {
      const deleteBtn = wrapper.find('[data-test="delete-account-btn"]');
      expect(deleteBtn.exists()).toBe(true);
      expect(deleteBtn.text()).toContain("Delete");
    });

    it("Opens delete dialog when button is clicked", async () => {
      const deleteBtn = wrapper.find('[data-test="delete-account-btn"]');
      await deleteBtn.trigger("click");

      const dialog = wrapper.findComponent({ name: "UserDelete" });
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("Shows warning dialog in community edition", async () => {
      wrapper.unmount();
      mountWrapper({ isCommunity: true });
      await flushPromises();

      const dialog = wrapper.findComponent({ name: "UserDeleteWarning" });
      expect(dialog.exists()).toBe(true);
    });
  });

  describe("Edit and cancel", () => {
    it("Enters edit mode when edit button is clicked", async () => {
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      await editBtn.trigger("click");

      const nameFields = wrapper.findAllComponents({ name: "v-text-field" });
      const nameField = nameFields.find((f) => f.props("modelValue") === "Test User");
      expect(nameField?.props("readonly")).toBe(false);
    });

    it("Exits edit mode and resets data when cancel is clicked", async () => {
      await flushPromises();
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("New Name");

      const cancelBtn = wrapper.find('[data-test="cancel-edit-button"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const nameFields = wrapper.findAllComponents({ name: "v-text-field" });
      const nameField = nameFields.find((f) => f.props("readonly") === true);
      expect(nameField?.props("modelValue")).toBe(mockUser.name);
    });
  });

  describe("Update user data", () => {
    it("Calls patchData with correct data", async () => {
      await flushPromises();
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("New Name");

      const saveBtn = wrapper.find('[data-test="save-changes-button"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(usersStore.patchData).toHaveBeenCalled();
    });

    it("Updates auth store after successful update", async () => {
      await flushPromises();
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("New Name");

      const saveBtn = wrapper.find('[data-test="save-changes-button"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(authStore.updateUserData).toHaveBeenCalled();
    });

    it("Shows success message after update", async () => {
      await flushPromises();
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("New Name");

      const saveBtn = wrapper.find('[data-test="save-changes-button"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Profile data updated successfully.");
    });

    it("Exits edit mode after successful update", async () => {
      await flushPromises();
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      await editBtn.trigger("click");

      const nameInput = wrapper.find('[data-test="name-input"] input');
      await nameInput.setValue("New Name");

      const saveBtn = wrapper.find('[data-test="save-changes-button"]');
      await saveBtn.trigger("click");
      await flushPromises();

      const editBtnAfter = wrapper.find('[data-test="edit-profile-button"]');
      expect(editBtnAfter.exists()).toBe(true);
    });
  });

  describe("Update errors", () => {
    it("Shows field error for 400 error", async () => {
      const error = createAxiosError(400, "Bad Request", ["username"]);
      vi.mocked(usersStore.patchData).mockRejectedValueOnce(error);

      await flushPromises();
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      await editBtn.trigger("click");

      const saveBtn = wrapper.find('[data-test="save-changes-button"]');
      await saveBtn.trigger("click");
      await flushPromises();

      const usernameFields = wrapper.findAllComponents({ name: "v-text-field" });
      const usernameField = usernameFields.find((f) =>
        f.props("modelValue") === mockUser.username,
      );
      expect(usernameField?.props("errorMessages")).toBeTruthy();
    });

    it("Shows field error for 409 error", async () => {
      const error = createAxiosError(409, "Conflict", ["email"]);
      vi.mocked(usersStore.patchData).mockRejectedValueOnce(error);

      await flushPromises();
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      await editBtn.trigger("click");

      const saveBtn = wrapper.find('[data-test="save-changes-button"]');
      await saveBtn.trigger("click");
      await flushPromises();

      const emailFields = wrapper.findAllComponents({ name: "v-text-field" });
      const emailField = emailFields.find((f) =>
        f.props("modelValue") === mockUser.email,
      );
      expect(emailField?.props("errorMessages")).toBeTruthy();
    });

    it("Shows generic error for other errors", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(usersStore.patchData).mockRejectedValueOnce(error);

      await flushPromises();
      const editBtn = wrapper.find('[data-test="edit-profile-button"]');
      await editBtn.trigger("click");

      const saveBtn = wrapper.find('[data-test="save-changes-button"]');
      await saveBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while updating user data.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("User info fetch on mount", () => {
    it("Fetches user info on mount", async () => {
      await flushPromises();
      expect(authStore.getUserInfo).toHaveBeenCalled();
    });
  });
});
