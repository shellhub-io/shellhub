import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { mockInvitation } from "@tests/mocks/invitation";
import InvitationCancel from "@/components/Team/Invitation/InvitationCancel.vue";
import useInvitationsStore from "@/store/modules/invitations";
import handleError from "@/utils/handleError";

describe("InvitationCancel", () => {
  let wrapper: VueWrapper<InstanceType<typeof InvitationCancel>>;
  let invitationsStore: ReturnType<typeof useInvitationsStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const openDialog = async () => {
    const listItem = wrapper.find('[data-test="invitation-cancel-btn"]');
    await listItem.trigger("click");
    await flushPromises();
  };

  const mountWrapper = ({
    invitation = mockInvitation,
    hasAuthorization = true,
  } = {}) => {
    wrapper = mountComponent(InvitationCancel, {
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
    it("renders cancel list item", () => {
      const listItem = wrapper.find('[data-test="invitation-cancel-btn"]');
      expect(listItem.exists()).toBe(true);
    });

    it("renders cancel title", () => {
      const title = wrapper.find('[data-test="invitation-cancel-title"]');
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Cancel");
    });

    it("disables list item when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper({ hasAuthorization: false });

      const listItem = wrapper.find('[data-test="invitation-cancel-btn"]');
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Dialog", () => {
    it("opens dialog when list item is clicked", async () => {
      await openDialog();

      const cancelDialog = dialog.find('[data-test="invitation-cancel-dialog"]');
      expect(cancelDialog.exists()).toBe(true);
    });

    it("displays invitation email in dialog description", async () => {
      await openDialog();

      const cancelDialog = dialog.find('[data-test="invitation-cancel-dialog"]');
      expect(cancelDialog.text()).toContain(mockInvitation.user.email);
    });

    it("renders dialog buttons", async () => {
      await openDialog();

      expect(dialog.find('[data-test="cancel-invitation-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    });
  });

  describe("Invitation cancellation", () => {
    it("calls cancelInvitation when confirming", async () => {
      await openDialog();

      const confirmBtn = dialog.find('[data-test="cancel-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(invitationsStore.cancelInvitation).toHaveBeenCalledWith({
        tenant: mockInvitation.namespace.tenant_id,
        user_id: mockInvitation.user.id,
      });
    });

    it("shows success snackbar on successful cancellation", async () => {
      await openDialog();

      const confirmBtn = dialog.find('[data-test="cancel-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Successfully cancelled invitation.");
    });

    it("emits update event on successful cancellation", async () => {
      await openDialog();

      const confirmBtn = dialog.find('[data-test="cancel-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog on successful cancellation", async () => {
      await openDialog();

      const confirmBtn = dialog.find('[data-test="cancel-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const cancelDialogContent = dialog.find('[data-test="invitation-cancel-dialog"] .v-overlay__content');
      expect(cancelDialogContent.attributes("style")).toContain("display: none");
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar when cancellation fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(invitationsStore.cancelInvitation).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="cancel-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to cancel invitation.");
    });

    it("shows error message for 400 status", async () => {
      const error = createAxiosError(400, "Bad Request");
      vi.mocked(invitationsStore.cancelInvitation).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="cancel-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Invalid invitation.");
    });

    it("shows error message for 403 status", async () => {
      const error = createAxiosError(403, "Forbidden");
      vi.mocked(invitationsStore.cancelInvitation).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="cancel-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("You don't have permission to cancel invitations.");
    });

    it("shows error message for 404 status", async () => {
      const error = createAxiosError(404, "Not Found");
      vi.mocked(invitationsStore.cancelInvitation).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="cancel-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Invitation not found.");
    });

    it("shows generic error message for other status codes", async () => {
      const error = createAxiosError(503, "Service Unavailable");
      vi.mocked(invitationsStore.cancelInvitation).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="cancel-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to cancel invitation.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("Dialog close", () => {
    it("closes dialog when cancel button is clicked", async () => {
      await openDialog();

      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      const cancelDialogContent = dialog.find('[data-test="invitation-cancel-dialog"] .v-overlay__content');
      expect(cancelDialogContent.attributes("style")).toContain("display: none");
    });
  });
});
