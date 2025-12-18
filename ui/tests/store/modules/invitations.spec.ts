import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { namespacesApi } from "@/api/http";
import useInvitationsStore from "@/store/modules/invitations";
import { IInvitation } from "@/interfaces/IInvitation";
import { getInvitationStatusFilter } from "@/utils/invitations";
import { buildUrl } from "../../utils/url";

const encodedPendingInvitationsFilter = encodeURIComponent(getInvitationStatusFilter("pending"));

const mockInvitationBase: IInvitation = {
  status: "pending",
  role: "administrator",
  invited_by: "admin@example.com",
  expires_at: "2026-01-30T10:00:00Z",
  created_at: "2026-01-01T10:00:00Z",
  updated_at: "2026-01-01T10:00:00Z",
  status_updated_at: "2026-01-01T10:00:00Z",
  namespace: {
    tenant_id: "tenant-123",
    name: "production",
  },
  user: {
    id: "user-456",
    email: "user@example.com",
  },
};

describe("Invitations Store", () => {
  let mockNamespacesApi: MockAdapter;
  let store: ReturnType<typeof useInvitationsStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
    store = useInvitationsStore();
  });

  afterEach(() => { mockNamespacesApi.reset(); });

  describe("Initial State", () => {
    it("should have correct default values", () => {
      expect(store.pendingInvitations).toEqual([]);
      expect(store.namespaceInvitations).toEqual([]);
      expect(store.invitationCount).toBe(0);
    });
  });

  describe("fetchUserPendingInvitationList", () => {
    const url = `http://localhost:3000/api/users/invitations?filter=${encodedPendingInvitationsFilter}&page=1&per_page=100`;

    it("should fetch user pending invitations successfully", async () => {
      const mockInvitations = [
        mockInvitationBase,
        {
          ...mockInvitationBase,
          namespace: { tenant_id: "tenant-456", name: "staging" },
          user: { id: "user-789", email: "user2@example.com" },
        },
      ];

      mockNamespacesApi
        .onGet(url)
        .reply(200, mockInvitations);

      await store.fetchUserPendingInvitationList();

      expect(store.pendingInvitations).toEqual(mockInvitations);
    });

    it("should handle empty pending invitations list", async () => {
      mockNamespacesApi
        .onGet(url)
        .reply(200, []);

      await store.fetchUserPendingInvitationList();

      expect(store.pendingInvitations).toEqual([]);
    });

    it("should handle forbidden error when fetching pending invitations", async () => {
      mockNamespacesApi
        .onGet(url)
        .reply(403, { message: "Insufficient permissions" });

      await expect(store.fetchUserPendingInvitationList()).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should handle network error when fetching pending invitations", async () => {
      mockNamespacesApi
        .onGet(url)
        .networkError();

      await expect(store.fetchUserPendingInvitationList()).rejects.toThrow();
    });
  });

  describe("fetchNamespaceInvitationList", () => {
    it("should fetch namespace invitations successfully with pagination", async () => {
      const mockInvitations = [
        mockInvitationBase,
        {
          ...mockInvitationBase,
          role: "observer",
          user: { id: "user-789", email: "observer@example.com" },
        },
      ];

      mockNamespacesApi
        .onGet(buildUrl("http://localhost:3000/api/namespaces/tenant-123/invitations", { page: 1, per_page: 10 }))
        .reply(200, mockInvitations, {
          "x-total-count": "2",
        });

      await store.fetchNamespaceInvitationList("tenant-123", 1, 10);

      expect(store.namespaceInvitations).toEqual(mockInvitations);
      expect(store.invitationCount).toBe(2);
    });

    it("should handle empty namespace invitations list", async () => {
      mockNamespacesApi
        .onGet(buildUrl("http://localhost:3000/api/namespaces/tenant-123/invitations", { page: 1, per_page: 10 }))
        .reply(200, [], {
          "x-total-count": "0",
        });

      await store.fetchNamespaceInvitationList("tenant-123", 1, 10);

      expect(store.namespaceInvitations).toEqual([]);
      expect(store.invitationCount).toBe(0);
    });

    it("should fetch namespace invitations with filter", async () => {
      const mockInvitations = [mockInvitationBase];

      mockNamespacesApi
        .onGet(buildUrl("http://localhost:3000/api/namespaces/tenant-123/invitations", { filter: "test-filter", page: 1, per_page: 10 }))
        .reply(200, mockInvitations, {
          "x-total-count": "1",
        });

      await store.fetchNamespaceInvitationList("tenant-123", 1, 10, "test-filter");

      expect(store.namespaceInvitations).toEqual(mockInvitations);
      expect(store.invitationCount).toBe(1);
    });

    it("should handle not found error when fetching namespace invitations", async () => {
      mockNamespacesApi
        .onGet(buildUrl("http://localhost:3000/api/namespaces/tenant-123/invitations", { page: 1, per_page: 10 }))
        .reply(404, { message: "Namespace not found" });

      await expect(
        store.fetchNamespaceInvitationList("tenant-123", 1, 10),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when fetching namespace invitations", async () => {
      mockNamespacesApi
        .onGet(buildUrl("http://localhost:3000/api/namespaces/tenant-123/invitations", { page: 1, per_page: 10 }))
        .networkError();

      await expect(store.fetchNamespaceInvitationList("tenant-123", 1, 10)).rejects.toThrow();
    });
  });

  describe("acceptInvitation", () => {
    it("should accept invitation successfully", async () => {
      mockNamespacesApi
        .onPatch("http://localhost:3000/api/namespaces/tenant-123/invitations/accept")
        .reply(200);

      await expect(store.acceptInvitation("tenant-123")).resolves.not.toThrow();
    });

    it("should handle not found error when accepting invitation", async () => {
      mockNamespacesApi
        .onPatch("http://localhost:3000/api/namespaces/tenant-123/invitations/accept")
        .reply(404, { message: "Invitation not found" });

      await expect(store.acceptInvitation("tenant-123")).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when accepting invitation", async () => {
      mockNamespacesApi
        .onPatch("http://localhost:3000/api/namespaces/tenant-123/invitations/accept")
        .networkError();

      await expect(store.acceptInvitation("tenant-123")).rejects.toThrow();
    });
  });

  describe("declineInvitation", () => {
    it("should decline invitation successfully", async () => {
      mockNamespacesApi
        .onPatch("http://localhost:3000/api/namespaces/tenant-123/invitations/decline")
        .reply(200);

      await expect(store.declineInvitation("tenant-123")).resolves.not.toThrow();
    });

    it("should handle not found error when declining invitation", async () => {
      mockNamespacesApi
        .onPatch("http://localhost:3000/api/namespaces/tenant-123/invitations/decline")
        .reply(404, { message: "Invitation not found" });

      await expect(store.declineInvitation("tenant-123")).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when declining invitation", async () => {
      mockNamespacesApi
        .onPatch("http://localhost:3000/api/namespaces/tenant-123/invitations/decline")
        .networkError();

      await expect(store.declineInvitation("tenant-123")).rejects.toThrow();
    });
  });

  describe("editInvitation", () => {
    it("should edit invitation role successfully", async () => {
      const editData = {
        tenant: "tenant-123",
        user_id: "user-456",
        role: "observer" as const,
      };

      mockNamespacesApi
        .onPatch("http://localhost:3000/api/namespaces/tenant-123/invitations/user-456")
        .reply(200);

      await expect(store.editInvitation(editData)).resolves.not.toThrow();
    });

    it("should handle not found error when editing invitation", async () => {
      const editData = {
        tenant: "tenant-123",
        user_id: "user-456",
        role: "observer" as const,
      };

      mockNamespacesApi
        .onPatch("http://localhost:3000/api/namespaces/tenant-123/invitations/user-456")
        .reply(404, { message: "Invitation not found" });

      await expect(store.editInvitation(editData)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when editing invitation", async () => {
      const editData = {
        tenant: "tenant-123",
        user_id: "user-456",
        role: "observer" as const,
      };

      mockNamespacesApi
        .onPatch("http://localhost:3000/api/namespaces/tenant-123/invitations/user-456")
        .networkError();

      await expect(store.editInvitation(editData)).rejects.toThrow();
    });
  });

  describe("cancelInvitation", () => {
    it("should cancel invitation successfully", async () => {
      const cancelData = {
        tenant: "tenant-123",
        user_id: "user-456",
      };

      mockNamespacesApi
        .onDelete("http://localhost:3000/api/namespaces/tenant-123/invitations/user-456")
        .reply(200);

      await expect(store.cancelInvitation(cancelData)).resolves.not.toThrow();
    });

    it("should handle not found error when canceling invitation", async () => {
      const cancelData = {
        tenant: "tenant-123",
        user_id: "user-456",
      };

      mockNamespacesApi
        .onDelete("http://localhost:3000/api/namespaces/tenant-123/invitations/user-456")
        .reply(404, { message: "Invitation not found" });

      await expect(store.cancelInvitation(cancelData)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle network error when canceling invitation", async () => {
      const cancelData = {
        tenant: "tenant-123",
        user_id: "user-456",
      };

      mockNamespacesApi
        .onDelete("http://localhost:3000/api/namespaces/tenant-123/invitations/user-456")
        .networkError();

      await expect(store.cancelInvitation(cancelData)).rejects.toThrow();
    });
  });

  describe("sendInvitationEmail", () => {
    it("should send invitation email successfully", async () => {
      const inviteData = {
        email: "newuser@example.com",
        role: "observer" as const,
        tenant_id: "tenant-123",
      };

      mockNamespacesApi
        .onPost("http://localhost:3000/api/namespaces/tenant-123/members")
        .reply(200);

      await expect(store.sendInvitationEmail(inviteData)).resolves.not.toThrow();
    });

    it("should handle validation error when sending invitation email", async () => {
      const inviteData = {
        email: "invalid-email",
        role: "observer" as const,
        tenant_id: "tenant-123",
      };

      mockNamespacesApi
        .onPost("http://localhost:3000/api/namespaces/tenant-123/members")
        .reply(400, { message: "Invalid email address" });

      await expect(store.sendInvitationEmail(inviteData)).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle network error when sending invitation email", async () => {
      const inviteData = {
        email: "newuser@example.com",
        role: "observer" as const,
        tenant_id: "tenant-123",
      };

      mockNamespacesApi
        .onPost("http://localhost:3000/api/namespaces/tenant-123/members")
        .networkError();

      await expect(store.sendInvitationEmail(inviteData)).rejects.toThrow();
    });
  });

  describe("generateInvitationLink", () => {
    it("should generate invitation link successfully", async () => {
      const inviteData = {
        email: "newuser@example.com",
        role: "observer" as const,
        tenant_id: "tenant-123",
      };

      // eslint-disable-next-line vue/max-len
      const mockLink = "http://localhost/accept-invite?email=newuser%40example.com&sig=fake-sig&tenant-id=00000000-0000-4000-0000-000000000000&user-id=fake-user";

      mockNamespacesApi
        .onPost("http://localhost:3000/api/namespaces/tenant-123/invitations/links")
        .reply(200, { link: mockLink });

      const result = await store.generateInvitationLink(inviteData);

      expect(result).toBe(mockLink);
    });

    it("should handle validation error when generating invitation link", async () => {
      const inviteData = {
        email: "invalid-email",
        role: "observer" as const,
        tenant_id: "tenant-123",
      };

      mockNamespacesApi
        .onPost("http://localhost:3000/api/namespaces/tenant-123/invitations/links")
        .reply(400, { message: "Invalid email address" });

      await expect(
        store.generateInvitationLink(inviteData),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle network error when generating invitation link", async () => {
      const inviteData = {
        email: "newuser@example.com",
        role: "observer" as const,
        tenant_id: "tenant-123",
      };

      mockNamespacesApi
        .onPost("http://localhost:3000/api/namespaces/tenant-123/invitations/links")
        .networkError();

      await expect(store.generateInvitationLink(inviteData)).rejects.toThrow();
    });
  });
});
