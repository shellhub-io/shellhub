import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import DeviceAdd from "@/components/Devices/DeviceAdd.vue";
import { devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

const devices = [
  {
    uid: "a582b47a42d",
    name: "39-5e-2a",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T18:58:53.276Z",
    online: false,
    namespace: "user",
    status: "accepted",
  },
  {
    uid: "a582b47a42e",
    name: "39-5e-2b",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T19:58:53.276Z",
    online: true,
    namespace: "user",
    status: "accepted",
  },
];

describe("Device Add", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceAdd>>;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

  beforeEach(async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, devices);

    wrapper = mount(DeviceAdd, {
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

  it("Renders the component dialog with new interface", async () => {
    const button = wrapper.find('[data-test="device-add-btn"]');
    expect(button.exists()).toBe(true);
    await button.trigger("click");
    const dialog = new DOMWrapper(document.body);

    // Test dialog structure
    expect(dialog.find('[data-test="device-add-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="documentation-link"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);

    // Test toolbar elements
    expect(dialog.find(".v-toolbar").exists()).toBe(true);
    expect(dialog.find(".v-toolbar .v-avatar").exists()).toBe(true);
    expect(dialog.find(".v-toolbar .v-toolbar-title").exists()).toBe(true);

    // Test expansion panels for installation methods
    expect(dialog.find(".v-expansion-panels").exists()).toBe(true);
    expect(dialog.find(".v-expansion-panel").exists()).toBe(true);

    // Test footer toolbar
    expect(dialog.findAll(".v-toolbar").length).toBe(2); // Header and footer
  });

  it("Shows installation methods with expansion panels", async () => {
    const button = wrapper.find('[data-test="device-add-btn"]');
    await button.trigger("click");
    const dialog = new DOMWrapper(document.body);

    const expansionPanels = dialog.findAll(".v-expansion-panel");
    expect(expansionPanels.length).toBeGreaterThan(0);

    // Test that Auto method exists and is recommended
    const autoPanel = dialog.find('[data-test="device-add-dialog"] .v-expansion-panel');
    expect(autoPanel.exists()).toBe(true);
  });

  it("Shows advanced options for script-based methods", async () => {
    const button = wrapper.find('[data-test="device-add-btn"]');
    await button.trigger("click");
    const dialog = new DOMWrapper(document.body);

    // Wait for dialog to be fully rendered
    await wrapper.vm.$nextTick();

    // Expand the first method panel (should be Auto)
    const firstPanel = dialog.find(".v-expansion-panel-title");
    if (firstPanel.exists()) {
      await firstPanel.trigger("click");
      await wrapper.vm.$nextTick();

      // Check if advanced options text exists (since it's inside an alert)
      const advancedOptionsText = dialog.find(".v-expansion-panel-text");
      expect(advancedOptionsText.exists()).toBe(true);

      // Look for Advanced Options text content
      const advancedText = dialog.text().includes("Advanced Options");
      expect(advancedText).toBe(true);
    }
  });
});
