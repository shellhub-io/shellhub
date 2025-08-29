import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import { nextTick } from "vue";
import TerminalLoginForm from "@/components/Terminal/TerminalLoginForm.vue";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import { TerminalAuthMethods } from "@/interfaces/ITerminal";
import usePrivateKeysStore from "@/store/modules/private_keys";

const mockPrivateKeys: Array<IPrivateKey> = [
  { id: 1, name: "test-key-1", data: "private-key-data-1", hasPassphrase: true, fingerprint: "fingerprint-1" },
  { id: 2, name: "test-key-2", data: "private-key-data-2", hasPassphrase: false, fingerprint: "fingerprint-2" },
];

describe("Terminal Login Form", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalLoginForm>>;
  setActivePinia(createPinia());
  const privateKeysStore = usePrivateKeysStore();
  const vuetify = createVuetify();

  beforeEach(async () => {
    privateKeysStore.privateKeys = mockPrivateKeys;

    wrapper = mount(TerminalLoginForm, {
      global: {
        plugins: [vuetify],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the components", async () => {
    expect(wrapper.find('[data-test="username-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="auth-method-select"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="password-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="private-keys-select"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="cancel-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="submit-btn"]').exists()).toBe(true);

    wrapper.vm.authenticationMethod = TerminalAuthMethods.PrivateKey;
    wrapper.vm.togglePassphraseField();
    await nextTick();

    expect(wrapper.find('[data-test="password-field"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="private-keys-select"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="passphrase-field"]').exists()).toBe(true);
  });

  it("toggles password visibility when eye icon is clicked", async () => {
    const passwordField = wrapper.find('[data-test="password-field"]');
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

    await wrapper.find('[data-test="username-field"] input').setValue(username);
    await wrapper.find('[data-test="password-field"] input').setValue(password);
    await wrapper.find("form").trigger("submit");
    await nextTick();

    expect(wrapper.emitted().submit[0]).toEqual([{
      username,
      password,
      authenticationMethod: TerminalAuthMethods.Password,
      privateKey: undefined,
    }]);
  });

  it("emits submit event with correct form data using password", async () => {
    const username = "testuser";
    const password = "";
    const privateKey = mockPrivateKeys[0];

    await wrapper.find('[data-test="username-field"] input').setValue(username);
    wrapper.vm.authenticationMethod = TerminalAuthMethods.PrivateKey;
    wrapper.vm.togglePassphraseField();
    await nextTick();
    await wrapper.find('[data-test="private-keys-select"] input').setValue(privateKey.name);
    await wrapper.find('[data-test="passphrase-field"] input').setValue("testpassphrase");
    await wrapper.find("form").trigger("submit");
    await nextTick();

    expect(wrapper.emitted().submit[0]).toEqual([{
      username,
      password,
      authenticationMethod: TerminalAuthMethods.PrivateKey,
      privateKey: privateKey.data,
      passphrase: "testpassphrase",
    }]);
  });

  it("emits close event when cancel button is clicked", async () => {
    const cancelButton = wrapper.find('[data-test="cancel-btn"]');

    await cancelButton.trigger("click");

    expect(wrapper.emitted()).toHaveProperty("close");
  });

  it("submits form when Enter is pressed", async () => {
    const username = "testuser";
    const password = "testpassword";
    const passwordField = wrapper.find('[data-test="password-field"] input');

    await wrapper.find('[data-test="username-field"] input').setValue(username);
    await passwordField.setValue(password);
    await passwordField.trigger("keydown.enter.prevent");
    await nextTick();

    expect(wrapper.emitted()).toHaveProperty("submit");
  });
});
