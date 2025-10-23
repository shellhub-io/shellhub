import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import PrivateKeyEdit from "@/components/PrivateKeys/PrivateKeyEdit.vue";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import usePrivateKeysStore from "@/store/modules/private_keys";

vi.mock("@/utils/sshKeys", () => ({
  convertToFingerprint: () => "XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX",
  parsePrivateKey: vi.fn(),
}));

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

const mockPrivateKey = {
  id: 1,
  name: "test-name",
  data: "test-data",
  hasPassphrase: false,
  fingerprint: "XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX",
};

describe("Private Key Edit", () => {
  let wrapper: VueWrapper<InstanceType<typeof PrivateKeyEdit>>;
  let privateKeysStore: ReturnType<typeof usePrivateKeysStore>;
  const vuetify = createVuetify();

  beforeEach(() => {
    setActivePinia(createPinia());
    privateKeysStore = usePrivateKeysStore();

    wrapper = mount(PrivateKeyEdit, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
        stubs: {
          "v-file-upload": true,
          "v-file-upload-item": true,
        },
      },
      props: {
        privateKey: mockPrivateKey,
      },
      attachTo: document.body,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component snapshot correctly", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("opens the dialog and initializes the form with existing data when clicked", async () => {
    await wrapper.find('[data-test="privatekey-edit-btn"]').trigger("click");
    await flushPromises();

    const dialog = new DOMWrapper(document.body);

    expect(dialog.find('[data-test="private-key-edit-dialog"]').exists()).toBe(true);

    const nameInput = dialog.find<HTMLInputElement>('[data-test="name-field"] input');
    const dataTextarea = dialog.find<HTMLTextAreaElement>('[data-test="private-key-field"] textarea');

    expect(nameInput.element.value).toBe(mockPrivateKey.name);
    expect(dataTextarea.element.value).toBe(mockPrivateKey.data);
  });

  it("disables the save button if the required name field is empty", async () => {
    await wrapper.find('[data-test="privatekey-edit-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const saveButton = dialog.find('[data-test="pk-edit-save-btn"]');
    const nameInput = dialog.find('[data-test="name-field"] input');

    expect(saveButton.attributes("disabled")).toBeUndefined();

    await nameInput.setValue("");
    await flushPromises();

    expect(saveButton.attributes("disabled")).toBeDefined();
  });

  it("updates the store, emits 'update', and shows a success message on valid submission", async () => {
    privateKeysStore.addPrivateKey(mockPrivateKey); // Ensure the key exists in the store
    const storeSpy = vi.spyOn(privateKeysStore, "editPrivateKey");

    await wrapper.find('[data-test="privatekey-edit-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const newName = "new-private-key-name";
    await dialog.find('[data-test="name-field"] input').setValue(newName);

    await dialog.find('[data-test="pk-edit-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledTimes(1);
    expect(storeSpy).toHaveBeenCalledWith({
      id: mockPrivateKey.id,
      name: newName,
      data: mockPrivateKey.data,
      hasPassphrase: false,
      fingerprint: "XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX:XX",
    });

    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Private key updated successfully.");
    expect(wrapper.emitted()).toHaveProperty("update");
    expect(dialog.findComponent('[data-test="private-key-edit-dialog"]').exists()).toBe(false); // Dialog should close
  });

  it("shows an error message and keeps the dialog open if the update fails", async () => {
    const storeSpy = vi.spyOn(privateKeysStore, "editPrivateKey").mockImplementation(() => {
      throw new Error("API Error");
    });

    await wrapper.find('[data-test="privatekey-edit-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    await dialog.find('[data-test="pk-edit-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledTimes(1);
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update private key.");
    expect(wrapper.emitted().update).toBeUndefined();
    expect(dialog.find('[data-test="private-key-edit-dialog"]').exists()).toBe(true);
  });
});
