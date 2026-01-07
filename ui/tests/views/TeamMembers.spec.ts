import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import TeamMembers from "@/views/TeamMembers.vue";
import { namespacesApi } from "@/api/http";
import { mockNamespace } from "./mocks/namespace";
import useNamespacesStore from "@/store/modules/namespaces";

describe("TeamMembers", () => {
  let wrapper: VueWrapper<InstanceType<typeof TeamMembers>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let mockNamespacesApi: MockAdapter;
  const tenantId = mockNamespace.tenant_id;

  const mountWrapper = async (mockError = false) => {
    localStorage.setItem("tenant", tenantId);
    if (mockError) vi.mocked(namespacesStore?.fetchNamespace).mockRejectedValueOnce(mockError);

    wrapper = mountComponent(TeamMembers, {
      piniaOptions: {
        initialState: { namespaces: { currentNamespace: mockNamespace } },
        stubActions: !mockError,
      },
    });

    namespacesStore = useNamespacesStore();

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
