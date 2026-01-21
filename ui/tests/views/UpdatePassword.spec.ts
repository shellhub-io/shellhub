import { VueWrapper, flushPromises } from "@vue/test-utils";
import { Router } from "vue-router";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import UpdatePassword from "@/views/UpdatePassword.vue";
import useUsersStore from "@/store/modules/users";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/store/api/users");

describe("UpdatePassword View", () => {
  let wrapper: VueWrapper<InstanceType<typeof UpdatePassword>>;
  let router: Router;
  let usersStore: ReturnType<typeof useUsersStore>;

  const uid = "testID";
  const token = "testtoken";

  const mountWrapper = async (mockError?: Error) => {
    wrapper?.unmount();

    router = createCleanRouter();
    await router.push({ name: "UpdatePassword", query: { id: uid, token } });
    await router.isReady();
    vi.spyOn(router, "push").mockResolvedValue();

    wrapper = mountComponent(UpdatePassword, {
      global: { plugins: [router] },
      piniaOptions: { stubActions: !mockError },
    });

    usersStore = useUsersStore();
    if (mockError) vi.mocked(usersStore.updatePassword).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when page loads", () => {
    beforeEach(() => mountWrapper());

    it("renders the update password form", () => {
      expect(wrapper.find('[data-test="title"]').text()).toContain("Reset your password");
      expect(wrapper.find('[data-test="sub-text"]').text()).toContain("Please insert your new password.");

      const passwordField = wrapper.find('[data-test="password-text"]');
      const confirmPasswordField = wrapper.find('[data-test="password-confirm-text"]');

      expect(passwordField.exists()).toBe(true);
      expect(passwordField.text()).toContain("Password");
      expect(confirmPasswordField.exists()).toBe(true);
      expect(confirmPasswordField.text()).toContain("Confirm Password");
    });

    it("displays the update password button", () => {
      const updateBtn = wrapper.find('[data-test="update-password-btn"]');
      expect(updateBtn.exists()).toBe(true);
      expect(updateBtn.text()).toContain("UPDATE PASSWORD");
    });

    it("displays back to login link", () => {
      const backToLogin = wrapper.find('[data-test="back-to-login"]');
      expect(backToLogin.exists()).toBe(true);
      expect(backToLogin.text()).toContain("Back to");
      expect(backToLogin.text()).toContain("Login");
    });
  });

  describe("when password update succeeds", () => {
    beforeEach(() => mountWrapper());

    it("calls updatePassword with correct parameters and redirects to login", async () => {
      const pushSpy = vi.spyOn(router, "push").mockResolvedValue();

      await wrapper.find('[data-test="password-text"] input').setValue("12345678");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("12345678");
      await wrapper.find('[data-test="update-password-btn"]').trigger("click");
      await flushPromises();

      expect(usersStore.updatePassword).toHaveBeenCalledWith({
        password: "12345678",
        token: token,
        id: uid,
      });
      expect(pushSpy).toHaveBeenCalledWith({ name: "Login" });
    });
  });

  describe("when password update fails", () => {
    beforeEach(() => mountWrapper(createAxiosError(400, "Bad Request")));

    it("displays error message on failure", async () => {
      await wrapper.find('[data-test="password-text"] input').setValue("12345678");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("12345678");
      await wrapper.find('[data-test="update-password-btn"]').trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update password.");
    });

    it("does not redirect to login on error", async () => {
      const pushSpy = vi.spyOn(router, "push");

      await wrapper.find('[data-test="password-text"] input').setValue("12345678");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("12345678");
      await wrapper.find('[data-test="update-password-btn"]').trigger("click");
      await flushPromises();

      expect(pushSpy).not.toHaveBeenCalled();
    });
  });

  describe("form validation", () => {
    beforeEach(() => mountWrapper());

    it("shows error when passwords do not match", async () => {
      await wrapper.find('[data-test="password-text"] input').setValue("12345678");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("different");
      await flushPromises();

      const confirmPasswordField = wrapper.find('[data-test="password-confirm-text"]');
      expect(confirmPasswordField.text()).toContain("Passwords do not match");
    });

    it("does not call updatePassword when passwords do not match", async () => {
      await wrapper.find('[data-test="password-text"] input').setValue("12345678");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("different");
      await flushPromises();
      await wrapper.find('[data-test="update-password-btn"]').trigger("click");

      expect(usersStore.updatePassword).not.toHaveBeenCalled();
    });

    it("calls updatePassword when passwords match", async () => {
      await wrapper.find('[data-test="password-text"] input').setValue("12345678");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("12345678");
      await wrapper.find('[data-test="update-password-btn"]').trigger("click");
      await flushPromises();

      expect(usersStore.updatePassword).toHaveBeenCalledWith({
        password: "12345678",
        token: token,
        id: uid,
      });
    });

    it("shows error when password is too short", async () => {
      await wrapper.find('[data-test="password-text"] input').setValue("123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("123");
      await flushPromises();

      const passwordField = wrapper.find('[data-test="password-text"]');
      expect(passwordField.text()).toContain("Your password should be 5-32 characters long");
    });

    it("does not call updatePassword when password is too short", async () => {
      await wrapper.find('[data-test="password-text"] input').setValue("123");
      await wrapper.find('[data-test="password-confirm-text"] input').setValue("123");
      await flushPromises();
      await wrapper.find('[data-test="update-password-btn"]').trigger("click");

      expect(usersStore.updatePassword).not.toHaveBeenCalled();
    });
  });
});
