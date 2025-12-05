import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import InvitationCancel from "@/components/Team/Invitation/InvitationCancel.vue";
import { namespacesApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useInvitationsStore from "@/store/modules/invitations";
import { IInvitation } from "@/interfaces/IInvitation";

type InvitationCancelWrapper = VueWrapper<InstanceType<typeof InvitationCancel>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

const invitation: IInvitation = {
  status: "pending",
  role: "operator",
  invited_by: "user1",
  expires_at: "2025-12-31T23:59:59Z",
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

describe("InvitationCancel", () => {
  let wrapper: InvitationCancelWrapper;
  setActivePinia(createPinia());
  const invitationsStore = useInvitationsStore();
  const vuetify = createVuetify();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  beforeEach(() => {
    wrapper = mount(InvitationCancel, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        invitation,
        hasAuthorization: true,
      },
    });
  });

  it("Cancel invitation success", async () => {
    mockNamespacesApi.onDelete("http://localhost:3000/api/namespaces/fake-tenant/invitations/user123").reply(200);

    const storeSpy = vi.spyOn(invitationsStore, "cancelInvitation");

    await wrapper.findComponent('[data-test="invitation-cancel-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="cancel-invitation-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      tenant: "fake-tenant",
      user_id: "user123",
    });

    expect(mockSnackbar.showSuccess).toBeCalledWith("Successfully cancelled invitation.");
  });

  it("Cancel invitation error", async () => {
    mockNamespacesApi.onDelete("http://localhost:3000/api/namespaces/fake-tenant/invitations/user123").reply(404);

    const storeSpy = vi.spyOn(invitationsStore, "cancelInvitation");

    await wrapper.findComponent('[data-test="invitation-cancel-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="cancel-invitation-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      tenant: "fake-tenant",
      user_id: "user123",
    });

    expect(mockSnackbar.showError).toBeCalledWith("Invitation not found.");
  });
});
