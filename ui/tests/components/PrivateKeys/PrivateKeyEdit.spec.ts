import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import PrivateKeyEdit from "@/components/PrivateKeys/PrivateKeyEdit.vue";
import { mockPrivateKey } from "@tests/mocks/privateKey";
import usePrivateKeysStore from "@/store/modules/private_keys";
import * as handleErrorModule from "@/utils/handleError";
import { generateKeyPairSync } from "crypto";
import * as sshKeysUtils from "@/utils/sshKeys";

const generatedPrivateKey = generateKeyPairSync("ed25519").privateKey;

// Valid test private key (unencrypted)
const VALID_PRIVATE_KEY = generatedPrivateKey.export({ type: "pkcs8", format: "pem" });

// Encrypted private key
const ENCRYPTED_PRIVATE_KEY = generatedPrivateKey.export({
  type: "pkcs8",
  format: "pem",
  cipher: "aes-256-cbc",
  passphrase: "test-passphrase",
});

describe("PrivateKeyEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof PrivateKeyEdit>>;
  let dialog: DOMWrapper<Element>;
  let privateKeysStore: ReturnType<typeof usePrivateKeysStore>;

  const openEditDialog = async () => {
    const editBtn = wrapper.find('[data-test="private-key-edit-btn"]');
    await editBtn.trigger("click");
    await flushPromises();
  };

  const mountWrapper = (privateKey = mockPrivateKey) => {
    wrapper = mountComponent(PrivateKeyEdit, {
      global: { stubs: ["v-file-upload", "v-file-upload-item"] },
      props: { privateKey },
      attachTo: document.body,
    });
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => {
    mountWrapper();
    privateKeysStore = usePrivateKeysStore();
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Edit button", () => {
    it("Renders edit button", () => {
      const editBtn = wrapper.find('[data-test="private-key-edit-btn"]');
      expect(editBtn.exists()).toBe(true);
    });

    it("Shows edit icon", () => {
      const editBtn = wrapper.find('[data-test="private-key-edit-btn"]');
      const icon = editBtn.find(".mdi-pencil");
      expect(icon.exists()).toBe(true);
    });

    it("Shows 'Edit' text", () => {
      const editBtn = wrapper.find('[data-test="private-key-edit-btn"]');
      expect(editBtn.text()).toBe("Edit");
    });

    it("Opens dialog when clicked", async () => {
      const editBtn = wrapper.find('[data-test="private-key-edit-btn"]');
      await editBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(true);
    });
  });

  describe("Dialog display", () => {
    beforeEach(() => openEditDialog());

    it("Shows FormDialog with correct props", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("title")).toBe("Edit Private Key");
      expect(formDialog.props("icon")).toBe("mdi-key");
      expect(formDialog.props("confirmText")).toBe("Save");
      expect(formDialog.props("cancelText")).toBe("Cancel");
    });

    it("Initializes name field with existing value", () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      expect((nameField.element as HTMLInputElement).value).toBe(mockPrivateKey.name);
    });

    it("Initializes private key field with existing value", () => {
      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      expect(fileText.props("modelValue")).toBe(mockPrivateKey.data);
    });

    it("Shows passphrase field when privateKey has passphrase", async () => {
      wrapper.unmount();
      mountWrapper({ ...mockPrivateKey, hasPassphrase: true });

      await openEditDialog();

      const passphraseField = dialog.find('[data-test="passphrase-field"]');
      expect(passphraseField.exists()).toBe(true);
    });
  });

  describe("Form validation", () => {
    beforeEach(() => openEditDialog());

    it("Shows error when name is empty", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const nameFieldComponent = wrapper.findComponent({ name: "VTextField" });
      expect(nameFieldComponent.props("errorMessages")).toContain("Name is required");
    });

    it("Validates passphrase is required for encrypted keys", async () => {
      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", ENCRYPTED_PRIVATE_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const passphraseField = dialog.find('[data-test="passphrase-field"]');
      expect(passphraseField.find(".v-messages__message").text()).toContain("Passphrase for this private key is required");
    });
  });

  describe("Edit private key", () => {
    beforeEach(() => openEditDialog());

    it("Calls editPrivateKey with updated data", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("updated-key");
      await flushPromises();

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PRIVATE_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(privateKeysStore.editPrivateKey).toHaveBeenCalledWith(
        expect.objectContaining({
          id: mockPrivateKey.id,
          name: "updated-key",
          data: VALID_PRIVATE_KEY,
          hasPassphrase: false,
          fingerprint: sshKeysUtils.convertToFingerprint(VALID_PRIVATE_KEY as string),
        }),
      );
    });

    it("Shows success message after edit", async () => {
      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Private key updated successfully.");
    });

    it("Emits update event after edit", async () => {
      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("Closes dialog after edit", async () => {
      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Error handling", () => {
    beforeEach(() => openEditDialog());

    it("Handles duplicate name error", async () => {
      const error = new Error("name");
      vi.mocked(privateKeysStore.editPrivateKey).mockImplementation(() => {
        throw error;
      });

      const nameField = dialog.find('[data-test="name-field"]');
      await nameField.find("input").setValue("existing-key");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(nameField.find(".v-messages__message").text()).toContain("Name is already used");
    });

    it("Handles duplicate private key error", async () => {
      const error = new Error("private_key");
      vi.mocked(privateKeysStore.editPrivateKey).mockImplementation(() => {
        throw error;
      });

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const keyFieldError = dialog.find('[data-test="private-key-field"] .v-messages__message');
      expect(keyFieldError.text()).toContain("Private key data is already used");
    });

    it("Handles both name and key duplicate error", async () => {
      const error = new Error("both");
      vi.mocked(privateKeysStore.editPrivateKey).mockImplementation(() => {
        throw error;
      });

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const nameFieldError = dialog.find('[data-test="name-field"] .v-messages__message');
      const keyFieldError = dialog.find('[data-test="private-key-field"] .v-messages__message');
      expect(nameFieldError.text()).toContain("Name is already used");
      expect(keyFieldError.text()).toContain("Private key data is already used");
    });

    it("Handles incorrect passphrase error", async () => {
      const handleErrorSpy = vi.spyOn(handleErrorModule, "default").mockImplementation(() => { });

      const encryptedError = new Error("KeyEncryptedError");
      encryptedError.name = "KeyEncryptedError";
      const parseError = new Error("KeyParseError");
      parseError.name = "KeyParseError";

      // First call to parsePrivateKey throws KeyEncryptedError and shows passphrase field
      vi.spyOn(sshKeysUtils, "parsePrivateKey").mockImplementationOnce(() => { throw encryptedError; });

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", ENCRYPTED_PRIVATE_KEY);
      await flushPromises();

      // Second call to parsePrivateKey throws KeyParseError due to wrong passphrase
      vi.spyOn(sshKeysUtils, "parsePrivateKey").mockImplementationOnce(() => { throw parseError; });

      const passphraseField = dialog.find('[data-test="passphrase-field"]');
      await passphraseField.find("input").setValue("wrong-passphrase");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(passphraseField.find(".v-messages__message").text()).toContain("Incorrect passphrase");
      expect(handleErrorSpy).not.toHaveBeenCalled();
    });

    it("Handles generic error", async () => {
      const handleErrorSpy = vi.spyOn(handleErrorModule, "default").mockImplementation(() => { });

      const error = new Error("Unknown error");
      vi.mocked(privateKeysStore.editPrivateKey).mockImplementation(() => {
        throw error;
      });

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update private key.");
      expect(handleErrorSpy).toHaveBeenCalledWith(error);
    });
  });

  describe("Dialog actions", () => {
    beforeEach(() => openEditDialog());

    it("Closes dialog when cancel is clicked", async () => {
      const cancelBtn = dialog.find('[data-test="private-key-cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(false);
    });

    it("Resets form to original values when dialog is reopened", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("changed-name");
      await flushPromises();

      const cancelBtn = dialog.find('[data-test="private-key-cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      await openEditDialog();

      dialog = new DOMWrapper(document.body);
      const nameFieldAfter = dialog.find('[data-test="name-field"] input');
      expect((nameFieldAfter.element as HTMLInputElement).value).toBe(mockPrivateKey.name);
    });
  });
});
