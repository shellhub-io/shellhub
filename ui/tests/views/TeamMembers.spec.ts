import { VueWrapper, flushPromises } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import TeamMembers from "@/views/TeamMembers.vue";
import { mockNamespace } from "../mocks";
import { createAxiosError } from "@tests/utils/axiosError";
import { namespacesApi } from "@/api/http";

describe("TeamMembers", () => {
  let wrapper: VueWrapper<InstanceType<typeof TeamMembers>>;
  const tenantId = mockNamespace.tenant_id;

  const mountWrapper = async (mockError?: Error) => {
    localStorage.setItem("tenant", tenantId);

    if (mockError) {
      vi.spyOn(namespacesApi, "getNamespace").mockRejectedValueOnce(mockError);
    }

    wrapper = mountComponent(TeamMembers, {
      piniaOptions: {
        initialState: { namespaces: { currentNamespace: mockNamespace } },
        stubActions: !mockError,
      },
    });

    await flushPromises();
  };

  afterEach(() => {
    wrapper?.unmount();
    localStorage.clear();
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
    describe("403 - forbidden", () => {
      it("displays permission denied snackbar", async () => {
        await mountWrapper(createAxiosError(403, "Forbidden"));

        expect(mockSnackbar.showError).toHaveBeenCalledWith("You don't have permission to access this resource.");
      });
    });

    describe("500 - internal server error", () => {
      it("displays default error snackbar", async () => {
        await mountWrapper(createAxiosError(500, "Internal Server Error"));

        expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load namespaces.");
      });
    });
  });
});
