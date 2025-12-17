import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import InvitationList from "@/components/Team/Invitation/InvitationList.vue";
import { namespacesApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import useInvitationsStore from "@/store/modules/invitations";
import { IInvitation } from "@/interfaces/IInvitation";
import { formatShortDateTime } from "@/utils/date";
import { nextTick } from "vue";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

const mockInvitations: IInvitation[] = [
  {
    status: "pending",
    role: "operator",
    invited_by: "admin",
    expires_at: "2025-12-31T23:59:59Z",
    created_at: "2025-12-01T00:00:00Z",
    updated_at: "2025-12-01T00:00:00Z",
    status_updated_at: "2025-12-01T00:00:00Z",
    namespace: {
      tenant_id: "fake-tenant",
      name: "Test Namespace",
    },
    user: {
      id: "user1",
      email: "user1@example.com",
    },
  },
];

// eslint-disable-next-line vue/max-len
const mockInvitationsUrl = "http://localhost:3000/api/namespaces/fake-tenant/invitations?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJzdGF0dXMiLCJvcGVyYXRvciI6ImVxIiwidmFsdWUiOiJwZW5kaW5nIn19XQ%3D%3D&page=1&per_page=10";

const vuetify = createVuetify();

const mountWrapper = () => mount(InvitationList, {
  global: {
    plugins: [vuetify],
    provide: { [SnackbarInjectionKey]: mockSnackbar },
  },
});

describe("InvitationList", async () => {
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const invitationsStore = useInvitationsStore();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  mockNamespacesApi.onGet(mockInvitationsUrl).reply(200, mockInvitations, {
    "x-total-count": "1",
  });
  invitationsStore.namespaceInvitations = mockInvitations;
  invitationsStore.invitationCount = 1;
  authStore.tenantId = "fake-tenant";

  const storeSpy = vi.spyOn(invitationsStore, "fetchNamespaceInvitationList");

  const wrapper = mountWrapper();

  await flushPromises();

  it("Loads invitations on mount", () => {
    expect(storeSpy).toHaveBeenCalled();
  });

  it("Renders the status filter select", () => {
    const statusSelect = wrapper.find('[data-test="invitation-status-select"]');
    expect(statusSelect.exists()).toBe(true);
  });

  it("Shows 'Expires At' header for pending status filter by default", () => {
    expect(wrapper.text()).toContain("Expires At");
  });

  it("Shows 'Cancelled At' header for cancelled status filter", async () => {
    const statusSelect = wrapper.findComponent('[data-test="invitation-status-select"]');
    await statusSelect.setValue("cancelled");
    await flushPromises();
    expect(wrapper.text()).toContain("Cancelled At");
  });

  it("Shows 'Accepted At' header for accepted status filter", async () => {
    const statusSelect = wrapper.findComponent('[data-test="invitation-status-select"]');
    await statusSelect.setValue("accepted");
    await flushPromises();
    expect(wrapper.text()).toContain("Accepted At");
  });

  it("Shows 'Rejected At' header for rejected status filter", async () => {
    const statusSelect = wrapper.findComponent('[data-test="invitation-status-select"]');
    await statusSelect.setValue("rejected");
    await flushPromises();
    expect(wrapper.text()).toContain("Rejected At");
  });

  it("Shows expiration date for pending invitations", () => {
    const dateCell = wrapper.find('[data-test="invitation-date-cell"]');
    const formattedExpiresAt = formatShortDateTime(mockInvitations[0].expires_at); // December 31, 2025 11:59 PM
    expect(dateCell.text()).toContain(formattedExpiresAt);
  });

  it("Shows expired warning for expired pending invitations", async () => {
    const expiredInvitation: IInvitation = {
      ...mockInvitations[0],
      expires_at: "2020-01-01T00:00:00Z",
    };
    const formattedExpiresAt = formatShortDateTime(expiredInvitation.expires_at); // January 1, 2020 00:00 AM
    const expiredWrapper = mountWrapper();
    await flushPromises();
    invitationsStore.namespaceInvitations = [expiredInvitation];
    await nextTick();
    const dateCell = expiredWrapper.find('[data-test="invitation-date-cell"]');
    expect(dateCell.text()).toContain(formattedExpiresAt);
    expect(dateCell.html()).toContain("mdi-alert-circle");
    expect(dateCell.html()).toContain("text-error");
  });

  it("Shows status_updated_at for non-pending invitations", async () => {
    const cancelledInvitation: IInvitation = {
      ...mockInvitations[0],
      status: "cancelled",
      status_updated_at: "2025-12-15T10:30:00Z",
    };
    const formattedStatusUpdatedAt = formatShortDateTime(cancelledInvitation.status_updated_at); // December 15, 2025 10:30 AM

    const cancelledWrapper = mountWrapper();
    await flushPromises();

    const statusSelect = cancelledWrapper.findComponent('[data-test="invitation-status-select"]');
    await statusSelect.setValue("cancelled");
    await flushPromises();

    invitationsStore.namespaceInvitations = [cancelledInvitation];
    await nextTick();

    const dateCell = cancelledWrapper.find('[data-test="invitation-date-cell"]');
    expect(dateCell.text()).toContain(formattedStatusUpdatedAt);
  });

  it("Can change status filter and it updates the invitations", async () => {
    const statusSelect = wrapper.findComponent('[data-test="invitation-status-select"]');

    storeSpy.mockClear();
    await statusSelect.setValue("accepted");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalled();
  });
});
