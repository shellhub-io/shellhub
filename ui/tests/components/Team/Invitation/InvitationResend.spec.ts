import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { mockInvitation } from "@tests/mocks/invitation";
import InvitationResend from "@/components/Team/Invitation/InvitationResend.vue";
import useInvitationsStore from "@/store/modules/invitations";
import moment from "moment";
import handleError from "@/utils/handleError";

describe("InvitationResend", () => {
  let wrapper: VueWrapper<InstanceType<typeof InvitationResend>>;
  let invitationsStore: ReturnType<typeof useInvitationsStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const openDialog = async () => {
    const listItem = wrapper.find('[data-test="invitation-resend-btn"]');
    await listItem.trigger("click");
    await flushPromises();
  };

  const mountWrapper = ({
    invitation = mockInvitation,
    hasAuthorization = true,
  } = {}) => {
    wrapper = mountComponent(InvitationResend, {
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
    it("renders resend list item", () => {
      const listItem = wrapper.find('[data-test="invitation-resend-btn"]');
      expect(listItem.exists()).toBe(true);
    });

    it("renders resend title", () => {
      const title = wrapper.find('[data-test="invitation-resend-title"]');
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Resend");
    });

    it("disables list item when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper({ hasAuthorization: false });

      const listItem = wrapper.find('[data-test="invitation-resend-btn"]');
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Button disabled state", () => {
    it("disables button for non-expired pending invitation", () => {
      const futureDate = moment().add(7, "days").toISOString();
      const pendingInvitation = { ...mockInvitation, status: "pending" as const, expires_at: futureDate };
      mountWrapper({ invitation: pendingInvitation });

      const listItem = wrapper.find('[data-test="invitation-resend-btn"]');
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });

    it("enables button for expired pending invitation", () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitation: expiredInvitation });

      const listItem = wrapper.find('[data-test="invitation-resend-btn"]');
      expect(listItem.classes()).not.toContain("v-list-item--disabled");
    });

    it("enables button for cancelled invitation", () => {
      const cancelledInvitation = { ...mockInvitation, status: "cancelled" as const };
      mountWrapper({ invitation: cancelledInvitation });

      const listItem = wrapper.find('[data-test="invitation-resend-btn"]');
      expect(listItem.classes()).not.toContain("v-list-item--disabled");
    });
  });

  describe("Dialog", () => {
    it("opens dialog when list item is clicked on expired invitation", async () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitation: expiredInvitation });

      await openDialog();

      const resendDialog = dialog.find('[data-test="invitation-resend-dialog"]');
      expect(resendDialog.exists()).toBe(true);
    });

    it("displays invitation email in dialog description", async () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitation: expiredInvitation });

      await openDialog();

      const resendDialog = dialog.find('[data-test="invitation-resend-dialog"]');
      expect(resendDialog.text()).toContain(mockInvitation.user.email);
    });

    it("renders dialog buttons", async () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitation: expiredInvitation });

      await openDialog();

      expect(dialog.find('[data-test="resend-invitation-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    });
  });

  describe("Invitation resend", () => {
    it("calls sendInvitationEmail when confirming", async () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitation: expiredInvitation });

      await openDialog();

      const confirmBtn = dialog.find('[data-test="resend-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(invitationsStore.sendInvitationEmail).toHaveBeenCalledWith({
        tenant_id: mockInvitation.namespace.tenant_id,
        email: mockInvitation.user.email,
        role: mockInvitation.role,
      });
    });

    it("shows success snackbar on successful resend", async () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitation: expiredInvitation });

      await openDialog();

      const confirmBtn = dialog.find('[data-test="resend-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Successfully resent invitation email.");
    });

    it("emits update event on successful resend", async () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitation: expiredInvitation });

      await openDialog();

      const confirmBtn = dialog.find('[data-test="resend-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("closes dialog on successful resend", async () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitation: expiredInvitation });

      await openDialog();

      const confirmBtn = dialog.find('[data-test="resend-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const resendDialogContent = dialog.find('[data-test="invitation-resend-dialog"] .v-overlay__content');
      expect(resendDialogContent.attributes("style")).toContain("display: none");
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar when resend fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };

      mountWrapper({ invitation: expiredInvitation });
      vi.mocked(invitationsStore.sendInvitationEmail).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="resend-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to resend invitation.");
      expect(handleError).toHaveBeenCalledWith(error);
    });

    it("shows error message for 400 status", async () => {
      const error = createAxiosError(400, "Bad Request");
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };

      mountWrapper({ invitation: expiredInvitation });
      vi.mocked(invitationsStore.sendInvitationEmail).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="resend-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Invalid invitation.");
    });

    it("shows error message for 403 status", async () => {
      const error = createAxiosError(403, "Forbidden");
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };

      mountWrapper({ invitation: expiredInvitation });
      vi.mocked(invitationsStore.sendInvitationEmail).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="resend-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("You don't have permission to send invitations.");
    });

    it("shows error message for 404 status", async () => {
      const error = createAxiosError(404, "Not Found");
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };

      mountWrapper({ invitation: expiredInvitation });
      vi.mocked(invitationsStore.sendInvitationEmail).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="resend-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Invitation not found.");
    });

    it("shows error message for 409 status", async () => {
      const error = createAxiosError(409, "Conflict");
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };

      mountWrapper({ invitation: expiredInvitation });
      vi.mocked(invitationsStore.sendInvitationEmail).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="resend-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("This user is already invited or is a member of this namespace.");
    });

    it("shows generic error message for other status codes", async () => {
      const error = createAxiosError(503, "Service Unavailable");
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };

      mountWrapper({ invitation: expiredInvitation });
      vi.mocked(invitationsStore.sendInvitationEmail).mockRejectedValueOnce(error);

      await openDialog();

      const confirmBtn = dialog.find('[data-test="resend-invitation-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to resend invitation.");
    });
  });

  describe("Dialog close", () => {
    it("closes dialog when cancel button is clicked", async () => {
      const pastDate = moment().subtract(1, "day").toISOString();
      const expiredInvitation = { ...mockInvitation, status: "pending" as const, expires_at: pastDate };
      mountWrapper({ invitation: expiredInvitation });

      await openDialog();

      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      const resendDialogContent = dialog.find('[data-test="invitation-resend-dialog"] .v-overlay__content');
      expect(resendDialogContent.attributes("style")).toContain("display: none");
    });
  });
});
