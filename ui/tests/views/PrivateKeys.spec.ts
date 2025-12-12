import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import PrivateKeys from "@/views/PrivateKeys.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import usePrivateKeysStore from "@/store/modules/private_keys";

type PrivateKeysWrapper = VueWrapper<InstanceType<typeof PrivateKeys>>;

describe("Private Keys", () => {
  let wrapper: PrivateKeysWrapper;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const privateKeysStore = usePrivateKeysStore();

  const createWrapper = () => mount(PrivateKeys, {
    global: {
      plugins: [vuetify, SnackbarPlugin],
      stubs: {
        "v-file-upload": true,
        "v-file-upload-item": true,
      },
    },
  });

  it("Renders with private keys", () => {
    privateKeysStore.$patch({
      privateKeys: [
        {
          id: 1,
          name: "test-key",
          data: "fake-data",
          hasPassphrase: false,
          fingerprint: "aa:bb:cc:dd",
        },
      ],
    });

    wrapper = createWrapper();

    expect(wrapper.find('[data-test="private-keys-page-header"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="add-private-key-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="private-keys-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
  });

  it("Renders without private keys", () => {
    privateKeysStore.$patch({ privateKeys: [] });

    wrapper = createWrapper();

    expect(wrapper.find('[data-test="private-keys-page-header"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="add-private-key-btn"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="private-keys-list"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-add-private-key-btn"]').exists()).toBe(true);
  });

  it("Opens dialog when button is clicked", async () => {
    privateKeysStore.$patch({ privateKeys: [] });
    wrapper = createWrapper();

    const button = wrapper.find('[data-test="no-items-add-private-key-btn"]');
    await button.trigger("click");

    expect(wrapper.vm.showPrivateKeyAdd).toBe(true);
  });
});
