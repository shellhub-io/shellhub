import { describe, expect, it, afterEach, beforeEach } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import DeviceAdd from "@/components/Devices/DeviceAdd.vue";

describe("DeviceAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceAdd>>;
  let dialog: DOMWrapper<Element>;

  const openDialog = async () => {
    await wrapper.find('[data-test="device-add-btn"]').trigger("click");
    await flushPromises();
  };

  beforeEach(() => {
    wrapper = mountComponent(DeviceAdd, {
      piniaOptions: { initialState: { auth: { tenantId: "test-tenant-id" } } },
    });

    dialog = new DOMWrapper(document.body);
  });

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("add device button", () => {
    it("renders add device button", () => {
      expect(wrapper.find('[data-test="device-add-btn"]').exists()).toBe(true);
    });

    it("displays Add Device text", () => {
      expect(wrapper.find('[data-test="device-add-btn"]').text()).toContain("Add Device");
    });

    it("opens dialog when clicked", async () => {
      await openDialog();

      expect(dialog.find('[data-test="device-add-dialog"]').exists()).toBe(true);
    });
  });

  describe("device add dialog", () => {
    it("renders dialog with title and description", async () => {
      await openDialog();

      expect(dialog.find('[data-test="device-add-dialog"]').exists()).toBe(true);
      expect(dialog.text()).toContain("Adding a device");
      expect(dialog.text()).toContain("Choose an installation method");
    });

    it("displays documentation link in footer", async () => {
      await openDialog();

      const documentationLink = dialog.find('[data-test="documentation-link"]');
      expect(documentationLink.exists()).toBe(true);
      expect(documentationLink.attributes("href")).toBe("https://docs.shellhub.io/user-guides/devices/adding");
    });

    it("renders expansion panels for installation methods", async () => {
      await openDialog();

      expect(dialog.find(".v-expansion-panels").exists()).toBe(true);
      expect(dialog.findAll(".v-expansion-panel").length).toBeGreaterThan(0);
    });

    it("shows recommended method chip", async () => {
      await openDialog();

      expect(dialog.text()).toContain("recommended");
    });

    it("closes dialog when close button is clicked", async () => {
      await openDialog();

      await dialog.find('[data-test="close-btn-toolbar"]').trigger("click");
      await flushPromises();

      expect(dialog.find(".v-overlay__content").attributes("style")).toContain("display: none");
    });
  });

  describe("installation methods", () => {
    it("displays installation method options", async () => {
      await openDialog();

      const installationMethods = ["Auto", "Docker", "Podman", "Snap", "Standalone", "WSL", "Yocto Project", "Buildroot", "FreeBSD"];

      installationMethods.forEach((method) => {
        expect(dialog.text()).toContain(method);
      });
    });

    it("displays method requirements", async () => {
      await openDialog();

      expect(dialog.text()).toContain("Requirements:");
    });

    it("expands panel when clicked", async () => {
      await openDialog();

      const firstPanel = dialog.find(".v-expansion-panel-title");
      await firstPanel.trigger("click");
      await flushPromises();

      const panelText = dialog.find(".v-expansion-panel-text");
      expect(panelText.exists()).toBe(true);
    });

    it("shows the method's command when clicked", async () => {
      const command = "curl -sSf http://localhost:3000/install.sh | TENANT_ID=test-tenant-id SERVER_ADDRESS=http://localhost:3000 sh";
      await openDialog();

      const firstPanel = dialog.find(".v-expansion-panel-title");
      await firstPanel.trigger("click");
      await flushPromises();

      const copyCommandField = dialog.find('[data-test="copy-command-field"] input').element as HTMLInputElement;
      expect(copyCommandField.value).toContain(command);
    });
  });

  describe("advanced options", () => {
    it("shows advanced options expansion panel", async () => {
      await openDialog();

      expect(dialog.text()).toContain("Advanced Options");
    });
  });
});
