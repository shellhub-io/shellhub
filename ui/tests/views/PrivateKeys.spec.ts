import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import PrivateKeys from "@/views/PrivateKeys.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import usePrivateKeysStore from "@/store/modules/private_keys";

type PrivateKeysWrapper = VueWrapper<InstanceType<typeof PrivateKeys>>;

describe("Private Keys", () => {
  let wrapper: PrivateKeysWrapper;
  let privateKeysStore: ReturnType<typeof usePrivateKeysStore>;
  const vuetify = createVuetify();

  setActivePinia(createPinia());

  beforeEach(() => {
    privateKeysStore = usePrivateKeysStore();
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

    wrapper = mount(PrivateKeys, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
        stubs: {
          PrivateKeyList: {
            template: "<div data-test='private-key-list' />",
          },
        },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", () => {
    expect(wrapper.find('[data-test="private-keys-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="private-keys-components"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
  });
});
