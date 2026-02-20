import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import TerminalLoginForm from "@/components/Terminal/TerminalLoginForm.vue";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import { TerminalAuthMethods } from "@/interfaces/ITerminal";
import { mountComponent } from "@tests/utils/mount";
import * as sshUtils from "@/utils/sshKeys";

const mockPrivateKeys: Array<IPrivateKey> = [
  {
    id: 1,
    name: "test-key-1",
    data: "private-key-data-1",
    hasPassphrase: true,
    fingerprint: "fingerprint-1",
  },
  {
    id: 2,
    name: "test-key-2",
    data: "private-key-data-2",
    hasPassphrase: false,
    fingerprint: "fingerprint-2",
  },
  {
    id: 3,
    name: "test-key-3",
    data: "private-key-data-3",
    hasPassphrase: true,
    fingerprint: "fingerprint-3",
  },
];

describe("TerminalLoginForm", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalLoginForm>>;
  let dialog: DOMWrapper<Element>;

  const mountWrapper = (sshid?: string) => {
    wrapper = mountComponent(TerminalLoginForm, {
      props: {
        modelValue: true,
        sshid,
      },
      attachTo: document.body,
      piniaOptions: {
        initialState: { privateKeys: { privateKeys: mockPrivateKeys } },
      },
    });

    dialog = new DOMWrapper(document.body);
  };

  const selectPrivateKeyAuth = async (keyName?: string) => {
    const authMethodSelect = wrapper.findComponent({ name: "VSelect" });
    authMethodSelect.vm.$emit("update:modelValue", TerminalAuthMethods.PrivateKey);
    await flushPromises();

    if (keyName) {
      const privateKeySelect = wrapper.findComponent({ name: "PrivateKeySelectWithAdd" });
      privateKeySelect.vm.$emit("update:modelValue", keyName);
      await flushPromises();
    }
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering", () => {
    it("renders FormDialog with correct props", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.exists()).toBe(true);
      expect(formDialog.props("title")).toBe("Connect to Device");
      expect(formDialog.props("icon")).toBe("mdi-console");
      expect(formDialog.props("confirmText")).toBe("Connect");
      expect(formDialog.props("cancelText")).toBe("Cancel");
    });

    it("renders terminal login form", () => {
      expect(dialog.find('[data-test="terminal-login-form"]').exists()).toBe(true);
    });

    it("renders username field", () => {
      const usernameField = dialog.find('[data-test="username-field"]');
      expect(usernameField.exists()).toBe(true);
      expect(usernameField.find("label").text()).toBe("Username");
    });

    it("renders authentication method select", () => {
      const authMethodSelect = dialog.find('[data-test="auth-method-select"]');
      expect(authMethodSelect.exists()).toBe(true);
      expect(authMethodSelect.find("label").text()).toBe("Authentication method");
    });

    it("renders password field by default", () => {
      const passwordField = dialog.find('[data-test="password-field"]');
      expect(passwordField.exists()).toBe(true);
      expect(passwordField.find("label").text()).toBe("Password");
      expect(passwordField.find("input").attributes("type")).toBe("password");
    });

    it("does not render private key select by default", () => {
      expect(dialog.find('[data-test="private-keys-select"]').exists()).toBe(false);
    });

    it("renders submit and cancel buttons", () => {
      expect(dialog.find('[data-test="submit-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="cancel-btn"]').exists()).toBe(true);
    });
  });

  describe("Authentication method switching", () => {
    it("shows password field when Password method is selected", () => {
      expect(dialog.find('[data-test="password-field"]').exists()).toBe(true);
      expect(dialog.find('[data-test="private-keys-select"]').exists()).toBe(false);
    });

    it("shows private key select when Private Key method is selected", async () => {
      await selectPrivateKeyAuth();

      expect(dialog.find('[data-test="private-keys-select"]').exists()).toBe(true);
      expect(dialog.find('[data-test="password-field"]').exists()).toBe(false);
    });

    it("shows passphrase field when private key with passphrase is selected", async () => {
      await selectPrivateKeyAuth("test-key-1");

      expect(dialog.find('[data-test="passphrase-field"]').exists()).toBe(true);
    });

    it("does not show passphrase field for key without passphrase", async () => {
      await selectPrivateKeyAuth("test-key-2");

      expect(dialog.find('[data-test="passphrase-field"]').exists()).toBe(false);
    });
  });

  describe("Password visibility toggle", () => {
    it("toggles password visibility when eye icon is clicked", async () => {
      const passwordField = dialog.find('[data-test="password-field"]');
      expect(passwordField.find("input").attributes("type")).toBe("password");

      const eyeIcon = passwordField.find(".mdi-eye-off");
      await eyeIcon.trigger("click");
      await flushPromises();

      expect(passwordField.find("input").attributes("type")).toBe("text");

      const eyeIconOpen = passwordField.find(".mdi-eye");
      await eyeIconOpen.trigger("click");
      await flushPromises();

      expect(passwordField.find("input").attributes("type")).toBe("password");
    });

    it("toggles passphrase visibility when eye icon is clicked", async () => {
      await selectPrivateKeyAuth("test-key-1");

      const passphraseField = dialog.find('[data-test="passphrase-field"]');
      expect(passphraseField.find("input").attributes("type")).toBe("password");

      const eyeIcon = passphraseField.find(".mdi-eye-off");
      await eyeIcon.trigger("click");
      await flushPromises();

      expect(passphraseField.find("input").attributes("type")).toBe("text");
    });
  });

  describe("Form validation", () => {
    it("disables submit button when form is invalid", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("enables submit button when password form is valid", async () => {
      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const passwordField = dialog.find('[data-test="password-field"]');
      await passwordField.find("input").setValue("testpassword");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(false);
    });

    it("disables submit button when username is missing", async () => {
      const passwordField = dialog.find('[data-test="password-field"]');
      await passwordField.find("input").setValue("testpassword");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("disables submit button when password is missing", async () => {
      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("enables submit button when private key form is valid without passphrase", async () => {
      await selectPrivateKeyAuth("test-key-2");

      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(false);
    });

    it("disables submit button when private key form requires passphrase but it's missing", async () => {
      await selectPrivateKeyAuth("test-key-1");

      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("enables submit button when private key form with passphrase is valid", async () => {
      await selectPrivateKeyAuth("test-key-1");

      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const passphraseField = dialog.find('[data-test="passphrase-field"]');
      await passphraseField.find("input").setValue("testpassphrase");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(false);
    });
  });

  describe("Form submission with password", () => {
    it("emits submit event with correct data for password authentication", async () => {
      vi.spyOn(sshUtils, "isKeyValid").mockReturnValueOnce(true);

      const username = "testuser";
      const password = "testpassword";

      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue(username);
      await flushPromises();

      const passwordField = dialog.find('[data-test="password-field"]');
      await passwordField.find("input").setValue(password);
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(wrapper.emitted("submit")).toBeTruthy();
      expect(wrapper.emitted("submit")?.[0]).toEqual([
        {
          username,
          password,
          authenticationMethod: TerminalAuthMethods.Password,
          privateKey: undefined,
          passphrase: undefined,
        },
      ]);
    });

    it("does not submit form when invalid", async () => {
      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(wrapper.emitted("submit")).toBeFalsy();
    });

    it("does not emit submit when passphrase is invalid", async () => {
      vi.spyOn(sshUtils, "isKeyValid").mockReturnValueOnce(false);

      await selectPrivateKeyAuth("test-key-1");

      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const passphraseField = dialog.find('[data-test="passphrase-field"]');
      await passphraseField.find("input").setValue("wrongpassphrase");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(wrapper.emitted("submit")).toBeFalsy();
      expect(dialog.find('[data-test="passphrase-field"]').text()).toContain("Wrong passphrase");
    });

    it("resets fields after successful submission", async () => {
      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const passwordField = dialog.find('[data-test="password-field"]');
      await passwordField.find("input").setValue("testpassword");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(usernameField.find("input").element.value).toBe("");
      expect(passwordField.find("input").element.value).toBe("");
    });
  });

  describe("Form submission with private key", () => {
    it("emits submit event with correct data for private key without passphrase", async () => {
      const username = "testuser";

      await selectPrivateKeyAuth("test-key-2");

      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue(username);
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(wrapper.emitted("submit")).toBeTruthy();
      expect(wrapper.emitted("submit")?.[0]).toEqual([
        {
          username,
          password: "",
          authenticationMethod: TerminalAuthMethods.PrivateKey,
          privateKey: "private-key-data-2",
          passphrase: undefined,
        },
      ]);
    });

    it("emits submit event with correct data for private key with passphrase", async () => {
      const username = "testuser";
      const passphrase = "testpassphrase";

      await selectPrivateKeyAuth("test-key-1");

      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue(username);
      await flushPromises();

      const passphraseField = dialog.find('[data-test="passphrase-field"]');
      await passphraseField.find("input").setValue(passphrase);
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(wrapper.emitted("submit")).toBeTruthy();
      expect(wrapper.emitted("submit")?.[0]).toEqual([
        {
          username,
          password: "",
          authenticationMethod: TerminalAuthMethods.PrivateKey,
          privateKey: "private-key-data-1",
          passphrase,
        },
      ]);
    });
  });

  describe("Dialog close behavior", () => {
    it("emits close event when cancel is clicked", async () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });

      formDialog.vm.$emit("cancel");
      await flushPromises();

      expect(wrapper.emitted("close")).toBeTruthy();
    });

    it("emits close event when dialog is closed", async () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });

      formDialog.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("close")).toBeTruthy();
    });

    it("resets fields when dialog is closed", async () => {
      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const passwordField = dialog.find('[data-test="password-field"]');
      await passwordField.find("input").setValue("testpassword");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      formDialog.vm.$emit("close");
      await flushPromises();

      expect(usernameField.find("input").element.value).toBe("");
      expect(passwordField.find("input").element.value).toBe("");
    });

    it("resets authentication method to Password on close", async () => {
      await selectPrivateKeyAuth();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      formDialog.vm.$emit("close");
      await flushPromises();

      expect(dialog.find('[data-test="password-field"]').exists()).toBe(true);
      expect(dialog.find('[data-test="private-keys-select"]').exists()).toBe(false);
    });
  });

  describe("SSHID hint", () => {
    it("does not show SSHID hint when sshid prop is not provided", () => {
      expect(dialog.find('[data-test="sshid-hint"]').exists()).toBe(false);
    });

    it("shows SSHID hint when sshid prop is provided", () => {
      wrapper.unmount();
      document.body.innerHTML = "";

      mountWrapper("test-sshid");

      expect(dialog.find('[data-test="sshid-hint"]').exists()).toBe(true);
    });

    it("opens SSHIDHelper when 'Show me how' button is clicked", async () => {
      wrapper.unmount();
      document.body.innerHTML = "";

      mountWrapper("test-sshid");
      const showExamplesBtn = dialog.find('[data-test="show-sshid-examples-btn"]');
      await showExamplesBtn.trigger("click");
      await flushPromises();

      const sshidHelper = wrapper.findComponent({ name: "SSHIDHelper" });
      expect(sshidHelper.props("modelValue")).toBe(true);
      expect(sshidHelper.props("sshid")).toBe("test-sshid");
    });
  });

  describe("Loading state", () => {
    it("disables confirm button when loading", async () => {
      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const passwordField = dialog.find('[data-test="password-field"]');
      await passwordField.find("input").setValue("testpassword");
      await flushPromises();

      await wrapper.setProps({ loading: true });
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmLoading")).toBe(true);
    });
  });

  describe("Keyboard shortcuts", () => {
    it("submits form on Enter key in password field", async () => {
      vi.spyOn(sshUtils, "isKeyValid").mockReturnValueOnce(true);

      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const passwordField = dialog.find('[data-test="password-field"]');
      await passwordField.find("input").setValue("testpassword");
      await flushPromises();

      await passwordField.find("input").trigger("keydown.enter");
      await flushPromises();

      expect(wrapper.emitted("submit")).toBeTruthy();
    });

    it("submits form on Enter key in passphrase field", async () => {
      await selectPrivateKeyAuth("test-key-1");

      const usernameField = dialog.find('[data-test="username-field"]');
      await usernameField.find("input").setValue("testuser");
      await flushPromises();

      const passphraseField = dialog.find('[data-test="passphrase-field"]');
      await passphraseField.find("input").setValue("testpassphrase");
      await flushPromises();

      await passphraseField.find("input").trigger("keydown.enter");
      await flushPromises();

      expect(wrapper.emitted("submit")).toBeTruthy();
    });
  });
});
