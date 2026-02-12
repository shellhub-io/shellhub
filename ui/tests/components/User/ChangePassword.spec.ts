import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import ChangePassword from "@/components/User/ChangePassword.vue";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useUsersStore from "@/store/modules/users";

describe("ChangePassword", () => {
  let wrapper: VueWrapper<InstanceType<typeof ChangePassword>>;
  let dialog: DOMWrapper<Element>;
  let usersStore: ReturnType<typeof useUsersStore>;

  const mountWrapper = (modelValue = true) => {
    wrapper = mountComponent(ChangePassword, {
      props: { modelValue },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          auth: {
            name: "Test User",
            username: "testuser",
            email: "test@example.com",
            recoveryEmail: "recovery@example.com",
          },
        },
      },
    });

    usersStore = useUsersStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering", () => {
    it("renders FormDialog with correct props", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.exists()).toBe(true);
      expect(formDialog.props("title")).toBe("Change Password");
      expect(formDialog.props("icon")).toBe("mdi-lock");
      expect(formDialog.props("confirmText")).toBe("Save Password");
      expect(formDialog.props("cancelText")).toBe("Cancel");
    });

    it("renders all password input fields", () => {
      expect(dialog.find('[data-test="password-input"]').exists()).toBe(true);
      expect(dialog.find('[data-test="new-password-input"]').exists()).toBe(true);
      expect(dialog.find('[data-test="confirm-new-password-input"]').exists()).toBe(true);
    });

    it("renders submit and cancel buttons", () => {
      expect(dialog.find('[data-test="change-password-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    });

    it("initializes with all password fields hidden", () => {
      const currentPasswordField = dialog.find('[data-test="password-input"]');
      const newPasswordField = dialog.find('[data-test="new-password-input"]');
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');

      expect(currentPasswordField.find("input").attributes("type")).toBe("password");
      expect(newPasswordField.find("input").attributes("type")).toBe("password");
      expect(confirmPasswordField.find("input").attributes("type")).toBe("password");
    });
  });

  describe("Password visibility toggle", () => {
    it("toggles current password visibility when eye icon is clicked", async () => {
      const currentPasswordField = dialog.find('[data-test="password-input"]');
      const eyeIcon = currentPasswordField.find(".mdi-eye-off");

      expect(currentPasswordField.find("input").attributes("type")).toBe("password");

      await eyeIcon.trigger("click");
      await flushPromises();

      expect(currentPasswordField.find("input").attributes("type")).toBe("text");
      expect(currentPasswordField.find(".mdi-eye").exists()).toBe(true);
    });

    it("toggles new password visibility when eye icon is clicked", async () => {
      const newPasswordField = dialog.find('[data-test="new-password-input"]');
      const eyeIcon = newPasswordField.find(".mdi-eye-off");

      await eyeIcon.trigger("click");
      await flushPromises();

      expect(newPasswordField.find("input").attributes("type")).toBe("text");
    });

    it("toggles confirm password visibility when eye icon is clicked", async () => {
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');
      const eyeIcon = confirmPasswordField.find(".mdi-eye-off");

      await eyeIcon.trigger("click");
      await flushPromises();

      expect(confirmPasswordField.find("input").attributes("type")).toBe("text");
    });
  });

  describe("Form validation", () => {
    it("disables submit button when all fields are empty", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("disables submit button when current password is empty", async () => {
      const newPasswordField = dialog.find('[data-test="new-password-input"]');
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');

      await newPasswordField.find("input").setValue("newpass123");
      await confirmPasswordField.find("input").setValue("newpass123");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("disables submit button when new password is empty", async () => {
      const currentPasswordField = dialog.find('[data-test="password-input"]');
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');

      await currentPasswordField.find("input").setValue("oldpass123");
      await confirmPasswordField.find("input").setValue("newpass123");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("disables submit button when confirm password is empty", async () => {
      const currentPasswordField = dialog.find('[data-test="password-input"]');
      const newPasswordField = dialog.find('[data-test="new-password-input"]');

      await currentPasswordField.find("input").setValue("oldpass123");
      await newPasswordField.find("input").setValue("newpass123");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("shows error when new password is too short", async () => {
      const newPasswordField = dialog.find('[data-test="new-password-input"]');

      await newPasswordField.find("input").setValue("abc");
      await newPasswordField.find("input").trigger("blur");
      await flushPromises();

      expect(newPasswordField.text()).toContain("Your password should be 5-32 characters long");
    });

    it("shows error when new password is too long", async () => {
      const newPasswordField = dialog.find('[data-test="new-password-input"]');

      await newPasswordField.find("input").setValue("a".repeat(33));
      await newPasswordField.find("input").trigger("blur");
      await flushPromises();

      expect(newPasswordField.text()).toContain("Your password should be 5-32 characters long");
    });

    it("shows error when passwords do not match", async () => {
      const newPasswordField = dialog.find('[data-test="new-password-input"]');
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');

      await newPasswordField.find("input").setValue("newpass123");
      await confirmPasswordField.find("input").setValue("different123");
      await flushPromises();

      expect(confirmPasswordField.text()).toContain("Passwords do not match");
    });

    it("enables submit button when all fields are valid and match", async () => {
      const currentPasswordField = dialog.find('[data-test="password-input"]');
      const newPasswordField = dialog.find('[data-test="new-password-input"]');
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');

      await currentPasswordField.find("input").setValue("oldpass123");
      await newPasswordField.find("input").setValue("newpass123");
      await confirmPasswordField.find("input").setValue("newpass123");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(false);
    });
  });

  describe("Password update", () => {
    it("successfully updates password", async () => {
      const currentPasswordField = dialog.find('[data-test="password-input"]');
      const newPasswordField = dialog.find('[data-test="new-password-input"]');
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');
      const submitBtn = dialog.find('[data-test="change-password-btn"]');

      await currentPasswordField.find("input").setValue("oldpass123");
      await newPasswordField.find("input").setValue("newpass123");
      await confirmPasswordField.find("input").setValue("newpass123");
      await flushPromises();

      vi.mocked(usersStore.patchPassword).mockResolvedValueOnce();

      await submitBtn.trigger("click");
      await flushPromises();

      expect(usersStore.patchPassword).toHaveBeenCalledWith({
        name: "Test User",
        username: "testuser",
        email: "test@example.com",
        recovery_email: "recovery@example.com",
        currentPassword: "oldpass123",
        newPassword: "newpass123",
      });
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Password updated successfully.");
      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")![0]).toEqual([false]);
    });

    it("shows error when current password is incorrect (400)", async () => {
      const currentPasswordField = dialog.find('[data-test="password-input"]');
      const newPasswordField = dialog.find('[data-test="new-password-input"]');
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');
      const submitBtn = dialog.find('[data-test="change-password-btn"]');

      await currentPasswordField.find("input").setValue("wrongpass");
      await newPasswordField.find("input").setValue("newpass123");
      await confirmPasswordField.find("input").setValue("newpass123");
      await flushPromises();

      const error = createAxiosError(400, "Bad Request");
      vi.mocked(usersStore.patchPassword).mockRejectedValueOnce(error);

      await submitBtn.trigger("click");
      await flushPromises();

      expect(currentPasswordField.text()).toContain("Your current password is incorrect");
      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "An error occurred while updating the password.",
      );
    });

    it("shows error when passwords do not match on server validation (403)", async () => {
      const currentPasswordField = dialog.find('[data-test="password-input"]');
      const newPasswordField = dialog.find('[data-test="new-password-input"]');
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');
      const submitBtn = dialog.find('[data-test="change-password-btn"]');

      await currentPasswordField.find("input").setValue("oldpass123");
      await newPasswordField.find("input").setValue("newpass123");
      await confirmPasswordField.find("input").setValue("newpass123");
      await flushPromises();

      const error = createAxiosError(403, "Forbidden");
      vi.mocked(usersStore.patchPassword).mockRejectedValueOnce(error);

      await submitBtn.trigger("click");
      await flushPromises();

      expect(newPasswordField.text()).toContain("Your password doesn't match");
      expect(confirmPasswordField.text()).toContain("Your password doesn't match");
      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "An error occurred while updating the password.",
      );
    });

    it("shows generic error for other server errors", async () => {
      const currentPasswordField = dialog.find('[data-test="password-input"]');
      const newPasswordField = dialog.find('[data-test="new-password-input"]');
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');
      const submitBtn = dialog.find('[data-test="change-password-btn"]');

      await currentPasswordField.find("input").setValue("oldpass123");
      await newPasswordField.find("input").setValue("newpass123");
      await confirmPasswordField.find("input").setValue("newpass123");
      await flushPromises();

      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(usersStore.patchPassword).mockRejectedValueOnce(error);

      await submitBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "An error occurred while updating the password.",
      );
    });
  });

  describe("Dialog close behavior", () => {
    it("closes dialog and resets fields when cancel button is clicked", async () => {
      const currentPasswordField = dialog.find('[data-test="password-input"]');
      const newPasswordField = dialog.find('[data-test="new-password-input"]');
      const confirmPasswordField = dialog.find('[data-test="confirm-new-password-input"]');
      const cancelBtn = dialog.find('[data-test="close-btn"]');

      await currentPasswordField.find("input").setValue("oldpass123");
      await newPasswordField.find("input").setValue("newpass123");
      await confirmPasswordField.find("input").setValue("newpass123");
      await flushPromises();

      await cancelBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")![0]).toEqual([false]);

      // Remount to check reset
      wrapper.unmount();
      document.body.innerHTML = "";
      mountWrapper();

      expect(currentPasswordField.find("input").element.value).toBe("");
      expect(newPasswordField.find("input").element.value).toBe("");
      expect(confirmPasswordField.find("input").element.value).toBe("");
    });

    it("closes dialog when FormDialog emits close event", async () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });

      formDialog.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")![0]).toEqual([false]);
    });
  });
});
