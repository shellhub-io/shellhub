import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useUsersStore from "@admin/store/modules/users";
import UserResetPassword from "@admin/components/User/UserResetPassword.vue";

describe("UserResetPassword", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserResetPassword>>;
  let usersStore: ReturnType<typeof useUsersStore>;
  const mockUserId = "user-123";
  const mockGeneratedPassword = "generated-password-456";

  const mountWrapper = () => {
    wrapper = mountComponent(UserResetPassword, {
      props: { userId: mockUserId },
      attachTo: document.body,
    });

    usersStore = useUsersStore();
  };

  const getDialog = () => new DOMWrapper(document.body).find('[role="dialog"]');

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the trigger icon button", () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      expect(triggerBtn.exists()).toBe(true);
    });

    it("does not show dialog initially", () => {
      expect(getDialog().exists()).toBe(false);
    });
  });

  describe("opening dialog", () => {
    beforeEach(() => mountWrapper());

    it("shows dialog when clicking the trigger button", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      expect(dialog.exists()).toBe(true);
      expect(dialog.text()).toContain("Enable Local Authentication");
    });

    it("displays step 1 content with confirmation message", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      expect(dialog.text()).toContain("This action will enable local authentication");
      expect(dialog.text()).toContain("generate a new password");
    });

    it("shows enable and cancel buttons in step 1", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      expect(dialog.find('[data-test="enable-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="cancel-btn"]').exists()).toBe(true);
    });
  });

  describe("enabling local authentication", () => {
    beforeEach(() => {
      mountWrapper();
      vi.mocked(usersStore.resetUserPassword).mockResolvedValue(mockGeneratedPassword);
    });

    it("calls store action when clicking enable button", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      expect(usersStore.resetUserPassword).toHaveBeenCalledWith(mockUserId);
    });

    it("proceeds to step 2 after successful password reset", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      // Step 2 should show the warning alert
      expect(dialog.find('[data-test="password-warning"]').exists()).toBe(true);
    });

    it("displays generated password in step 2", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      const passwordField = dialog.find('[data-test="generated-password-field"] input');
      expect((passwordField.element as HTMLInputElement).value).toBe(mockGeneratedPassword);
    });

    it("shows password warning alert in step 2", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      const warning = dialog.find('[data-test="password-warning"]');
      expect(warning.exists()).toBe(true);
      expect(warning.text()).toContain("Users are strongly encouraged to change this password");
    });

    it("shows close button in step 2", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="enable-btn"]').exists()).toBe(false);
      expect(dialog.find('[data-test="cancel-btn"]').exists()).toBe(false);
    });

    it("password field is readonly", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      const passwordField = dialog.find('[data-test="generated-password-field"] input');
      expect(passwordField.attributes("readonly")).toBeDefined();
    });
  });

  describe("error handling", () => {
    beforeEach(() => mountWrapper());

    it("shows error message when password reset fails", async () => {
      vi.mocked(usersStore.resetUserPassword).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to reset user password. Please try again.");
    });

    it("stays on step 1 when password reset fails", async () => {
      vi.mocked(usersStore.resetUserPassword).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      // Should still be on step 1 with enable button visible
      expect(dialog.find('[data-test="enable-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="password-warning"]').exists()).toBe(false);
    });
  });

  describe("closing dialog", () => {
    beforeEach(() => {
      mountWrapper();
      vi.mocked(usersStore.resetUserPassword).mockResolvedValue(mockGeneratedPassword);
    });

    it("closes dialog and resets to step 1 when clicking cancel in step 1", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const dialogContent = getDialog().find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none;");
    });

    it("emits update event when closing from step 1", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog and resets to step 1 when clicking close in step 2", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      // Now in step 2
      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      const dialogContent = getDialog().find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none;");
    });

    it("emits update event when closing from step 2", async () => {
      const triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      const dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      // Now in step 2
      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("resets to step 1 when reopening dialog after closing from step 2", async () => {
      // Open and go to step 2
      let triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      let dialog = getDialog();
      const enableBtn = dialog.find('[data-test="enable-btn"]');
      await enableBtn.trigger("click");
      await flushPromises();

      // Close from step 2
      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      // Reopen
      triggerBtn = wrapper.find('[data-test="open-dialog-icon"]');
      await triggerBtn.trigger("click");
      await flushPromises();

      dialog = getDialog();

      // Should be back to step 1 with enable button
      expect(dialog.find('[data-test="enable-btn"]').exists()).toBe(true);
    });
  });
});
