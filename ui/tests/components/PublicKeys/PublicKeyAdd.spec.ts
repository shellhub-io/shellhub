import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PublicKeyAdd from "@/components/PublicKeys/PublicKeyAdd.vue";
import FileTextComponent from "@/components/Fields/FileTextComponent.vue";
import { sshApi, tagsApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import usePublicKeysStore from "@/store/modules/public_keys";

vi.mock("@/utils/sshKeys", () => ({
  isKeyValid: () => true,
  convertToFingerprint: () => "MOCK:FINGERPRINT",
}));

type PublicKeyAddWrapper = VueWrapper<InstanceType<typeof PublicKeyAdd>>;

describe("Public Key Add", () => {
  let wrapper: PublicKeyAddWrapper;
  let authStore: ReturnType<typeof useAuthStore>;
  let publicKeysStore: ReturnType<typeof usePublicKeysStore>;

  const vuetify = createVuetify();
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  const mockSshApi = new MockAdapter(sshApi.getAxios());

  const pasteIntoFTC = async (text: string) => {
    const ftc = wrapper.findComponent(FileTextComponent);
    await ftc.trigger("paste", {
      clipboardData: {
        getData: (type: string) => (type === "text/plain" ? text : ""),
        files: [],
      },
    });
  };

  const getTextareaValue = () => {
    const dialog = new DOMWrapper(document.body);
    const host = dialog.find('[data-test="ftc-textarea"]');
    expect(host.exists()).toBe(true);
    const ta = host.find("textarea");
    expect(ta.exists()).toBe(true);
    return (ta.element as HTMLTextAreaElement).value;
  };

  beforeEach(() => {
    setActivePinia(createPinia());
    authStore = useAuthStore();
    publicKeysStore = usePublicKeysStore();

    mockTagsApi.onGet(/api\/namespaces\/.*\/tags/).reply(200, [
      { name: "tag1" }, { name: "tag2" }, { name: "tag3" }, { name: "tag4" },
    ]);
    mockSshApi.onPost(/api\/sshkeys\/public-keys/).reply(200);

    localStorage.setItem("tenant", "fake-tenant-data");
    authStore.role = "owner";

    wrapper = mount(PublicKeyAdd, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
        stubs: {
          "v-file-upload": true,
          "v-file-upload-item": true,
        },
      },
      attachTo: document.body,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
    mockTagsApi.reset();
    mockSshApi.reset();
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders all dialog components when opened", async () => {
    await wrapper.find('[data-test="public-key-add-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    expect(dialog.find('[data-test="public-key-add-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="name-field"]').exists()).toBe(true);
    expect(wrapper.findComponent({ name: "FileTextComponent" }).exists()).toBe(true);
  });

  it("Allows adding a public key with default settings", async () => {
    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");

    await wrapper.find('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("my new public key");

    await pasteIntoFTC("fakeish key");
    expect(getTextareaValue()).toBe("fakeish key");

    const dialog = new DOMWrapper(document.body);
    await dialog.find('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      name: "my new public key",
      data: Buffer.from("fakeish key", "utf-8").toString("base64"),
      username: ".*",
      filter: { hostname: ".*" },
    });
  });

  it("Saves with hostname restriction", async () => {
    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");

    await wrapper.find('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("host key");

    await pasteIntoFTC("ssh-rsa AAAAB3Nza...");

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await flushPromises();
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("web-.*");

    const dialog = new DOMWrapper(document.body);
    await dialog.find('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      name: "host key",
      data: Buffer.from("ssh-rsa AAAAB3Nza...", "utf-8").toString("base64"),
      username: ".*",
      filter: { hostname: "web-.*" },
    });
  });

  it("Saves with tags restriction", async () => {
    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");

    await wrapper.find('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("tags key");

    await pasteIntoFTC("ssh-ed25519 AAAAC3Nza...");

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();
    await wrapper.findComponent({ name: "VAutocomplete" }).setValue(["tag1", "tag2"]);
    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    await dialog.find('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      name: "tags key",
      data: Buffer.from("ssh-ed25519 AAAAC3Nza...", "utf-8").toString("base64"),
      username: ".*",
      filter: { tags: ["tag1", "tag2"] },
    });
  });

  it("Blocks save when username restriction has empty username", async () => {
    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");

    await wrapper.find('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("need user");

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await flushPromises();

    const saveButton = wrapper.findComponent('[data-test="pk-add-save-btn"]');
    expect(saveButton.attributes("disabled")).toBeDefined();
    await saveButton.trigger("click");
    expect(storeSpy).not.toHaveBeenCalled();
  });

  it("Conditional rendering: username + tags shows proper inputs", async () => {
    await wrapper.find('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();

    expect(wrapper.findComponent('[data-test="rule-field"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="tags-selector"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="hostname-field"]').exists()).toBe(false);
  });

  it("Conditional rendering: hostname filter shows hostname field only", async () => {
    await wrapper.find('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await flushPromises();

    expect(wrapper.findComponent('[data-test="hostname-field"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="tags-selector"]').exists()).toBe(false);
  });

  it("Displays error message if name is not provided", async () => {
    await wrapper.find('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    const nameInput = wrapper.findComponent('[data-test="name-field"]');
    await nameInput.setValue("foo");
    await nameInput.setValue("");
    await flushPromises();

    expect(wrapper.vm.nameError).toBeTruthy();
  });

  it("Displays error message if username is not provided when restriction is active", async () => {
    await wrapper.find('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await flushPromises();
    const ruleInput = wrapper.findComponent('[data-test="rule-field"]');
    await ruleInput.setValue("foo");
    await ruleInput.setValue("");
    await flushPromises();

    expect(wrapper.vm.usernameError).toBeTruthy();
  });

  it("Displays error message if hostname is not provided when filter=hostname", async () => {
    await wrapper.find('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await flushPromises();
    const hostnameInput = wrapper.findComponent('[data-test="hostname-field"]');
    await hostnameInput.setValue("foo");
    await hostnameInput.setValue("");
    await flushPromises();

    expect(wrapper.vm.hostnameError).toBeTruthy();
  });
});
