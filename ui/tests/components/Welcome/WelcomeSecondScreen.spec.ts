import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import WelcomeSecondScreen from "@/components/Welcome/WelcomeSecondScreen.vue";
import { mountComponent } from "@tests/utils/mount";

describe("WelcomeSecondScreen", () => {
  let wrapper: VueWrapper<InstanceType<typeof WelcomeSecondScreen>>;

  const mountWrapper = (tenantId = "test-tenant-123") => {
    wrapper = mountComponent(WelcomeSecondScreen, {
      piniaOptions: { initialState: { auth: { tenantId } } },
    });
  };

  beforeEach(() => { mountWrapper(); });

  afterEach(() => wrapper?.unmount());

  describe("Component rendering", () => {
    it("renders install agent heading", () => {
      expect(wrapper.text()).toContain("Install ShellHub Agent");
      expect(wrapper.text()).toContain("Connect your device to ShellHub in just one step");
    });

    it("renders download icon", () => {
      const avatar = wrapper.find(".v-avatar");
      expect(avatar.exists()).toBe(true);
      expect(wrapper.html()).toContain("mdi-download");
    });

    it("renders requirements list", () => {
      expect(wrapper.text()).toContain("Requirements:");
      expect(wrapper.text()).toContain("Linux system with curl");
      expect(wrapper.text()).toContain("Internet connection");
      expect(wrapper.text()).toContain("Tries: Docker → Podman → Snap → Standalone");
    });

    it("renders installation alert", () => {
      expect(wrapper.text()).toContain("Installation");
      expect(wrapper.text()).toContain("Ready to install?");
    });

    it("renders waiting for device card", () => {
      expect(wrapper.text()).toContain("Waiting for device...");
      expect(wrapper.text()).toContain("After running the command, your device will appear in the next step");
    });
  });

  describe("Installation command", () => {
    it("renders command with correct tenant ID and origin", () => {
      const commandField = wrapper.findComponent({ name: "CopyCommandField" });
      expect(commandField.exists()).toBe(true);

      // eslint-disable-next-line vue/max-len
      const expectedCommand = `curl -sSf ${window.location.origin}/install.sh | TENANT_ID=test-tenant-123 SERVER_ADDRESS=${window.location.origin} sh`;
      expect(commandField.props("command")).toBe(expectedCommand);
    });

    it("updates command when tenant ID changes", () => {
      wrapper.unmount();
      mountWrapper("different-tenant");

      const commandField = wrapper.findComponent({ name: "CopyCommandField" });
      // eslint-disable-next-line vue/max-len
      const expectedCommand = `curl -sSf ${window.location.origin}/install.sh | TENANT_ID=different-tenant SERVER_ADDRESS=${window.location.origin} sh`;
      expect(commandField.props("command")).toBe(expectedCommand);
    });
  });
});
