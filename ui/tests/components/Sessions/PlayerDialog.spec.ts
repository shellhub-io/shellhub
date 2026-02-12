import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import PlayerDialog from "@/components/Sessions/PlayerDialog.vue";

const mockPlayer = {
  play: vi.fn(),
  pause: vi.fn(),
  seek: vi.fn(),
  dispose: vi.fn(),
  getCurrentTime: vi.fn().mockResolvedValue(0),
  getDuration: vi.fn().mockResolvedValue(100),
  addEventListener: vi.fn(),
};

vi.mock("asciinema-player", () => ({
  create: vi.fn(() => mockPlayer),
}));

const mockLogs = '{"version":2,"width":80,"height":24}';

describe("PlayerDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof PlayerDialog>>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = (modelValue = true, logs = mockLogs) => {
    wrapper = mountComponent(PlayerDialog, {
      props: {
        modelValue,
        logs,
      },
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

    it("Renders dialog with fullscreen mode", () => {
      const dialog = wrapper.findComponent({ name: "BaseDialog" });
      expect(dialog.props("forceFullscreen")).toBe(true);
    });

    it("Renders dialog as not scrollable", () => {
      const dialog = wrapper.findComponent({ name: "BaseDialog" });
      expect(dialog.props("scrollable")).toBeUndefined();
    });

    it("Disables dialog transition", () => {
      const dialog = wrapper.findComponent({ name: "BaseDialog" });
      expect(dialog.props("transition")).toBeUndefined();
    });
  });

  describe("Close button", () => {
    it("Renders close button", () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');
      expect(closeBtn.exists()).toBe(true);
    });

    it("Close button has correct icon", () => {
      const closeBtnIcon = dialog.find('[data-test="close-btn"] i');
      expect(closeBtnIcon.classes()).toContain("mdi-close");
    });

    it("Closes dialog when close button is clicked", async () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");

      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none");
    });
  });

  describe("Player component", () => {
    it("Renders Player component", () => {
      const player = wrapper.findComponent({ name: "Player" });
      expect(player.exists()).toBe(true);
    });

    it("Passes logs prop to Player", () => {
      const player = wrapper.findComponent({ name: "Player" });
      expect(player.props("logs")).toBe(mockLogs);
    });

    it("Closes dialog when Player emits close event", async () => {
      const player = wrapper.findComponent({ name: "Player" });
      await player.vm.$emit("close");

      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none");
    });
  });
});
