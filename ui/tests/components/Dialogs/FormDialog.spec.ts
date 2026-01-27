import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";

vi.mock("vuetify", async () => {
  const actual = await vi.importActual<typeof import("vuetify")>("vuetify");

  return {
    ...actual,
    useDisplay: () => ({
      smAndDown: { value: false },
      thresholds: {
        value: {
          sm: 600,
          md: 960,
          lg: 1280,
          xl: 1920,
          xxl: 2560,
        },
      },
    }),
  };
});

describe("FormDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof FormDialog>>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = async (
    props: Partial<InstanceType<typeof FormDialog>["$props"]> = {},
    slots: Record<string, string> = {},
  ) => {
    wrapper = mountComponent(FormDialog, {
      props: { modelValue: true, ...props },
      slots: {
        default: "<input data-test='form-input' type='text' />",
        ...slots,
      },
    });
    dialog = new DOMWrapper(document.body).find('[role="dialog"]');
    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  it("Renders the dialog", async () => {
    await mountWrapper();

    expect(wrapper.exists()).toBe(true);
    expect(dialog.exists()).toBe(true);
  });

  describe("WindowDialog integration", () => {
    it("Wraps content in WindowDialog", async () => {
      await mountWrapper();

      const windowDialog = wrapper.findComponent(WindowDialog);
      expect(windowDialog.exists()).toBe(true);
    });

    it("Passes props to WindowDialog", async () => {
      await mountWrapper({
        title: "Form Title",
        description: "Form Description",
        icon: "mdi-form-select",
        iconColor: "primary",
      });

      const windowDialog = wrapper.findComponent(WindowDialog);
      const props = windowDialog.props();
      expect(props.title).toBe("Form Title");
      expect(props.description).toBe("Form Description");
      expect(props.icon).toBe("mdi-form-select");
      expect(props.iconColor).toBe("primary");
    });

    it("Passes showCloseButton prop to WindowDialog", async () => {
      await mountWrapper({ showCloseButton: false });

      const windowDialog = wrapper.findComponent(WindowDialog);
      expect(windowDialog.props().showCloseButton).toBe(false);
    });

    it("Emits close when WindowDialog emits close", async () => {
      await mountWrapper();

      const windowDialog = wrapper.findComponent(WindowDialog);
      windowDialog.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("close")).toBeTruthy();
    });
  });

  describe("Form content", () => {
    it("Renders form element", async () => {
      await mountWrapper();

      const form = dialog.find("form");
      expect(form.exists()).toBe(true);
    });

    it("Renders default slot content inside form", async () => {
      await mountWrapper();

      const formInput = dialog.find('[data-test="form-input"]');
      expect(formInput.exists()).toBe(true);
    });
  });

  describe("Footer buttons", () => {
    it("Shows confirm button when confirmText is provided", async () => {
      await mountWrapper({ confirmText: "Submit" });

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      expect(confirmBtn.exists()).toBe(true);
      expect(confirmBtn.text()).toBe("Submit");
    });

    it("Shows cancel button when cancelText is provided", async () => {
      await mountWrapper({ cancelText: "Cancel" });

      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      expect(cancelBtn.exists()).toBe(true);
      expect(cancelBtn.text()).toBe("Cancel");
    });

    it("Emits confirm when confirm button is clicked", async () => {
      await mountWrapper({ confirmText: "Submit" });

      const confirmBtn = dialog.get('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("confirm")).toBeTruthy();
    });

    it("Emits cancel when cancel button is clicked", async () => {
      await mountWrapper({ cancelText: "Cancel" });

      const cancelBtn = dialog.get('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("cancel")).toBeTruthy();
    });

    it("Disables confirm button when confirmDisabled is true", async () => {
      await mountWrapper({
        confirmText: "Submit",
        confirmDisabled: true,
      });

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      expect(confirmBtn.attributes("disabled")).toBeDefined();
    });

    it("Shows loading state on confirm button", async () => {
      await mountWrapper({
        confirmText: "Submit",
        confirmLoading: true,
      });

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      expect(confirmBtn.html()).toContain("v-btn--loading");
    });

    it("Uses custom data-test attributes", async () => {
      await mountWrapper({
        confirmText: "Submit",
        confirmDataTest: "custom-submit",
        cancelText: "Cancel",
        cancelDataTest: "custom-cancel",
      });

      expect(dialog.find('[data-test="custom-submit"]').exists()).toBe(true);
      expect(dialog.find('[data-test="custom-cancel"]').exists()).toBe(true);
    });

    it("Doesn't show buttons when alert is showing", async () => {
      await mountWrapper({
        confirmText: "Submit",
        alertMessage: "Error message",
      });

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      const cancelBtn = dialog.find('[data-test="cancel-btn"]');

      expect(confirmBtn.exists()).toBe(false);
      expect(cancelBtn.exists()).toBe(false);
    });

    it("Shows alert message when showAlert is true", async () => {
      await mountWrapper({
        alertMessage: "This is an error",
        alertType: "error",
      });

      const alert = dialog.find('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("This is an error");
    });
  });

  describe("Helper text and link", () => {
    it("Shows helper text in footer", async () => {
      await mountWrapper({
        footerHelperText: "Need help?",
      });

      expect(dialog.text()).toContain("Need help?");
    });

    it("Shows helper link when provided", async () => {
      await mountWrapper({
        footerHelperText: "Need help?",
        footerHelperLink: "https://example.com/help",
        footerHelperLinkText: "Read docs",
      });

      const link = dialog.find('a[href="https://example.com/help"]');
      expect(link.exists()).toBe(true);
      expect(link.text()).toContain("Read docs");
    });
  });
});
