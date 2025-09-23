import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, vi } from "vitest";
import ApiKeyGenerate from "@/components/Team/ApiKeys/ApiKeyGenerate.vue";
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

vi.mock("@/utils/permission", () => ({
  default: () => true,
}));

vi.mock("@/store/modules/api_keys", () => ({
  default: () => ({
    generateApiKey: vi.fn().mockResolvedValue("new-generated-api-key"),
  }),
}));

describe("Api Key Generate", () => {
  let wrapper: VueWrapper<InstanceType<typeof ApiKeyGenerate>>;
  const vuetify = createVuetify();
  setActivePinia(createPinia());

  beforeEach(async () => {
    vi.clearAllMocks();
    wrapper = mount(ApiKeyGenerate, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    expect(dialog.find('[data-test="api-key-generate-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="key-name-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="api-key-expiration-date"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="add-btn"]').exists()).toBe(true);
  });

  it("Opens dialog when button is clicked", async () => {
    expect(wrapper.vm.showDialog).toBe(false);

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");
    await flushPromises();

    expect(wrapper.vm.showDialog).toBe(true);
  });

  it("Shows error message when errorMessage is set", async () => {
    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");

    // Set error message directly
    wrapper.vm.errorMessage = "Test error message";
    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="form-dialog-alert"]').exists()).toBe(true);
  });

  it("Clears error message when dialog is closed", async () => {
    wrapper.vm.errorMessage = "Test error message";

    wrapper.vm.close();
    await flushPromises();

    expect(wrapper.vm.errorMessage).toBe("");
  });

  it("Handles form submission", async () => {
    const mockSubmitData = { name: "test-key", role: "administrator" as const, expires_in: 30 };

    await wrapper.findComponent('[data-test="api-key-generate-main-btn"]').trigger("click");
    await wrapper.vm.generateKey(mockSubmitData);

    // Should call the store's generateApiKey method
    // This test validates the method structure and flow
    expect(wrapper.vm.showDialog).toBe(false); // Dialog should close on success
  });
});
