import { describe, expect, it, afterEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import InvitationAccept from "@/components/Invitations/InvitationAccept.vue";
import useInvitationsStore from "@/store/modules/invitations";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";
import handleError from "@/utils/handleError";

const triggerButtonTemplate = `
  <template #default="{ openDialog }">
    <button 
      data-test="trigger-button" 
      @click="openDialog"
    >
      Accept
    </button>
  </template>
`;

describe("InvitationAccept", () => {
  let wrapper: VueWrapper<InstanceType<typeof InvitationAccept>>;
  let dialog: DOMWrapper<HTMLElement>;
  let invitationsStore: ReturnType<typeof useInvitationsStore>;
  let authStore: ReturnType<typeof useAuthStore>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const mountWrapper = (namespaceName = "Test Namespace", role = "operator", onSuccess?: () => void) => {
    wrapper = mountComponent(InvitationAccept, {
      props: {
        tenant: "tenant-123",
        namespaceName,
        role,
        dataTest: "accept-invitation-dialog",
        onSuccess,
      },
      slots: { default: triggerButtonTemplate },
    });

    invitationsStore = useInvitationsStore();
    authStore = useAuthStore();
    namespacesStore = useNamespacesStore();
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

  describe("Accept dialog", () => {
    it("Shows confirmation dialog with correct title", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toBe("Accept Invitation");
    });

    it("Shows description with namespace and role", async () => {
      mountWrapper("My Namespace", "administrator");

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("description")).toContain("My Namespace");
      expect(messageDialog.props("description")).toContain("administrator");
    });

    it("Shows generic description when namespace name and role are not provided", async () => {
      mountWrapper("", "");

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("description")).toBe(
        "Accepting this invitation will allow you to collaborate with the namespace collaborators.",
      );
    });

    it("Displays primary icon", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("icon")).toBe("mdi-account-check");
      expect(messageDialog.props("iconColor")).toBe("primary");
    });

    it("Shows Accept and Cancel buttons", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("confirmText")).toBe("Accept");
      expect(messageDialog.props("cancelText")).toBe("Cancel");
    });
  });

  describe("Invitation acceptance", () => {
    it("Calls acceptInvitation when confirmed", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(invitationsStore.acceptInvitation).toHaveBeenCalledWith("tenant-123");
    });

    it("Calls enterInvitedNamespace after successful acceptance", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(authStore.enterInvitedNamespace).toHaveBeenCalledWith("tenant-123");
    });

    it("Fetches namespace list after acceptance", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.fetchNamespaceList).toHaveBeenCalled();
    });

    it("Closes dialog after successful acceptance", async () => {
      mountWrapper();

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("Calls onSuccess callback after successful acceptance", async () => {
      const onSuccess = vi.fn();
      mountWrapper("Test Namespace", "operator", onSuccess);

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
      vi.mocked(invitationsStore.acceptInvitation).mockRejectedValueOnce(error);

      await wrapper.find('[data-test="trigger-button"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(invitationsStore.acceptInvitation).toHaveBeenCalledWith("tenant-123");
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to accept invitation");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
