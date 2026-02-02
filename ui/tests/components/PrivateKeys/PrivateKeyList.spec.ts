import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { mockPrivateKey, mockPrivateKeys } from "@tests/mocks";
import PrivateKeyList from "@/components/PrivateKeys/PrivateKeyList.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";

describe("PrivateKeyList", () => {
  let wrapper: VueWrapper<InstanceType<typeof PrivateKeyList>>;
  let privateKeysStore: ReturnType<typeof usePrivateKeysStore>;

  const mountWrapper = (privateKeys = mockPrivateKeys) => {
    wrapper = mountComponent(PrivateKeyList, {
      piniaOptions: { initialState: { privateKeys: { privateKeys } } },
    });

    privateKeysStore = usePrivateKeysStore();
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

    it("Renders private key items", () => {
      const items = wrapper.findAll('[data-test="private-key-item"]');
      expect(items.length).toBe(mockPrivateKeys.length);
    });

    it("Displays private key name", () => {
      const nameCell = wrapper.find('[data-test="private-key-name"]');
      expect(nameCell.text()).toBe(mockPrivateKey.name);
    });

    it("Displays private key fingerprint", () => {
      const fingerprintCell = wrapper.find('[data-test="private-key-fingerprint"]');
      expect(fingerprintCell.text()).toBe(mockPrivateKey.fingerprint);
    });

    it("Displays actions column", () => {
      const actionsCell = wrapper.find('[data-test="private-key-actions"]');
      expect(actionsCell.exists()).toBe(true);
    });
  });

  describe("Actions menu", () => {
    it("Renders action button for each private key", () => {
      const actionButtons = wrapper.findAll('[data-test="private-key-actions"] button');
      expect(actionButtons.length).toBe(mockPrivateKeys.length);
    });

    it("Opens actions menu when button is clicked", async () => {
      const actionButton = wrapper.find('[data-test="private-key-actions"] button');
      await actionButton.trigger("click");
      await flushPromises();

      expect(wrapper.findComponent({ name: "PrivateKeyEdit" }).exists()).toBe(true);
      expect(wrapper.findComponent({ name: "PrivateKeyDelete" }).exists()).toBe(true);
    });
  });

  describe("Empty state", () => {
    it("Shows no data when private keys list is empty", () => {
      wrapper.unmount();
      mountWrapper([]);
      const items = wrapper.findAll('[data-test="private-key-item"]');
      expect(items.length).toBe(0);
    });
  });

  describe("Data fetching", () => {
    it("Fetches private keys on mount", () => {
      expect(privateKeysStore.getPrivateKeyList).toHaveBeenCalled();
    });

    it("Refetches when page changes", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      dataTable.vm.$emit("update:page", 2);
      await flushPromises();

      expect(privateKeysStore.getPrivateKeyList).toHaveBeenCalled();
    });

    it("Refetches when items per page changes", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      dataTable.vm.$emit("update:itemsPerPage", 20);
      await flushPromises();

      expect(privateKeysStore.getPrivateKeyList).toHaveBeenCalled();
    });
  });

  describe("Update handling", () => {
    it("Calls getPrivateKeysList when PrivateKeyEdit emits update", async () => {
      const actionButton = wrapper.find('[data-test="private-key-actions"] button');
      await actionButton.trigger("click");
      await flushPromises();

      const editComponent = wrapper.findComponent({ name: "PrivateKeyEdit" });
      editComponent.vm.$emit("update");
      await flushPromises();

      expect(privateKeysStore.getPrivateKeyList).toHaveBeenCalled();
    });

    it("Calls getPrivateKeysList when PrivateKeyDelete emits update", async () => {
      const actionButton = wrapper.find('[data-test="private-key-actions"] button');
      await actionButton.trigger("click");
      await flushPromises();

      const deleteComponent = wrapper.findComponent({ name: "PrivateKeyDelete" });
      deleteComponent.vm.$emit("update");
      await flushPromises();

      expect(privateKeysStore.getPrivateKeyList).toHaveBeenCalled();
    });
  });
});
