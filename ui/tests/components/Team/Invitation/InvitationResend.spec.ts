import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import InvitationResend from "@/components/Team/Invitation/InvitationResend.vue";
import { namespacesApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useInvitationsStore from "@/store/modules/invitations";
import { IInvitation } from "@/interfaces/IInvitation";

type InvitationResendWrapper = VueWrapper<InstanceType<typeof InvitationResend>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

const pendingExpiredInvitation: IInvitation = {
  status: "pending",
  role: "operator",
  invited_by: "user1",
  expires_at: "2020-12-31T23:59:59Z",
  created_at: "2025-12-01T00:00:00Z",
  updated_at: "2025-12-01T00:00:00Z",
  status_updated_at: "2025-12-01T00:00:00Z",
  namespace: {
    tenant_id: "fake-tenant",
    name: "Test Namespace",
  },
  user: {
    id: "user123",
    email: "test@example.com",
  },
};

const cancelledInvitation: IInvitation = {
  ...pendingExpiredInvitation,
  status: "cancelled",
};

describe("InvitationResend", () => {
  let wrapper: InvitationResendWrapper;
  setActivePinia(createPinia());
  const invitationsStore = useInvitationsStore();
  const vuetify = createVuetify();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  beforeEach(() => {
    wrapper = mount(InvitationResend, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        invitation: pendingExpiredInvitation,
        hasAuthorization: true,
      },
    });
  });

  it("Resend invitation success for expired invitation", async () => {
    mockNamespacesApi.onPost("http://localhost:3000/api/namespaces/fake-tenant/members").reply(200);

    const storeSpy = vi.spyOn(invitationsStore, "sendInvitationEmail");

    await wrapper.findComponent('[data-test="invitation-resend-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="resend-invitation-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      tenant_id: "fake-tenant",
      email: "test@example.com",
      role: "operator",
    });

    expect(mockSnackbar.showSuccess).toBeCalledWith("Successfully resent invitation email.");
  });

  it("Resend invitation success for cancelled invitation", async () => {
    await wrapper.setProps({ invitation: cancelledInvitation });
    mockNamespacesApi.onPost("http://localhost:3000/api/namespaces/fake-tenant/members").reply(200);

    const storeSpy = vi.spyOn(invitationsStore, "sendInvitationEmail");

    await wrapper.findComponent('[data-test="invitation-resend-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="resend-invitation-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      tenant_id: "fake-tenant",
      email: "test@example.com",
      role: "operator",
    });

    expect(mockSnackbar.showSuccess).toBeCalledWith("Successfully resent invitation email.");
  });

  it("Resend invitation error", async () => {
    mockNamespacesApi.onPost("http://localhost:3000/api/namespaces/fake-tenant/members").reply(409);

    const storeSpy = vi.spyOn(invitationsStore, "sendInvitationEmail");

    await wrapper.findComponent('[data-test="invitation-resend-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="resend-invitation-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      tenant_id: "fake-tenant",
      email: "test@example.com",
      role: "operator",
    });

    expect(mockSnackbar.showError).toBeCalledWith("This user is already invited or is a member of this namespace.");
  });
});
