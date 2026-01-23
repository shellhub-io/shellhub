import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import MfaForceRecoveryMail from "@/components/AuthMFA/MfaForceRecoveryMail.vue";
import useUsersStore from "@/store/modules/users";

describe("MfaForceRecoveryMail", () => {
  let dialog: DOMWrapper<HTMLElement>;
  let wrapper: VueWrapper<InstanceType<typeof MfaForceRecoveryMail>>;
  let usersStore: ReturnType<typeof useUsersStore>;

  const mountWrapper = () => {
    wrapper = mountComponent(MfaForceRecoveryMail, {
      props: { modelValue: true },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          auth: {
            email: "test@test.com",
            recoveryEmail: null,
          },
        },
      },
    });

    usersStore = useUsersStore();
    dialog = new DOMWrapper(document.body).find('[role="dialog"]');
  };

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
    vi.clearAllMocks();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("displays MFA recovery email setup dialog", () => {
      expect(dialog.text()).toContain("Multi-Factor Authentication Enabled");
      expect(dialog.text()).toContain("Add a recovery email to secure your account access");
    });

    it("shows recovery email explanation", () => {
      expect(dialog.text()).toContain("In case you lose access to all your MFA credentials");
      expect(dialog.text()).toContain("we'll need a recovery email to verify your identity");
    });

    it("displays recovery email input field", () => {
      expect(dialog.find('[data-test="recovery-email-text"]').exists()).toBe(true);
    });

    it("displays save button", () => {
      expect(dialog.find('[data-test="save-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="save-btn"]').text()).toBe("Save Recovery Email");
    });

    it("is a persistent dialog without close button", () => {
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(false);
    });
  });

  describe("form validation", () => {
    beforeEach(() => mountWrapper());

    it("disables save button when email is empty", () => {
      const saveBtn = dialog.find('[data-test="save-btn"]');
      expect(saveBtn.attributes("disabled")).toBeDefined();
    });

    it("validates email format", async () => {
      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("invalid-email");
      await flushPromises();
      expect(dialog.text()).toContain("Please enter a valid email address");
    });

    it("validates recovery email is not same as current email", async () => {
      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("test@test.com");
      await flushPromises();
      expect(dialog.text()).toContain("Recovery email must not be the same as your current email");
    });

    it("enables save button with valid email", async () => {
      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("recovery@test.com");

      const saveBtn = dialog.find('[data-test="save-btn"]');
      expect(saveBtn.attributes("disabled")).toBeUndefined();
    });
  });

  describe("email submission", () => {
    beforeEach(() => mountWrapper());

    it("successfully saves recovery email", async () => {
      vi.mocked(usersStore.patchData).mockResolvedValueOnce();

      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("recovery@test.com");
      await dialog.find('[data-test="save-btn"]').trigger("click");

      expect(usersStore.patchData).toHaveBeenCalledWith({ recovery_email: "recovery@test.com" });
    });

    it("displays error when recovery email is already in use", async () => {
      vi.mocked(usersStore.patchData).mockRejectedValueOnce(createAxiosError(409, "Conflict"));

      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("taken@test.com");
      await dialog.find('[data-test="save-btn"]').trigger("click");

      expect(dialog.text()).toContain("This recovery email is already in use");
    });

    it("displays error when recovery email is invalid", async () => {
      vi.mocked(usersStore.patchData).mockRejectedValueOnce(createAxiosError(400, "Bad Request"));

      await dialog.findComponent('[data-test="recovery-email-text"]').setValue("bad@test.com");
      await dialog.find('[data-test="save-btn"]').trigger("click");

      expect(dialog.text()).toContain("This recovery email is invalid");
    });
  });
});
