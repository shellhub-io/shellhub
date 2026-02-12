import { describe, expect, it, afterEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import InvitationDecline from "@/components/Invitations/InvitationDecline.vue";
import useInvitationsStore from "@/store/modules/invitations";
import handleError from "@/utils/handleError";

const triggerButtonTemplate = `
  <template #default="{ openDialog }">
    <button 
      data-test="trigger-button" 
      @click="openDialog"
    >
      Decline
    </button>
  </template>
`;

describe("InvitationDecline", () => {
  let wrapper: VueWrapper<InstanceType<typeof InvitationDecline>>;
  let dialog: DOMWrapper<HTMLElement>;
  let invitationsStore: ReturnType<typeof useInvitationsStore>;

  const mountWrapper = (namespaceName = "Test Namespace", onSuccess?: () => void) => {
    wrapper = mountComponent(InvitationDecline, {
      props: {
        tenant: "tenant-123",
        namespaceName,
        dataTest: "decline-invitation-dialog",
        onSuccess,
      },
      slots: { default: triggerButtonTemplate },
    });

    invitationsStore = useInvitationsStore();
    dialog = new DOMWrapper(document.body);
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Slot content", () => {
    it("Renders slot content", () => {
      mountWrapper();

      expect(wrapper.find('[data-test="trigger-button"]').exists()).toBe(true);
    });

    it("Opens dialog when slot content is clicked", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(true);
    });
  });

  describe("Decline dialog", () => {
    it("Shows confirmation dialog with correct title", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toBe("Decline Invitation");
    });

    it("Shows description with namespace name", async () => {
      mountWrapper("My Namespace");

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("description")).toContain("My Namespace");
      expect(messageDialog.props("description")).toContain("This action cannot be undone");
    });

    it("Shows generic description when namespace name is not provided", async () => {
      mountWrapper("");

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("description")).toBe(
        "You are about to decline this invitation. This action cannot be undone.",
      );
    });

    it("Displays error icon", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("icon")).toBe("mdi-account-remove");
      expect(messageDialog.props("iconColor")).toBe("error");
    });

    it("Shows Decline and Cancel buttons", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("confirmText")).toBe("Decline");
      expect(messageDialog.props("confirmColor")).toBe("error");
      expect(messageDialog.props("cancelText")).toBe("Cancel");
    });
  });

  describe("Invitation decline", () => {
    it("Calls declineInvitation when confirmed", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(invitationsStore.declineInvitation).toHaveBeenCalledWith("tenant-123");
    });

    it("Closes dialog after successful decline", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("Calls onSuccess callback after successful decline", async () => {
      const onSuccess = vi.fn();
      mountWrapper("Test Namespace", onSuccess);

      invitationsStore = useInvitationsStore();
      dialog = new DOMWrapper(document.body);

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(onSuccess).toHaveBeenCalled();
    });
  });

  describe("Error handling", () => {
    it("Handles generic server error", async () => {
      const error = createAxiosError(500, "Internal server error");
      mountWrapper();

      vi.mocked(invitationsStore.declineInvitation).mockRejectedValueOnce(error);

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(invitationsStore.declineInvitation).toHaveBeenCalledWith("tenant-123");
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to decline invitation");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
