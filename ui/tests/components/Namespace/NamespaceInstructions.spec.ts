import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import NamespaceInstructions from "@/components/Namespace/NamespaceInstructions.vue";
import { envVariables } from "@/envVariables";

vi.mock("@/envVariables", () => ({ envVariables: { isCommunity: true } }));

describe("NamespaceInstructions", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceInstructions>>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = () => {
    wrapper = mountComponent(NamespaceInstructions, {
      props: { modelValue: true },
      attachTo: document.body,
    });
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Dialog display", () => {
    it("Renders WindowDialog component", () => {
      expect(dialog.exists()).toBe(true);
    });

    it("Shows correct title", () => {
      const titlebar = dialog.find('[data-test="window-dialog-titlebar"]');
      expect(titlebar.text()).toBe("You have no namespaces associated");
    });

    it("Shows warning icon", () => {
      const titlebarAvatar = dialog.find('[data-test="window-dialog-titlebar"] .v-avatar');
      expect(titlebarAvatar.find("i").classes()).toContain("mdi-folder-alert");
      expect(titlebarAvatar.classes()).toContain("text-warning");
    });
  });

  describe("Content", () => {
    it("Displays main message about namespace requirement", () => {
      expect(dialog.text()).toContain(
        "In order to use ShellHub, you must have a namespace associated with your account or join an existing one.",
      );
    });

    it("Shows CLI instructions in community version", () => {
      expect(dialog.text()).toContain("CLI script");
      expect(dialog.find("#cli-instructions").exists()).toBe(true);
    });

    it("Displays link to ShellHub Administration Guide", () => {
      const link = dialog.find('[data-test="docs-link"]');
      expect(link.exists()).toBe(true);
      expect(link.text()).toContain("ShellHub Administration Guide");
      expect(link.attributes("href")).toBe("https://docs.shellhub.io/self-hosted/administration");
      expect(link.attributes("target")).toBe("_blank");
    });
  });

  describe("NamespaceAdd integration", () => {
    it("Renders NamespaceAdd component", () => {
      const namespaceAdd = wrapper.findComponent({ name: "NamespaceAdd" });
      expect(namespaceAdd.exists()).toBe(true);
    });

    it("Opens NamespaceAdd when Add Namespace button is clicked in non-community version", async () => {
      wrapper?.unmount();
      envVariables.isCommunity = false;
      mountWrapper();

      const addButton = dialog.find('[data-test="add-namespace-btn"]');
      await addButton.trigger("click");
      await flushPromises();

      const namespaceAdd = wrapper.findComponent({ name: "NamespaceAdd" });
      expect(namespaceAdd.props("modelValue")).toBe(true);
    });
  });

  describe("Model value", () => {
    it("Closes dialog when close event is emitted", async () => {
      mountWrapper();

      const dialog = wrapper.findComponent({ name: "WindowDialog" });
      dialog.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });
  });
});
