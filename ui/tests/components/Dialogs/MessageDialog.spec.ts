import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";

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
  let wrapper: MessageDialogWrapper;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = async (
    props: Partial<InstanceType<typeof MessageDialog>["$props"]> = {},
    slots: Record<string, string> = {},
  ) => {
    wrapper = mountComponent(MessageDialog, {
      props: { modelValue: true, ...props },
      slots: {
        default: "<div data-test='default-slot'>Message content</div>",
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

  describe("Titlebar mode", () => {
    it("Renders titlebar with title when showTitlebar is true", async () => {
      await mountWrapper({
        showTitlebar: true,
        title: "Titlebar Title",
      });

      const toolbar = dialog.find(".v-toolbar");
      expect(toolbar.exists()).toBe(true);

      const title = dialog.find(".v-toolbar-title");
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Titlebar Title");
    });

    it("Renders description in titlebar", async () => {
      await mountWrapper({
        showTitlebar: true,
        title: "Test",
        description: "Titlebar Description",
      });

      expect(dialog.text()).toContain("Titlebar Description");
    });

    it("Renders icon in titlebar", async () => {
      await mountWrapper({
        showTitlebar: true,
        icon: "mdi-alert",
        iconColor: "error",
      });

      const avatar = dialog.find(".v-avatar");
      expect(avatar.exists()).toBe(true);

      const icon = dialog.find(".v-avatar .v-icon");
      expect(icon.html()).toContain("mdi-alert");
    });

    it("Shows close button in titlebar when showCloseButton is true", async () => {
      await mountWrapper({
        showTitlebar: true,
        showCloseButton: true,
      });

      const closeBtn = dialog.find('[data-test="close-btn-toolbar"]');
      expect(closeBtn.exists()).toBe(true);
    });

    it("Hides close button when showCloseButton is false", async () => {
      await mountWrapper({
        showTitlebar: true,
        showCloseButton: false,
      });

      const closeBtn = dialog.find('[data-test="close-btn-toolbar"]');
      expect(closeBtn.exists()).toBe(false);
    });

    it("Emits close when close button is clicked", async () => {
      await mountWrapper({
        showTitlebar: true,
        showCloseButton: true,
      });

      const closeBtn = dialog.get('[data-test="close-btn-toolbar"]');
      await closeBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("close")).toHaveLength(1);
    });
  });

  describe("Content mode", () => {
    it("Renders icon in content area when showTitlebar is false", async () => {
      await mountWrapper({
        showTitlebar: false,
        icon: "mdi-check-circle",
        iconColor: "success",
      });

      const contentIcon = dialog.find(".v-card-text .v-icon");
      expect(contentIcon.exists()).toBe(true);
      expect(contentIcon.html()).toContain("mdi-check-circle");
    });

    it("Renders title in content area when showTitlebar is false", async () => {
      await mountWrapper({
        showTitlebar: false,
        title: "Content Title",
      });

      const contentTitle = dialog.find(".v-card-text .text-h5");
      expect(contentTitle.exists()).toBe(true);
      expect(contentTitle.text()).toBe("Content Title");
    });

    it("Renders description in content area when showTitlebar is false", async () => {
      await mountWrapper({
        showTitlebar: false,
        description: "Content Description",
      });

      const contentDesc = dialog.find(".v-card-text .text-body-2");
      expect(contentDesc.exists()).toBe(true);
      expect(contentDesc.text()).toBe("Content Description");
    });

    it("Does not render titlebar when showTitlebar is false", async () => {
      await mountWrapper({
        showTitlebar: false,
        title: "Title",
      });

      const titlebar = dialog.find("[data-test='titlebar']");
      expect(titlebar.exists()).toBe(false);
    });
  });

  describe("Default slot", () => {
    it("Renders slot content", async () => {
      await mountWrapper();

      const slotNode = dialog.find('[data-test="default-slot"]');
      expect(slotNode.exists()).toBe(true);
      expect(slotNode.text()).toBe("Message content");
    });
  });

  describe("Footer with action buttons", () => {
    it("Shows footer with confirm button", async () => {
      await mountWrapper({
        showFooter: true,
        confirmText: "Confirm",
      });

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      expect(confirmBtn.exists()).toBe(true);
      expect(confirmBtn.text()).toBe("Confirm");
    });

    it("Shows footer with cancel button", async () => {
      await mountWrapper({
        showFooter: true,
        cancelText: "Cancel",
      });

      const cancelBtn = dialog.find('[data-test="cancel-btn"]');
      expect(cancelBtn.exists()).toBe(true);
      expect(cancelBtn.text()).toBe("Cancel");
    });

    it("Emits confirm when confirm button is clicked", async () => {
      await mountWrapper({
        showFooter: true,
        confirmText: "Confirm",
      });

      const confirmBtn = dialog.get('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("confirm")).toHaveLength(1);
    });

    it("Emits cancel when cancel button is clicked", async () => {
      await mountWrapper({
        showFooter: true,
        cancelText: "Cancel",
      });

      const cancelBtn = dialog.get('[data-test="cancel-btn"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("cancel")).toHaveLength(1);
    });

    it("Disables confirm button when confirmDisabled is true", async () => {
      await mountWrapper({
        showFooter: true,
        confirmText: "Confirm",
        confirmDisabled: true,
      });

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      expect(confirmBtn.attributes("disabled")).toBeDefined();
    });

    it("Shows loading state on confirm button", async () => {
      await mountWrapper({
        showFooter: true,
        confirmText: "Confirm",
        confirmLoading: true,
      });

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      expect(confirmBtn.html()).toContain("v-btn--loading");
    });

    it("Uses custom data-test attributes", async () => {
      await mountWrapper({
        showFooter: true,
        confirmText: "Submit",
        confirmDataTest: "custom-submit",
        cancelText: "Cancel",
        cancelDataTest: "custom-cancel",
      });

      expect(dialog.find('[data-test="custom-submit"]').exists()).toBe(true);
      expect(dialog.find('[data-test="custom-cancel"]').exists()).toBe(true);
    });
  });
});
