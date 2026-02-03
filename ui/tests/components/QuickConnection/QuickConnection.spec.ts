import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import QuickConnection from "@/components/QuickConnection/QuickConnection.vue";
import { createCleanRouter } from "@tests/utils/router";

describe("QuickConnection", () => {
  let wrapper: VueWrapper<InstanceType<typeof QuickConnection>>;
  let dialog: DOMWrapper<Element>;

  const mountWrapper = (disabled = false) => {
    wrapper = mountComponent(QuickConnection, {
      global: { plugins: [createCleanRouter()] },
      props: { disabled },
      attachTo: document.body,
    });
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Quick Connect button", () => {
    it("Renders quick connect button", () => {
      const btn = wrapper.find('[data-test="quick-connection-open-btn"]');
      expect(btn.exists()).toBe(true);
    });

    it("Shows 'Quick Connect' text", () => {
      const btn = wrapper.find('[data-test="quick-connection-open-btn"]');
      expect(btn.text()).toBe("Quick Connect");
    });

    it("Shows console icon", () => {
      const btn = wrapper.find('[data-test="quick-connection-open-btn"]');
      const icon = btn.find(".mdi-console");
      expect(icon.exists()).toBe(true);
    });

    it("Opens dialog when clicked", async () => {
      const btn = wrapper.find('[data-test="quick-connection-open-btn"]');
      await btn.trigger("click");
      await flushPromises();

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(true);
    });

    it("Is disabled when disabled prop is true", () => {
      wrapper.unmount();
      mountWrapper(true);

      const btn = wrapper.find('[data-test="quick-connection-open-btn"]');
      expect(btn.attributes("disabled")).toBeDefined();
    });
  });

  describe("Instructions", () => {
    it("Shows keyboard shortcut instructions when not disabled", () => {
      const instructions = wrapper.find('[data-test="quick-connect-instructions"]');
      expect(instructions.exists()).toBe(true);
      expect(instructions.text()).toContain("Press");
      expect(instructions.text()).toContain("Ctrl+K");
      expect(instructions.text()).toContain("to Quick Connect!");
    });

    it("Hides instructions when disabled", () => {
      wrapper.unmount();
      mountWrapper(true);

      const instructions = wrapper.find('[data-test="quick-connect-instructions"]');
      expect(instructions.exists()).toBe(false);
    });
  });

  describe("Dialog display", () => {
    beforeEach(async () => {
      const btn = wrapper.find('[data-test="quick-connection-open-btn"]');
      await btn.trigger("click");
      await flushPromises();
    });

    it("Shows WindowDialog with correct props", () => {
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("title")).toBe("Quick Connect");
      expect(windowDialog.props("description")).toBe("Search and connect to your online devices");
      expect(windowDialog.props("icon")).toBe("mdi-console");
      expect(windowDialog.props("iconColor")).toBe("primary");
      expect(windowDialog.props("showFooter")).toBe(true);
    });

    it("Renders search text field", () => {
      const searchField = dialog.find('[data-test="search-text"]');
      expect(searchField.exists()).toBe(true);
    });

    it("Shows correct column headers", () => {
      const hostnameHeader = dialog.find('[data-test="hostname-header"]');
      const osHeader = dialog.find('[data-test="operating-system-header"]');
      const sshidHeader = dialog.find('[data-test="sshid-header"]');
      const tagsHeader = dialog.find('[data-test="tags-header"]');

      expect(hostnameHeader.text()).toBe("Hostname");
      expect(osHeader.exists()).toBe(true);
      expect(sshidHeader.text()).toBe("SSHID");
      expect(tagsHeader.text()).toBe("Tags");
    });

    it("Renders QuickConnectionList component", () => {
      const list = wrapper.findComponent({ name: "QuickConnectionList" });
      expect(list.exists()).toBe(true);
    });

    it("Passes filter prop to QuickConnectionList", async () => {
      const searchField = dialog.find('[data-test="search-text"] input');
      await searchField.setValue("test-device");
      await flushPromises();

      const list = wrapper.findComponent({ name: "QuickConnectionList" });
      expect(list.props("filter")).toBe("test-device");
    });
  });

  describe("Dialog footer", () => {
    beforeEach(async () => {
      const btn = wrapper.find('[data-test="quick-connection-open-btn"]');
      await btn.trigger("click");
      await flushPromises();
    });

    it("Shows keyboard instructions in footer", () => {
      const connectIcon = dialog.find('[data-test="connect-icon"]');
      const navigateUpIcon = dialog.find('[data-test="navigate-up-icon"]');
      const navigateDownIcon = dialog.find('[data-test="navigate-down-icon"]');
      const copyInstructions = dialog.find('[data-test="copy-sshid-instructions"]');

      expect(connectIcon.exists()).toBe(true);
      expect(navigateUpIcon.exists()).toBe(true);
      expect(navigateDownIcon.exists()).toBe(true);
      expect(copyInstructions.exists()).toBe(true);
      expect(copyInstructions.text()).toContain("Ctrl + C");
      expect(copyInstructions.text()).toContain("To copy SSHID");
    });

    it("Shows close button", () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');
      expect(closeBtn.exists()).toBe(true);
      expect(closeBtn.text()).toBe("Close");
    });

    it("Closes dialog when close button is clicked", async () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Search functionality", () => {
    beforeEach(async () => {
      const btn = wrapper.find('[data-test="quick-connection-open-btn"]');
      await btn.trigger("click");
      await flushPromises();
    });

    it("Updates filter when search text changes", async () => {
      const searchField = dialog.find('[data-test="search-text"] input');
      await searchField.setValue("my-device");
      await flushPromises();

      const list = wrapper.findComponent({ name: "QuickConnectionList" });
      expect(list.props("filter")).toBe("my-device");
    });

    it("Trims search text", async () => {
      const searchField = dialog.find('[data-test="search-text"] input');
      await searchField.setValue("  device-name  ");
      await flushPromises();

      const list = wrapper.findComponent({ name: "QuickConnectionList" });
      expect(list.props("filter")).toBe("device-name");
    });
  });
});
