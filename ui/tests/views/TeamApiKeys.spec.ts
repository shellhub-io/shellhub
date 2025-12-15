import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import TeamApiKeys from "@/views/TeamApiKeys.vue";
import { apiKeysApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";
import useApiKeysStore from "@/store/modules/api_keys";

type TeamApiKeysWrapper = VueWrapper<InstanceType<typeof TeamApiKeys>>;

describe("Team Api Keys", () => {
  let wrapper: TeamApiKeysWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockApiKeysApi = new MockAdapter(apiKeysApi.getAxios());
  const apiKeysStore = useApiKeysStore();

  const mockApiKeys = [
    {
      name: "fake-api-key",
      tenant_id: "00000000-0000-4000-0000-000000000000",
      role: "administrator" as const,
      created_by: "xxxxxxxx",
      created_at: "",
      updated_at: "",
      expires_in: 1753815353,
    },
  ];

  beforeEach(() => {
    mockApiKeysApi.reset();

    mockApiKeysApi
      .onGet("http://localhost:3000/api/namespaces/api-key?page=1&per_page=10")
      .reply(200, mockApiKeys, { "x-total-count": "1" });

    apiKeysStore.$patch({
      apiKeys: mockApiKeys,
      apiKeysCount: 1,
    });

    wrapper = mount(TeamApiKeys, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
    mockApiKeysApi.reset();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", async () => {
    await flushPromises();
    expect(wrapper.find('[data-test="api-key-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="api-key-list"]').exists()).toBe(true);
  });
});
