import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import MfaDisable from "@/components/AuthMFA/MfaDisable.vue";
import useAuthStore from "@/store/modules/auth";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/store/api/auth", () => ({
  disableMfa: vi.fn(),
  requestMfaReset: vi.fn(),
}));

describe("MfaDisable", () => {
  let dialog: DOMWrapper<HTMLElement>;
  let wrapper: VueWrapper<InstanceType<typeof MfaDisable>>;
  let authStore: ReturnType<typeof useAuthStore>;

  const mountWrapper = () => {
    wrapper = mountComponent(MfaDisable, {
      props: { modelValue: true },
      attachTo: document.body,
    });

    authStore = useAuthStore();

    dialog = new DOMWrapper(document.body).find('[role="dialog"]');
  };

  afterEach(() => {
    wrapper.unmount();
    document.body.innerHTML = "";
    vi.clearAllMocks();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("displays verification code window by default", () => {
      expect(dialog.exists()).toBe(true);
      expect(dialog.find('[data-test="verification-code"]').exists()).toBe(true);
      expect(dialog.find('[data-test="verify-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="use-recovery-code-btn"]').exists()).toBe(true);
    });

    it("shows correct title and description for verification step", () => {
      expect(dialog.text()).toContain("Disable Multi-Factor Authentication");
      expect(dialog.text()).toContain("Verify your identity using your authenticator app");
    });

    it("displays close button", () => {
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    });
  });

  describe("window navigation", () => {
    beforeEach(() => mountWrapper());

    it("switches to recovery code window when user clicks link", async () => {
      await dialog.find('[data-test="use-recovery-code-btn"]').trigger("click");

      expect(dialog.find('[data-test="recovery-code"]').exists()).toBe(true);
      expect(dialog.find('[data-test="recover-btn"]').exists()).toBe(true);
      expect(dialog.text()).toContain("Use Recovery Code");
      expect(dialog.text()).toContain("Enter one of your backup recovery codes");
    });

    it("displays email recovery option in recovery code window", async () => {
      await dialog.find('[data-test="use-recovery-code-btn"]').trigger("click");

      expect(dialog.text()).toContain("If you lost your recovery codes");
      expect(dialog.find('[data-test="recover-email-btn"]').exists()).toBe(true);
    });
  });

  describe("verification code flow", () => {
    beforeEach(() => mountWrapper());

    it("disables verify button when code is empty", () => {
      const verifyBtn = dialog.find('[data-test="verify-btn"]');
      expect(verifyBtn.attributes("disabled")).toBeDefined();
    });

    it("enables verify button when code is entered", async () => {
      await dialog.findComponent('[data-test="verification-code"]').setValue("123456");

      const verifyBtn = dialog.find('[data-test="verify-btn"]');
      expect(verifyBtn.attributes("disabled")).toBeUndefined();
    });

    it("successfully disables MFA with verification code", async () => {
      vi.mocked(authStore.disableMfa).mockResolvedValueOnce();

      await dialog.findComponent('[data-test="verification-code"]').setValue("123456");
      await dialog.find('[data-test="verify-btn"]').trigger("click");

      expect(authStore.disableMfa).toHaveBeenCalledWith({ code: "123456" });
      expect(dialog.find('[data-test="alert-message"]').exists()).toBe(false);
    });

    it("displays error alert when verification code is invalid", async () => {
      vi.mocked(authStore.disableMfa).mockRejectedValueOnce(createAxiosError(403, "Invalid verification code"));

      await dialog.findComponent('[data-test="verification-code"]').setValue("999999");
      await dialog.find('[data-test="verify-btn"]').trigger("click");

      expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
      expect(dialog.text()).toContain("verification code sent in your MFA verification is invalid");
    });

    it("displays generic error message for other errors", async () => {
      vi.mocked(authStore.disableMfa).mockRejectedValueOnce(createAxiosError(500, "Internal server error"));

      await dialog.findComponent('[data-test="verification-code"]').setValue("123456");
      await dialog.find('[data-test="verify-btn"]').trigger("click");

      expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
      expect(dialog.text()).toContain("An error occurred during your MFA verification");
    });
  });

  describe("recovery code flow", () => {
    beforeEach(async () => {
      mountWrapper();
      await dialog.find('[data-test="use-recovery-code-btn"]').trigger("click");
    });

    it("disables recover button when recovery code is empty", () => {
      const recoverBtn = dialog.find('[data-test="recover-btn"]');
      expect(recoverBtn.attributes("disabled")).toBeDefined();
    });

    it("enables recover button when recovery code is entered", async () => {
      await dialog.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");

      const recoverBtn = dialog.find('[data-test="recover-btn"]');
      expect(recoverBtn.attributes("disabled")).toBeUndefined();
    });

    it("successfully disables MFA with recovery code", async () => {
      await dialog.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");
      await dialog.find('[data-test="recover-btn"]').trigger("click");

      expect(authStore.disableMfa).toHaveBeenCalledWith({ recovery_code: "RMS32SAK521A" });
      expect(dialog.find('[data-test="alert-message"]').exists()).toBe(false);
    });

    it("displays error alert when recovery code is invalid", async () => {
      vi.mocked(authStore.disableMfa).mockRejectedValueOnce(createAxiosError(403, "Invalid recovery code"));

      await dialog.findComponent('[data-test="recovery-code"]').setValue("INVALID123");
      await dialog.find('[data-test="recover-btn"]').trigger("click");

      expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
    });
  });

  describe("email recovery flow", () => {
    beforeEach(async () => {
      localStorage.setItem("email", "test@test.com");
      mountWrapper();
      await dialog.find('[data-test="use-recovery-code-btn"]').trigger("click");
    });

    it("shows email sent confirmation window", async () => {
      await dialog.find('[data-test="recover-email-btn"]').trigger("click");

      expect(authStore.requestMfaReset).toHaveBeenCalled();
      expect(dialog.text()).toContain("Email Verification Sent");
      expect(dialog.text()).toContain("Check your email to complete MFA removal");
      expect(dialog.text()).toContain("test@test.com");
    });

    it("displays success icon in email sent window", async () => {
      await dialog.find('[data-test="recover-email-btn"]').trigger("click");

      expect(dialog.find('[data-test="sub-title"]').text()).toContain("test@test.com");
    });

    it("displays error alert when email sending fails", async () => {
      vi.mocked(authStore.requestMfaReset).mockRejectedValueOnce(createAxiosError(403, "Forbidden"));

      await dialog.find('[data-test="recover-email-btn"]').trigger("click");

      expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
      expect(dialog.text()).toContain("error occurred sending your recovery mail");
    });
  });

  describe("dialog controls", () => {
    beforeEach(() => mountWrapper());
    it("closes dialog and resets state", async () => {
      await dialog.findComponent('[data-test="verification-code"]').setValue("123456");
      await dialog.find('[data-test="close-btn"]').trigger("click");

      expect(wrapper.emitted("update:modelValue")).toEqual([[false]]);
    });

    it("hides alert when closing", async () => {
      vi.mocked(authStore.disableMfa).mockRejectedValueOnce(createAxiosError(403, "Forbidden"));

      await dialog.findComponent('[data-test="verification-code"]').setValue("999999");
      await dialog.find('[data-test="verify-btn"]').trigger("click");

      expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);

      await dialog.find('[data-test="use-recovery-code-btn"]').trigger("click");

      expect(dialog.find('[data-test="alert-message"]').exists()).toBe(false);
    });
  });

  describe("keyboard interactions", () => {
    beforeEach(() => mountWrapper());

    it("submits verification code on Enter key", async () => {
      const codeInput = dialog.findComponent('[data-test="verification-code"]');
      await codeInput.setValue("123456");
      await codeInput.trigger("keyup.enter");

      expect(authStore.disableMfa).toHaveBeenCalledWith({ code: "123456" });
    });

    it("submits recovery code on Enter key", async () => {
      await dialog.find('[data-test="use-recovery-code-btn"]').trigger("click");

      const recoveryInput = dialog.findComponent('[data-test="recovery-code"]');
      await recoveryInput.setValue("RMS32SAK521A");
      await recoveryInput.trigger("keyup.enter");

      expect(authStore.disableMfa).toHaveBeenCalledWith({ recovery_code: "RMS32SAK521A" });
    });

    it("navigates with keyboard on recovery link", async () => {
      await dialog.find('[data-test="use-recovery-code-btn"]').trigger("keyup.enter");

      expect(dialog.find('[data-test="recovery-code"]').exists()).toBe(true);
    });

    it("navigates with keyboard on email recovery link", async () => {
      await dialog.find('[data-test="use-recovery-code-btn"]').trigger("click");
      await dialog.find('[data-test="recover-email-btn"]').trigger("keyup.enter");

      expect(authStore.requestMfaReset).toHaveBeenCalled();
    });
  });
});
