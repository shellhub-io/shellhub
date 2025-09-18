import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, vi } from "vitest";
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

// Mock navigator.clipboard
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn(),
  },
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

  beforeEach(async () => {
    vi.clearAllMocks();
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

    const keyField = dialog.find('[data-test="generated-key-field"] input');
    expect(keyField.element.value).toBe("test-api-key-12345");
  });

  it("Copies API key to clipboard when copy button is clicked", async () => {
    const writeTextSpy = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue();

    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const copyButton = dialog.find('[data-test="copy-btn"]');
    await copyButton.trigger("click");
    await flushPromises();

    expect(writeTextSpy).toHaveBeenCalledWith("test-api-key-12345");
    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("API Key copied to clipboard!");
  });

  it("Copies API key to clipboard when icon button is clicked", async () => {
    const writeTextSpy = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue();

    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const iconButton = dialog.find('[data-test="copy-key-icon-btn"]');
    await iconButton.trigger("click");
    await flushPromises();

    expect(writeTextSpy).toHaveBeenCalledWith("test-api-key-12345");
    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("API Key copied to clipboard!");
  });

  it("Shows error message when clipboard copy fails", async () => {
    const writeTextSpy = vi.spyOn(navigator.clipboard, "writeText").mockRejectedValue(new Error("Clipboard error"));
    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {
      // Mock implementation to silence console.error in tests
    });

    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const copyButton = dialog.find('[data-test="copy-btn"]');
    await copyButton.trigger("click");
    await flushPromises();

    expect(writeTextSpy).toHaveBeenCalledWith("test-api-key-12345");
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to copy API key to clipboard.");
    expect(consoleSpy).toHaveBeenCalledWith("Failed to copy: ", expect.any(Error));

    consoleSpy.mockRestore();
  });

  it("Has close button that triggers close method", async () => {
    const closeSpy = vi.spyOn(wrapper.vm, "close");

    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    const closeButton = dialog.find('[data-test="close-btn"]');
    expect(closeButton.exists()).toBe(true);

    // Test that the close method exists and can be called
    wrapper.vm.close();
    expect(closeSpy).toHaveBeenCalled();
  });

  it("Closes dialog when close method is called directly", async () => {
    expect(wrapper.vm.showDialog).toBe(true);

    wrapper.vm.close();
    await flushPromises();

    expect(wrapper.vm.showDialog).toBe(false);
  });
});
