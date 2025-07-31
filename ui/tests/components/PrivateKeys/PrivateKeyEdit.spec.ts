import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PrivateKeyEdit from "@/components/PrivateKeys/PrivateKeyEdit.vue";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type PrivateKeyEditWrapper = VueWrapper<InstanceType<typeof PrivateKeyEdit>>;

vi.mock("@/utils/validate", () => ({
  createKeyFingerprint: () => "XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX",
  validateKey: () => true,
}));

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Private Key Edit", () => {
  let wrapper: PrivateKeyEditWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  const mockObject = {
    id: 1,
    name: "test-name",
    data: "test-data",
    hasPassphrase: false,
    fingerprint: "XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX",
  };

  beforeEach(async () => {
    wrapper = mount(PrivateKeyEdit, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        privateKey: mockObject,
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
    expect(wrapper.find('[data-test="privatekey-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privatekey-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privatekey-edit-btn"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="privatekey-edit-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="name-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-edit-cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-edit-save-btn"]').exists()).toBe(true);
  });

  it("Checks if the private key data is valid", () => {
    wrapper.vm.initializeFormData();
    const privateKeyData = wrapper.vm.keyLocal;
    expect(privateKeyData).toBeDefined();
  });

  it("Checks if the name field is valid", async () => {
    await flushPromises();
    const nameField = wrapper.vm.name;
    expect(nameField).toBeDefined();
  });

  it("Checks if the update function emits an update event", () => {
    wrapper.vm.update();
    expect(wrapper.emitted().update).toBeTruthy();
  });

  it("Checks if the edit function updates the store on success", async () => {
    const storeSpy = vi.spyOn(store, "dispatch");
    wrapper.vm.initializeFormData();
    const privateKeyPayload = {
      name: wrapper.vm.name,
      data: wrapper.vm.keyLocal,
      id: 1,
      fingerprint: "XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX",
      hasPassphrase: wrapper.vm.hasPassphrase,
    };
    await wrapper.vm.edit();
    expect(storeSpy).toHaveBeenCalledWith("privateKey/edit", privateKeyPayload);
    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Private key updated successfully.");
  });

  it("Checks if the edit function handles error on failure", async () => {
    wrapper.vm.initializeFormData();
    await wrapper.vm.edit();
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update private key.");
  });
});
