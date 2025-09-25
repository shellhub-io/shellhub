import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PublicKeyAdd from "@/components/PublicKeys/PublicKeyAdd.vue";
import { sshApi, tagsApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import usePublicKeysStore from "@/store/modules/public_keys";

type PublicKeyAddWrapper = VueWrapper<InstanceType<typeof PublicKeyAdd>>;

describe("Public Key Add", () => {
  let wrapper: PublicKeyAddWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const publicKeysStore = usePublicKeysStore();
  const vuetify = createVuetify();

  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  const mockSshApi = new MockAdapter(sshApi.getAxios());

  beforeEach(async () => {
    document.body.innerHTML = "";

    localStorage.setItem("tenant", "fake-tenant-data");
    mockTagsApi.resetHandlers();
    mockTagsApi.onGet(/\/api\/namespaces\/fake-tenant-data\/tags.*/).reply(200, [
      { name: "1" }, { name: "2" }, { name: "3" }, { name: "4" },
    ]);
    mockSshApi.resetHandlers();

    authStore.role = "owner";
    wrapper = mount(PublicKeyAdd, {
      global: { plugins: [vuetify, router, SnackbarPlugin] },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    expect(wrapper.find('[data-test="public-key-add-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    expect(formDialog.exists()).toBe(true);
    expect(formDialog.props("title")).toBe("New Public Key");
    expect(formDialog.props("icon")).toBe("mdi-key-outline");
    expect(formDialog.props("confirmText")).toBe("Save");
    expect(formDialog.props("cancelText")).toBe("Cancel");

    expect(dialog.find('[data-test="name-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="username-restriction-field"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await flushPromises();
    expect(dialog.find('[data-test="rule-field"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();
    expect(dialog.find('[data-test="filter-restriction-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await flushPromises();
    expect(dialog.find('[data-test="hostname-field"]').exists()).toBe(true);

    expect(dialog.find('[data-test="data-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-add-cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-add-save-btn"]').exists()).toBe(true);
  });

  it("Conditional rendering: username + tags shows proper inputs", async () => {
    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();

    expect(dialog.find('[data-test="rule-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(true);
    expect(dialog.find('[data-test="hostname-field"]').exists()).toBe(false);
  });

  it("Conditional rendering: hostname filter shows hostname field only", async () => {
    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await flushPromises();

    expect(dialog.find('[data-test="hostname-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(false);
  });

  it("Allows adding a public key (default: all devices and any username)", async () => {
    mockSshApi.onPost("http://localhost:3000/api/sshkeys/public-keys").reply(200);
    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("my new public key");
    await wrapper.findComponent('[data-test="data-field"]').setValue("fakeish key");
    await flushPromises();

    await wrapper.findComponent('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      data: Buffer.from("fakeish key", "utf-8").toString("base64"),
      filter: { hostname: ".*" },
      name: "my new public key",
      username: ".*",
    });
  });

  it("Saves with hostname restriction", async () => {
    mockSshApi.onPost("http://localhost:3000/api/sshkeys/public-keys").reply(200);
    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("host key");
    await wrapper.findComponent('[data-test="data-field"]').setValue("ssh-rsa AAAAB3Nza...");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("web-.*");
    await flushPromises();

    await wrapper.findComponent('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      data: btoa("ssh-rsa AAAAB3Nza..."),
      filter: { hostname: "web-.*" },
      name: "host key",
      username: ".*",
    });
  });

  it("Saves with tags restriction (up to 3 tags)", async () => {
    mockSshApi.onPost("http://localhost:3000/api/sshkeys/public-keys").reply(200);
    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("tags key");
    await wrapper.findComponent('[data-test="data-field"]').setValue("ssh-ed25519 AAAAC3Nza...");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();

    await wrapper.findComponent('[data-test="tags-selector"]').setValue(["1", "2"]);
    await flushPromises();

    await wrapper.findComponent('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      data: btoa("ssh-ed25519 AAAAC3Nza..."),
      filter: { tags: ["1", "2"] },
      name: "tags key",
      username: ".*",
    });
  });

  it("Blocks selecting more than 3 tags and shows error", async () => {
    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();

    await wrapper.findComponent('[data-test="tags-selector"]').setValue(["1", "2", "3", "4"]);
    await flushPromises();

    expect(wrapper.vm.errMsg).toBe("The maximum capacity has reached");

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    expect(formDialog.props("confirmDisabled")).toBe(true);
  });

  it("Blocks save when username restriction has empty username", async () => {
    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("need user");
    await wrapper.findComponent('[data-test="data-field"]').setValue("ssh-ed25519 AAAA...");
    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await flushPromises();

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    expect(formDialog.props("confirmDisabled")).toBe(true);

    await wrapper.findComponent('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).not.toHaveBeenCalled();
  });

  it("Blocks save when hostname filter has empty hostname", async () => {
    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("need host");
    await wrapper.findComponent('[data-test="data-field"]').setValue("ssh-ed25519 BBBB...");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await flushPromises();

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    expect(formDialog.props("confirmDisabled")).toBe(true);

    await wrapper.findComponent('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).not.toHaveBeenCalled();
  });

  it("Displays error message if name is not provided", async () => {
    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="name-field"]').setValue("foo");
    await wrapper.findComponent('[data-test="name-field"]').setValue("");
    await flushPromises();

    expect(wrapper.vm.nameError).toBeTruthy();
  });

  it("Displays error message if username is not provided when restriction is active", async () => {
    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await wrapper.findComponent('[data-test="rule-field"]').setValue("foo");
    await wrapper.findComponent('[data-test="rule-field"]').setValue("");
    await flushPromises();

    expect(wrapper.vm.usernameError).toBeTruthy();
  });

  it("Displays error message if hostname is not provided when filter=hostname", async () => {
    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("foo");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("");
    await flushPromises();

    expect(wrapper.vm.hostnameError).toBeTruthy();
  });
});
