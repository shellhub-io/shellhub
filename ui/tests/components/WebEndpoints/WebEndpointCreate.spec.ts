import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import WebEndpointCreate from "@/components/WebEndpoints/WebEndpointCreate.vue";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useDevicesStore from "@/store/modules/devices";
import useWebEndpointsStore from "@/store/modules/web_endpoints";
import { mockDevice } from "@tests/mocks/device";
import { VAutocomplete } from "vuetify/lib/components";

describe("WebEndpointCreate", () => {
  let wrapper: VueWrapper<InstanceType<typeof WebEndpointCreate>>;
  let dialog: DOMWrapper<Element>;
  let devicesStore: ReturnType<typeof useDevicesStore>;
  let webEndpointsStore: ReturnType<typeof useWebEndpointsStore>;

  const mountWrapper = async (props = {}, initialState = {}) => {
    wrapper = mountComponent(WebEndpointCreate, {
      props: {
        modelValue: true,
        uid: "device-123",
        useDevicesList: false,
        ...props,
      },
      attachTo: document.body,
      piniaOptions: {
        initialState,
      },
    });

    devicesStore = useDevicesStore();
    webEndpointsStore = useWebEndpointsStore();

    await flushPromises();

    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering", () => {
    it("renders FormDialog with correct props", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.exists()).toBe(true);
      expect(formDialog.props("title")).toBe("Create Device Web Endpoint");
      expect(formDialog.props("icon")).toBe("mdi-lan");
      expect(formDialog.props("confirmText")).toBe("Create Web Endpoint");
      expect(formDialog.props("cancelText")).toBe("Close");
    });

    it("renders all form fields", () => {
      expect(dialog.find('[data-test="host-text"]').exists()).toBe(true);
      expect(dialog.find('[data-test="port-text"]').exists()).toBe(true);
      expect(dialog.find('[data-test="timeout-combobox"]').exists()).toBe(true);
      expect(dialog.find('[data-test="tls-enabled-checkbox"]').exists()).toBe(true);
    });

    it("does not render device autocomplete when useDevicesList is false", () => {
      expect(dialog.find('[data-test="web-endpoint-autocomplete"]').exists()).toBe(false);
    });

    it("renders device autocomplete when useDevicesList is true", async () => {
      wrapper.unmount();
      document.body.innerHTML = "";

      await mountWrapper(
        { useDevicesList: true, uid: undefined },
        { devices: { devices: [mockDevice] } },
      );

      expect(dialog.find('[data-test="web-endpoint-autocomplete"]').exists()).toBe(true);
    });

    it("does not show custom timeout field initially", () => {
      expect(dialog.find('[data-test="custom-timeout"]').exists()).toBe(false);
    });

    it("shows custom timeout field when custom option selected", async () => {
      const timeoutCombobox = wrapper.findComponent('[data-test="timeout-combobox"]');
      await timeoutCombobox.setValue("custom");
      await flushPromises();

      expect(dialog.find('[data-test="custom-timeout"]').exists()).toBe(true);
    });

    it("does not show TLS fields initially", () => {
      expect(dialog.find('[data-test="tls-accordion"]').attributes("style")).toContain("display: none;");
    });

    it("shows TLS fields when TLS is enabled", async () => {
      const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
      await tlsCheckbox.setValue(true);
      await flushPromises();

      expect(dialog.find('[data-test="tls-accordion"]').attributes("style")).not.toContain("display: none;");
    });
  });

  describe("Form validation", () => {
    it("disables confirm button when port is empty", async () => {
      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue("");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("shows error when host is invalid", async () => {
      const hostField = wrapper.findComponent('[data-test="host-text"]');
      await hostField.setValue("invalid-host");
      await flushPromises();

      expect(dialog.text()).toContain("Enter a valid IPv4 or IPv6 address");
    });

    it.each([
      ["127.0.0.1", true],
      ["192.168.1.1", true],
      ["255.255.255.255", true],
      ["::1", true],
      ["2001:0db8:85a3:0000:0000:8a2e:0370:7334", true],
      ["invalid", false],
      ["256.256.256.256", false],
      ["192.168.1", false],
    ])("validates IPv4/IPv6 address: %s (valid: %s)", async (ip, isValid) => {
      const hostField = wrapper.findComponent('[data-test="host-text"]');
      await hostField.setValue(ip);
      await flushPromises();

      const hasHostError = dialog.text().includes("Enter a valid IPv4 or IPv6 address");
      expect(hasHostError).toBe(!isValid);
    });

    it("validates port is between 1 and 65535", async () => {
      const portField = wrapper.findComponent('[data-test="port-text"]');

      await portField.setValue(0);
      await flushPromises();
      expect(dialog.text()).toContain("Port must be greater than or equal to 1");

      await portField.setValue(65536);
      await flushPromises();
      expect(dialog.text()).toContain("Port must be less than or equal to 65535");

      await portField.setValue(8080);
      await flushPromises();
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(false);
    });

    it("validates TLS domain when TLS is enabled", async () => {
      const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
      await tlsCheckbox.setValue(true);
      await flushPromises();

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(443);
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);

      const tlsDomainField = wrapper.findComponent('[data-test="tls-domain-text"]');
      await tlsDomainField.setValue("example.com");
      await flushPromises();

      expect(formDialog.props("confirmDisabled")).toBe(false);
    });

    it.each([
      ["example.com", true],
      ["sub.example.com", true],
      ["device.local", true],
      ["my-device.example.com", true],
      ["invalid domain", false],
      ["192.168.1.1", false],
    ])("validates FQDN: %s (valid: %s)", async (domain, isValid) => {
      const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
      await tlsCheckbox.setValue(true);
      await flushPromises();

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(443);
      await flushPromises();

      const tlsDomainField = wrapper.findComponent('[data-test="tls-domain-text"]');
      await tlsDomainField.setValue(domain);
      await flushPromises();

      const hasError = dialog.text().includes("Enter a valid FQDN");
      expect(hasError).toBe(!isValid);
    });

    it("validates custom timeout when custom option is selected", async () => {
      const timeoutCombobox = wrapper.findComponent('[data-test="timeout-combobox"]');
      await timeoutCombobox.setValue("custom");
      await flushPromises();

      const customTimeoutField = wrapper.findComponent('[data-test="custom-timeout"]');
      await customTimeoutField.setValue(0);
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);

      await customTimeoutField.setValue(100);
      await flushPromises();

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(8080);
      await flushPromises();

      expect(formDialog.props("confirmDisabled")).toBe(false);
    });

    it("requires device selection when useDevicesList is true", async () => {
      wrapper.unmount();
      document.body.innerHTML = "";

      await mountWrapper(
        { useDevicesList: true, uid: undefined },
        { devices: { devices: [mockDevice] } },
      );

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(8080);
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });
  });

  describe("Device autocomplete (useDevicesList)", () => {
    beforeEach(async () => {
      wrapper.unmount();
      document.body.innerHTML = "";

      await mountWrapper(
        { useDevicesList: true, uid: undefined },
        { devices: { devices: [mockDevice], deviceListFilter: undefined } },
      );
    });

    it("fetches devices on mount", () => {
      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith({ filter: undefined });
    });

    it("filters devices on search input", async () => {
      const autocomplete = wrapper.findComponent('[data-test="web-endpoint-autocomplete"]') as VueWrapper<VAutocomplete>;
      autocomplete.vm.$emit("update:search", "test");
      await flushPromises();

      const expectedFilter = Buffer.from(
        JSON.stringify([
          {
            type: "property",
            params: { name: "name", operator: "contains", value: "test" },
          },
        ]),
      ).toString("base64");

      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith({ filter: expectedFilter });
    });

    it("clears filter when search is empty", async () => {
      const autocomplete = wrapper.findComponent('[data-test="web-endpoint-autocomplete"]') as VueWrapper<VAutocomplete>;
      autocomplete.vm.$emit("update:search", "");
      await flushPromises();

      expect(devicesStore.deviceListFilter).toBeUndefined();
      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith({ filter: undefined });
    });

    it("shows error when fetching devices fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(devicesStore.fetchDeviceList).mockRejectedValueOnce(error);

      const autocomplete = wrapper.findComponent('[data-test="web-endpoint-autocomplete"]') as VueWrapper<VAutocomplete>;
      autocomplete.vm.$emit("update:search", "test");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load devices.");
    });
  });

  describe("Web endpoint creation", () => {
    it("creates web endpoint with default timeout (unlimited)", async () => {
      const hostField = wrapper.findComponent('[data-test="host-text"]');
      await hostField.setValue("192.168.1.1");

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(8080);

      const confirmBtn = dialog.find('[data-test="create-tunnel-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(webEndpointsStore.createWebEndpoint).toHaveBeenCalledWith({
        uid: "device-123",
        host: "192.168.1.1",
        port: 8080,
        ttl: -1,
      });

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Web Endpoint created successfully.");
      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("creates web endpoint with predefined timeout", async () => {
      const hostField = wrapper.findComponent('[data-test="host-text"]');
      await hostField.setValue("127.0.0.1");

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(3000);

      const timeoutCombobox = wrapper.findComponent('[data-test="timeout-combobox"]');
      await timeoutCombobox.setValue(3600);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="create-tunnel-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(webEndpointsStore.createWebEndpoint).toHaveBeenCalledWith({
        uid: "device-123",
        host: "127.0.0.1",
        port: 3000,
        ttl: 3600,
      });
    });

    it("creates web endpoint with custom timeout", async () => {
      const hostField = wrapper.findComponent('[data-test="host-text"]');
      await hostField.setValue("10.0.0.1");

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(5000);

      const timeoutCombobox = wrapper.findComponent('[data-test="timeout-combobox"]');
      await timeoutCombobox.setValue("custom");
      await flushPromises();

      const customTimeoutField = wrapper.findComponent('[data-test="custom-timeout"]');
      await customTimeoutField.setValue(7200);
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="create-tunnel-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(webEndpointsStore.createWebEndpoint).toHaveBeenCalledWith({
        uid: "device-123",
        host: "10.0.0.1",
        port: 5000,
        ttl: 7200,
      });
    });

    it("creates web endpoint with TLS enabled", async () => {
      const hostField = wrapper.findComponent('[data-test="host-text"]');
      await hostField.setValue("192.168.1.100");

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(443);

      const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
      await tlsCheckbox.setValue(true);
      await flushPromises();

      const tlsDomainField = wrapper.findComponent('[data-test="tls-domain-text"]');
      await tlsDomainField.setValue("example.com");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="create-tunnel-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(webEndpointsStore.createWebEndpoint).toHaveBeenCalledWith({
        uid: "device-123",
        host: "192.168.1.100",
        port: 443,
        ttl: -1,
        tls: {
          enabled: true,
          verify: false,
          domain: "example.com",
        },
      });
    });

    it("creates web endpoint with TLS verify enabled", async () => {
      const hostField = wrapper.findComponent('[data-test="host-text"]');
      await hostField.setValue("192.168.1.200");

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(443);

      const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
      await tlsCheckbox.setValue(true);
      await flushPromises();

      const tlsVerifyCheckbox = wrapper.findComponent('[data-test="tls-verify-checkbox"]');
      await tlsVerifyCheckbox.setValue(true);

      const tlsDomainField = wrapper.findComponent('[data-test="tls-domain-text"]');
      await tlsDomainField.setValue("secure.example.com");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="create-tunnel-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(webEndpointsStore.createWebEndpoint).toHaveBeenCalledWith({
        uid: "device-123",
        host: "192.168.1.200",
        port: 443,
        ttl: -1,
        tls: {
          enabled: true,
          verify: true,
          domain: "secure.example.com",
        },
      });
    });

    it("uses selected device when useDevicesList is true", async () => {
      wrapper.unmount();
      document.body.innerHTML = "";

      await mountWrapper(
        { useDevicesList: true, uid: undefined },
        { devices: { devices: [mockDevice] } },
      );

      const autocomplete = wrapper.findComponent('[data-test="web-endpoint-autocomplete"]');
      await autocomplete.setValue(mockDevice);
      await flushPromises();

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(8080);

      const confirmBtn = dialog.find('[data-test="create-tunnel-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(webEndpointsStore.createWebEndpoint).toHaveBeenCalledWith({
        uid: "a582b47a42d",
        host: "127.0.0.1",
        port: 8080,
        ttl: -1,
      });
    });
  });

  describe("Error handling", () => {
    it("shows alert when 403 error (max endpoints reached)", async () => {
      const error = createAxiosError(403, "Forbidden");
      vi.mocked(webEndpointsStore.createWebEndpoint).mockRejectedValueOnce(error);

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(8080);

      const confirmBtn = dialog.find('[data-test="create-tunnel-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("alertMessage")).toContain(
        "This device has reached the maximum allowed number of Web Endpoints",
      );
    });

    it("shows generic error for other error codes", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(webEndpointsStore.createWebEndpoint).mockRejectedValueOnce(error);

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(8080);

      const confirmBtn = dialog.find('[data-test="create-tunnel-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to create Web Endpoint.");
    });
  });

  describe("Dialog close behavior", () => {
    it("closes dialog and resets fields on cancel", async () => {
      const hostField = wrapper.findComponent('[data-test="host-text"]');
      await hostField.setValue("10.10.10.10");

      const portField = wrapper.findComponent('[data-test="port-text"]');
      await portField.setValue(9000);

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      formDialog.vm.$emit("cancel");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });

    it("clears filter and refetches devices on close", async () => {
      wrapper.unmount();
      document.body.innerHTML = "";

      vi.mocked(devicesStore.fetchDeviceList).mockResolvedValue();

      await mountWrapper(
        { useDevicesList: true, uid: undefined },
        { devices: { devices: [mockDevice], deviceListFilter: "some-filter" } },
      );

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      formDialog.vm.$emit("close");
      await flushPromises();

      expect(devicesStore.deviceListFilter).toBeUndefined();
      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith({ filter: undefined });
    });

    it("resets TLS fields when TLS is disabled", async () => {
      const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
      await tlsCheckbox.setValue(true);
      await flushPromises();

      const tlsDomainField = wrapper.findComponent('[data-test="tls-domain-text"]');
      await tlsDomainField.setValue("example.com");

      const tlsVerifyCheckbox = wrapper.findComponent('[data-test="tls-verify-checkbox"]');
      await tlsVerifyCheckbox.setValue(true);
      await flushPromises();

      await tlsCheckbox.setValue(false);
      await flushPromises();

      expect(dialog.find('[data-test="tls-accordion"]').attributes("style")).toContain("display: none;");
    });
  });
});
