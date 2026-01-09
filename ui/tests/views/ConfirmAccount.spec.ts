import { VueWrapper, flushPromises } from "@vue/test-utils";
import { Router } from "vue-router";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import ConfirmAccount from "@/views/ConfirmAccount.vue";
import useUsersStore from "@/store/modules/users";
import createCleanRouter from "@tests/utils/router";

type ConfirmAccountWrapper = VueWrapper<InstanceType<typeof ConfirmAccount>>;

describe("Confirm Account View", () => {
  let wrapper: ConfirmAccountWrapper;
  let router: Router;

  const username = "test-user";

  const setupTest = async ({
    queryUsername = username,
    mockError,
  }: {
    queryUsername?: string;
    mockError?: Error;
  } = {}) => {
    router = createCleanRouter();
    await router.push(`/confirm-account?username=${queryUsername}`);
    await router.isReady();

    wrapper = mountComponent(ConfirmAccount, {
      global: { plugins: [router] },
      piniaOptions: { stubActions: !mockError },
    }) as ConfirmAccountWrapper;

    const usersStore = useUsersStore();

    if (mockError) {
      vi.spyOn(console, "error").mockImplementation(() => { });
      vi.spyOn(usersStore, "resendEmail").mockRejectedValueOnce(mockError);
    }

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when page loads", () => {
    beforeEach(() => setupTest());

    it("renders the account activation message", () => {
      const title = wrapper.find('[data-test="title"]');
      const subtitle = wrapper.find('[data-test="subtitle"]');

      expect(title.text()).toContain("Account Activation Required");
      expect(subtitle.text()).toContain("Thank you for registering an account on ShellHub");
      expect(subtitle.text()).toContain("An email was sent with a confirmation link");
    });

    it("displays the resend email button", () => {
      expect(wrapper.find('[data-test="resend-email-btn"]').text()).toContain("Resend Email");
    });

    it("displays the back to login link", () => {
      const loginLink = wrapper.find('[data-test="back-to-login-link"]');
      expect(loginLink.text()).toContain("Back to");
      expect(loginLink.text()).toContain("Login");
    });
  });

  describe("when resend email succeeds", () => {
    it("displays success message and redirects to login", async () => {
      await setupTest();
      const pushSpy = vi.spyOn(router, "push").mockImplementation(() => Promise.resolve());

      await wrapper.find('[data-test="resend-email-btn"]').trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("The email has been sent.");
      expect(pushSpy).toHaveBeenCalledWith({ name: "Login" });
    });
  });

  describe("when resend email fails", () => {
    beforeEach(() => setupTest({ mockError: new Error("Failed to send email") }));

    it("displays error snackbar notification", async () => {
      await wrapper.find('[data-test="resend-email-btn"]').trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while sending the email. Please try again.");
    });

    it("does not redirect to login on error", async () => {
      const pushSpy = vi.spyOn(router, "push");

      await wrapper.find('[data-test="resend-email-btn"]').trigger("click");
      await flushPromises();

      expect(pushSpy).not.toHaveBeenCalled();
    });
  });
});
