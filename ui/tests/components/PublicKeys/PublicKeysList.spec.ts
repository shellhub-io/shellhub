import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import PublicKeysList from "@/components/PublicKeys/PublicKeysList.vue";
import { mockPublicKey, mockPublicKeys } from "@tests/mocks/publicKey";
import usePublicKeysStore from "@/store/modules/public_keys";
import handleError from "@/utils/handleError";
import { formatAbbreviatedDateTime } from "@/utils/date";
import { createAxiosError } from "@tests/utils/axiosError";

describe("PublicKeysList", () => {
  let wrapper: VueWrapper<InstanceType<typeof PublicKeysList>>;
  let publicKeysStore: ReturnType<typeof usePublicKeysStore>;

  const mountWrapper = (publicKeys = mockPublicKeys) => {
    wrapper = mountComponent(PublicKeysList, {
      piniaOptions: {
        initialState: {
          publicKeys: {
            publicKeys,
            publicKeyCount: publicKeys.length,
          },
        },
      },
    });

    publicKeysStore = usePublicKeysStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Table rendering", () => {
    it("Renders DataTable component", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.exists()).toBe(true);
    });

    it("Displays public key items", () => {
      const items = wrapper.findAll('[data-test="public-key-item"]');
      expect(items).toHaveLength(mockPublicKeys.length);
    });

    it("Displays public key name", () => {
      const name = wrapper.find('[data-test="public-key-name"]');
      expect(name.text()).toBe(mockPublicKey.name);
    });

    it("Displays public key fingerprint", () => {
      const fingerprint = wrapper.find('[data-test="public-key-fingerprint"]');
      expect(fingerprint.text()).toBe(mockPublicKey.fingerprint);
    });

    it("Displays public key filter", () => {
      const filter = wrapper.find('[data-test="public-key-filter"]');
      expect(filter.text()).toBe("All devices");
    });

    it("Displays public key username", () => {
      const username = wrapper.find('[data-test="public-key-username"]');
      expect(username.text()).toBe("All users");
    });

    it("Displays public key created_at date", () => {
      const createdAt = wrapper.find('[data-test="public-key-created-at"]');
      expect(createdAt.text()).toBe(formatAbbreviatedDateTime(mockPublicKey.created_at));
    });

    it("Displays actions column", () => {
      const actions = wrapper.find('[data-test="public-key-actions"]');
      expect(actions.exists()).toBe(true);
    });
  });

  describe("Actions menu", () => {
    it("Renders action button", () => {
      const actionBtn = wrapper.find('[data-test="public-key-actions"]');
      expect(actionBtn.exists()).toBe(true);
    });

    it("Shows edit component in menu", () => {
      const editComponent = wrapper.findComponent({ name: "PublicKeyEdit" });
      expect(editComponent.exists()).toBe(true);
    });

    it("Shows delete component in menu", () => {
      const deleteComponent = wrapper.findComponent({ name: "PublicKeyDelete" });
      expect(deleteComponent.exists()).toBe(true);
    });

    it("Passes publicKey to edit component", () => {
      const editComponent = wrapper.findComponent({ name: "PublicKeyEdit" });
      expect(editComponent.props("publicKey")).toEqual(mockPublicKey);
    });

    it("Passes fingerprint to delete component", () => {
      const deleteComponent = wrapper.findComponent({ name: "PublicKeyDelete" });
      expect(deleteComponent.props("fingerprint")).toBe(mockPublicKey.fingerprint);
    });
  });

  describe("Empty state", () => {
    it("Shows no items when publicKeys is empty", () => {
      wrapper.unmount();
      mountWrapper([]);

      const items = wrapper.findAll('[data-test="public-key-item"]');
      expect(items).toHaveLength(0);
    });
  });

  describe("Data fetching", () => {
    it("Refetches when page changes", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      dataTable.vm.$emit("update:page", 2);
      await flushPromises();

      expect(publicKeysStore.fetchPublicKeyList).toHaveBeenCalledWith({
        page: 2,
        perPage: 10,
      });
    });

    it("Refetches when itemsPerPage changes", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      dataTable.vm.$emit("update:itemsPerPage", 20);
      await flushPromises();

      expect(publicKeysStore.fetchPublicKeyList).toHaveBeenCalledWith({
        page: 1,
        perPage: 20,
      });
    });

    it("Handles fetch error", async () => {
      wrapper.unmount();
      mountWrapper([]);
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(publicKeysStore.fetchPublicKeyList).mockRejectedValueOnce(error);

      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load public keys.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("Update handling", () => {
    it("Refreshes list when edit component emits update", async () => {
      const editComponent = wrapper.findComponent({ name: "PublicKeyEdit" });
      editComponent.vm.$emit("update");
      await flushPromises();

      expect(publicKeysStore.fetchPublicKeyList).toHaveBeenCalled();
    });

    it("Refreshes list when delete component emits update", async () => {
      const deleteComponent = wrapper.findComponent({ name: "PublicKeyDelete" });
      deleteComponent.vm.$emit("update");
      await flushPromises();

      expect(publicKeysStore.fetchPublicKeyList).toHaveBeenCalled();
    });
  });
});
