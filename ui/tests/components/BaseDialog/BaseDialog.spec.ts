import { createVuetify } from "vuetify";
import { DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import BaseDialog from "@/components/BaseDialog.vue";

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

type BaseDialogWrapper = VueWrapper<InstanceType<typeof BaseDialog>>;

describe("BaseDialog", () => {
  const vuetify = createVuetify();
  let dialog: DOMWrapper<HTMLElement>;
  let wrapper: BaseDialogWrapper;

  const breakpointTests = [
    { breakpoint: "sm", expectedWidth: 600 },
    { breakpoint: "md", expectedWidth: 960 },
    { breakpoint: "lg", expectedWidth: 1280 },
    { breakpoint: "xl", expectedWidth: 1920 },
    { breakpoint: "xxl", expectedWidth: 2560 },
  ] as const;

  const slotContent = "<div>Test content</div>";

  const mountWrapper = (props: { forceFullscreen?: boolean; breakpoint?: "sm" | "md" | "lg" | "xl" | "xxl" } = {}) => mount(BaseDialog, {
    global: { plugins: [vuetify] },
    slots: { default: slotContent },
    props: { modelValue: true, ...props },
  });

  beforeEach(async () => {
    wrapper = mountWrapper();
    dialog = new DOMWrapper(document.body);
  });

  afterEach(() => {
    vi.clearAllMocks();
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Renders slot content", () => {
    expect(dialog.html()).toContain("Test content");
  });

  it("Uses default sm breakpoint (600px) when no breakpoint is specified", () => {
    expect(wrapper.vm.maxWidth).toBe(600);
    expect(wrapper.vm.fullscreen).toBe(false);
  });

  it("Uses fullscreen mode when forceFullscreen is true", () => {
    wrapper = mountWrapper({ forceFullscreen: true });
    expect(wrapper.vm.maxWidth).toBeUndefined();
    expect(wrapper.vm.fullscreen).toBe(true);
  });

  breakpointTests.forEach(({ breakpoint, expectedWidth }) => {
    it(`Uses correct width for ${breakpoint} breakpoint (${expectedWidth}px)`, () => {
      wrapper = mountWrapper({ breakpoint });

      expect(wrapper.vm.maxWidth).toBe(expectedWidth);
      expect(wrapper.vm.fullscreen).toBe(false);
    });
  });
});
