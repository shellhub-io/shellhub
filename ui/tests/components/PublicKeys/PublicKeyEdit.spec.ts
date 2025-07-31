import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PublicKeyEdit from "@/components/PublicKeys/PublicKeyEdit.vue";
import { sshApi, tagsApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type PublicKeyEditWrapper = VueWrapper<InstanceType<typeof PublicKeyEdit>>;

describe("Public Key Edit", () => {
  let wrapper: PublicKeyEditWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  const mockSshApi = new MockAdapter(sshApi.getAxios());

  const mockPublicKey = {
    name: "test-name",
    data: "",
    filter: {
      hostname: ".*",
    },
    username: ".*",
    fingerprint: "fake-fingerprint",
    created_at: "2023-01-01T00:00:00Z",
    tenant_id: "fake-tenant",
  };

  beforeEach(async () => {
    mockTagsApi.onGet("http://localhost:3000/api/tags").reply(200, []);

    wrapper = mount(PublicKeyEdit, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      props: {
        publicKey: mockPublicKey,
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
    expect(wrapper.find('[data-test="public-key-edit-title-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-edit-icon"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="public-key-edit-title"]').exists()).toBe(true);
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
    expect(dialog.find('[data-test="pk-edit-cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-edit-save-btn"]').exists()).toBe(true);
  });

  it("Allows editing a public key with username restriction", async () => {
    await wrapper.setProps({
      publicKey: {
        data: "fake key",
        filter: {
          hostname: ".*",
        },
        name: "my edited public key",
        username: ".*",
        fingerprint: "fingerprint123",
        created_at: "2023-01-01T00:00:00Z",
        tenant_id: "fake-tenant",
      },
    });
    mockSshApi.onPut("http://localhost:3000/api/sshkeys/public-keys/fingerprint123").reply(200);
    const pkEdit = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");
    await flushPromises();
    await wrapper.findComponent('[data-test="name-field"]').setValue("my edited public key");
    await wrapper.findComponent('[data-test="data-field"]').setValue("fakeish key");
    await wrapper.findComponent('[data-test="pk-edit-save-btn"]').trigger("click");
    await flushPromises();
    expect(pkEdit).toHaveBeenCalledWith("publicKeys/put", {
      data: btoa("fake key"),
      filter: {
        hostname: ".*",
      },
      name: "my edited public key",
      username: ".*",
      fingerprint: "fingerprint123",
      created_at: "2023-01-01T00:00:00Z",
      tenant_id: "fake-tenant",
    });
  });

  it("Displays error message if name is not provided", async () => {
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="name-field"]').setValue("foo");
    await wrapper.findComponent('[data-test="name-field"]').setValue("");
    await flushPromises();
    expect(wrapper.vm.nameError).toBe("this is a required field");
  });

  it("Displays error message if username is not provided", async () => {
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await wrapper.findComponent('[data-test="rule-field"]').setValue("foo");
    await wrapper.findComponent('[data-test="rule-field"]').setValue("");
    await flushPromises();
    expect(wrapper.vm.usernameError).toBe("this is a required field");
  });

  it("Displays error message if hostname is not provided", async () => {
    await wrapper.findComponent('[data-test="public-key-edit-title-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("foo");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("");
    await flushPromises();
    expect(wrapper.vm.hostnameError).toBe("this is a required field");
  });
});
