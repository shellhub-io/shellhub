import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PrivateKeyDelete from "@/components/PrivateKeys/PrivateKeyDelete.vue";
import { store, key } from "@/store";
import { SnackbarPlugin } from "@/plugins/snackbar";

type PrivateKeyDeleteWrapper = VueWrapper<InstanceType<typeof PrivateKeyDelete>>;

describe("Private Key Delete", () => {
  let wrapper: PrivateKeyDeleteWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(PrivateKeyDelete, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
      props: {
        id: 1,
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
    expect(wrapper.find('[data-test="privatekey-delete-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privatekey-delete-btn-title"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="privatekey-delete-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="privatekey-dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="privatekey-dialog-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="privatekey-close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="privatekey-remove-btn"]').exists()).toBe(true);
  });

  it("Checks if the remove function updates the store on success", async () => {
    const storeSpy = vi.spyOn(store, "dispatch");
    await wrapper.setProps({ id: 1 });
    await wrapper.findComponent('[data-test="privatekey-delete-btn"]').trigger("click");
    await flushPromises();
    await wrapper.findComponent('[data-test="privatekey-remove-btn"]').trigger("click");
    expect(storeSpy).toHaveBeenCalledWith("privateKey/remove", 1);
  });
});
