import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import Snackbar from "@/components/Snackbar/Snackbar.vue";
import { plugin } from "@/plugins/snackbar";
import { nextTick } from "vue";

describe("Snackbar", () => {
  let wrapper: VueWrapper<InstanceType<typeof Snackbar>>;
  let snackbarDom: DOMWrapper<HTMLElement>;

  const mountWrapper = () => {
    wrapper = mountComponent(Snackbar, { attachTo: document.body });
    snackbarDom = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("Rendering", () => {
    it("Renders v-snackbar component", () => {
      const snackbar = wrapper.findComponent({ name: "v-snackbar" });
      expect(snackbar.exists()).toBe(true);
    });

    it("Has correct location prop", () => {
      const snackbar = wrapper.findComponent({ name: "v-snackbar" });
      expect(snackbar.props("location")).toBe("top center");
    });

    it("Has correct timeout prop", () => {
      const snackbar = wrapper.findComponent({ name: "v-snackbar" });
      expect(snackbar.props("timeout")).toBe(4000);
    });

    it("Has correct transition prop", () => {
      const snackbar = wrapper.findComponent({ name: "v-snackbar" });
      expect(snackbar.props("transition")).toBe("slide-x-transition");
    });
  });

  describe("Message display", () => {
    it.each([
      { type: "success", method: "showSuccess", message: "Operation successful" },
      { type: "error", method: "showError", message: "Operation failed" },
      { type: "info", method: "showInfo", message: "Information message" },
      { type: "warning", method: "showWarning", message: "Warning message" },
    ])("Displays $type message", async ({ method, message }) => {
      plugin[method as keyof typeof plugin](message);
      await nextTick();
      expect(snackbarDom.text()).toBe(message);
    });
  });

  describe("Color based on type", () => {
    it.each([
      { type: "success", method: "showSuccess", color: "success" },
      { type: "error", method: "showError", color: "error" },
      { type: "info", method: "showInfo", color: "info" },
      { type: "warning", method: "showWarning", color: "warning" },
    ])("Shows $color color for $type type", async ({ method, color }) => {
      plugin[method as keyof typeof plugin]("Test");
      await nextTick();
      const snackbar = wrapper.findComponent({ name: "v-snackbar" });
      expect(snackbar.props("color")).toBe(color);
    });
  });

  describe("Show/Hide behavior", () => {
    it("Hides snackbar after timeout", async () => {
      vi.useFakeTimers();
      plugin.showSuccess("Test message");
      await nextTick();

      const snackbar = wrapper.findComponent({ name: "v-snackbar" });
      expect(snackbar.props("modelValue")).toBe(true);

      vi.advanceTimersByTime(4000);
      await flushPromises();
      expect(snackbar.props("modelValue")).toBe(false);

      vi.useRealTimers();
    });

    it("Shows snackbar when plugin triggers show", async () => {
      plugin.showSuccess("Test message");
      await nextTick();
      const snackbar = wrapper.findComponent({ name: "v-snackbar" });
      expect(snackbar.props("modelValue")).toBe(true);
    });

    it("Does not close prematurely when triggered multiple times quickly", async () => {
      vi.useFakeTimers();
      const snackbar = wrapper.findComponent({ name: "v-snackbar" });

      plugin.showSuccess("First message");
      vi.advanceTimersByTime(2000);
      plugin.showError("Second message");
      await nextTick();

      // Still visible — the first timeout was cancelled by the second call
      expect(snackbar.props("modelValue")).toBe(true);

      vi.advanceTimersByTime(4000);
      await flushPromises();
      expect(snackbar.props("modelValue")).toBe(false);

      vi.useRealTimers();
    });
  });
});
