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

  beforeEach(() => {
    wrapper = mount(PrivateKeyAdd, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
        stubs: {
          "v-file-upload": true,
          "v-file-upload-item": true,
        },
      },
      props: { modelValue: true },
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

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    expect(formDialog.exists()).toBe(true);
    expect(formDialog.props("title")).toBe("New Private Key");
    expect(formDialog.props("icon")).toBe("mdi-key");
    expect(formDialog.props("confirmText")).toBe("Save");
    expect(formDialog.props("cancelText")).toBe("Cancel");

    const dialog = new DOMWrapper(document.body);
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

    expect(wrapper.vm.privateKeyDataError).toEqual("Invalid private key data");
  });
});
