import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import PrivateKeyAdd from "@/components/PrivateKeys/PrivateKeyAdd.vue";
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

describe("PrivateKeyAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof PrivateKeyAdd>>;
  let dialog: DOMWrapper<Element>;
  let privateKeysStore: ReturnType<typeof usePrivateKeysStore>;

  beforeEach(() => {
    wrapper = mountComponent(PrivateKeyAdd, {
      global: { stubs: ["v-file-upload", "v-file-upload-item"] },
      props: { modelValue: true },
      attachTo: document.body,
    });

    privateKeysStore = usePrivateKeysStore();
    dialog = new DOMWrapper(document.body);
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Dialog display", () => {
    it("Shows FormDialog with correct props", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("title")).toBe("New Private Key");
      expect(formDialog.props("icon")).toBe("mdi-key");
      expect(formDialog.props("confirmText")).toBe("Save");
      expect(formDialog.props("cancelText")).toBe("Cancel");
    });

    it("Shows privacy policy alert", () => {
      const alert = dialog.find('[data-test="privacy-policy-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("ShellHub never stores your private keys");
    });

    it("Renders name field", () => {
      const nameField = dialog.find('[data-test="name-field"]');
      expect(nameField.exists()).toBe(true);
    });

    it("Renders private key field", () => {
      const keyField = dialog.find('[data-test="private-key-field"]');
      expect(keyField.exists()).toBe(true);
    });

    it("Does not show passphrase field by default", () => {
      const passphraseField = dialog.find('[data-test="passphrase-field"]');
      expect(passphraseField.exists()).toBe(false);
    });
  });

  describe("Form validation - Name field", () => {
    it("Shows error when name is empty", async () => {
      const nameField = wrapper.findComponent({ name: "VTextField" });

      await nameField.setValue("name");
      await nameField.setValue("");
      await flushPromises();

      expect(nameField.props("errorMessages")).toContain("Name is required");
    });

    it("Disables confirm button when name is empty", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });
  });

  describe("Form validation - Private key field", () => {
    it("Shows passphrase field when encrypted key is detected", async () => {
      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", ENCRYPTED_PRIVATE_KEY);
      await flushPromises();

      const passphraseField = dialog.find('[data-test="passphrase-field"]');
      expect(passphraseField.exists()).toBe(true);
    });

    it("Validates passphrase is required for encrypted keys", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("test-key");

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

  describe("Create private key", () => {
    it("Calls addPrivateKey when form is valid", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PRIVATE_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(privateKeysStore.addPrivateKey).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "my-key",
          data: VALID_PRIVATE_KEY,
          hasPassphrase: false,
        }),
      );
    });

    it("Shows success message after creation", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PRIVATE_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Private key created successfully.");
    });

    it("Emits update event after creation", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PRIVATE_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("Closes dialog after creation", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PRIVATE_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });
  });

  describe("Error handling", () => {
    it("Handles duplicate name error", async () => {
      const error = new Error("name");
      vi.mocked(privateKeysStore.addPrivateKey).mockImplementation(() => {
        throw error;
      });

      const nameField = dialog.find('[data-test="name-field"]');
      await nameField.find("input").setValue("existing-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PRIVATE_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(nameField.find(".v-messages__message").text()).toContain("Name is already used");
    });

    it("Handles duplicate private key error", async () => {
      const error = new Error("private_key");
      vi.mocked(privateKeysStore.addPrivateKey).mockImplementation(() => {
        throw error;
      });

      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PRIVATE_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const keyFieldError = dialog.find('[data-test="ftc-file-error"]');
      expect(keyFieldError.text()).toContain("Private key data is already used");
    });

    it("Handles both name and key duplicate error", async () => {
      const error = new Error("both");
      vi.mocked(privateKeysStore.addPrivateKey).mockImplementation(() => {
        throw error;
      });

      const nameField = dialog.find('[data-test="name-field"]');
      await nameField.find("input").setValue("existing-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PRIVATE_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const keyFieldError = dialog.find('[data-test="ftc-file-error"]');

      expect(nameField.find(".v-messages__message").text()).toContain("Name is already used");
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

      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

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
      vi.mocked(privateKeysStore.addPrivateKey).mockImplementation(() => {
        throw error;
      });

      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("my-key");

      const fileText = wrapper.findComponent({ name: "FileTextComponent" });
      fileText.vm.$emit("update:modelValue", VALID_PRIVATE_KEY);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="private-key-save-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to create private key.");
      expect(handleErrorSpy).toHaveBeenCalledWith(error);
    });
  });

  describe("Dialog actions", () => {
    it("Closes dialog when cancel is clicked", async () => {
      const cancelBtn = dialog.find('[data-test="private-key-cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[wrapper.emitted("update:modelValue")!.length - 1]).toEqual([false]);
    });

    it("Resets form when dialog is closed", async () => {
      const nameField = dialog.find('[data-test="name-field"] input');
      await nameField.setValue("test");

      const cancelBtn = dialog.find('[data-test="private-key-cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const nameFieldAfter = dialog.find('[data-test="name-field"] input');
      expect((nameFieldAfter.element as HTMLInputElement).value).toBe("");
    });
  });
});
