import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { mockNamespaces } from "../mocks";
import Namespaces from "@admin/views/Namespaces.vue";

vi.mock("@admin/store/api/namespaces");

describe("Namespaces", () => {
  let wrapper: VueWrapper<InstanceType<typeof Namespaces>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const mountWrapper = async (mockError?: Error) => {
    const router = createCleanAdminRouter();
    await router.push({ name: "namespaces" });
    await router.isReady();

    wrapper = mountComponent(Namespaces, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: {
          adminNamespaces: {
            namespaces: mockError ? [] : mockNamespaces,
            namespaceCount: mockError ? 0 : mockNamespaces.length,
          },
        },
        stubActions: !mockError,
      },
    });

    namespacesStore = useNamespacesStore();
    if (mockError) vi.mocked(namespacesStore.fetchNamespaceList).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when namespaces load successfully", () => {
    beforeEach(() => mountWrapper());

    it("displays the page header with correct title", () => {
      expect(wrapper.text()).toContain("Namespaces");
      expect(wrapper.text()).toContain("Namespace Management");
    });

    it("displays the page header description", () => {
      expect(wrapper.text()).toContain("Track every tenant, search by name, and export namespace data for audits.");
    });

    it("displays the search input field", () => {
      const searchInput = wrapper.find('[data-test="search-input"]');
      expect(searchInput.exists()).toBe(true);
      expect(searchInput.text()).toContain("Search by name"); // Placeholder
    });

    it("displays the export namespaces button", () => {
      expect(wrapper.find('[data-test="namespaces-export-btn"]').exists()).toBe(true);
    });

    it("displays the namespaces list component", () => {
      expect(wrapper.find('[data-test="namespaces-list"]').exists()).toBe(true);
    });
  });

  describe("when searching for namespaces", () => {
    beforeEach(() => mountWrapper());

    it("triggers search on keyup event", async () => {
      const searchInput = wrapper.find('[data-test="search-input"] input');
      await searchInput.setValue("test-namespace");
      await searchInput.trigger("keyup");
      await flushPromises();

      expect(namespacesStore.fetchNamespaceList).toHaveBeenCalled();
    });

    it("encodes filter correctly when searching", async () => {
      const searchInput = wrapper.find('[data-test="search-input"] input');
      await searchInput.setValue("dev");
      await searchInput.trigger("keyup");
      await flushPromises();

      expect(namespacesStore.setFilter).toHaveBeenCalled();
      expect(namespacesStore.fetchNamespaceList).toHaveBeenCalledWith(
        expect.objectContaining({
          filter: expect.any(String),
          page: 1,
        }),
      );
    });

    it("clears filter when search is empty", async () => {
      const searchInput = wrapper.find('[data-test="search-input"] input');
      await searchInput.setValue("");
      await searchInput.trigger("keyup");
      await flushPromises();

      expect(namespacesStore.setFilter).toHaveBeenCalledWith("");
      expect(namespacesStore.fetchNamespaceList).toHaveBeenCalledWith(
        expect.objectContaining({
          filter: "",
          page: 1,
        }),
      );
    });
  });
});
