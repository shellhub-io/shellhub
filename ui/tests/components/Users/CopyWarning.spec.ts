import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, afterEach, vi } from "vitest";
import { h } from "vue";
import CopyWarning from "@/components/User/CopyWarning.vue";

describe("CopyWarning", () => {
  let wrapper: VueWrapper<InstanceType<typeof CopyWarning>>;
  let dialog: DOMWrapper<Element>;
  const vuetify = createVuetify();

  const mountComponent = (props = {}) => mount(CopyWarning, {
    props: {
      onSuccess: vi.fn(),
      ...props,
    },
    global: {
      plugins: [vuetify],
    },
    slots: {
      default: ({ copyText }: { copyText: (text: string) => void }) => h("button", {
        "data-test": "copy-btn",
        onClick: () => copyText("copy-text"),
      }, "Copy"),
    },
    attachTo: document.body,
  });

  beforeEach(() => {
    vi.resetAllMocks();

    Object.defineProperty(global.navigator, "clipboard", {
      value: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
      writable: true,
    });

    Object.defineProperty(window, "location", {
      value: {
        protocol: "https:",
        hostname: "localhost",
      },
      configurable: true,
    });
  });

  afterEach(() => {
    document.body.innerHTML = "";
  });

  it("copies to clipboard when secure", async () => {
    const onSuccess = vi.fn();

    wrapper = mount(CopyWarning, {
      props: { onSuccess },
      global: {
        plugins: [vuetify],
      },
      slots: {
        default: ({ copyText }: { copyText: (text: string) => void }) => h("button", {
          "data-test": "copy-btn",
          onClick: () => copyText("copied-text"),
        }, "Copy"),
      },
      attachTo: document.body,
    });

    await wrapper.find('[data-test="copy-btn"]').trigger("click");
    await flushPromises();

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith("copied-text");
    expect(onSuccess).toHaveBeenCalled();
  });

  it("shows dialog when insecure", async () => {
    Object.defineProperty(window, "location", {
      value: {
        protocol: "http:",
        hostname: "unsecure.com",
      },
      configurable: true,
    });

    wrapper = mountComponent();
    dialog = new DOMWrapper(document.body);

    await wrapper.find('[data-test="copy-btn"]').trigger("click");
    await flushPromises();

    expect(dialog.text()).toContain("Clipboard access is only permitted");
  });

  it("executes macro copy on ctrl + c", async () => {
    const onSuccess = vi.fn();
    const macro = "macro@example.com";

    wrapper = mount(CopyWarning, {
      props: {
        onSuccess,
        macro,
      },
      global: {
        plugins: [vuetify],
      },
      attachTo: document.body,
    });

    const event = new KeyboardEvent("keydown", {
      key: "c",
      ctrlKey: true,
      bubbles: true,
    });

    window.dispatchEvent(event);
    await flushPromises();

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(macro);
    expect(onSuccess).toHaveBeenCalled();
  });
});
