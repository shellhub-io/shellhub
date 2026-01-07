import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import createCleanRouter from "@tests/utils/router";
import ForgotPassword from "@/views/ForgotPassword.vue";
import useUsersStore from "@/store/modules/users";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/store/api/users");

describe("Forgot Password View", () => {
  let wrapper: VueWrapper<InstanceType<typeof ForgotPassword>>;

  const mountWrapper = (mockError?: Error) => {
    wrapper = mountComponent(ForgotPassword, {
      global: { plugins: [createCleanRouter()] },
      piniaOptions: { stubActions: !mockError },
    });

    const usersStore = useUsersStore();
    if (mockError) vi.mocked(usersStore.recoverPassword).mockRejectedValueOnce(mockError);
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when page loads", () => {
    beforeEach(() => mountWrapper());

    it("renders the forgot password form", () => {
      const title = wrapper.find('[data-test="title-text"]');
      const body = wrapper.find('[data-test="body-text"]');
      const accountField = wrapper.find('[data-test="account-text"]');

      expect(title.text()).toContain("Forgot your password");
      expect(body.text()).toContain("Please insert the email associated with the account");
      expect(accountField.exists()).toBe(true);
    });

    it("displays the reset password button", () => {
      const resetBtn = wrapper.find('[data-test="forgot-password-btn"]');
      expect(resetBtn.exists()).toBe(true);
      expect(resetBtn.text()).toContain("RESET PASSWORD");
    });

    it("displays the back to login link", () => {
      const backLink = wrapper.find('[data-test="back-to-login"]');
      expect(backLink.text()).toContain("Back to");
      expect(backLink.text()).toContain("Login");
    });

    it("does not show success message initially", () => {
      expect(wrapper.find('[data-test="success-text"]').exists()).toBe(false);
    });
  });

  describe("when password reset succeeds", () => {
    beforeEach(() => mountWrapper());

    it("displays success message after submitting valid email", async () => {
      const accountField = wrapper.find('[data-test="account-text"] input');
      const form = wrapper.find("form");

      await accountField.setValue("testuser@example.com");
      await form.trigger("submit");
      await flushPromises();

      const successText = wrapper.find('[data-test="success-text"]');
      expect(successText.exists()).toBe(true);
      expect(successText.text()).toContain("An email with password reset instructions has been sent");
    });

    it("hides the form after successful submission", async () => {
      const accountField = wrapper.find('[data-test="account-text"] input');
      const form = wrapper.find("form");

      await accountField.setValue("testuser");
      await form.trigger("submit");
      await flushPromises();

      expect(wrapper.find('[data-test="title-text"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="account-text"]').exists()).toBe(false);
    });
  });

  describe("when password reset fails", () => {
    beforeEach(() => mountWrapper(createAxiosError(404, "User not found")));

    it("displays error snackbar notification", async () => {
      const accountField = wrapper.find('[data-test="account-text"] input');
      const form = wrapper.find("form");

      await accountField.setValue("nonexistent@example.com");
      await form.trigger("submit");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith(
        "Failed to send password reset email. Please ensure the email/username is correct and try again.",
      );
    });

    it("does not show success message on error", async () => {
      const accountField = wrapper.find('[data-test="account-text"] input');
      const form = wrapper.find("form");

      await accountField.setValue("invalid");
      await form.trigger("submit");
      await flushPromises();

      expect(wrapper.find('[data-test="success-text"]').exists()).toBe(false);
    });
  });

  describe("form validation", () => {
    beforeEach(() => mountWrapper());

    it("disables submit button when account field is empty", () => {
      const resetBtn = wrapper.find('[data-test="forgot-password-btn"]');
      expect(resetBtn.attributes("disabled")).toBeDefined();
    });

    it("enables submit button when valid account is entered", async () => {
      const accountField = wrapper.find('[data-test="account-text"] input');
      await accountField.setValue("validuser@example.com");
      await flushPromises();

      const resetBtn = wrapper.find('[data-test="forgot-password-btn"]');
      expect(resetBtn.attributes("disabled")).toBeUndefined();
    });
  });
});
