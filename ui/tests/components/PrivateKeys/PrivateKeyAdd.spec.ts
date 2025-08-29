import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import PrivateKeyAdd from "@/components/PrivateKeys/PrivateKeyAdd.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type PrivateKeyAddWrapper = VueWrapper<InstanceType<typeof PrivateKeyAdd>>;

describe("Setting Private Keys", () => {
  let wrapper: PrivateKeyAddWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(PrivateKeyAdd, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
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
    wrapper.vm.showDialog = true;
    await flushPromises();
    const dialog = new DOMWrapper(document.body);

    expect(dialog.find('[data-test="card-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="name-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-save-btn"]').exists()).toBe(true);
  });

  it("Sets private key name error message", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("not-working-name");

    await wrapper.findComponent('[data-test="name-field"]').setValue("");

    await flushPromises();

    expect(wrapper.vm.nameError).toEqual("Name is required");
  });

  it("Sets private key data error message", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="private-key-field"]').setValue("not-working-key");

    await wrapper.findComponent('[data-test="private-key-field"]').setValue("");

    await flushPromises();

    expect(wrapper.vm.privateKeyDataError).toEqual("Private key data is required");
  });
});
