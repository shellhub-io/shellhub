import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { mockInvitation } from "@tests/mocks/invitation";
import InvitationEdit from "@/components/Team/Invitation/InvitationEdit.vue";
import useInvitationsStore from "@/store/modules/invitations";
import handleError from "@/utils/handleError";

describe("InvitationEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof InvitationEdit>>;
  let invitationsStore: ReturnType<typeof useInvitationsStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const openDialog = async () => {
    const listItem = wrapper.find('[data-test="invitation-edit-btn"]');
    await listItem.trigger("click");
    await flushPromises();
  };

  const triggerUpdateButton = async () => {
    const updateBtn = dialog.find('[data-test="update-btn"]');
    await updateBtn.trigger("click");
    await flushPromises();
  };

  const mountWrapper = ({
    invitation = mockInvitation,
    hasAuthorization = true,
  } = {}) => {
    wrapper = mountComponent(InvitationEdit, {
      props: { invitation, hasAuthorization },
      attachTo: document.body,
    });
    invitationsStore = useInvitationsStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Rendering", () => {
    it("renders edit list item", () => {
      const listItem = wrapper.find('[data-test="invitation-edit-btn"]');
      expect(listItem.exists()).toBe(true);
    });

    it("renders edit title", () => {
      const title = wrapper.find('[data-test="invitation-edit-title"]');
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Edit Role");
    });

    it("disables list item when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper({ hasAuthorization: false });

      const listItem = wrapper.find('[data-test="invitation-edit-btn"]');
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Dialog", () => {
    it("opens dialog when list item is clicked", async () => {
      await openDialog();

      const editDialog = dialog.find('[data-test="invitation-edit-dialog"]');
      expect(editDialog.exists()).toBe(true);
    });

    it("renders role selector with initial value", async () => {
      await openDialog();

      const roleSelect = wrapper.findComponent({ name: "RoleSelect" });
      expect(roleSelect.exists()).toBe(true);
    });

    it("renders dialog buttons", async () => {
      await openDialog();

      expect(dialog.find('[data-test="update-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="cancel-btn"]').exists()).toBe(true);
    });
  });

  describe("Invitation role update", () => {
    it("calls editInvitation when confirming", async () => {
      await openDialog();
      await triggerUpdateButton();

      expect(invitationsStore.editInvitation).toHaveBeenCalledWith({
        tenant: mockInvitation.namespace.tenant_id,
        user_id: mockInvitation.user.id,
        role: mockInvitation.role,
      });
    });

    it("shows success snackbar on successful update", async () => {
      await openDialog();
      await triggerUpdateButton();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Successfully updated invitation role.");
    });

    it("emits update event on successful edit", async () => {
      await openDialog();
      await triggerUpdateButton();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog on successful update", async () => {
      await openDialog();
      await triggerUpdateButton();

      const editDialogContent = dialog.find('[data-test="invitation-edit-dialog"] .v-overlay__content');
      expect(editDialogContent.attributes("style")).toContain("display: none");
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar when update fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(invitationsStore.editInvitation).mockRejectedValueOnce(error);

      await openDialog();
      await triggerUpdateButton();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update invitation role.");
    });

    it("shows error message for 400 status", async () => {
      const error = createAxiosError(400, "Bad Request");
      vi.mocked(invitationsStore.editInvitation).mockRejectedValueOnce(error);

      await openDialog();
      await triggerUpdateButton();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Invalid invitation or role.");
    });

    it("shows error message for 403 status", async () => {
      const error = createAxiosError(403, "Forbidden");
      vi.mocked(invitationsStore.editInvitation).mockRejectedValueOnce(error);

      await openDialog();
      await triggerUpdateButton();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("You don't have permission to edit invitations.");
    });

    it("shows error message for 404 status", async () => {
      const error = createAxiosError(404, "Not Found");
      vi.mocked(invitationsStore.editInvitation).mockRejectedValueOnce(error);

      await openDialog();
      await triggerUpdateButton();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Invitation not found.");
    });

    it("shows generic error message for other status codes", async () => {
      const error = createAxiosError(503, "Service Unavailable");
      vi.mocked(invitationsStore.editInvitation).mockRejectedValueOnce(error);

      await openDialog();
      await triggerUpdateButton();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update invitation role.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("Dialog close", () => {
    it("closes dialog when cancel button is clicked", async () => {
      await openDialog();

      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const editDialogContent = dialog.find('[data-test="invitation-edit-dialog"] .v-overlay__content');
      expect(editDialogContent.attributes("style")).toContain("display: none");
    });

    it("resets role to initial value when dialog is closed", async () => {
      wrapper.unmount();
      const customInvitation = { ...mockInvitation, role: "administrator" as const };
      mountWrapper({ invitation: customInvitation });

      await openDialog();

      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      // Open again to verify reset
      await openDialog();

      const roleSelect = wrapper.findComponent({ name: "RoleSelect" });
      expect(roleSelect.props("modelValue")).toBe("administrator");
    });
  });
});
