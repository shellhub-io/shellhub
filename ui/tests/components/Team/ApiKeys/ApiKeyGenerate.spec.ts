import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { createAxiosError } from "@tests/utils/axiosError";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import ApiKeyGenerate from "@/components/Team/ApiKeys/ApiKeyGenerate.vue";
import useApiKeysStore from "@/store/modules/api_keys";

vi.mock("@/utils/permission", () => ({
  default: () => true,
}));

describe("ApiKeyGenerate", () => {
  let wrapper: VueWrapper<InstanceType<typeof ApiKeyGenerate>>;
  let apiKeysStore: ReturnType<typeof useApiKeysStore>;
  let dialog: DOMWrapper<HTMLElement>;

  const openDialog = async () => {
    const button = wrapper.find('[data-test="api-key-generate-main-btn"]');
    await button.trigger("click");
    await flushPromises();
  };

  const mountWrapper = () => {
    wrapper = mountComponent(ApiKeyGenerate, {
      attachTo: document.body,
    });
    apiKeysStore = useApiKeysStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Rendering", () => {
    it("renders generate button", () => {
      const button = wrapper.find('[data-test="api-key-generate-main-btn"]');
      expect(button.exists()).toBe(true);
    });

    it("shows dialog when button is clicked", async () => {
      await openDialog();

      const formDialog = dialog.find('[data-test="api-key-generate-dialog"]');
      expect(formDialog.exists()).toBe(true);
    });

    it("renders form fields inside dialog", async () => {
      await openDialog();

      expect(dialog.find('[data-test="key-name-text"]').exists()).toBe(true);
      expect(dialog.find('[data-test="api-key-expiration-date"]').exists()).toBe(true);
      expect(dialog.find('[data-test="api-key-role"]').exists()).toBe(true);
    });

    it("renders dialog buttons", async () => {
      await openDialog();

      expect(dialog.find('[data-test="add-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    });
  });

  describe("Form validation", () => {
    it("disables generate button when form is invalid", async () => {
      await openDialog();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      expect(generateBtn.attributes("disabled")).toBeDefined();
    });

    it("enables generate button when form is valid", async () => {
      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      expect(generateBtn.attributes("disabled")).toBeUndefined();
    });
  });

  describe("API key generation", () => {
    it("calls generateApiKey when submitting valid form", async () => {
      vi.mocked(apiKeysStore.generateApiKey).mockResolvedValueOnce("generated-key-123");

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      expect(apiKeysStore.generateApiKey).toHaveBeenCalledWith({
        name: "my-api-key",
        role: "administrator",
        expires_in: 30,
      });
    });

    it("shows success dialog after successful generation", async () => {
      vi.mocked(apiKeysStore.generateApiKey).mockResolvedValueOnce("generated-key-123");

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      const successDialog = dialog.find('[data-test="api-key-success-dialog"]');
      expect(successDialog.exists()).toBe(true);
    });

    it("closes generate dialog after successful generation", async () => {
      vi.mocked(apiKeysStore.generateApiKey).mockResolvedValueOnce("generated-key-123");

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none;");
    });

    it("emits update event when success dialog is closed", async () => {
      vi.mocked(apiKeysStore.generateApiKey).mockResolvedValueOnce("generated-key-123");

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      const closeBtn = dialog.find('[data-test="api-key-success-dialog"] [data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });
  });

  describe("Error handling", () => {
    it("shows error snackbar when generation fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(apiKeysStore.generateApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to generate API Key.");
    });

    it("shows error message for 400 status", async () => {
      const error = createAxiosError(400, "Bad Request");
      vi.mocked(apiKeysStore.generateApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      const alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.text()).toContain("Please provide a name for the API key.");
    });

    it("shows error message for 401 status", async () => {
      const error = createAxiosError(401, "Unauthorized");
      vi.mocked(apiKeysStore.generateApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      const alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.text()).toContain("You are not authorized to create an API key.");
    });

    it("shows error message for 409 status", async () => {
      const error = createAxiosError(409, "Conflict");
      vi.mocked(apiKeysStore.generateApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      const alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.text()).toContain("An API key with the same name already exists.");
    });

    it("shows generic error message for other status codes", async () => {
      const error = createAxiosError(503, "Service Unavailable");
      vi.mocked(apiKeysStore.generateApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      const alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.text()).toContain("An error occurred while generating your API key. Please try again later.");
    });

    it("dismisses error message when alert is dismissed", async () => {
      const error = createAxiosError(409, "Conflict");
      vi.mocked(apiKeysStore.generateApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      let alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(true);

      const closeAlertBtn = alert.find('[data-test="alert-got-it-btn"]');
      await closeAlertBtn.trigger("click");
      await flushPromises();

      alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(false);
    });
  });

  describe("Dialog close", () => {
    it("closes dialog when cancel button is clicked", async () => {
      await openDialog();

      const cancelBtn = dialog.find('[data-test="close-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none;");
    });

    it("resets form when dialog is closed", async () => {
      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const cancelBtn = dialog.find('[data-test="close-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      await openDialog();

      const nameInputAfterReopen = dialog.find('[data-test="key-name-text"] input').element as HTMLInputElement;
      expect(nameInputAfterReopen.value).toBe("");
    });

    it("clears error message when dialog is closed", async () => {
      const error = createAxiosError(409, "Conflict");
      vi.mocked(apiKeysStore.generateApiKey).mockRejectedValueOnce(error);

      await openDialog();

      const nameInput = dialog.find('[data-test="key-name-text"] input');
      await nameInput.setValue("my-api-key");
      await flushPromises();

      const generateBtn = dialog.find('[data-test="add-btn"]');
      await generateBtn.trigger("click");
      await flushPromises();

      let alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(true);

      const cancelBtn = dialog.find('[data-test="close-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      await openDialog();

      alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(false);
    });
  });
});
