import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, vi } from "vitest";
import type { Mock } from "vitest";
import { useClipboard } from "@vueuse/core";
import ApiKeySuccess from "@/components/Team/ApiKeys/ApiKeySuccess.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
  showWarning: vi.fn(),
  showInfo: vi.fn(),
};

vi.mock("@/helpers/snackbar", () => ({
  default: () => mockSnackbar,
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

const mockCopy = vi.fn();
(useClipboard as unknown as Mock).mockReturnValue({
  copy: mockCopy,
});

Object.defineProperty(globalThis, "isSecureContext", {
  writable: true,
  configurable: true,
  value: true,
});

describe("Api Key Success", () => {
  let wrapper: VueWrapper<InstanceType<typeof ApiKeySuccess>>;
  const vuetify = createVuetify();
  setActivePinia(createPinia());

  const defaultProps = {
    modelValue: true,
    apiKey: "test-api-key-12345",
    keyName: "test-key",
  };

  beforeEach(() => {
    wrapper = mount(ApiKeySuccess, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: defaultProps,
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components correctly", async () => {
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    expect(dialog.find('[data-test="api-key-success-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="generated-key-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="copy-key-icon-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="copy-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("Displays the API key in the text field", async () => {
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const keyField = dialog.find('[data-test="generated-key-field"] input').element as HTMLInputElement;
    expect(keyField.value).toBe("test-api-key-12345");
  });

  it("Copies API key to clipboard when copy button is clicked", async () => {
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const copyButton = dialog.find('[data-test="copy-btn"]');
    await copyButton.trigger("click");
    await flushPromises();

    expect(mockCopy).toHaveBeenCalledWith("test-api-key-12345");
    expect(mockSnackbar.showInfo).toHaveBeenCalledWith("API Key copied to clipboard!");
  });

  it("Copies API key to clipboard when icon button is clicked", async () => {
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const iconButton = dialog.find('[data-test="copy-key-icon-btn"]');
    await iconButton.trigger("click");
    await flushPromises();

    expect(mockCopy).toHaveBeenCalledWith("test-api-key-12345");
    expect(mockSnackbar.showInfo).toHaveBeenCalledWith("API Key copied to clipboard!");
  });

  it("Shows error message when clipboard copy fails", async () => {
    mockCopy.mockRejectedValueOnce(new Error("Clipboard error"));

    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const copyButton = dialog.find('[data-test="copy-btn"]');
    await copyButton.trigger("click");
    await flushPromises();

    expect(mockCopy).toHaveBeenCalledWith("test-api-key-12345");

    const warningDialog = dialog.find('[data-test="copy-warning-dialog"]');
    expect(warningDialog.exists()).toBe(true);
  });

  it("Has close button that triggers close method", async () => {
    expect(wrapper.vm.showDialog).toBe(true);

    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const closeButton = dialog.find('[data-test="close-btn"]');
    expect(closeButton.exists()).toBe(true);

    wrapper.vm.close();
    expect(wrapper.vm.showDialog).toBe(false);
  });
});
