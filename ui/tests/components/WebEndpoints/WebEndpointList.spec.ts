import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper } from "@vue/test-utils";
import WebEndpointList from "@/components/WebEndpoints/WebEndpointList.vue";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import useWebEndpointsStore from "@/store/modules/web_endpoints";
import {
  mockWebEndpoint,
  mockExpiredWebEndpoint,
  mockNeverExpiresWebEndpoint,
  mockWebEndpointWithTLS,
  mockWebEndpoints,
} from "@tests/mocks/webEndpoint";
import { Router } from "vue-router";

describe("WebEndpointList", () => {
  let wrapper: VueWrapper<InstanceType<typeof WebEndpointList>>;
  let webEndpointsStore: ReturnType<typeof useWebEndpointsStore>;
  let router: Router;

  const mountWrapper = async (webEndpoints = mockWebEndpoints) => {
    router = createCleanRouter();

    wrapper = mountComponent(WebEndpointList, {
      global: { plugins: [router] },
      piniaOptions: { initialState: { webEndpoints: { webEndpoints: webEndpoints, webEndpointCount: webEndpoints.length } } },
    });

    webEndpointsStore = useWebEndpointsStore();

    await flushPromises();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe("Component initialization", () => {
    it("fetches web endpoints on mount", () => {
      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalledWith({
        page: 1,
        perPage: 10,
        sortField: undefined,
        sortOrder: undefined,
      });
    });

    it("renders DataTable component", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      expect(dataTable.exists()).toBe(true);
      expect(dataTable.props("tableName")).toBe("webEndpoints");
    });
  });

  describe("Table rendering", () => {
    it("renders table headers correctly", () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      const headers = dataTable.props("headers");

      expect(headers).toHaveLength(7);
      expect(headers[0].text).toBe("Device");
      expect(headers[1].text).toBe("Address");
      expect(headers[2].text).toBe("Host");
      expect(headers[3].text).toBe("Port");
      expect(headers[4].text).toBe("Domain");
      expect(headers[5].text).toBe("Expiration Date");
      expect(headers[6].text).toBe("Actions");
    });

    it("renders web endpoint row with device info", () => {
      expect(wrapper.text()).toContain("39-5e-2a");
      expect(wrapper.text()).toContain("Linux Mint 19.3");
    });

    it("renders web endpoint URL as clickable link", () => {
      const urlLink = wrapper.find('[data-test="web-endpoint-url"] a');
      expect(urlLink.exists()).toBe(true);
      expect(urlLink.attributes("href")).toContain("endpoint-123.example.com");
      expect(urlLink.attributes("target")).toBe("_blank");
      expect(urlLink.attributes("rel")).toBe("noopener noreferrer");
    });

    it("renders host and port", () => {
      expect(wrapper.text()).toContain("192.168.1.1");
      expect(wrapper.text()).toContain("8080");
    });

    it("renders TLS disabled chip when TLS is not enabled", () => {
      const tlsChip = wrapper.find('[data-test="web-endpoint-tls"] .v-chip');
      expect(tlsChip.exists()).toBe(true);
      expect(tlsChip.text()).toBe("Disabled");
    });

    it("renders TLS domain chip when TLS is enabled", async () => {
      wrapper.unmount();
      await mountWrapper([mockWebEndpointWithTLS]);

      const tlsChip = wrapper.find('[data-test="web-endpoint-tls"] .v-chip');
      expect(tlsChip.exists()).toBe(true);
      expect(tlsChip.text()).toBe("secure.example.com");
    });

    it("renders WebEndpointDelete component", () => {
      const deleteComponent = wrapper.findComponent({ name: "WebEndpointDelete" });
      expect(deleteComponent.exists()).toBe(true);
      expect(deleteComponent.props("address")).toBe("endpoint-123");
    });
  });

  describe("Expiration date formatting", () => {
    it("shows future expiration date", () => {
      expect(wrapper.text()).toContain("Expires on");
    });

    it("shows never expires for zero date", async () => {
      wrapper.unmount();
      await mountWrapper([mockNeverExpiresWebEndpoint]);

      expect(wrapper.text()).toContain("Never Expires");
    });

    it("shows expired date and applies warning class", async () => {
      wrapper.unmount();
      await mountWrapper([mockExpiredWebEndpoint]);

      expect(wrapper.text()).toContain("Expired on");
      const row = wrapper.find('[data-test="web-endpoint-row"]');
      expect(row.classes()).toContain("text-warning");
    });
  });

  describe("Pagination", () => {
    it("handles page change", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      await dataTable.vm.$emit("update:page", 2);
      await flushPromises();

      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalledWith({
        page: 2,
        perPage: 10,
        sortField: undefined,
        sortOrder: undefined,
      });
    });

    it("handles items per page change", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      await dataTable.vm.$emit("update:itemsPerPage", 20);
      await flushPromises();

      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalledWith({
        page: 1,
        perPage: 20,
        sortField: undefined,
        sortOrder: undefined,
      });
    });
  });

  describe("Sorting", () => {
    it("sorts by address field", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      await dataTable.vm.$emit("update:sort", "address");
      await flushPromises();

      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalledWith({
        page: 1,
        perPage: 10,
        sortField: "address",
        sortOrder: "asc",
      });
    });

    it("toggles sort order on second click", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });
      await dataTable.vm.$emit("update:sort", "host");
      await flushPromises();

      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalledWith({
        page: 1,
        perPage: 10,
        sortField: "host",
        sortOrder: "asc",
      });

      await dataTable.vm.$emit("update:sort", "host");
      await flushPromises();

      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalledWith({
        page: 1,
        perPage: 10,
        sortField: "host",
        sortOrder: "desc",
      });
    });

    it("sorts by different fields", async () => {
      const dataTable = wrapper.findComponent({ name: "DataTable" });

      await dataTable.vm.$emit("update:sort", "expires_in");
      await flushPromises();

      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalledWith(
        expect.objectContaining({
          sortField: "expires_in",
        }),
      );

      await dataTable.vm.$emit("update:sort", "port");
      await flushPromises();

      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalledWith(
        expect.objectContaining({
          sortField: "port",
        }),
      );
    });
  });

  describe("Empty state", () => {
    it("renders empty state when no endpoints", async () => {
      wrapper.unmount();
      await mountWrapper([]);

      expect(wrapper.text()).toContain("No data available");
    });
  });

  describe("Multiple endpoints", () => {
    it("renders multiple web endpoints", () => {
      const rows = wrapper.findAll('[data-test="web-endpoint-row"]');
      expect(rows).toHaveLength(3);
    });
  });

  describe("Refresh functionality", () => {
    it("refetches data when WebEndpointDelete emits update", async () => {
      const deleteComponent = wrapper.findComponent({ name: "WebEndpointDelete" });
      deleteComponent.vm.$emit("update");
      await flushPromises();

      expect(webEndpointsStore.fetchWebEndpointsList).toHaveBeenCalled();
    });
  });

  describe("Device navigation", () => {
    it("navigates to device details when device name is clicked", async () => {
      wrapper.unmount();
      await mountWrapper([mockWebEndpoint]);

      const pushSpy = vi.spyOn(router, "push");

      const deviceLink = wrapper.find(".link");
      await deviceLink.trigger("click");
      await flushPromises();

      expect(pushSpy).toHaveBeenCalledWith({
        name: "DeviceDetails",
        params: { identifier: "device-123" },
      });
    });
  });
});
