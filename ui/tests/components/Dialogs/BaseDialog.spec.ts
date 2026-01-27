import { DOMWrapper, VueWrapper } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { flushPromises } from "@vue/test-utils";
import BaseDialog from "@/components/Dialogs/BaseDialog.vue";
import { mountComponent } from "@tests/utils/mount";

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

describe("BaseDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof BaseDialog>>;
  let dialog: DOMWrapper<HTMLElement>;

  const getDialogContent = () => dialog.find(".v-overlay__content");

  const mountWrapper = async (props = {}, slots = {}) => {
    wrapper = mountComponent(BaseDialog, {
      slots: {
        default: "<div>Test content</div>",
        ...slots,
      },
      props: {
        modelValue: true,
        ...props,
      },
    });

    await flushPromises();
    dialog = new DOMWrapper(document.body).find('[role="dialog"]');
  };

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  it("Renders slot content", async () => {
    await mountWrapper();

    expect(dialog.html()).toContain("Test content");
  });

  it("Renders the dialog with default sm threshold", async () => {
    await mountWrapper();
    const dialogContent = getDialogContent();
    expect(dialogContent.attributes("style")).toContain("max-width: 600px");
  });

  it("Applies fullscreen mode when forceFullscreen is true", async () => {
    await mountWrapper({ forceFullscreen: true });

    expect(dialog.classes()).toContain("v-dialog--fullscreen");
  });

  it("Applies correct max-width for md threshold", async () => {
    await mountWrapper({ threshold: "md" });

    const dialogContent = getDialogContent();
    expect(dialogContent.attributes("style")).toContain("max-width: 960px");
  });

  it("Applies correct max-width for lg threshold", async () => {
    await mountWrapper({ threshold: "lg" });

    const dialogContent = getDialogContent();
    expect(dialogContent.attributes("style")).toContain("max-width: 1280px");
  });

  it("Applies correct max-width for xl threshold", async () => {
    await mountWrapper({ threshold: "xl" });

    const dialogContent = getDialogContent();
    expect(dialogContent.attributes("style")).toContain("max-width: 1920px");
  });

  it("Applies correct max-width for xxl threshold", async () => {
    await mountWrapper({ threshold: "xxl" });

    const dialogContent = getDialogContent();
    expect(dialogContent.attributes("style")).toContain("max-width: 2560px");
  });

  it("Hides dialog when modelValue is false", async () => {
    await mountWrapper({ modelValue: false });

    expect(dialog.exists()).toBe(false);
  });

  it("Renders content slot when provided", async () => {
    await mountWrapper({}, { content: "<div data-test='content-slot'>Content slot</div>" });

    const contentSlot = dialog.find('[data-test="content-slot"]');
    expect(contentSlot.exists()).toBe(true);
    expect(contentSlot.text()).toBe("Content slot");
  });
});
