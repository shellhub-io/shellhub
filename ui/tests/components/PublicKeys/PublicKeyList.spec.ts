import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PublicKeysList from "@/components/PublicKeys/PublicKeysList.vue";
import { sshApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

type PublicKeysListWrapper = VueWrapper<InstanceType<typeof PublicKeysList>>;

describe("Public Key List", () => {
  let wrapper: PublicKeysListWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();
  const mockSshApi = new MockAdapter(sshApi.getAxios());

  const mockPublicKeys = {
    data: [
      {
        data: "",
        fingerprint: "fake-fingerprint",
        created_at: "2020-05-01T00:00:00.000Z",
        tenant_id: "fake-tenant",
        name: "example",
        filter: {
          hostname: ".*",
        },
        username: ".*",
      },
    ],
    headers: {
      "x-total-count": 1,
    },
  };

  beforeEach(async () => {
    mockSshApi.onGet("http://localhost:3000/api/sshkeys/public-keys?filter=&page=1&per_page=10").reply(200, mockPublicKeys);
    store.commit("publicKeys/setPublicKeys", mockPublicKeys);
    wrapper = mount(PublicKeysList, {
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
    expect(wrapper.find('[data-test="public-keys-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-item"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-fingerprint"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-filter"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-username"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-created-at"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-actions"]').exists()).toBe(true);
  });

  it("Handles authorization for editing and removing keys", async () => {
    authStore.role = "owner";
    expect(wrapper.vm.hasAuthorizationFormDialogEdit).toBeTruthy();
    expect(wrapper.vm.hasAuthorizationFormDialogRemove).toBeTruthy();

    authStore.role = "observer";
    expect(wrapper.vm.hasAuthorizationFormDialogEdit).toBeFalsy();
    expect(wrapper.vm.hasAuthorizationFormDialogRemove).toBeFalsy();
  });

  it("Checks if the public key list is not empty", () => {
    expect(wrapper.vm.publicKeys.length).toBeGreaterThan(0);
  });

  it("Checks if the public key has correct properties", () => {
    const publicKey = wrapper.vm.publicKeys[0];
    expect(publicKey).toHaveProperty("data");
    expect(publicKey).toHaveProperty("fingerprint");
    expect(publicKey).toHaveProperty("created_at");
    expect(publicKey).toHaveProperty("tenant_id");
    expect(publicKey).toHaveProperty("name");
    expect(publicKey).toHaveProperty("filter");
    expect(publicKey).toHaveProperty("username");
  });

  it("Checks if the public key filter is a hostname", () => {
    const publicKey = wrapper.vm.publicKeys[0];
    expect(publicKey.filter).toHaveProperty("hostname");
  });
});
