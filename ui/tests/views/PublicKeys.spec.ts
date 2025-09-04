import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PublicKeys from "@/views/PublicKeys.vue";
import { sshApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import usePublicKeysStore from "@/store/modules/public_keys";

type PublicKeysWrapper = VueWrapper<InstanceType<typeof PublicKeys>>;

describe("Public Keys", () => {
  let wrapper: PublicKeysWrapper;
  setActivePinia(createPinia());
  const publicKeysStore = usePublicKeysStore();
  const vuetify = createVuetify();

  const mockSshApi = new MockAdapter(sshApi.getAxios());

  const mockPublicKeys = [{
    data: "",
    fingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:01",
    created_at: "2025-01-01T00:00:00.000Z",
    tenant_id: "00000000-0000-4000-0000-000000000000",
    name: "public-key-test",
    username: ".*",
    filter: {
      hostname: ".*",
    },
  }];

  beforeEach(async () => {
    mockSshApi.onGet("http://localhost:3000/api/sshkeys/public-keys?page=1&per_page=10").reply(200, mockPublicKeys, { "x-total-count": 1 });
    publicKeysStore.publicKeys = mockPublicKeys;

    wrapper = mount(PublicKeys, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="public-keys-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-keys-components"]').exists()).toBe(true);
  });

  it("Renders the PublicKeyAdd component", () => {
    expect(wrapper.findComponent({ name: "PublicKeyAdd" }).exists()).toBe(true);
  });

  it("Shows the no items message when there are no public keys", async () => {
    mockSshApi.onGet("http://localhost:3000/api/sshkeys/public-keys?page=1&per_page=10").reply(200, [], { "x-total-count": 0 });
    await wrapper.vm.refresh();
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-message-component"]').text()).toContain("Looks like you don't have any Public Keys");
  });
});
