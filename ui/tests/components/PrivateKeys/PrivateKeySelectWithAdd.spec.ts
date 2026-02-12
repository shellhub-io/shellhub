import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { mockPrivateKey, mockPrivateKeys } from "@tests/mocks";
import PrivateKeySelectWithAdd from "@/components/PrivateKeys/PrivateKeySelectWithAdd.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";

describe("PrivateKeySelectWithAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof PrivateKeySelectWithAdd>>;
  let privateKeysStore: ReturnType<typeof usePrivateKeysStore>;

  const mountWrapper = (privateKeys = mockPrivateKeys) => {
    wrapper = mountComponent(PrivateKeySelectWithAdd, {
      global: { stubs: ["v-file-upload", "v-file-upload-item"] },
      props: { modelValue: "" },
      piniaOptions: { initialState: { privateKeys: { privateKeys } } },
      attachTo: document.body,
    });

    privateKeysStore = usePrivateKeysStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Select field", () => {
    it("Renders select component", () => {
      const select = wrapper.find('[data-test="private-keys-select"]');
      expect(select.exists()).toBe(true);
    });

    it("Shows Private Key label", () => {
      const select = wrapper.findComponent({ name: "VSelect" });
      expect(select.props("label")).toBe("Private Key");
    });

    it("Shows hint text", () => {
      const select = wrapper.findComponent({ name: "VSelect" });
      expect(select.props("hint")).toBe("Select a private key file for authentication");
      expect(select.props("persistentHint")).toBe(true);
    });

    it("Displays available private key names", () => {
      const select = wrapper.findComponent({ name: "VSelect" });
      expect(select.props("items")).toEqual([mockPrivateKey.name]);
    });

    it("Updates model value when selection changes", async () => {
      const select = wrapper.findComponent({ name: "VSelect" });
      await select.setValue(mockPrivateKey.name);
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([mockPrivateKey.name]);
    });
  });

  describe("Add New Private Key button", () => {
    let menu: DOMWrapper<HTMLElement>;
    beforeEach(() => {
      const select = wrapper.findComponent({ name: "VSelect" });
      select.vm.menu = true;
      menu = new DOMWrapper(document.body);
    });
    it("Renders Add New Private Key button in select menu", () => {
      const addButton = menu.find('[data-test="add-private-key-btn"]');
      expect(addButton.exists()).toBe(true);
    });

    it("Shows plus icon", () => {
      const addButton = menu.find('[data-test="add-private-key-btn"]');
      const icon = addButton.find(".v-icon");
      expect(icon.classes()).toContain("mdi-plus");
    });

    it("Shows Add New Private Key text", () => {
      const addButton = menu.find('[data-test="add-private-key-btn"]');
      expect(addButton.text()).toContain("Add New Private Key");
    });

    it("Opens PrivateKeyAdd dialog when clicked", async () => {
      const addButton = menu.find('[data-test="add-private-key-btn"]');
      await addButton.trigger("click");
      await flushPromises();

      const privateKeyAdd = wrapper.findComponent({ name: "PrivateKeyAdd" });
      expect(privateKeyAdd.props("modelValue")).toBe(true);
    });
  });

  describe("Handle private key added", () => {
    it("Refreshes private keys list when new key is added", async () => {
      const privateKeyAdd = wrapper.findComponent({ name: "PrivateKeyAdd" });
      privateKeyAdd.vm.$emit("update");
      await flushPromises();

      expect(privateKeysStore.getPrivateKeyList).toHaveBeenCalled();
    });

    it("Selects newest private key after adding", async () => {
      const newKey = { ...mockPrivateKey, id: 2, name: "new-key" };
      vi.mocked(privateKeysStore.getPrivateKeyList).mockImplementation(() => {
        privateKeysStore.privateKeys.push(newKey);
      });

      const privateKeyAdd = wrapper.findComponent({ name: "PrivateKeyAdd" });
      privateKeyAdd.vm.$emit("update");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
    });

    it("Emits key-added event after adding", async () => {
      const newKey = { ...mockPrivateKey, id: 3, name: "another-key" };
      vi.mocked(privateKeysStore.getPrivateKeyList).mockImplementation(() => {
        privateKeysStore.privateKeys.push(newKey);
      });

      const privateKeyAdd = wrapper.findComponent({ name: "PrivateKeyAdd" });
      privateKeyAdd.vm.$emit("update");
      await flushPromises();

      expect(wrapper.emitted("key-added")).toBeTruthy();
    });
  });

  describe("Empty state", () => {
    it("Works with no private keys", () => {
      wrapper.unmount();

      mountWrapper([]);

      const select = wrapper.findComponent({ name: "VSelect" });
      expect(select.props("items")).toEqual([]);
    });
  });
});
