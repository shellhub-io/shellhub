import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import MfaSettings from "@/components/AuthMFA/MfaSettings.vue";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";

const mfaGenerateData = {
  secret: "secret-mfa",
  link: "link-mfa",
  recovery_codes: [
    "HW2wlxV40B",
    "2xsmMUHHHb",
    "DTQgVsaVac",
    "KXPBoXvuWD",
    "QQYTPfotBi",
    "XWiKBEPyb4",
  ],
};

describe("MfaSettings", () => {
  let dialog: DOMWrapper<HTMLElement>;
  let wrapper: VueWrapper<InstanceType<typeof MfaSettings>>;
  let authStore: ReturnType<typeof useAuthStore>;
  let usersStore: ReturnType<typeof useUsersStore>;

  const mountWrapper = (hasRecoveryEmail = false, modelValue = true) => {
    wrapper = mountComponent(MfaSettings, {
      props: { modelValue },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          auth: {
            email: "user@test.com",
            recoveryEmail: hasRecoveryEmail ? "existing@test.com" : null,
          },
        },
      },
    });

    authStore = useAuthStore();
    usersStore = useUsersStore();

    vi.mocked(authStore.generateMfa).mockResolvedValue(mfaGenerateData);

    dialog = new DOMWrapper(document.body).find('[role="dialog"]');
  };

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("initial rendering - step 1 (recovery email)", () => {
    beforeEach(() => mountWrapper());

    it("displays recovery email setup window", () => {
      expect(dialog.exists()).toBe(true);
      expect(dialog.text()).toContain("Recovery Email Setup");
      expect(dialog.text()).toContain("Add a recovery email to secure your MFA process");
    });

    it("shows recovery email explanation", () => {
      expect(dialog.text()).toContain("In case you lose access to all your MFA credentials");
      expect(dialog.text()).toContain("we'll need a recovery email");
    });

    it("displays recovery email input field", () => {
      expect(dialog.find('[data-test="recovery-email-text"]').exists()).toBe(true);
    });

    it("disables save button when email is empty", () => {
      const saveBtn = dialog.find('[data-test="disable-btn"]');
      expect(saveBtn.attributes("disabled")).toBeDefined();
    });

    it("validates recovery email format", async () => {
      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("invalid-email");
      await flushPromises();
      expect(dialog.text()).toContain("Please enter a valid email address");
    });

    it("validates recovery email is not same as current email", async () => {
      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("user@test.com");
      await flushPromises();
      expect(dialog.text()).toContain("Recovery email must not be the same as your current email");
    });

    it("enables save button with valid email", async () => {
      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("recovery@test.com");

      const saveBtn = dialog.find('[data-test="disable-btn"]');
      expect(saveBtn.attributes("disabled")).toBeUndefined();
    });
  });

  describe("recovery email submission", () => {
    beforeEach(() => mountWrapper());

    it("saves recovery email and proceeds to next step", async () => {
      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("recovery@test.com");
      await dialog.find('[data-test="disable-btn"]').trigger("click");

      expect(usersStore.patchData).toHaveBeenCalledWith({ recovery_email: "recovery@test.com" });
      expect(dialog.text()).toContain("Save Your Recovery Codes");
    });

    it("displays error when recovery email is already in use", async () => {
      vi.mocked(usersStore.patchData).mockRejectedValueOnce(createAxiosError(409, "Conflict"));

      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("taken@test.com");
      await dialog.find('[data-test="disable-btn"]').trigger("click");

      expect(dialog.text()).toContain("This recovery email is already in use");
    });

    it("displays error when recovery email is invalid", async () => {
      vi.mocked(usersStore.patchData).mockRejectedValueOnce(createAxiosError(400, "Bad Request"));

      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("bad@test.com");
      await dialog.find('[data-test="disable-btn"]').trigger("click");

      expect(dialog.text()).toContain("This recovery email is invalid");
    });
  });

  describe("step 2 (recovery codes)", () => {
    beforeEach(async () => {
      mountWrapper();

      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("recovery@test.com");
      await dialog.find('[data-test="disable-btn"]').trigger("click");
    });

    it("displays recovery codes warning", () => {
      expect(dialog.find('[data-test="alert-first-page"]').exists()).toBe(true);
      expect(dialog.text()).toContain("Please tick the box below when you're confident");
      expect(dialog.text()).toContain("Without them, you won't be able to get back into your account");
    });

    it("displays all recovery codes", () => {
      const codes = dialog.findAll('[data-test="recovery-codes"]');
      expect(codes).toHaveLength(6);
      expect(codes[0].text()).toBe("HW2wlxV40B");
      expect(codes[5].text()).toBe("XWiKBEPyb4");
    });

    it("displays download button for recovery codes", () => {
      expect(dialog.find('[data-test="download-recovery-codes-btn"]').exists()).toBe(true);
    });

    it("displays copy button for recovery codes", () => {
      expect(dialog.find('[data-test="copy-recovery-codes-btn"]').exists()).toBe(true);
    });

    it("disables next button until checkbox is confirmed", () => {
      const nextBtn = dialog.find('[data-test="next-btn"]');
      expect(nextBtn.attributes("disabled")).toBeDefined();
    });

    it("enables next button after confirming codes are saved", async () => {
      await dialog.findComponent('[data-test="checkbox-recovery"]').setValue(true);

      const nextBtn = dialog.find('[data-test="next-btn"]');
      expect(nextBtn.attributes("disabled")).toBeUndefined();
    });
  });

  describe("step 3 (QR code and verification)", () => {
    beforeEach(async () => {
      mountWrapper();

      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("recovery@test.com");
      await dialog.find('[data-test="disable-btn"]').trigger("click");
      await dialog.findComponent('[data-test="checkbox-recovery"]').setValue(true);
      await dialog.find('[data-test="next-btn"]').trigger("click");
    });

    it("displays QR code setup instructions", () => {
      expect(dialog.text()).toContain("Configure MFA Device");
      expect(dialog.text()).toContain("Scan QR code and verify your authenticator app");
    });

    it("shows setup steps", () => {
      expect(dialog.text()).toContain("Step 1:");
      expect(dialog.text()).toContain("Scan the QR code above or manually enter the Secret Key");
      expect(dialog.text()).toContain("Step 2:");
      expect(dialog.text()).toContain("Enter the 6-digit code from your TOTP provider");
    });

    it("displays verification code input", () => {
      expect(dialog.find('[data-test="verification-code"]').exists()).toBe(true);
    });

    it("displays back button", () => {
      expect(dialog.find('[data-test="back-btn"]').exists()).toBe(true);
    });

    it("disables verify button when code is empty", () => {
      const verifyBtn = dialog.find('[data-test="verify-btn"]');
      expect(verifyBtn.attributes("disabled")).toBeDefined();
    });

    it("enables verify button when code is entered", async () => {
      await dialog.findComponent('[data-test="verification-code"]').setValue("123456");

      const verifyBtn = dialog.find('[data-test="verify-btn"]');
      expect(verifyBtn.attributes("disabled")).toBeUndefined();
    });

    it("goes back to recovery codes step when back button is clicked", async () => {
      await dialog.find('[data-test="back-btn"]').trigger("click");

      expect(dialog.text()).toContain("Save Your Recovery Codes");
      expect(dialog.find('[data-test="recovery-codes"]').exists()).toBe(true);
    });
  });

  describe("MFA enablement", () => {
    beforeEach(async () => {
      mountWrapper();

      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("recovery@test.com");
      await dialog.find('[data-test="disable-btn"]').trigger("click");
      await dialog.findComponent('[data-test="checkbox-recovery"]').setValue(true);
      await dialog.find('[data-test="next-btn"]').trigger("click");
    });

    it("successfully enables MFA with valid code", async () => {
      await dialog.findComponent('[data-test="verification-code"]').setValue("123456");
      await dialog.find('[data-test="verify-btn"]').trigger("click");

      expect(authStore.enableMfa).toHaveBeenCalledWith({
        code: "123456",
        secret: mfaGenerateData.secret,
        recovery_codes: mfaGenerateData.recovery_codes,
      });
      expect(dialog.text()).toContain("Congratulations! You've successfully enabled MFA");
    });

    it("displays error when verification code is invalid", async () => {
      vi.mocked(authStore.enableMfa).mockRejectedValueOnce(createAxiosError(401, "Unauthorized"));

      await dialog.findComponent('[data-test="verification-code"]').setValue("999999");
      await dialog.find('[data-test="verify-btn"]').trigger("click");

      expect(dialog.find('[data-test="error-alert"]').exists()).toBe(true);
      expect(dialog.text()).toContain("verification code sent in your MFA verification is invalid");
    });

    it("displays generic error for server errors", async () => {
      vi.mocked(authStore.enableMfa).mockRejectedValueOnce(createAxiosError(500, "Internal Server Error"));

      await dialog.findComponent('[data-test="verification-code"]').setValue("123456");
      await dialog.find('[data-test="verify-btn"]').trigger("click");

      expect(dialog.find('[data-test="error-alert"]').exists()).toBe(true);
      expect(dialog.text()).toContain("error occurred during your MFA verification");
    });
  });

  describe("step 4 (success)", () => {
    beforeEach(async () => {
      mountWrapper();

      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("recovery@test.com");
      await dialog.find('[data-test="disable-btn"]').trigger("click");
      await dialog.findComponent('[data-test="checkbox-recovery"]').setValue(true);
      await dialog.find('[data-test="next-btn"]').trigger("click");
      await dialog.findComponent('[data-test="verification-code"]').setValue("123456");
      await dialog.find('[data-test="verify-btn"]').trigger("click");
    });

    it("displays success message", () => {
      expect(dialog.text()).toContain("MFA Setup Complete");
      expect(dialog.text()).toContain("Congratulations! You've successfully enabled MFA");
    });

    it("displays success icon", () => {
      expect(dialog.find('[data-test="green-cloud-icon"]').exists()).toBe(true);
    });

    it("displays security benefits", () => {
      const benefits = dialog.find('[data-test="congratulation-bullet-point"]');
      expect(benefits.text()).toContain("Two-step verification");
      expect(benefits.text()).toContain("Reduced risk of unauthorized access");
      expect(benefits.text()).toContain("Enhanced security against phishing attacks");
    });
  });

  describe("dialog with existing recovery email", () => {
    it("skips to step 2 when recovery email already exists", async () => {
      mountWrapper(true, false);
      await wrapper.setProps({ modelValue: true });
      await flushPromises();
      dialog = new DOMWrapper(document.body).find('[role="dialog"]');
      expect(dialog.text()).toContain("Save Your Recovery Codes");
      expect(dialog.find('[data-test="recovery-codes"]').exists()).toBe(true);
      expect(authStore.generateMfa).toHaveBeenCalled();
    });
  });

  describe("dialog controls", () => {
    it("closes dialog when close button is clicked", async () => {
      mountWrapper();

      await dialog.find('[data-test="close-btn"]').trigger("click");

      expect(wrapper.emitted("update:modelValue")).toEqual([[false]]);
    });
  });

  describe("keyboard interactions", () => {
    it("submits verification code on Enter key", async () => {
      mountWrapper();

      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("recovery@test.com");
      await dialog.find('[data-test="disable-btn"]').trigger("click");
      await dialog.findComponent('[data-test="checkbox-recovery"]').setValue(true);
      await dialog.find('[data-test="next-btn"]').trigger("click");

      const codeInput = dialog.findComponent('[data-test="verification-code"]');
      await codeInput.setValue("123456");
      await codeInput.trigger("keyup.enter");

      expect(authStore.enableMfa).toHaveBeenCalled();
    });
  });
});
