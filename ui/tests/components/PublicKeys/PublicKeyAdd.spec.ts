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
    localStorage.setItem("tenant", "fake-tenant-data");

    mockTagsApi
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
      .reply(200, [{ name: "1" }, { name: "2" }]);

    authStore.role = "owner";
    wrapper = mount(PublicKeyAdd, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
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
    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    expect(formDialog.exists()).toBe(true);
    expect(formDialog.props("title")).toBe("New Public Key");
    expect(formDialog.props("icon")).toBe("mdi-key-outline");
    expect(formDialog.props("confirmText")).toBe("Save");
    expect(formDialog.props("cancelText")).toBe("Cancel");

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

    const storeSpy = vi.spyOn(publicKeysStore, "createPublicKey");

    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");

    await flushPromises();

    // Set the name to match what the test expects
    await wrapper.findComponent('[data-test="name-field"]').setValue("my new public key");
    await wrapper.findComponent('[data-test="data-field"]').setValue("fakeish key");

    // Wait for validation to complete
    await flushPromises();

    await wrapper.findComponent('[data-test="pk-add-save-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      data: btoa("fakeish key"),
      filter: { hostname: ".*" },
      name: "my new public key",
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

  it("Displays error message if username is not provided", async () => {
    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="username-restriction-field"]').setValue("username");
    await wrapper.findComponent('[data-test="rule-field"]').setValue("foo");
    await wrapper.findComponent('[data-test="rule-field"]').setValue("");
    await flushPromises();
    expect(wrapper.vm.usernameError).toBe("this is a required field");
  });

  it("Displays error message if hostname is not provided", async () => {
    await wrapper.findComponent('[data-test="public-key-add-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="filter-restriction-field"]').setValue("hostname");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("foo");
    await wrapper.findComponent('[data-test="hostname-field"]').setValue("");
    await flushPromises();
    expect(wrapper.vm.hostnameError).toBe("this is a required field");
  });
});
