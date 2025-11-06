import { createVuetify } from "vuetify";
import { DOMWrapper, mount, VueWrapper, flushPromises } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import BaseDialog from "@/components/Dialogs/BaseDialog.vue";

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

type MessageDialogWrapper = VueWrapper<InstanceType<typeof MessageDialog>>;

describe("MessageDialog", () => {
  const vuetify = createVuetify();
  let wrapper: MessageDialogWrapper;
  let dialogDom: DOMWrapper<HTMLElement>;

  const mountWrapper = (
    props: Partial<InstanceType<typeof MessageDialog>["$props"]> = {},
    slots: Record<string, string> = {},
  ) => mount(MessageDialog, {
    global: { plugins: [vuetify] },
    props: { modelValue: true, ...props },
    slots: {
      default: "<div data-test='default-slot'>Message content</div>",
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

  describe("Titlebar mode (showTitlebar=true)", () => {
    it("Renders titlebar with title and description", async () => {
      wrapper = mountWrapper({
        showTitlebar: true,
        title: "Titlebar Title",
        description: "Titlebar Description",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const toolbar = dialogDom.find(".v-toolbar");
      expect(toolbar.exists()).toBe(true);

      const title = dialogDom.find(".v-toolbar-title");
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Titlebar Title");

      expect(dialogDom.text()).toContain("Titlebar Description");
    });

    it("Renders icon in titlebar when showTitlebar is true", async () => {
      wrapper = mountWrapper({
        showTitlebar: true,
        icon: "mdi-alert",
        iconColor: "error",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const avatar = dialogDom.find(".v-avatar");
      expect(avatar.exists()).toBe(true);

      const icon = dialogDom.find(".v-avatar .v-icon");
      expect(icon.exists()).toBe(true);
      expect(icon.html()).toContain("mdi-alert");
    });

    it("Shows close button in titlebar when showCloseButton is true", async () => {
      wrapper = mountWrapper({
        showTitlebar: true,
        showCloseButton: true,
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      expect(dialogDom.find('[data-test="close-btn-toolbar"]').exists()).toBe(true);
    });

    it("Hides close button in titlebar when showCloseButton is false", async () => {
      wrapper = mountWrapper({
        showTitlebar: true,
        showCloseButton: false,
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      expect(dialogDom.find('[data-test="close-btn-toolbar"]').exists()).toBe(false);
    });

    it("Emits close when titlebar close button is clicked", async () => {
      wrapper = mountWrapper({
        showTitlebar: true,
        showCloseButton: true,
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const btn = dialogDom.get('[data-test="close-btn-toolbar"]');
      await btn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("close")).toBeTruthy();
      expect(wrapper.emitted("close")).toHaveLength(1);
    });
  });

  describe("Content mode (showTitlebar=false)", () => {
    it("Renders icon in content area when showTitlebar is false", async () => {
      wrapper = mountWrapper({
        showTitlebar: false,
        icon: "mdi-check-circle",
        iconColor: "success",
        iconSize: 64,
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const contentIcon = dialogDom.find(".v-card-text .v-icon");
      expect(contentIcon.exists()).toBe(true);
      expect(contentIcon.html()).toContain("mdi-check-circle");
    });

    it("Renders title in content area when showTitlebar is false", async () => {
      wrapper = mountWrapper({
        showTitlebar: false,
        title: "Content Title",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const contentTitle = dialogDom.find(".v-card-text .text-h5");
      expect(contentTitle.exists()).toBe(true);
      expect(contentTitle.text()).toBe("Content Title");
    });

    it("Renders description in content area when showTitlebar is false", async () => {
      wrapper = mountWrapper({
        showTitlebar: false,
        description: "Content Description",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const contentDesc = dialogDom.find(".v-card-text .text-body-2");
      expect(contentDesc.exists()).toBe(true);
      expect(contentDesc.text()).toBe("Content Description");
    });

    it("Does not render titlebar when showTitlebar is false", async () => {
      wrapper = mountWrapper({
        showTitlebar: false,
        title: "Title",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const toolbar = dialogDom.findComponent(".v-toolbar");
      expect(toolbar.exists()).toBe(false);
    });
  });

  describe("Default slot", () => {
    it("Renders default slot content in card text", async () => {
      wrapper = mountWrapper();
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const slotNode = dialogDom.find('[data-test="default-slot"]');
      expect(slotNode.exists()).toBe(true);
      expect(slotNode.text()).toBe("Message content");
    });
  });

  describe("Footer with action buttons", () => {
    it("Shows footer with confirm button when confirmText is provided", async () => {
      wrapper = mountWrapper({
        showFooter: true,
        confirmText: "Confirm",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const confirmBtn = dialogDom.find('[data-test="confirm-btn"]');
      expect(confirmBtn.exists()).toBe(true);
      expect(confirmBtn.text()).toBe("Confirm");
    });

    it("Shows footer with cancel button when cancelText is provided", async () => {
      wrapper = mountWrapper({
        showFooter: true,
        cancelText: "Cancel",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const cancelBtn = dialogDom.find('[data-test="cancel-btn"]');
      expect(cancelBtn.exists()).toBe(true);
      expect(cancelBtn.text()).toBe("Cancel");
    });

    it("Shows both confirm and cancel buttons", async () => {
      wrapper = mountWrapper({
        showFooter: true,
        confirmText: "Yes",
        cancelText: "No",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      expect(dialogDom.find('[data-test="confirm-btn"]').exists()).toBe(true);
      expect(dialogDom.find('[data-test="cancel-btn"]').exists()).toBe(true);
    });

    it("Emits confirm when confirm button is clicked", async () => {
      wrapper = mountWrapper({
        showFooter: true,
        confirmText: "Confirm",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const confirmBtn = dialogDom.get('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("confirm")).toBeTruthy();
      expect(wrapper.emitted("confirm")).toHaveLength(1);
    });

    it("Emits cancel when cancel button is clicked", async () => {
      wrapper = mountWrapper({
        showFooter: true,
        cancelText: "Cancel",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const cancelBtn = dialogDom.get('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("cancel")).toBeTruthy();
      expect(wrapper.emitted("cancel")).toHaveLength(1);
    });

    it("Disables confirm button when confirmDisabled is true", async () => {
      wrapper = mountWrapper({
        showFooter: true,
        confirmText: "Confirm",
        confirmDisabled: true,
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      const confirmBtn = dialogDom.get('[data-test="confirm-btn"]');
      expect(confirmBtn.attributes("disabled")).toBeDefined();
    });

    it("Shows loading state on confirm button when confirmLoading is true", async () => {
      wrapper = mountWrapper({
        showFooter: true,
        confirmText: "Confirm",
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
        showFooter: true,
        confirmText: "Confirm",
        confirmDataTest: "custom-confirm",
        cancelText: "Cancel",
        cancelDataTest: "custom-cancel",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      expect(dialogDom.find('[data-test="custom-confirm"]').exists()).toBe(true);
      expect(dialogDom.find('[data-test="custom-cancel"]').exists()).toBe(true);
    });

    it("Hides footer when showFooter is false", async () => {
      wrapper = mountWrapper({
        showFooter: false,
        confirmText: "Confirm",
        cancelText: "Cancel",
      });
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      expect(dialogDom.find('[data-test="confirm-btn"]').exists()).toBe(false);
      expect(dialogDom.find('[data-test="cancel-btn"]').exists()).toBe(false);
    });
  });

  describe("Named slots", () => {
    it("Renders titlebar-content slot", async () => {
      wrapper = mountWrapper(
        { showTitlebar: true },
        {
          "titlebar-content": "<div data-test='titlebar-content-slot'>Extra Content</div>",
        },
      );
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      expect(dialogDom.find('[data-test="titlebar-content-slot"]').exists()).toBe(true);
      expect(dialogDom.find('[data-test="titlebar-content-slot"]').text()).toBe("Extra Content");
    });

    it("Renders titlebar-actions slot", async () => {
      wrapper = mountWrapper(
        { showTitlebar: true },
        {
          "titlebar-actions": "<button data-test='titlebar-actions-slot'>Action</button>",
        },
      );
      dialogDom = new DOMWrapper(document.body);
      await flushPromises();

      expect(dialogDom.find('[data-test="titlebar-actions-slot"]').exists()).toBe(true);
    });
  });

  describe("BaseDialog integration", () => {
    it("Emits close when BaseDialog emits close", async () => {
      wrapper = mountWrapper();
      await flushPromises();

      const base = wrapper.findComponent(BaseDialog);
      expect(base.exists()).toBe(true);

      base.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("close")).toBeTruthy();
    });

    it("Passes threshold and forceFullscreen props to BaseDialog", async () => {
      wrapper = mountWrapper({ threshold: "md", forceFullscreen: true });
      await flushPromises();

      const base = wrapper.findComponent(BaseDialog);
      expect(base.exists()).toBe(true);

      const props = base.props();
      expect(props.threshold).toBe("md");
      expect(props.forceFullscreen).toBe(true);
    });
  });
});
