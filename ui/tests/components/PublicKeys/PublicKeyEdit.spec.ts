import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PublicKeyEdit from "@/components/PublicKeys/PublicKeyEdit.vue";
import { sshApi, tagsApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import usePublicKeysStore from "@/store/modules/public_keys";

type PublicKeyEditWrapper = VueWrapper<InstanceType<typeof PublicKeyEdit>>;

const toB64 = (s: string) => Buffer.from(s, "utf-8").toString("base64");
const RAW_KEY = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFAKE";
const B64_KEY = toB64(RAW_KEY);

describe("Public Key Edit", () => {
  let wrapper: PublicKeyEditWrapper;
  setActivePinia(createPinia());
  const publicKeysStore = usePublicKeysStore();
  const vuetify = createVuetify();

  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  const mockSshApi = new MockAdapter(sshApi.getAxios());

  const mockPublicKey = {
    name: "test-name",
    data: B64_KEY,
    filter: { hostname: ".*" },
    username: ".*",
    fingerprint: "fake-fingerprint",
    created_at: "2023-01-01T00:00:00Z",
    tenant_id: "fake-tenant",
  };

  beforeEach(() => {
    document.body.innerHTML = "";

    localStorage.setItem("tenant", "fake-tenant-data");

    mockTagsApi.resetHandlers();
    mockTagsApi
      .onGet(/\/api\/namespaces\/fake-tenant-data\/tags.*/)
      .reply(200, [{ name: "1" }, { name: "2" }, { name: "3" }, { name: "4" }]);

    mockSshApi.resetHandlers();

    wrapper = mount(PublicKeyEdit, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
        stubs: {
          "v-file-upload": true,
          "v-file-upload-item": true,
        },
      },
      props: { publicKey: mockPublicKey },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components and conditional fields", async () => {
    expect(wrapper.find('[data-test="public-key-edit-title-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-edit-icon"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    expect(dialog.find('[data-test="name-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="username-restriction-field"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await flushPromises();
    await new Promise(requestAnimationFrame);
    expect(dialog.find('[data-test="rule-field"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();
    await new Promise(requestAnimationFrame);
    expect(dialog.find('[data-test="filter-restriction-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await flushPromises();
    await new Promise(requestAnimationFrame);
    expect(dialog.find('[data-test="hostname-field"]').exists()).toBe(true);

    expect(dialog.find('[data-test="data-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-edit-cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-edit-save-btn"]').exists()).toBe(true);
  });

  it("Conditional rendering: username + tags shows proper inputs", async () => {
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();
    await new Promise(requestAnimationFrame);

    expect(dialog.find('[data-test="rule-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(true);
    expect(dialog.find('[data-test="hostname-field"]').exists()).toBe(false);
  });

  it("Conditional rendering: hostname filter shows hostname field only", async () => {
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await flushPromises();
    await new Promise(requestAnimationFrame);

    expect(dialog.find('[data-test="hostname-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(false);
  });

  it("Saves with default (all devices + any username)", async () => {
    mockSshApi
      .onPut(`http://localhost:3000/api/sshkeys/public-keys/${mockPublicKey.fingerprint}`)
      .reply(200);
    const storeSpy = vi.spyOn(publicKeysStore, "updatePublicKey");

    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("my edited public key");
    await wrapper.findComponent('[data-test="pk-edit-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      data: mockPublicKey.data,
      filter: { hostname: ".*" },
      name: "my edited public key",
      username: ".*",
      fingerprint: mockPublicKey.fingerprint,
      created_at: mockPublicKey.created_at,
      tenant_id: mockPublicKey.tenant_id,
    });
  });

  it("Saves with hostname restriction", async () => {
    mockSshApi
      .onPut(`http://localhost:3000/api/sshkeys/public-keys/${mockPublicKey.fingerprint}`)
      .reply(200);
    const storeSpy = vi.spyOn(publicKeysStore, "updatePublicKey");

    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("host key");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("web-.*");
    await flushPromises();

    await wrapper.findComponent('[data-test="pk-edit-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      data: mockPublicKey.data,
      filter: { hostname: "web-.*" },
      name: "host key",
      username: ".*",
      fingerprint: mockPublicKey.fingerprint,
      created_at: mockPublicKey.created_at,
      tenant_id: mockPublicKey.tenant_id,
    });
  });

  it("Saves with tags restriction (up to 3 tags)", async () => {
    mockSshApi
      .onPut(`http://localhost:3000/api/sshkeys/public-keys/${mockPublicKey.fingerprint}`)
      .reply(200);
    const storeSpy = vi.spyOn(publicKeysStore, "updatePublicKey");

    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="name-field"]').setValue("tags key");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();

    await wrapper.findComponent('[data-test="tags-selector"]').setValue(["1", "2"]);
    await flushPromises();

    await wrapper.findComponent('[data-test="pk-edit-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      data: mockPublicKey.data,
      filter: { tags: ["1", "2"] },
      name: "tags key",
      username: ".*",
      fingerprint: mockPublicKey.fingerprint,
      created_at: mockPublicKey.created_at,
      tenant_id: mockPublicKey.tenant_id,
    });
  });

  it("Blocks selecting more than 3 tags and shows error", async () => {
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");
    await flushPromises();

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();

    await wrapper.findComponent('[data-test="tags-selector"]').setValue(["1", "2", "3", "4"]);
    await flushPromises();

    expect(wrapper.vm.tagSelectorErrorMessage).toBe("You can select up to three tags only");

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    expect(formDialog.props("confirmDisabled")).toBe(true);
  });

  it("Displays error message if name is not provided", async () => {
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="name-field"]').setValue("foo");
    await wrapper.findComponent('[data-test="name-field"]').setValue("");
    await flushPromises();

    expect(wrapper.vm.nameError).toBeTruthy();
  });

  it("Displays error message if username is not provided when restriction is active", async () => {
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await wrapper.findComponent('[data-test="rule-field"]').setValue("bar");
    await wrapper.findComponent('[data-test="rule-field"]').setValue("");
    await flushPromises();

    expect(wrapper.vm.usernameError).toBeTruthy();
  });

  it("Displays error message if hostname is not provided when filter=hostname", async () => {
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("web-.*");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("");
    await flushPromises();

    expect(wrapper.vm.hostnameError).toBeTruthy();
  });
});
