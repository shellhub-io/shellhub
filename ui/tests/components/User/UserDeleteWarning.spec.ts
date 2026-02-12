import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import UserDeleteWarning from "@/components/User/UserDeleteWarning.vue";
import { mountComponent } from "@tests/utils/mount";
import { envVariables } from "@/envVariables";

describe("UserDeleteWarning", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserDeleteWarning>>;
  let dialog: DOMWrapper<Element>;
  const originalIsCommunity = envVariables.isCommunity;
  const originalIsEnterprise = envVariables.isEnterprise;

  const mountWrapper = (modelValue = false) => {
    wrapper = mountComponent(UserDeleteWarning, {
      props: { modelValue },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          auth: {
            username: "testuser",
          },
        },
      },
    });

    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => {
    envVariables.isCommunity = true;
    envVariables.isEnterprise = false;
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
    envVariables.isCommunity = originalIsCommunity;
    envVariables.isEnterprise = originalIsEnterprise;
  });

  describe("Component rendering when closed", () => {
    it("does not render dialog when modelValue is false", () => {
      mountWrapper(false);
      expect(dialog.find('[data-test="user-delete-dialog"]').exists()).toBe(false);
    });
  });

  describe("Community instance display", () => {
    beforeEach(async () => {
      envVariables.isCommunity = true;
      envVariables.isEnterprise = false;
      mountWrapper(true);
      await flushPromises();
    });

    it("renders WindowDialog with correct props for Community", () => {
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.exists()).toBe(true);
      expect(windowDialog.props("title")).toBe("Account Deletion");
      expect(windowDialog.props("description")).toBe("CLI Required");
      expect(windowDialog.props("icon")).toBe("mdi-console");
      expect(windowDialog.props("iconColor")).toBe("primary");
    });

    it("displays Community-specific instructions", () => {
      expect(dialog.text()).toContain("Community instances");
      expect(dialog.text()).toContain("user accounts can only be deleted via the CLI");
    });

    it("renders documentation link", () => {
      const docsLink = dialog.find('[data-test="docs-link"]');
      expect(docsLink.exists()).toBe(true);
      expect(docsLink.attributes("href")).toBe(
        "https://docs.shellhub.io/self-hosted/administration#delete-a-user",
      );
      expect(docsLink.attributes("target")).toBe("_blank");
      expect(docsLink.attributes("rel")).toBe("noopener noreferrer");
      expect(docsLink.text()).toContain("administration documentation");
    });

    it("renders CLI command with correct username", () => {
      const commandField = dialog.find('[data-test="copy-command-field"]');
      expect(commandField.exists()).toBe(true);

      const input = commandField.find("input");
      expect(input.attributes("value")).toBe("./bin/cli user delete testuser");
    });

    it("generates correct CLI command for different usernames", async () => {
      wrapper.unmount();
      document.body.innerHTML = "";

      wrapper = mountComponent(UserDeleteWarning, {
        props: { modelValue: true },
        attachTo: document.body,
        piniaOptions: {
          initialState: {
            auth: {
              username: "another-user",
            },
          },
        },
      });

      await flushPromises();
      dialog = new DOMWrapper(document.body);

      const commandField = dialog.find('[data-test="copy-command-field"]');
      const input = commandField.find("input");
      expect(input.attributes("value")).toBe("./bin/cli user delete another-user");
    });

    it("does not display Enterprise-specific content", () => {
      expect(dialog.text()).not.toContain("Enterprise instances");
      expect(dialog.text()).not.toContain("Admin Console");
    });
  });

  describe("Enterprise instance display", () => {
    beforeEach(async () => {
      envVariables.isCommunity = false;
      envVariables.isEnterprise = true;
      mountWrapper(true);
      await flushPromises();
    });

    it("renders WindowDialog with correct props for Enterprise", () => {
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });
      expect(windowDialog.exists()).toBe(true);
      expect(windowDialog.props("title")).toBe("Account Deletion");
      expect(windowDialog.props("description")).toBe("Admin Console Required");
      expect(windowDialog.props("icon")).toBe("mdi-shield-account");
      expect(windowDialog.props("iconColor")).toBe("primary");
    });

    it("displays Enterprise-specific instructions", () => {
      expect(dialog.text()).toContain("Enterprise instances");
      expect(dialog.text()).toContain("user accounts can only be deleted via the Admin Console");
    });

    it("renders Admin Console link", () => {
      const adminLink = dialog.find('a[href="/admin/users"]');
      expect(adminLink.exists()).toBe(true);
      expect(adminLink.attributes("target")).toBe("_blank");
      expect(adminLink.attributes("rel")).toBe("noopener noreferrer");
      expect(adminLink.text()).toBe("Admin Console");
    });

    it("mentions contacting system administrator", () => {
      expect(dialog.text()).toContain("contact your system administrator");
    });

    it("does not display Community-specific content", () => {
      expect(dialog.text()).not.toContain("Community instances");
      expect(dialog.find('[data-test="docs-link"]').exists()).toBe(false);
      expect(dialog.find('[data-test="copy-command-field"]').exists()).toBe(false);
    });
  });

  describe("Dialog close behavior", () => {
    beforeEach(async () => {
      mountWrapper(true);
      await flushPromises();
    });

    it("closes dialog when Close button is clicked", async () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');

      await closeBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")![0]).toEqual([false]);
    });

    it("closes dialog when WindowDialog emits close event", async () => {
      const windowDialog = wrapper.findComponent({ name: "WindowDialog" });

      windowDialog.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")![0]).toEqual([false]);
    });
  });

  describe("Component footer", () => {
    beforeEach(async () => {
      mountWrapper(true);
      await flushPromises();
    });

    it("renders close button in footer", () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');
      expect(closeBtn.exists()).toBe(true);
      expect(closeBtn.text()).toBe("Close");
    });
  });
});
