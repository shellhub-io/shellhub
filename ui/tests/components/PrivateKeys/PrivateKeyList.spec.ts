import { setActivePinia, createPinia } from "pinia";
import { beforeEach, describe, expect, it } from "vitest";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import PrivateKeyList from "@/components/PrivateKeys/PrivateKeyList.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import usePrivateKeysStore from "@/store/modules/private_keys";

type PrivateKeyListWrapper = VueWrapper<InstanceType<typeof PrivateKeyList>>;

const mockPrivateKeys = [
  { id: 1, name: "test-key-1", data: "private-key-data-1", hasPassphrase: true, fingerprint: "aa:bb:cc:dd:ee:ff" },
  { id: 2, name: "test-key-2", data: "private-key-data-2", hasPassphrase: false, fingerprint: "11:22:33:44:55:66" },
  { id: 3, name: "test-key-3", data: "private-key-data-3", hasPassphrase: false, fingerprint: "77:88:99:aa:bb:cc" },
];

describe("Private Key List", () => {
  let wrapper: PrivateKeyListWrapper;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const privateKeysStore = usePrivateKeysStore();

  beforeEach(() => {
    wrapper = mount(PrivateKeyList, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
    privateKeysStore.$patch({ privateKeys: mockPrivateKeys });
  });

  it("Renders all private key items", () => {
    const items = wrapper.findAll('[data-test="private-key-item"]');
    expect(items).toHaveLength(mockPrivateKeys.length);
  });

  it("Displays private key names", () => {
    const names = wrapper.findAll('[data-test="private-key-name"]');
    expect(names).toHaveLength(mockPrivateKeys.length);
    expect(names[0].text()).toBe("test-key-1");
    expect(names[1].text()).toBe("test-key-2");
    expect(names[2].text()).toBe("test-key-3");
  });

  it("Displays private key fingerprints", () => {
    const fingerprints = wrapper.findAll('[data-test="private-key-fingerprint"]');
    expect(fingerprints).toHaveLength(mockPrivateKeys.length);
    expect(fingerprints[0].text()).toBe("aa:bb:cc:dd:ee:ff");
    expect(fingerprints[1].text()).toBe("11:22:33:44:55:66");
    expect(fingerprints[2].text()).toBe("77:88:99:aa:bb:cc");
  });

  it("Renders action buttons for each private key", () => {
    const actionButtons = wrapper.findAll('[data-test="private-key-actions"]');
    expect(actionButtons).toHaveLength(mockPrivateKeys.length);
  });
});
