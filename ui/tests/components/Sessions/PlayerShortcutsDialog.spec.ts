import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import PlayerShortcutsDialog from "@/components/Sessions/PlayerShortcutsDialog.vue";

describe("PlayerShortcutsDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof PlayerShortcutsDialog>>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = (modelValue = true) => {
    wrapper = mountComponent(PlayerShortcutsDialog, {
      props: { modelValue },
      attachTo: document.body,
    });
    dialog = new DOMWrapper(document.body).find('[role="dialog"]');
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("Dialog rendering", () => {
    it("Renders dialog when modelValue is true", () => {
      const dialog = wrapper.findComponent({ name: "BaseDialog" });
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("modelValue")).toBe(true);
    });

    it("Does not render dialog when modelValue is false", () => {
      wrapper.unmount();
      mountWrapper(false);

      const dialog = wrapper.findComponent({ name: "BaseDialog" });
      expect(dialog.props("modelValue")).toBe(false);
    });

    it("Shows dialog title", () => {
      const title = dialog.find(".v-card-title");
      expect(title.text()).toBe("Keyboard Shortcuts");
    });
  });

  describe("Keyboard shortcuts display", () => {
    it("Shows space key shortcut for pause/resume", () => {
      const shortcuts = dialog.findAll(".shortcut");
      expect(shortcuts[0].text()).toContain("space");
      expect(shortcuts[0].text()).toContain("pause / resume");
    });

    it("Shows f key shortcut for fullscreen", () => {
      const shortcuts = dialog.findAll(".shortcut");
      expect(shortcuts[1].text()).toContain("f");
      expect(shortcuts[1].text()).toContain("toggle fullscreen mode");
    });

    it("Shows arrow keys for rewind/fast-forward by 5 seconds", () => {
      const shortcuts = dialog.findAll(".shortcut");
      expect(shortcuts[2].text()).toContain("←");
      expect(shortcuts[2].text()).toContain("→");
      expect(shortcuts[2].text()).toContain("rewind / fast-forward by 5 seconds");
    });

    it("Shows Shift + arrow keys for rewind/fast-forward by 10%", () => {
      const shortcuts = dialog.findAll(".shortcut");
      expect(shortcuts[3].text()).toContain("Shift");
      expect(shortcuts[3].text()).toContain("←");
      expect(shortcuts[3].text()).toContain("→");
      expect(shortcuts[3].text()).toContain("rewind / fast-forward by 10%");
    });

    it("Shows number keys for jumping to percentages", () => {
      const shortcuts = dialog.findAll(".shortcut");
      expect(shortcuts[4].text()).toContain("0");
      expect(shortcuts[4].text()).toContain("9");
      expect(shortcuts[4].text()).toContain("jump to 0%, 10%, 20% ... 90%");
    });

    it("Shows comma/period keys for stepping frames", () => {
      const shortcuts = dialog.findAll(".shortcut");
      expect(shortcuts[5].text()).toContain(",");
      expect(shortcuts[5].text()).toContain(".");
      expect(shortcuts[5].text()).toContain("step back / forward, a frame at a time");
      expect(shortcuts[5].text()).toContain("only when paused");
    });
  });
});
