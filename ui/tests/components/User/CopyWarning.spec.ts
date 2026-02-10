import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import { h } from "vue";
import { useClipboard, useMagicKeys } from "@vueuse/core";
import CopyWarning from "@/components/User/CopyWarning.vue";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";

vi.mock("@vueuse/core", async () => {
  const actual = await vi.importActual<typeof import("@vueuse/core")>("@vueuse/core");
  return {
    ...actual,
    useClipboard: vi.fn(),
    useMagicKeys: vi.fn(),
  };
});

describe("CopyWarning", () => {
  let wrapper: VueWrapper<InstanceType<typeof CopyWarning>>;
  let dialog: DOMWrapper<Element>;
  const mockCopy = vi.fn();
  let mockMagicKeysCallback: ((e: KeyboardEvent) => void) | null = null;

  const mountWrapper = (props = {}) => {
    mockCopy.mockResolvedValue(undefined);

    vi.mocked(useClipboard).mockReturnValue({
      copy: mockCopy,
    } as never);

    vi.mocked(useMagicKeys).mockImplementation(({ onEventFired }) => {
      if (onEventFired) {
        mockMagicKeysCallback = onEventFired as (e: KeyboardEvent) => void;
      }
      return {} as never;
    });

    wrapper = mountComponent(CopyWarning, {
      props,
      attachTo: document.body,
      slots: {
        default: ({ copyText }: { copyText: (text: string) => void }) => h(
          "button",
          {
            "data-test": "copy-btn",
            onClick: () => copyText("test-text"),
          },
          "Copy",
        ),
      },
    });

    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => {
    Object.defineProperty(globalThis, "isSecureContext", {
      writable: true,
      configurable: true,
      value: true,
    });
  });

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
    vi.clearAllMocks();
    mockMagicKeysCallback = null;
  });

  describe("Component rendering", () => {
    it("renders slot content with copyText function", () => {
      mountWrapper();
      expect(dialog.find('[data-test="copy-btn"]').exists()).toBe(true);
    });

    it("does not render dialog initially", () => {
      mountWrapper();
      expect(dialog.find('[data-test="copy-warning-dialog"]').exists()).toBe(false);
    });
  });

  describe("Copy functionality in secure context", () => {
    it("copies text to clipboard when copyText is called", async () => {
      mountWrapper();

      const copyBtn = dialog.find('[data-test="copy-btn"]');
      await copyBtn.trigger("click");
      await flushPromises();

      expect(mockCopy).toHaveBeenCalledWith("test-text");
      expect(mockSnackbar.showInfo).toHaveBeenCalledWith("Successfully copied to clipboard!");
    });

    it("shows custom success message when copiedItem prop is provided", async () => {
      mountWrapper({ copiedItem: "Device ID" });

      const copyBtn = dialog.find('[data-test="copy-btn"]');
      await copyBtn.trigger("click");
      await flushPromises();

      expect(mockCopy).toHaveBeenCalledWith("test-text");
      expect(mockSnackbar.showInfo).toHaveBeenCalledWith("Device ID copied to clipboard!");
    });

    it("exposes copyFn and allows manual copy", async () => {
      mountWrapper();

      await wrapper.vm.copyFn("manual-copy-text");
      await flushPromises();

      expect(mockCopy).toHaveBeenCalledWith("manual-copy-text");
      expect(mockSnackbar.showInfo).toHaveBeenCalledWith("Successfully copied to clipboard!");
    });
  });

  describe("Copy functionality in insecure context", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        writable: true,
        configurable: true,
        value: false,
      });
    });

    it("shows warning dialog instead of copying when not secure", async () => {
      mountWrapper();

      const copyBtn = dialog.find('[data-test="copy-btn"]');
      await copyBtn.trigger("click");
      await flushPromises();

      expect(mockCopy).not.toHaveBeenCalled();
      expect(dialog.find('[data-test="copy-warning-dialog"]').exists()).toBe(true);
      expect(dialog.text()).toContain("Copying is not allowed");
      expect(dialog.text()).toContain("Clipboard access is only permitted on secure (HTTPS) or localhost origins");
    });

    it("closes warning dialog when OK button is clicked", async () => {
      mountWrapper();

      const copyBtn = dialog.find('[data-test="copy-btn"]');
      await copyBtn.trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="copy-warning-dialog"]').exists()).toBe(true);

      const okBtn = dialog.find('[data-test="copy-warning-ok-btn"]');
      await okBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Clipboard API error handling", () => {
    it("shows warning dialog when clipboard API fails", async () => {
      mockCopy.mockRejectedValueOnce(new Error("Clipboard error"));
      mountWrapper();

      const copyBtn = dialog.find('[data-test="copy-btn"]');
      await copyBtn.trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="copy-warning-dialog"]').exists()).toBe(true);
      expect(dialog.text()).toContain("Clipboard access is only permitted");
    });
  });

  describe("Bypass prop", () => {
    it("does nothing when bypass prop is true", async () => {
      mountWrapper({ bypass: true });

      const copyBtn = dialog.find('[data-test="copy-btn"]');
      await copyBtn.trigger("click");
      await flushPromises();

      expect(mockCopy).not.toHaveBeenCalled();
      expect(mockSnackbar.showInfo).not.toHaveBeenCalled();
    });

    it("bypasses keyboard shortcut when bypass prop is true", async () => {
      mountWrapper({ bypass: true, macro: "test-macro" });

      if (mockMagicKeysCallback) {
        const ctrlCEvent = new KeyboardEvent("keydown", {
          key: "c",
          ctrlKey: true,
          bubbles: true,
        });

        mockMagicKeysCallback(ctrlCEvent);
        await flushPromises();
      }

      expect(mockCopy).not.toHaveBeenCalled();
    });
  });

  describe("Keyboard shortcut (macro)", () => {
    it("sets up keyboard listener when macro prop is provided", () => {
      mountWrapper({ macro: "test-macro" });
      expect(useMagicKeys).toHaveBeenCalled();
    });

    it("does not set up keyboard listener when macro prop is not provided", () => {
      vi.mocked(useMagicKeys).mockClear();
      mountWrapper();
      expect(useMagicKeys).not.toHaveBeenCalled();
    });

    it("copies macro text on Ctrl+C keydown", async () => {
      mountWrapper({ macro: "ssh user@device" });

      if (mockMagicKeysCallback) {
        const ctrlCEvent = new KeyboardEvent("keydown", {
          key: "c",
          ctrlKey: true,
          bubbles: true,
        });

        mockMagicKeysCallback(ctrlCEvent);
        await flushPromises();

        expect(mockCopy).toHaveBeenCalledWith("ssh user@device");
      }
    });

    it("prevents default behavior on Ctrl+C when macro is set", async () => {
      mountWrapper({ macro: "test-macro" });

      if (mockMagicKeysCallback) {
        const ctrlCEvent = new KeyboardEvent("keydown", {
          key: "c",
          ctrlKey: true,
          bubbles: true,
          cancelable: true,
        });
        const preventDefaultSpy = vi.spyOn(ctrlCEvent, "preventDefault");

        mockMagicKeysCallback(ctrlCEvent);
        await flushPromises();

        expect(preventDefaultSpy).toHaveBeenCalled();
      }
    });

    it("only triggers on keydown, not keyup", async () => {
      mountWrapper({ macro: "test-macro" });

      if (mockMagicKeysCallback) {
        const keyupEvent = new KeyboardEvent("keyup", {
          key: "c",
          ctrlKey: true,
          bubbles: true,
        });

        mockMagicKeysCallback(keyupEvent);
        await flushPromises();

        expect(mockCopy).not.toHaveBeenCalled();
      }
    });

    it("does not trigger without Ctrl key", async () => {
      mountWrapper({ macro: "test-macro" });

      if (mockMagicKeysCallback) {
        const cEvent = new KeyboardEvent("keydown", {
          key: "c",
          ctrlKey: false,
          bubbles: true,
        });

        mockMagicKeysCallback(cEvent);
        await flushPromises();

        expect(mockCopy).not.toHaveBeenCalled();
      }
    });
  });
});
