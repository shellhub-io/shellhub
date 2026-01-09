import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import TeamMembers from "@/views/TeamMembers.vue";
import { INamespaceMember } from "@/interfaces/INamespace";
import { namespacesApi } from "@/api/http";

describe("TeamMembers", () => {
  let wrapper: VueWrapper<InstanceType<typeof TeamMembers>>;
  let mockNamespacesApi: MockAdapter;
  const tenantId = "fake-tenant-data";

  const members = [
    {
      id: "507f1f77bcf86cd799439011",
      role: "owner" as const,
    },
  ] as INamespaceMember[];

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: tenantId,
    members,
    max_devices: 3,
    devices_count: 3,
    created_at: "",
    billing: null,
    settings: {
      session_record: true,
    },
    devices_accepted_count: 3,
    devices_rejected_count: 0,
    devices_pending_count: 0,
    type: "team" as const,
  };

  const mountWrapper = async (mockError = false) => {
    localStorage.setItem("tenant", tenantId);

    if (mockError) vi.spyOn(console, "error").mockImplementation(() => {});

    wrapper = mountComponent(TeamMembers, {
      piniaOptions: {
        initialState: { namespaces: { currentNamespace: namespaceData } },
        stubActions: !mockError,
      },
    });

    await flushPromises();
  };

  afterEach(() => {
    wrapper?.unmount();
    localStorage.clear();
    mockNamespacesApi?.restore();
    vi.restoreAllMocks();
  });

  describe("successful render", () => {
    beforeEach(async () => { await mountWrapper(); });

    it("renders the page header with correct content", () => {
      const pageHeader = wrapper.find('[data-test="title"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Members");
      expect(pageHeader.text()).toContain("Team Management");
      expect(pageHeader.text()).toContain("Manage team members and their access to this namespace");
    });

    it("displays the member list", () => {
      expect(wrapper.find('[data-test="member-list"]').exists()).toBe(true);
    });
  });

  describe("error handling", () => {
    beforeEach(() => { mockNamespacesApi = new MockAdapter(namespacesApi.getAxios()); });

    describe("403 - forbidden", () => {
      it("displays permission denied snackbar", async () => {
        mockNamespacesApi.onGet(`http://localhost:3000/api/namespaces/${tenantId}`).reply(403);

        await mountWrapper(true);

        expect(mockSnackbar.showError).toHaveBeenCalledWith("You don't have permission to access this resource.");
      });
    });

    describe("500 - internal server error", () => {
      it("displays default error snackbar", async () => {
        mockNamespacesApi.onGet(`http://localhost:3000/api/namespaces/${tenantId}`).reply(500);

        await mountWrapper(true);

        expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load namespaces.");
      });
    });
  });
});
