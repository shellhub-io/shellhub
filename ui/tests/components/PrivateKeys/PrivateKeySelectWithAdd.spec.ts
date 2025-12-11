import { setActivePinia, createPinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import PrivateKeySelectWithAdd from "@/components/PrivateKeys/PrivateKeySelectWithAdd.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import usePrivateKeysStore from "@/store/modules/private_keys";
import { nextTick } from "vue";

type PrivateKeySelectWithAddWrapper = VueWrapper<InstanceType<typeof PrivateKeySelectWithAdd>>;

const mockPrivateKeys = [
  { id: 1, name: "test-key-1", data: "private-key-data-1", hasPassphrase: true, fingerprint: "fingerprint-1" },
  { id: 2, name: "test-key-2", data: "private-key-data-2", hasPassphrase: false, fingerprint: "fingerprint-2" },
  { id: 3, name: "test-key-3", data: "private-key-data-3", hasPassphrase: false, fingerprint: "fingerprint-3" },
];

describe("Private Key Select With Add", () => {
  let wrapper: PrivateKeySelectWithAddWrapper;
  setActivePinia(createPinia());
  const privateKeysStore = usePrivateKeysStore();
  const vuetify = createVuetify();

  beforeEach(() => {
    privateKeysStore.privateKeys = mockPrivateKeys;

    wrapper = mount(PrivateKeySelectWithAdd, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
        stubs: {
          "v-file-upload": true,
          "v-file-upload-item": true,
        },
      },
      props: { modelValue: "test-key-1" },
    });
  });

  it("Renders the private key select", () => {
    const select = wrapper.find('[data-test="private-keys-select"]');
    expect(select.exists()).toBe(true);
  });

  it("Displays all private keys in the select", () => {
    const select = wrapper.findComponent({ name: "VSelect" });
    expect(select.props("items")).toEqual(["test-key-1", "test-key-2", "test-key-3"]);
  });

  it("Auto-selects newly added key and emits key-added event", async () => {
    const newKey = { id: 4, name: "new-test-key", data: "new-key-data", hasPassphrase: false, fingerprint: "new-fingerprint" };

    const getPrivateKeyListSpy = vi.spyOn(privateKeysStore, "getPrivateKeyList").mockImplementation(() => {
      privateKeysStore.privateKeys = [...mockPrivateKeys, newKey];
    });

    const privateKeyAdd = wrapper.findComponent({ name: "PrivateKeyAdd" });
    await privateKeyAdd.vm.$emit("update");
    await nextTick();
    await flushPromises();

    expect(getPrivateKeyListSpy).toHaveBeenCalled();
    expect(wrapper.emitted("key-added")).toBeTruthy();
    expect(wrapper.vm.selectedPrivateKeyName).toBe("new-test-key");
  });

  it("Handles empty private keys list", async () => {
    privateKeysStore.privateKeys = [];

    await nextTick();

    const select = wrapper.findComponent({ name: "VSelect" });
    expect(select.props("items")).toEqual([]);
  });

  it("Updates model value when selecting a key", async () => {
    const select = wrapper.findComponent({ name: "VSelect" });
    await select.setValue("test-key-2");
    await flushPromises();

    expect(wrapper.emitted("update:modelValue")).toBeTruthy();
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["test-key-2"]);
  });
});
