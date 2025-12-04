import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import InvitationEdit from "@/components/Team/Invitation/InvitationEdit.vue";
import { namespacesApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useInvitationsStore from "@/store/modules/invitations";
import { IInvitation } from "@/interfaces/IInvitation";

type InvitationEditWrapper = VueWrapper<InstanceType<typeof InvitationEdit>>;

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

describe("InvitationEdit", () => {
  let wrapper: InvitationEditWrapper;
  setActivePinia(createPinia());
  const invitationsStore = useInvitationsStore();
  const vuetify = createVuetify();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  beforeEach(() => {
    wrapper = mount(InvitationEdit, {
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

  it("Successfully edits invitation", async () => {
    mockNamespacesApi.onPatch("http://localhost:3000/api/namespaces/fake-tenant/invitations/user123").reply(200);

    const storeSpy = vi.spyOn(invitationsStore, "editInvitation");

    await wrapper.findComponent('[data-test="invitation-edit-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="update-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      tenant: "fake-tenant",
      user_id: "user123",
      role: "operator",
    });

    expect(mockSnackbar.showSuccess).toBeCalledWith("Successfully updated invitation role.");
  });

  it("Fails to edit invitation due to permission error", async () => {
    mockNamespacesApi.onPatch("http://localhost:3000/api/namespaces/fake-tenant/invitations/user123").reply(403);

    const storeSpy = vi.spyOn(invitationsStore, "editInvitation");

    await wrapper.findComponent('[data-test="invitation-edit-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="update-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      tenant: "fake-tenant",
      user_id: "user123",
      role: "operator",
    });

    expect(mockSnackbar.showError).toBeCalledWith("You don't have permission to edit invitations.");
  });
});
