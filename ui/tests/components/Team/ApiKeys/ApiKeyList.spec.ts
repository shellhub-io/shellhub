import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import moment from "moment";
import { createPinia, setActivePinia } from "pinia";
import ApiKeyList from "@/components/Team/ApiKeys/ApiKeyList.vue";
import { apiKeysApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useApiKeysStore from "@/store/modules/api_keys";

type ApiKeyListWrapper = VueWrapper<InstanceType<typeof ApiKeyList>>;

describe("Api Key List", () => {
  let wrapper: ApiKeyListWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockApiKeysApi = new MockAdapter(apiKeysApi.getAxios());
  const apiKeysStore = useApiKeysStore();

  const mockApiKeys = [
    {
      name: "aaaa2",
      tenant_id: "00000-0000-0000-0000-00000000000",
      role: "administrator" as const,
      created_by: "66562f80daba745a106393b5",
      created_at: "2024-06-07T12:10:56.531Z",
      updated_at: "2024-06-07T12:31:03.505Z",
      expires_in: 1720354256,
    },
    {
      name: "aaaa2",
      tenant_id: "00000-0000-0000-0000-00000000000",
      role: "administrator" as const,
      created_by: "66562f80daba745a106393b5",
      created_at: "2024-06-07T12:10:56.531Z",
      updated_at: "2024-06-07T12:31:03.505Z",
      expires_in: 1720354256,
    },
  ];

  beforeEach(() => {
    mockApiKeysApi.onGet("http://localhost:3000/api/namespaces/api-key?page=1&per_page=10").reply(200, mockApiKeys, { "x-total-count": 2 });
    apiKeysStore.$patch({
      apiKeys: mockApiKeys,
      apiKeysCount: 2,
    });

    wrapper = mount(ApiKeyList, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", () => {
    expect(wrapper.find('[data-test="api-key-list"]').exists()).toBe(true);
  });

  describe("expiration formatting", () => {
    it("Returns 'Never' when unixTime is -1", () => {
      const result = wrapper.vm.formatDate(-1);
      expect(result).toBe("Never");
    });

    it("Formats unixTime into 'Expires on <date>' when unixTime is in the future", () => {
      const futureUnixTime = moment().add(1, "day").unix();
      const result = wrapper.vm.formatDate(futureUnixTime);
      const expected = `Expires on ${moment.unix(futureUnixTime).format("MMM D YYYY")}.`;
      expect(result).toBe(expected);
    });

    it("Formats unixTime into 'Expired on <date>' when unixTime is in the past", () => {
      const pastUnixTime = moment().subtract(1, "day").unix();
      const result = wrapper.vm.formatDate(pastUnixTime);
      const expected = `Expired on ${moment.unix(pastUnixTime).format("MMM D YYYY")}.`;
      expect(result).toBe(expected);
    });

    it("Returns false when unixTime is -1", () => {
      const result = wrapper.vm.hasKeyExpired(-1);
      expect(result).toBe(false);
    });

    it("Returns true when unixTime is in the past", () => {
      const pastUnixTime = moment().subtract(1, "day").unix();
      const result = wrapper.vm.hasKeyExpired(pastUnixTime);
      expect(result).toBe(true);
    });

    it("Returns false when unixTime is in the future", () => {
      const futureUnixTime = moment().add(1, "day").unix();
      const result = wrapper.vm.hasKeyExpired(futureUnixTime);
      expect(result).toBe(false);
    });

    describe("expired vs will expire items formatting", () => {
      it("Formats items that will expire correctly", () => {
        const futureUnixTime = moment().add(1, "day").unix();
        const item = { expires_in: futureUnixTime };

        const formatted = wrapper.vm.formatDate(item.expires_in);
        const expected = `Expires on ${moment.unix(futureUnixTime).format("MMM D YYYY")}.`;

        expect(formatted).toBe(expected);
      });

      it("Formats expired items correctly", () => {
        const pastUnixTime = moment().subtract(1, "day").unix();
        const item = { expires_in: pastUnixTime };

        const formatted = wrapper.vm.formatDate(item.expires_in);
        const expected = `Expired on ${moment.unix(pastUnixTime).format("MMM D YYYY")}.`;

        expect(formatted).toBe(expected);
      });

      it("Formats items that will not expire correctly", () => {
        const item = { expires_in: -1 };

        const formatted = wrapper.vm.formatDate(item.expires_in);
        const expected = "Never";

        expect(formatted).toBe(expected);
      });

      it("Formats expired items with formatKey correctly", () => {
        const pastUnixTime = moment().subtract(1, "day").unix();
        const item = { expires_in: pastUnixTime };

        const formatted = wrapper.vm.hasKeyExpired(item.expires_in);
        expect(formatted).toBe(true);
      });

      it("Formats items that will not expire with formatKey correctly", () => {
        const item = { expires_in: -1 };

        const formatted = wrapper.vm.hasKeyExpired(item.expires_in);
        expect(formatted).toBe(false);
      });
    });
  });
});
