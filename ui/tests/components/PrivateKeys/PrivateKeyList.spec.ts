import { setActivePinia, createPinia } from "pinia";
import { beforeEach, describe, expect, it } from "vitest";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import PrivateKeyList from "@/components/PrivateKeys/PrivateKeyList.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import usePrivateKeysStore from "@/store/modules/private_keys";

type PrivateKeyListWrapper = VueWrapper<InstanceType<typeof PrivateKeyList>>;

const mockPrivateKeys = [
  { id: 1, name: "test-key-1", data: "private-key-data-1", hasPassphrase: true, fingerprint: "fingerprint-1" },
  { id: 2, name: "test-key-2", data: "private-key-data-2", hasPassphrase: false, fingerprint: "fingerprint-2" },
  { id: 3, name: "test-key-3", data: "private-key-data-3", hasPassphrase: false, fingerprint: "fingerprint-3" },
];

describe("Private Key List", () => {
  let wrapper: PrivateKeyListWrapper;
  setActivePinia(createPinia());
  const privateKeysStore = usePrivateKeysStore();
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(PrivateKeyList, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });

    privateKeysStore.privateKeys = mockPrivateKeys;
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
