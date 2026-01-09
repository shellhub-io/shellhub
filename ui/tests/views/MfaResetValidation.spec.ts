import { VueWrapper, flushPromises } from "@vue/test-utils";
import { Router } from "vue-router";
import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import createCleanRouter from "@tests/utils/router";
import MfaResetValidation from "@/views/MfaResetValidation.vue";
import useAuthStore from "@/store/modules/auth";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/store/api/auth");

type MfaResetValidationWrapper = VueWrapper<InstanceType<typeof MfaResetValidation>>;

describe("MFA Reset Validation View", () => {
  let wrapper: MfaResetValidationWrapper;
  let router: Router;
  let authStore: ReturnType<typeof useAuthStore>;

  const mountWrapper = async ({
    resetId = "xxxxxx",
    mockError,
  }: {
    resetId?: string;
    mockError?: Error;
  } = {}) => {
    wrapper?.unmount();

    router = createCleanRouter();
    await router.push(`/reset-mfa?id=${resetId}`);
    await router.isReady();

    wrapper = mountComponent(MfaResetValidation, {
      global: { plugins: [router] },
      piniaOptions: { stubActions: !mockError },
    }) as MfaResetValidationWrapper;

    authStore = useAuthStore(wrapper.vm.$pinia);

    if (mockError) vi.spyOn(authStore, "resetMfa").mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when page loads", () => {
    beforeEach(() => mountWrapper());

    it("renders the MFA reset form with all fields", () => {
      expect(wrapper.find('[data-test="verification-title"]').text()).toContain("MFA Disabling");
      expect(wrapper.find('[data-test="verification-subtitle"]').text()).toContain(
        "Please, paste the codes we've sent on your primary and recovery email");

      const emailField = wrapper.find('[data-test="email-text"]');
      const recoveryEmailField = wrapper.find('[data-test="recovery-email-text"]');

      expect(emailField.exists()).toBe(true);
      expect(emailField.text()).toContain("Primary Email Code");
      expect(recoveryEmailField.exists()).toBe(true);
      expect(recoveryEmailField.text()).toContain("Recovery Email Code");
    });

    it("displays the submit button", () => {
      const submitBtn = wrapper.find('[data-test="save-mail-btn"]');
      expect(submitBtn.exists()).toBe(true);
      expect(submitBtn.text()).toContain("Disable MFA");
    });

    it("displays the back to login link", () => {
      expect(wrapper.find('[data-test="back-to-login"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
    });

    it("does not show success or error messages initially", () => {
      expect(wrapper.find('[data-test="verification-success"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="verification-error"]').exists()).toBe(false);
    });
  });

  describe("when MFA reset succeeds", () => {
    beforeEach(() => mountWrapper());

    it("calls resetMfa with correct parameters", async () => {
      await wrapper.find('[data-test="email-text"] input').setValue("123");
      await wrapper.find('[data-test="recovery-email-text"] input').setValue("1234");
      await wrapper.find('[data-test="save-mail-btn"]').trigger("click");
      await flushPromises();

      expect(authStore.resetMfa).toHaveBeenCalledWith({
        id: "xxxxxx",
        recovery_email_code: "1234",
        main_email_code: "123",
      });
    });

    it("displays success message after successful reset", async () => {
      await wrapper.find('[data-test="email-text"] input').setValue("123");
      await wrapper.find('[data-test="recovery-email-text"] input').setValue("1234");
      await wrapper.find('[data-test="save-mail-btn"]').trigger("click");
      await flushPromises();

      expect(wrapper.find('[data-test="verification-success"]').exists()).toBe(true);
    });
  });

  describe("when MFA reset fails", () => {
    beforeEach(() => mountWrapper({ mockError: createAxiosError(403, "Invalid codes") }));

    it("displays error message on failure", async () => {
      await wrapper.find('[data-test="email-text"] input').setValue("123");
      await wrapper.find('[data-test="recovery-email-text"] input').setValue("123");
      await wrapper.find('[data-test="save-mail-btn"]').trigger("click");
      await flushPromises();

      expect(wrapper.find('[data-test="verification-error"]').exists()).toBe(true);
    });

    it("does not show success message on error", async () => {
      await wrapper.find('[data-test="email-text"] input').setValue("123");
      await wrapper.find('[data-test="recovery-email-text"] input').setValue("123");
      await wrapper.find('[data-test="save-mail-btn"]').trigger("click");
      await flushPromises();

      expect(wrapper.find('[data-test="verification-success"]').exists()).toBe(false);
    });
  });
});
