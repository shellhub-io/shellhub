import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { Router } from "vue-router";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceList from "@admin/components/Namespace/NamespaceList.vue";
import { mockNamespaces } from "../../mocks";

describe("NamespaceList", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceList>>;
  let router: Router;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const mountWrapper = (mockNamespaceCount?: number) => {
    router = createCleanAdminRouter();

    wrapper = mountComponent(NamespaceList, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: {
          adminNamespaces: {
            namespaces: mockNamespaces,
            namespaceCount: mockNamespaceCount ?? mockNamespaces.length,
          },
        },
      },
    });

    namespacesStore = useNamespacesStore();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the data table", () => {
      expect(wrapper.find('[data-test="namespaces-list"]').exists()).toBe(true);
    });

    it("displays namespace names", () => {
      expect(wrapper.text()).toContain(mockNamespaces[0].name);
      expect(wrapper.text()).toContain(mockNamespaces[1].name);
    });

    it("displays namespace tenant IDs", () => {
      expect(wrapper.text()).toContain(mockNamespaces[0].tenant_id);
      expect(wrapper.text()).toContain(mockNamespaces[1].tenant_id);
    });

    it("displays owner links", () => {
      const ownerLinks = wrapper.findAll('[data-test="owner-link"]');
      expect(ownerLinks).toHaveLength(mockNamespaces.length);
    });

    it("displays device counts", () => {
      const firstNamespaceCount = mockNamespaces[0].devices_accepted_count
        + mockNamespaces[0].devices_pending_count
        + mockNamespaces[0].devices_rejected_count;

      expect(wrapper.text()).toContain(firstNamespaceCount.toString());
    });

    it("displays info buttons for each namespace", () => {
      const infoButtons = wrapper.findAll('[data-test="info-button"]');
      expect(infoButtons).toHaveLength(mockNamespaces.length);
    });

    it("displays edit buttons for each namespace", () => {
      const editButtons = wrapper.findAll('[data-test="namespace-edit-dialog-btn"]');
      expect(editButtons).toHaveLength(mockNamespaces.length);
    });

    it("displays delete buttons for each namespace", () => {
      const deleteButtons = wrapper.findAll('[data-test="namespace-delete-dialog-btn"]');
      expect(deleteButtons).toHaveLength(mockNamespaces.length);
    });
  });

  describe("fetching namespaces", () => {
    it("fetches namespaces on mount", () => {
      mountWrapper();

      expect(namespacesStore.fetchNamespaceList).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 10,
          page: 1,
        }),
      );
    });

    it("refetches namespaces when page changes", async () => {
      mountWrapper(11); // Mock total count to 11 to enable pagination

      // Click next page button
      const nextPageBtn = wrapper.find('[data-test="pager-next"]');
      await nextPageBtn.trigger("click");
      await flushPromises();

      expect(namespacesStore.fetchNamespaceList).toHaveBeenCalledWith(
        expect.objectContaining({
          page: 2,
        }),
      );
    });

    it("refetches namespaces when items per page changes", async () => {
      mountWrapper(20);

      // Change items per page via combobox
      const ippCombo = wrapper.find('[data-test="ipp-combo"] input');
      await ippCombo.setValue(20);
      await flushPromises();

      expect(namespacesStore.fetchNamespaceList).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 20,
        }),
      );
    });
  });

  describe("navigating to namespace details", () => {
    it("navigates when clicking info button", async () => {
      mountWrapper();

      const pushSpy = vi.spyOn(router, "push");
      const infoButton = wrapper.findAll('[data-test="info-button"]')[0];

      await infoButton.trigger("click");

      expect(pushSpy).toHaveBeenCalledWith({
        name: "namespaceDetails",
        params: { id: mockNamespaces[0].tenant_id },
      });
    });
  });

  describe("navigating to user details", () => {
    it("navigates when clicking owner link", async () => {
      mountWrapper();

      const pushSpy = vi.spyOn(router, "push");
      const ownerLink = wrapper.findAll('[data-test="owner-link"]')[0];

      await ownerLink.trigger("click");

      expect(pushSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "userDetails",
          params: { id: mockNamespaces[0].owner },
        }),
      );
    });
  });

  describe("opening dialogs", () => {
    beforeEach(() => mountWrapper());

    it("opens edit dialog when clicking edit button", async () => {
      const editButton = wrapper.findAll('[data-test="namespace-edit-dialog-btn"]')[0];
      await editButton.trigger("click");
      await flushPromises();

      // NamespaceEdit component should be rendered with modelValue true
      expect(wrapper.findComponent({ name: "NamespaceEdit" }).exists()).toBe(true);
    });

    it("opens delete dialog when clicking delete button", async () => {
      const deleteButton = wrapper.findAll('[data-test="namespace-delete-dialog-btn"]')[0];
      await deleteButton.trigger("click");
      await flushPromises();

      // NamespaceDelete component should be rendered with modelValue true
      expect(wrapper.findComponent({ name: "NamespaceDelete" }).exists()).toBe(true);
    });
  });

  describe("error handling", () => {
    it("shows error snackbar when fetching namespaces fails", async () => {
      mountWrapper(11);
      vi.mocked(namespacesStore.fetchNamespaceList).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      // Trigger refetch by changing page
      const nextPageBtn = wrapper.find('[data-test="pager-next"]');
      await nextPageBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch namespaces.");
    });
  });
});
