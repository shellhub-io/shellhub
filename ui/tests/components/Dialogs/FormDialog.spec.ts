import { createVuetify } from "vuetify";
import { DOMWrapper, mount, VueWrapper, flushPromises } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
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

type FormDialogWrapper = VueWrapper<InstanceType<typeof FormDialog>>;

describe("FormDialog", () => {
  const vuetify = createVuetify();
  let wrapper: FormDialogWrapper;
  let dialogDom: DOMWrapper<HTMLElement>;

  const mountWrapper = (
    props: Partial<InstanceType<typeof FormDialog>["$props"]> = {},
    slots: Record<string, string> = {},
  ) => mount(FormDialog, {
    global: { plugins: [vuetify] },
    props: { modelValue: true, ...props },
    slots: {
      default: "<input data-test='form-input' type='text' />",
      ...slots,
    },
    attachTo: document.body,
  });

  beforeEach(async () => {
    document.body.innerHTML = "";
    wrapper = mountWrapper();
    dialogDom = new DOMWrapper(document.body);
    await flushPromises();
  });

  afterEach(() => {
    vi.clearAllMocks();
    wrapper.unmount();
    document.body.innerHTML = "";
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(dialogDom.html()).toMatchSnapshot();
  });

  describe("WindowDialog integration", () => {
    it("Wraps content in WindowDialog", async () => {
      const windowDialog = wrapper.findComponent(WindowDialog);
      expect(windowDialog.exists()).toBe(true);
    });

    it("Passes title, description and icon props to WindowDialog", async () => {
      wrapper = mountWrapper({
        title: "Form Title",
        description: "Form Description",
        icon: "mdi-form-select",
        iconColor: "primary",
      });
      await flushPromises();

      const windowDialog = wrapper.findComponent(WindowDialog);
      const props = windowDialog.props();
      expect(props.title).toBe("Form Title");
      expect(props.description).toBe("Form Description");
      expect(props.icon).toBe("mdi-form-select");
      expect(props.iconColor).toBe("primary");
    });

    it("Passes showCloseButton prop to WindowDialog", async () => {
      wrapper = mountWrapper({ showCloseButton: false });
      await flushPromises();

      const windowDialog = wrapper.findComponent(WindowDialog);
      expect(windowDialog.props().showCloseButton).toBe(false);
    });

    it("Emits close when WindowDialog emits close", async () => {
      const windowDialog = wrapper.findComponent(WindowDialog);
      expect(windowDialog.exists()).toBe(true);

      windowDialog.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("close")).toBeTruthy();
    });

    it("Passes threshold and forceFullscreen props to WindowDialog", async () => {
      wrapper = mountWrapper({ threshold: "md", forceFullscreen: true });
      await flushPromises();

      const windowDialog = wrapper.findComponent(WindowDialog);
      const props = windowDialog.props();
      expect(props.threshold).toBe("md");
      expect(props.forceFullscreen).toBe(true);
    });
  });

  describe("Form content", () => {
    it("Renders form element", async () => {
      const form = dialogDom.find("form");
      expect(form.exists()).toBe(true);
    });

    it("Renders default slot content inside form", async () => {
      const formInput = dialogDom.find('[data-test="form-input"]');
      expect(formInput.exists()).toBe(true);
    });
  });

  describe("Footer buttons", () => {
    it("Shows confirm button when confirmText is provided", async () => {
      wrapper = mountWrapper({ confirmText: "Submit" });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const confirmBtn = wrapper.findComponent('[data-test="confirm-btn"]');
      expect(confirmBtn.exists()).toBe(true);
      expect(confirmBtn.text()).toBe("Submit");
    });

    it("Shows cancel button with default text 'Close'", async () => {
      wrapper = mountWrapper();
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const cancelBtn = wrapper.findComponent('[data-test="cancel-btn"]');
      expect(cancelBtn.exists()).toBe(true);
      expect(cancelBtn.text()).toBe("Close");
    });

    it("Emits confirm when confirm button is clicked", async () => {
      wrapper = mountWrapper({ confirmText: "Submit" });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const confirmBtn = wrapper.findComponent('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("confirm")).toBeTruthy();
      expect(wrapper.emitted("confirm")).toHaveLength(1);
    });

    it("Disables confirm button when confirmDisabled is true", async () => {
      wrapper = mountWrapper({
        confirmText: "Submit",
        confirmDisabled: true,
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const confirmBtn = wrapper.findComponent('[data-test="confirm-btn"]');
      expect(confirmBtn.attributes("disabled")).toBeDefined();
    });

    it("Shows loading state on confirm button when confirmLoading is true", async () => {
      wrapper = mountWrapper({
        confirmText: "Submit",
        confirmLoading: true,
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const confirmBtn = dialogDom.find('[data-test="confirm-btn"]');
      expect(confirmBtn.exists()).toBe(true);
      expect(confirmBtn.html()).toContain("v-btn--loading");
    });

    it("Uses custom data-test attributes when provided", async () => {
      wrapper = mountWrapper({
        confirmText: "Submit",
        confirmDataTest: "custom-submit",
        cancelText: "Cancel",
        cancelDataTest: "custom-cancel",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      expect(dialogDom.find('[data-test="custom-submit"]').exists()).toBe(true);
      expect(dialogDom.find('[data-test="custom-cancel"]').exists()).toBe(true);
    });
  });

  describe("Footer helper text", () => {
    it("Shows footer helper text when provided", async () => {
      wrapper = mountWrapper({
        footerHelperText: "Need help? Visit our docs",
        confirmText: "Submit",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      expect(dialogDom.text()).toContain("Need help? Visit our docs");
    });

    it("Renders footer helper link when provided", async () => {
      wrapper = mountWrapper({
        footerHelperText: "Need help?",
        footerHelperLinkText: "Read docs",
        footerHelperLink: "https://example.com/docs",
        confirmText: "Submit",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const link = dialogDom.find("a[href='https://example.com/docs']");
      expect(link.exists()).toBe(true);
      expect(link.text()).toContain("Read docs");
    });

    it("Opens link in new tab by default", async () => {
      wrapper = mountWrapper({
        footerHelperText: "Help",
        footerHelperLinkText: "Docs",
        footerHelperLink: "https://example.com",
        confirmText: "Submit",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const link = dialogDom.find("a");
      expect(link.attributes("target")).toBe("_blank");
      expect(link.attributes("rel")).toBe("noopener noreferrer");
    });

    it("Opens link in same tab when footerHelperTarget is _self", async () => {
      wrapper = mountWrapper({
        footerHelperText: "Help",
        footerHelperLinkText: "Docs",
        footerHelperLink: "https://example.com",
        footerHelperTarget: "_self",
        confirmText: "Submit",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const link = dialogDom.find("a");
      expect(link.attributes("target")).toBe("_self");
    });

    it("Shows only helper text when no buttons are present", async () => {
      wrapper = mountWrapper({
        footerHelperText: "Information text",
        confirmText: "",
        cancelText: "",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      expect(dialogDom.text()).toContain("Information text");
      expect(dialogDom.findComponent('[data-test="confirm-btn"]').exists()).toBe(false);
      expect(dialogDom.findComponent('[data-test="cancel-btn"]').exists()).toBe(false);
    });
  });

  describe("Alert system", () => {
    it("Shows alert when alertMessage is provided", async () => {
      wrapper = mountWrapper({
        confirmText: "Submit",
        alertMessage: "Something went wrong",
      });

      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const alert = wrapper.findComponent('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("Something went wrong");
    });

    it("Shows alert with custom type", async () => {
      wrapper = mountWrapper({
        confirmText: "Submit",
        alertMessage: "Success message",
        alertType: "success",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const alert = wrapper.findComponent('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.html()).toContain("success");
    });

    it("Shows alert with custom button text", async () => {
      wrapper = mountWrapper({
        confirmText: "Submit",
        alertMessage: "Error occurred",
        alertButtonText: "Understood",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const alertBtn = wrapper.findComponent('[data-test="alert-got-it-btn"]');
      expect(alertBtn.exists()).toBe(true);
      expect(alertBtn.text()).toBe("Understood");
    });

    it("Emits alert-dismissed when alert button is clicked", async () => {
      wrapper = mountWrapper({
        confirmText: "Submit",
        alertMessage: "Error message",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const alertBtn = wrapper.findComponent('[data-test="alert-got-it-btn"]');
      await alertBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("alert-dismissed")).toBeTruthy();
      expect(wrapper.emitted("alert-dismissed")).toHaveLength(1);
    });

    it("Shows alert automatically when alertMessage prop changes", async () => {
      wrapper = mountWrapper({ confirmText: "Submit" });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      let alert = dialogDom.find('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(false);

      await wrapper.setProps({ alertMessage: "New error" });
      await flushPromises();

      dialogDom = new DOMWrapper(document.body);
      alert = dialogDom.find('[data-test="form-dialog-alert"]');
      expect(alert.exists()).toBe(true);
      expect(alert.text()).toContain("New error");
    });
  });

  describe("Footer-right slot", () => {
    it("Renders footer-right slot instead of default buttons", async () => {
      wrapper = mountWrapper(
        { confirmText: "Submit" },
        {
          "footer-right": "<button data-test='custom-footer-btn'>Custom Action</button>",
        },
      );
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const customBtn = dialogDom.find('[data-test="custom-footer-btn"]');
      expect(customBtn.exists()).toBe(true);
      expect(customBtn.text()).toBe("Custom Action");

      expect(dialogDom.find('[data-test="confirm-btn"]').exists()).toBe(false);
    });
  });

  describe("Footer visibility", () => {
    it("Shows footer by default", async () => {
      wrapper = mountWrapper({ confirmText: "Submit" });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const windowDialog = wrapper.findComponent(WindowDialog);
      expect(windowDialog.props().showFooter).toBe(true);
    });

    it("Hides footer when showFooter is false", async () => {
      wrapper = mountWrapper({
        confirmText: "Submit",
        showFooter: false,
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const windowDialog = wrapper.findComponent(WindowDialog);
      expect(windowDialog.props().showFooter).toBe(false);
    });
  });
});
