import { setActivePinia, createPinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import PrivateKeyList from "@/components/PrivateKeys/PrivateKeyList.vue";
import { store, key } from "@/store";
import { SnackbarPlugin } from "@/plugins/snackbar";

type PrivateKeyListWrapper = VueWrapper<InstanceType<typeof PrivateKeyList>>;

describe("Private Key List", () => {
  let wrapper: PrivateKeyListWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  const privateKeys = [
    {
      name: "",
      data: "",
      id: 1,
    },
    {
      name: "",
      data: "",
      id: 2,
    },
    {
      name: "",
      data: "",
      id: 3,
    },
  ];

  vi.mock("@/utils/validate", () => ({
    convertToFingerprint: () => "XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX",
  }));

  beforeEach(async () => {
    wrapper = mount(PrivateKeyList, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    expect(wrapper.find('[data-test="no-private-key-warning"]').exists()).toBe(true);
    store.commit("privateKey/fetchPrivateKey", privateKeys);
    await flushPromises();
    expect(wrapper.find('[data-test="privateKey-thead"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privateKey-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privateKey-fingerprint"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privateKey-actions"]').exists()).toBe(true);
  });
});
