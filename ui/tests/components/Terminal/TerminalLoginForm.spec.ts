import { setActivePinia, createPinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import { nextTick } from "vue";
import TerminalLoginForm from "@/components/Terminal/TerminalLoginForm.vue";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import { TerminalAuthMethods } from "@/interfaces/ITerminal";
import usePrivateKeysStore from "@/store/modules/private_keys";
import { SnackbarPlugin } from "@/plugins/snackbar";

const mockPrivateKeys: Array<IPrivateKey> = [
  { id: 1, name: "test-key-1", data: "private-key-data-1", hasPassphrase: true, fingerprint: "fingerprint-1" },
  { id: 2, name: "test-key-2", data: "private-key-data-2", hasPassphrase: false, fingerprint: "fingerprint-2" },
];

describe("Terminal Login Form", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalLoginForm>>;
  let dialog: DOMWrapper<HTMLElement>;
  setActivePinia(createPinia());
  const privateKeysStore = usePrivateKeysStore();
  const vuetify = createVuetify();

  beforeEach(() => {
    privateKeysStore.privateKeys = mockPrivateKeys;

    wrapper = mount(TerminalLoginForm, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: {
        modelValue: true,
      },
    });
    dialog = new DOMWrapper(document.body);
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Renders the components", async () => {
    expect(dialog.find('[data-test="username-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="auth-method-select"]').exists()).toBe(true);
    expect(dialog.find('[data-test="password-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-keys-select"]').exists()).toBe(false);
    expect(dialog.find('[data-test="cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="submit-btn"]').exists()).toBe(true);

    wrapper.vm.authenticationMethod = TerminalAuthMethods.PrivateKey;
    wrapper.vm.togglePassphraseField();
    await nextTick();

    expect(dialog.find('[data-test="private-keys-select"]').exists()).toBe(true);
    expect(dialog.find('[data-test="passphrase-field"]').exists()).toBe(true);
  });

  it("toggles password visibility when eye icon is clicked", async () => {
    const passwordField = dialog.find('[data-test="password-field"]');
    const appendIcon = passwordField.find(".mdi-eye-off");

    await appendIcon.trigger("click");
    await nextTick();

    expect(passwordField.find("input").attributes("type")).toBe("text");

    await passwordField.find(".mdi-eye").trigger("click");
    await nextTick();

    expect(passwordField.find("input").attributes("type")).toBe("password");
  });

  it("emits submit event with correct form data using password", async () => {
    const username = "testuser";
    const password = "testpassword";

    wrapper.vm.username = username;
    wrapper.vm.password = password;
    await nextTick();
    await flushPromises();
    wrapper.vm.submitForm();
    await flushPromises();

    expect(wrapper.emitted("submit")).toBeTruthy();
    expect(wrapper.emitted("submit")?.[0]).toEqual([{
      username,
      password,
      authenticationMethod: TerminalAuthMethods.Password,
      privateKey: undefined,
      passphrase: undefined,
    }]);
  });

  it("emits submit event with correct form data using private key", async () => {
    const username = "testuser";
    const privateKey = mockPrivateKeys[0];

    wrapper.vm.username = username;
    await flushPromises();

    wrapper.vm.authenticationMethod = TerminalAuthMethods.PrivateKey;
    wrapper.vm.togglePassphraseField();
    await nextTick();

    wrapper.vm.selectedPrivateKeyName = privateKey.name;
    await nextTick();

    wrapper.vm.passphrase = "testpassphrase";
    await flushPromises();

    wrapper.vm.submitForm();
    await flushPromises();

    expect(wrapper.emitted("submit")).toBeTruthy();
    expect(wrapper.emitted("submit")?.[0]).toEqual([{
      username,
      password: "",
      authenticationMethod: TerminalAuthMethods.PrivateKey,
      privateKey: privateKey.data,
      passphrase: "testpassphrase",
    }]);
  });
});
