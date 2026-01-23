import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import MfaRecover from "@/components/AuthMFA/MfaRecover.vue";
import useAuthStore from "@/store/modules/auth";

describe("MfaRecover", () => {
  let wrapper: VueWrapper<InstanceType<typeof MfaRecover>>;
  let router: ReturnType<typeof createCleanRouter>;
  let authStore: ReturnType<typeof useAuthStore>;

  const mountWrapper = () => {
    router = createCleanRouter();
    wrapper = mountComponent(MfaRecover, { global: { plugins: [router] } });
    authStore = useAuthStore();
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("displays MFA recovery form", () => {
      expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="title"]').text()).toContain("Multi-factor Authentication");
      expect(wrapper.find('[data-test="sub-title"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("If you lost your access to your MFA TOTP provider");
    });

    it("displays recovery code input field", () => {
      expect(wrapper.find('[data-test="recovery-code"]').exists()).toBe(true);
    });

    it("displays recover button", () => {
      expect(wrapper.find('[data-test="recover-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="recover-btn"]').text()).toBe("Recover Account");
    });

    it("shows email recovery option", () => {
      expect(wrapper.text()).toContain("If you lost your recovery codes");
      expect(wrapper.text()).toContain("we'll send you an e-mail");
    });
  });

  describe("form validation", () => {
    beforeEach(() => mountWrapper());

    it("disables recover button when recovery code is empty", () => {
      const recoverBtn = wrapper.find('[data-test="recover-btn"]');
      expect(recoverBtn.attributes("disabled")).toBeDefined();
    });

    it("enables recover button when recovery code is entered", async () => {
      await wrapper.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");

      const recoverBtn = wrapper.find('[data-test="recover-btn"]');
      expect(recoverBtn.attributes("disabled")).toBeUndefined();
    });
  });

  describe("recovery submission", () => {
    beforeEach(() => mountWrapper());

    it("successfully recovers account and redirects to home", async () => {
      const pushSpy = vi.spyOn(router, "push");

      await wrapper.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");
      await wrapper.find('[data-test="recover-btn"]').trigger("click");

      expect(authStore.recoverMfa).toHaveBeenCalledWith("RMS32SAK521A");
      expect(pushSpy).toHaveBeenCalledWith("/");
    });

    it("displays error alert when recovery code is invalid", async () => {
      vi.mocked(authStore.recoverMfa).mockRejectedValueOnce(createAxiosError(403, "Forbidden"));

      await wrapper.findComponent('[data-test="recovery-code"]').setValue("INVALID123");
      await wrapper.find('[data-test="recover-btn"]').trigger("click");

      expect(wrapper.find('[data-test="alert-message"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("verification code sent in your MFA verification is invalid");
    });

    it("displays generic error message for server errors", async () => {
      vi.mocked(authStore.recoverMfa).mockRejectedValueOnce(createAxiosError(500, "Internal Server Error"));

      await wrapper.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");
      await wrapper.find('[data-test="recover-btn"]').trigger("click");

      expect(wrapper.find('[data-test="alert-message"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("error occurred during your MFA verification");
    });
  });

  describe("keyboard interactions", () => {
    beforeEach(() => mountWrapper());

    it("submits recovery code on Enter key", async () => {
      const codeInput = wrapper.findComponent('[data-test="recovery-code"]');
      await codeInput.setValue("RMS32SAK521A");
      await codeInput.trigger("keyup.enter");

      expect(authStore.recoverMfa).toHaveBeenCalledWith("RMS32SAK521A");
    });
  });

  describe("alert dismissal", () => {
    beforeEach(() => mountWrapper());

    it("allows closing error alert", async () => {
      vi.mocked(authStore.recoverMfa).mockRejectedValueOnce(createAxiosError(403, "Forbidden"));

      await wrapper.findComponent('[data-test="recovery-code"]').setValue("INVALID");
      await wrapper.find('[data-test="recover-btn"]').trigger("click");

      expect(wrapper.find('[data-test="alert-message"]').exists()).toBe(true);
    });
  });
});
