import { DOMWrapper, VueWrapper, flushPromises } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { mountComponent } from "@tests/utils/mount";
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

describe("WindowDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof WindowDialog>>;
  let dialog: DOMWrapper<HTMLElement>;

  const mountWrapper = async (
    props: Partial<InstanceType<typeof WindowDialog>["$props"]> = {},
    slots: Record<string, string> = {},
  ) => {
    wrapper = mountComponent(WindowDialog, {
      props: { modelValue: true, ...props },
      slots: {
        default: "<div data-test='dialog-content'>Test content</div>",
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
    expect(dialog.exists()).toBe(true);
  });

  describe("Titlebar", () => {
    it("Renders titlebar with title", async () => {
      await mountWrapper({ title: "Test Title" });

      const titlebar = dialog.find('[data-test="window-dialog-titlebar"]');
      expect(titlebar.exists()).toBe(true);
      expect(titlebar.text()).toContain("Test Title");
    });

    it("Renders titlebar with description", async () => {
      await mountWrapper({
        title: "Test Title",
        description: "Test Description",
      });

      const titlebar = dialog.find('[data-test="window-dialog-titlebar"]');
      expect(titlebar.text()).toContain("Test Description");
    });

    it("Renders icon in titlebar when provided", async () => {
      await mountWrapper({
        icon: "mdi-alert",
        iconColor: "warning",
      });

      const avatar = dialog.find(".v-avatar");
      expect(avatar.exists()).toBe(true);

      const icon = dialog.find(".v-avatar .v-icon");
      expect(icon.html()).toContain("mdi-alert");
    });

    it("Shows close button by default", async () => {
      await mountWrapper();

      const closeBtn = dialog.find('[data-test="close-btn-toolbar"]');
      expect(closeBtn.exists()).toBe(true);
    });

    it("Hides close button when showCloseButton is false", async () => {
      await mountWrapper({ showCloseButton: false });

      const closeBtn = dialog.find('[data-test="close-btn-toolbar"]');
      expect(closeBtn.exists()).toBe(false);
    });

    it("Emits close event when close button is clicked", async () => {
      await mountWrapper();

      const closeBtn = dialog.get('[data-test="close-btn-toolbar"]');
      await closeBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("close")).toHaveLength(1);
    });

    it("Renders titlebar-content slot", async () => {
      await mountWrapper({}, {
        "titlebar-content": "<div data-test='custom-content'>Custom</div>",
      });

      const customContent = dialog.find('[data-test="custom-content"]');
      expect(customContent.exists()).toBe(true);
      expect(customContent.text()).toBe("Custom");
    });

    it("Renders titlebar-actions slot", async () => {
      await mountWrapper({}, {
        "titlebar-actions": "<button data-test='custom-action'>Action</button>",
      });

      const customAction = dialog.find('[data-test="custom-action"]');
      expect(customAction.exists()).toBe(true);
      expect(customAction.text()).toBe("Action");
    });
  });

  describe("Content", () => {
    it("Renders default slot content", async () => {
      await mountWrapper();

      const content = dialog.find('[data-test="dialog-content"]');
      expect(content.exists()).toBe(true);
      expect(content.text()).toBe("Test content");
    });
  });

  describe("Footer", () => {
    it("Renders footer by default", async () => {
      await mountWrapper({}, {
        footer: "<button data-test='footer-btn'>Footer Button</button>",
      });

      const footer = dialog.find('[data-test="window-dialog-footer"]');
      expect(footer.exists()).toBe(true);

      const footerBtn = dialog.find('[data-test="footer-btn"]');
      expect(footerBtn.exists()).toBe(true);
    });

    it("Hides footer when showFooter is false", async () => {
      await mountWrapper({ showFooter: false }, {
        footer: "<button data-test='footer-btn'>Footer Button</button>",
      });

      const footer = dialog.find('[data-test="window-dialog-footer"]');
      expect(footer.exists()).toBe(false);
    });

    it("Renders footer slot content", async () => {
      await mountWrapper({}, {
        footer: "<div data-test='custom-footer'>Custom footer</div>",
      });

      const customFooter = dialog.find('[data-test="custom-footer"]');
      expect(customFooter.exists()).toBe(true);
      expect(customFooter.text()).toBe("Custom footer");
    });
  });
});
