import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PublicKeyAdd from "@/components/PublicKeys/PublicKeyAdd.vue";
import { sshApi, tagsApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

type PublicKeyAddWrapper = VueWrapper<InstanceType<typeof PublicKeyAdd>>;

describe("Public Key Add", () => {
  let wrapper: PublicKeyAddWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();

  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  const mockSshApi = new MockAdapter(sshApi.getAxios());

  beforeEach(async () => {
    mockTagsApi.onGet("http://localhost:3000/api/tags").reply(200, []);

    authStore.role = "owner";
    wrapper = mount(PublicKeyAdd, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
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
    expect(wrapper.find('[data-test="public-key-add-btn"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="pk-add-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="name-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="username-restriction-field"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await flushPromises();
    expect(dialog.find('[data-test="rule-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="filter-restriction-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await flushPromises();
    expect(dialog.find('[data-test="hostname-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="data-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-add-cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-add-save-btn"]').exists()).toBe(true);
  });

  it("Allows adding a public key with username restriction", async () => {
    mockSshApi.onPost("http://localhost:3000/api/sshkeys/public-keys").reply(200);

    const pkAdd = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="name-field"]').setValue("my public key");
    await wrapper.findComponent('[data-test="data-field"]').setValue("fake key");
    await wrapper.findComponent('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();
    expect(pkAdd).toHaveBeenCalledWith("publicKeys/post", {
      data: btoa("fake key"),
      filter: {
        hostname: ".*",
      },
      name: "my public key",
      username: ".*",
    });
  });

  it("Displays error message if public key creation fails", async () => {
    mockSshApi.onPost("http://localhost:3000/api/sshkeys/public-keys").reply(409);

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="name-field"]').setValue("test");
    await wrapper.findComponent('[data-test="data-field"]').setValue("fake key");
    await wrapper.findComponent('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();
    expect(wrapper.vm.publicKeyDataError).toBe("Public Key data already exists");
  });

  it("Allows adding a public key with hostname restriction", async () => {
    mockSshApi.onPost("http://localhost:3000/api/sshkeys/public-keys").reply(200);

    const pkAdd = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="name-field"]').setValue("my public key");
    await wrapper.findComponent('[data-test="data-field"]').setValue("fake key");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("example.com");
    await wrapper.findComponent('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();
    expect(pkAdd).toHaveBeenCalledWith("publicKeys/post", {
      data: btoa("fake key"),
      filter: {
        hostname: "example.com",
      },
      name: "my public key",
      username: ".*",
    });
  });

  it("Allows adding a public key with tag restriction", async () => {
    mockSshApi.onPost("http://localhost:3000/api/sshkeys/public-keys").reply(200);

    const pkAdd = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="name-field"]').setValue("my public key");
    await wrapper.findComponent('[data-test="data-field"]').setValue("fake key");
    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("tags");
    await wrapper.findComponent('[data-test="tags-selector"]').setValue(["tag1", "tag2"]);
    await wrapper.findComponent('[data-test="pk-add-save-btn"]').trigger("click");
    await flushPromises();
    expect(pkAdd).toHaveBeenCalledWith("publicKeys/post", {
      data: btoa("fake key"),
      filter: {
        tags: ["tag1", "tag2"],
      },
      name: "my public key",
      username: ".*",
    });
  });

  it("Displays error message if name is not provided", async () => {
    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="name-field"]').setValue("foo");
    await wrapper.findComponent('[data-test="name-field"]').setValue("");
    await flushPromises();
    expect(wrapper.vm.nameError).toBe("this is a required field");
  });
});
