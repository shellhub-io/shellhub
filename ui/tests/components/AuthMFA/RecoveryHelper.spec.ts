import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import RecoveryHelper from "@/components/AuthMFA/RecoveryHelper.vue";
import useAuthStore from "@/store/modules/auth";
import { createAxiosError } from "@tests/utils/axiosError";

describe("RecoveryHelper", () => {
  let dialog: DOMWrapper<HTMLElement>;
  let wrapper: VueWrapper<InstanceType<typeof RecoveryHelper>>;
  let authStore: ReturnType<typeof useAuthStore>;

  const mountWrapper = () => {
    wrapper = mountComponent(RecoveryHelper, {
      props: { modelValue: true },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          auth: {
            disableTimeout: 300000,
            recoveryCode: "test-recovery-code",
          },
        },
      },
    });

    authStore = useAuthStore();
    dialog = new DOMWrapper(document.body).find('[role="dialog"]');
  };

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("displays MFA recovery verification dialog", () => {
      expect(dialog.text()).toContain("MFA Recovery Verification");
      expect(dialog.text()).toContain("Verify access to your authentication device");
    });

    it("shows countdown warning alert", () => {
      expect(dialog.find('[data-test="invalid-login-alert"]').exists()).toBe(true);
      expect(dialog.text()).toContain("Your recovery code will expire in");
    });

    it("displays recovery explanation text", () => {
      expect(dialog.text()).toContain("Recovery codes prove useful when you must access your account");
      expect(dialog.text()).toContain("if you lose access to the device, it is advisable to disable Multi-Factor Authentication");
    });

    it("displays confirmation checkbox", () => {
      expect(dialog.find('[data-test="checkbox-recovery"]').exists()).toBe(true);
      expect(dialog.text()).toContain("I have access to my authentication device");
    });
  });

  describe("dialog controls", () => {
    beforeEach(() => mountWrapper());

    it("disables close button when checkbox is not checked", () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');
      expect(closeBtn.attributes("disabled")).toBeDefined();
    });

    it("enables close button when checkbox is checked", async () => {
      await dialog.findComponent('[data-test="checkbox-recovery"]').setValue(true);

      const closeBtn = dialog.find('[data-test="close-btn"]');
      expect(closeBtn.attributes("disabled")).toBeUndefined();
    });

    it("displays both close and disable buttons", () => {
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="disable-btn"]').exists()).toBe(true);
    });
  });

  describe("MFA disablement", () => {
    beforeEach(() => mountWrapper());

    it("successfully disables MFA with recovery code", async () => {
      vi.mocked(authStore.disableMfa).mockResolvedValueOnce();

      await dialog.find('[data-test="disable-btn"]').trigger("click");

      expect(authStore.disableMfa).toHaveBeenCalledWith({ recovery_code: "test-recovery-code" });
    });

    it("shows success message when MFA is disabled", async () => {
      vi.mocked(authStore.disableMfa).mockResolvedValueOnce();

      await dialog.find('[data-test="disable-btn"]').trigger("click");

      expect(wrapper.emitted("update:modelValue")).toEqual([[false]]);
    });

    it("displays error message when disabling MFA fails", async () => {
      vi.mocked(authStore.disableMfa).mockRejectedValueOnce(createAxiosError(403, "Invalid recovery code"));

      await dialog.find('[data-test="disable-btn"]').trigger("click");

      expect(authStore.disableMfa).toHaveBeenCalledWith({ recovery_code: "test-recovery-code" });
    });
  });

  describe("dialog persistence", () => {
    beforeEach(() => mountWrapper());

    it("prevents closing dialog until checkbox is checked", () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');
      expect(closeBtn.attributes("disabled")).toBeDefined();
    });
  });
});
