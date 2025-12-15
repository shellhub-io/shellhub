import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import SettingPrivateKeys from "@/components/Setting/SettingPrivateKeys.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import usePrivateKeysStore from "@/store/modules/private_keys";

type SettingPrivateKeysWrapper = VueWrapper<InstanceType<typeof SettingPrivateKeys>>;

describe("Setting Private Keys", () => {
  let wrapper: SettingPrivateKeysWrapper;
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

    wrapper = mount(SettingPrivateKeys, {
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
    expect(wrapper.find('[data-test="card"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="card-button"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="private-key-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
  });
});
