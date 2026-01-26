import { describe, expect, it, afterEach } from "vitest";
import { VueWrapper, DOMWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import ContainerAdd from "@/components/Containers/ContainerAdd.vue";

describe("ContainerAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof ContainerAdd>>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = () => {
    wrapper = mountComponent(ContainerAdd, {
      piniaOptions: { initialState: { auth: { tenantId: "test-tenant-id" } } },
      attachTo: document.body,
    });

    dialog = new DOMWrapper(document.body);
  };

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("button rendering", () => {
    it("displays add Docker host button", () => {
      mountWrapper();

      const button = wrapper.find('[data-test="container-add-btn"]');
      expect(button.exists()).toBe(true);
      expect(button.text()).toContain("Add Docker Host");
    });
  });

  describe("dialog interaction", () => {
    it("does not show dialog initially", () => {
      mountWrapper();
      expect(dialog.find('[role="dialog"]').exists()).toBe(false);
    });

    it("opens dialog when button is clicked", async () => {
      mountWrapper();

      const button = wrapper.find('[data-test="container-add-btn"]');
      await button.trigger("click");

      expect(dialog.find('[role="dialog"]').exists()).toBe(true);
    });

    it("displays dialog title and description", async () => {
      mountWrapper();

      await wrapper.find('[data-test="container-add-btn"]').trigger("click");

      expect(dialog.text()).toContain("Registering a Docker host");
      expect(dialog.text()).toContain("Install the ShellHub Connector to add Docker containers");
    });
  });

  describe("installation instructions", () => {
    it("displays installation instructions", async () => {
      mountWrapper();

      await wrapper.find('[data-test="container-add-btn"]').trigger("click");

      const dialogText = dialog.find('[data-test="dialog-text"]');
      expect(dialogText.exists()).toBe(true);
      expect(dialogText.text()).toContain("In order to add Docker containers to ShellHub");
      expect(dialogText.text()).toContain("you need to install the ShellHub Connector");
    });

    it("displays one-line installation script information", async () => {
      mountWrapper();

      await wrapper.find('[data-test="container-add-btn"]').trigger("click");

      expect(dialog.text()).toContain("easiest way to install");
      expect(dialog.text()).toContain("automatic one-line installation script");
    });

    it("displays command with tenant ID", async () => {
      mountWrapper();

      await wrapper.find('[data-test="container-add-btn"]').trigger("click");

      const copyCommandField = dialog.find('[data-test="copy-command-field"] input').element as HTMLInputElement;

      expect(dialog.text()).toContain("Run the following command");
      expect(copyCommandField.value).toContain("TENANT_ID=test-tenant-id");
    });
  });

  describe("dialog closing", () => {
    it("closes dialog when close button is clicked", async () => {
      mountWrapper();

      await wrapper.find('[data-test="container-add-btn"]').trigger("click");
      expect(dialog.exists()).toBe(true);

      await dialog.find('[data-test="close-btn"]').trigger("click");

      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none");
    });
  });
});
