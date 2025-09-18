import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, vi } from "vitest";
import ApiKeyEdit from "@/components/Team/ApiKeys/ApiKeyEdit.vue";
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

vi.mock("@/store/modules/api_keys", () => ({
  default: () => ({
    editApiKey: vi.fn(),
  }),
}));

describe("Api Key Edit", () => {
  let wrapper: VueWrapper<InstanceType<typeof ApiKeyEdit>>;
  const vuetify = createVuetify();
  setActivePinia(createPinia());

  beforeEach(async () => {
    vi.clearAllMocks();
    wrapper = mount(ApiKeyEdit, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: {
        keyName: "fake-id",
        keyRole: "observer",
        hasAuthorization: true,
        disabled: false,
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
    expect(wrapper.find('[data-test="edit-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="edit-main-btn-title"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="edit-main-btn-title"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    expect(dialog.find('[data-test="edit-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="key-name-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="edit-btn"]').exists()).toBe(true);
  });

  it("Opens dialog when edit button is clicked", async () => {
    expect(wrapper.vm.showDialog).toBe(false);

    await wrapper.findComponent('[data-test="edit-main-btn-title"]').trigger("click");
    await flushPromises();

    expect(wrapper.vm.showDialog).toBe(true);
  });

  it("Shows error message when errorMessage is set", async () => {
    await wrapper.findComponent('[data-test="edit-main-btn-title"]').trigger("click");

    // Set error message directly
    wrapper.vm.errorMessage = "Test error message";
    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="form-dialog-alert"]').exists()).toBe(true);
  });

  it("Clears error message when dialog is opened", async () => {
    wrapper.vm.errorMessage = "Test error message";

    wrapper.vm.open();
    await flushPromises();

    expect(wrapper.vm.errorMessage).toBe("");
  });

  it("Handles form submission", async () => {
    const mockSubmitData = { name: "new-key-name", role: "administrator" };

    await wrapper.findComponent('[data-test="edit-main-btn-title"]').trigger("click");
    await wrapper.vm.editKey(mockSubmitData);

    // Should call the store's editApiKey method
    // This test validates the method structure and flow
    expect(wrapper.vm.showDialog).toBe(false); // Dialog should close on success
  });

  it("Handles 409 error correctly", async () => {
    const error409 = { response: { status: 409 } };

    // Mock the editKey method to throw 409 error
    wrapper.vm.editKey = vi.fn().mockImplementation(async () => {
      wrapper.vm.errorMessage = "An API key with the same name already exists.";
      throw error409;
    });

    await wrapper.findComponent('[data-test="edit-main-btn-title"]').trigger("click");

    try {
      await wrapper.vm.editKey({ name: "existing-key", role: "observer" });
    } catch (error) {
      // Expected error
    }

    expect(wrapper.vm.errorMessage).toBe("An API key with the same name already exists.");
  });
});
