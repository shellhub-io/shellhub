import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import type { Mock } from "vitest";
import { createVuetify } from "vuetify";
import { h } from "vue";
import { useClipboard } from "@vueuse/core";
import CopyWarning from "@/components/User/CopyWarning.vue";

const mockShowInfo = vi.fn();

vi.mock("@/helpers/snackbar", () => ({
  default: () => ({
    showInfo: mockShowInfo,
  }),
}));

vi.mock("@vueuse/core", async () => {
  const actual = await vi.importActual<typeof import("@vueuse/core")>("@vueuse/core");
  return {
    ...actual,
    useClipboard: vi.fn(() => ({
      copy: vi.fn().mockResolvedValue(undefined),
    })),
    useMagicKeys: vi.fn(),
  };
});

describe("CopyWarning.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof CopyWarning>>;
  const vuetify = createVuetify();
  const mockCopy = vi.fn();

  beforeEach(() => {
    vi.resetAllMocks();

    Object.defineProperty(globalThis, "isSecureContext", {
      writable: true,
      configurable: true,
      value: true,
    });

    (useClipboard as unknown as Mock).mockReturnValue({
      copy: mockCopy,
    });
  });

  const mountComponent = (props = {}) => mount(CopyWarning, {
    props,
    global: {
      plugins: [vuetify],
    },
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

  it("copies to clipboard when in secure context", async () => {
    wrapper = mountComponent({ copiedItem: "Something" });

    const button = wrapper.get('[data-test="copy-btn"]');
    await button.trigger("click");
    await flushPromises();

    expect(mockCopy).toHaveBeenCalledWith("test-text");
    expect(mockShowInfo).toHaveBeenCalledWith("Something copied to clipboard!");
  });

  it("shows dialog when not in secure context", async () => {
    (globalThis as unknown as { isSecureContext: boolean }).isSecureContext = false;

    wrapper = mountComponent();
    const button = wrapper.get('[data-test="copy-btn"]');
    await button.trigger("click");
    await flushPromises();

    const dialog = new DOMWrapper(document.body.querySelector(".v-dialog"));
    expect(dialog.text()).toContain("Copying is not allowed");
  });

  it("shows fallback message when copiedItem is not defined", async () => {
    wrapper = mountComponent();

    const button = wrapper.get('[data-test="copy-btn"]');
    await button.trigger("click");
    await flushPromises();

    expect(mockCopy).toHaveBeenCalledWith("test-text");
    expect(mockShowInfo).toHaveBeenCalledWith("Successfully copied to clipboard!");
  });

  it("exposes copyFn and can be triggered manually", async () => {
    wrapper = mountComponent();

    await wrapper.vm.copyFn("manual-copy");
    await flushPromises();

    expect(mockCopy).toHaveBeenCalledWith("manual-copy");
  });

  it("shows dialog when clipboard API fails", async () => {
    mockCopy.mockRejectedValueOnce(new Error("Clipboard error"));

    wrapper = mountComponent();
    const button = wrapper.get('[data-test="copy-btn"]');
    await button.trigger("click");
    await flushPromises();

    const dialog = new DOMWrapper(document.body.querySelector(".v-dialog"));
    expect(dialog.text()).toContain("Clipboard access is only permitted");
  });
});
