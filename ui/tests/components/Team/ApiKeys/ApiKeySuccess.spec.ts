import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises, DOMWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import ApiKeySuccess from "@/components/Team/ApiKeys/ApiKeySuccess.vue";

describe("ApiKeySuccess", () => {
  let wrapper: VueWrapper<InstanceType<typeof ApiKeySuccess>>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = ({ apiKey = "generated-api-key-123", modelValue = true } = {}) => {
    wrapper = mountComponent(ApiKeySuccess, {
      props: {
        apiKey,
        modelValue,
      },
      attachTo: document.body,
    });
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Rendering", () => {
    it("shows success dialog when modelValue is true", () => {
      const successDialog = dialog.find('[data-test="api-key-success-dialog"]');
      expect(successDialog.exists()).toBe(true);
    });

    it("hides dialog when modelValue is false", () => {
      wrapper.unmount();
      mountWrapper({ modelValue: false });

      const successDialog = dialog.find('[data-test="api-key-success-dialog"]');
      expect(successDialog.exists()).toBe(false);
    });

    it("displays the generated API key", () => {
      const keyField = dialog.find('[data-test="generated-key-field"] input').element as HTMLInputElement;
      expect(keyField.value).toBe("generated-api-key-123");
    });

    it("renders generated key field as readonly", () => {
      const keyField = dialog.find('[data-test="generated-key-field"] input');
      expect(keyField.attributes("readonly")).toBeDefined();
    });

    it("renders copy and close buttons", () => {
      expect(dialog.find('[data-test="copy-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    });

    it("renders copy icon button inside field", () => {
      expect(dialog.find('[data-test="copy-key-icon-btn"]').exists()).toBe(true);
    });
  });

  describe("Dialog actions", () => {
    it("emits update:modelValue with false when close button is clicked", async () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });

    it("closes dialog when close button is clicked", async () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      // Re-mount with modelValue false to simulate the v-model update
      wrapper.unmount();
      mountWrapper({ modelValue: false });

      const successDialog = dialog.find('[data-test="api-key-success-dialog"]');
      expect(successDialog.exists()).toBe(false);
    });
  });
});
