import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MemberInvite from "@/components/Team/Member/MemberInvite.vue";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import useInvitationsStore from "@/store/modules/invitations";
import * as hasPermissionModule from "@/utils/permission";
import { envVariables } from "@/envVariables";
import handleError from "@/utils/handleError";

describe("MemberInvite", () => {
  let wrapper: VueWrapper<InstanceType<typeof MemberInvite>>;
  let invitationsStore: ReturnType<typeof useInvitationsStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = (hasPermission = true) => {
    vi.spyOn(hasPermissionModule, "default").mockReturnValue(hasPermission);

    wrapper = mountComponent(MemberInvite, {
      attachTo: document.body,
      piniaOptions: {
        initialState: { auth: { tenantId: "fake-tenant-data" } },
      },
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

  it("renders the invite button", () => {
    expect(wrapper.find('[data-test="invite-dialog-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="invite-dialog-btn"]').text()).toBe("Invite Member");
  });

  it("disables invite button when user doesn't have permission", () => {
    wrapper.unmount();
    mountWrapper(false);

    const inviteButton = wrapper.find('[data-test="invite-dialog-btn"]');
    expect(inviteButton.attributes("disabled")).toBeDefined();
  });

  it("opens dialog when clicking invite button", async () => {
    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    expect(dialog.text()).toContain("Invite Member");
    expect(dialog.find('[data-test="email-text"] input').exists()).toBe(true);
    expect(dialog.find('[data-test="role-select"]').exists()).toBe(true);
  });

  it("validates email field", async () => {
    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("invalid-email");
    await flushPromises();

    const errorMessage = dialog.find('[data-test="email-text"]').text();
    expect(errorMessage).toBeTruthy();
  });

  it("displays link request checkbox on cloud environment", async () => {
    envVariables.isCloud = true;
    wrapper.unmount();
    mountWrapper();
    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const checkbox = dialog.find('[data-test="link-request-checkbox"]');
    expect(checkbox.exists()).toBe(true);
  });

  it("successfully sends email invitation", async () => {
    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("newuser@example.com");
    await flushPromises();

    const roleSelect = dialog.find('[data-test="role-select"] input');
    await roleSelect.setValue("administrator");
    await flushPromises();

    await dialog.find('[data-test="invite-btn"]').trigger("click");
    await flushPromises();

    expect(invitationsStore.sendInvitationEmail).toHaveBeenCalledWith({
      email: "newuser@example.com",
      tenant_id: "fake-tenant-data",
      role: "administrator",
    });
    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Invitation email sent successfully.");
    expect(wrapper.emitted("update")).toBeTruthy();
  });

  it("successfully generates invitation link", async () => {
    vi.mocked(invitationsStore.generateInvitationLink).mockResolvedValueOnce("https://example.com/invite/token");

    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("newuser@example.com");
    await flushPromises();

    const checkbox = dialog.find('[data-test="link-request-checkbox"] input');
    await checkbox.setValue(true);
    await flushPromises();

    await dialog.find('[data-test="invite-btn"]').trigger("click");
    await flushPromises();

    expect(invitationsStore.generateInvitationLink).toHaveBeenCalledWith({
      email: "newuser@example.com",
      tenant_id: "fake-tenant-data",
      role: "administrator",
    });
    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Invitation link generated successfully.");

    const linkField = dialog.find('[data-test="invitation-link"]');
    expect(linkField.exists()).toBe(true);
  });

  it("displays invitation link in second window after generation", async () => {
    vi.mocked(invitationsStore.generateInvitationLink).mockResolvedValueOnce("https://example.com/invite/token");

    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("newuser@example.com");
    await flushPromises();

    const checkbox = dialog.find('[data-test="link-request-checkbox"] input');
    await checkbox.setValue(true);
    await flushPromises();

    await dialog.find('[data-test="invite-btn"]').trigger("click");
    await flushPromises();

    const linkField = dialog.find('[data-test="invitation-link"] input').element as HTMLInputElement;
    expect(linkField.value).toBe("https://example.com/invite/token");
    expect(dialog.text()).toContain("Share this link with the person you want to invite");
  });

  it("handles 409 error when user is already a member", async () => {
    vi.mocked(invitationsStore.sendInvitationEmail).mockRejectedValueOnce(
      createAxiosError(409, "Conflict"),
    );

    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("existing@example.com");
    await flushPromises();

    await dialog.find('[data-test="invite-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to send invitation.");
    const errorMessage = dialog.find('[data-test="email-text"]').text();
    expect(errorMessage).toContain("This user is already a member of this namespace.");
    expect(handleError).not.toHaveBeenCalled();
  });

  it("handles 404 error when user doesn't exist", async () => {
    vi.mocked(invitationsStore.sendInvitationEmail).mockRejectedValueOnce(
      createAxiosError(404, "Not Found"),
    );

    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("nonexistent@example.com");
    await flushPromises();

    await dialog.find('[data-test="invite-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to send invitation.");
    const errorMessage = dialog.find('[data-test="email-text"]').text();
    expect(errorMessage).toContain("This user does not exist.");
    expect(handleError).not.toHaveBeenCalled();
  });

  it("handles generic error when sending invitation fails", async () => {
    const error = createAxiosError(500, "Internal Server Error");
    vi.mocked(invitationsStore.sendInvitationEmail).mockRejectedValueOnce(error);

    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("newuser@example.com");
    await flushPromises();

    await dialog.find('[data-test="invite-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to send invitation.");
    const errorMessage = dialog.find('[data-test="email-text"]').text();
    expect(errorMessage).toContain("An error occurred while sending the invitation.");
    expect(handleError).toHaveBeenCalledWith(error);
  });

  it("resets fields when closing dialog", async () => {
    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("test@example.com");
    await flushPromises();

    const roleSelect = dialog.find('[data-test="role-select"] input');
    await roleSelect.setValue("operator");
    await flushPromises();

    await dialog.find('[data-test="close-btn"]').trigger("click");
    await flushPromises();

    // Reopen dialog to verify reset
    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailFieldReopened = dialog.find('[data-test="email-text"] input').element as HTMLInputElement;
    expect(emailFieldReopened.value).toBe("");
  });

  it("disables invite button when email or role is missing", async () => {
    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const inviteButton = dialog.find('[data-test="invite-btn"]');
    expect(inviteButton.attributes("disabled")).toBeDefined();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("test@example.com");
    await flushPromises();

    expect(dialog.find('[data-test="invite-btn"]').attributes("disabled")).toBeUndefined();
  });

  it("closes dialog after successful email invitation", async () => {
    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("newuser@example.com");
    await flushPromises();

    await dialog.find('[data-test="invite-btn"]').trigger("click");
    await flushPromises();

    expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
  });

  it("keeps dialog open after generating link", async () => {
    vi.mocked(invitationsStore.generateInvitationLink).mockResolvedValueOnce("https://example.com/invite/token");

    await wrapper.find('[data-test="invite-dialog-btn"]').trigger("click");
    await flushPromises();

    const emailField = dialog.find('[data-test="email-text"] input');
    await emailField.setValue("newuser@example.com");
    await flushPromises();

    const checkbox = dialog.find('[data-test="link-request-checkbox"] input');
    await checkbox.setValue(true);
    await flushPromises();

    await dialog.find('[data-test="invite-btn"]').trigger("click");
    await flushPromises();

    expect(dialog.text()).toContain("Invite Member");
    expect(dialog.find('[data-test="invitation-link"]').exists()).toBe(true);
  });
});
