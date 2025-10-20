import { createVuetify } from "vuetify";
import { DOMWrapper, mount, VueWrapper, flushPromises } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";
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

type WindowDialogWrapper = VueWrapper<InstanceType<typeof WindowDialog>>;

describe("WindowDialog", () => {
  const vuetify = createVuetify();
  let wrapper: WindowDialogWrapper;
  let dialogDom: DOMWrapper<HTMLElement>;

  const mountWrapper = (
    props: Partial<InstanceType<typeof WindowDialog>["$props"]> = {},
    slots: Record<string, string> = {},
  ) => mount(WindowDialog, {
    global: { plugins: [vuetify] },
    props: { modelValue: true, ...props },
    slots: {
      default: "<div data-test='default-slot'>Default content</div>",
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

  it("Renders title and description when provided", async () => {
    wrapper.unmount();
    document.body.innerHTML = "";

    wrapper = mountWrapper({ title: "My Dialog", description: "Extra context" });
    dialogDom = new DOMWrapper(document.body);
    await flushPromises();

    const title = dialogDom.find(".v-toolbar-title");
    expect(title.exists()).toBe(true);
    expect(title.text()).toBe("My Dialog");

    expect(dialogDom.text()).toContain("Extra context");
  });

  it("Renders icon avatar when icon is set and uses iconColor", async () => {
    wrapper.unmount();
    document.body.innerHTML = "";

    wrapper = mountWrapper({ icon: "mdi-key", iconColor: "success" });
    dialogDom = new DOMWrapper(document.body);
    await flushPromises();

    const avatar = dialogDom.find(".v-avatar");
    expect(avatar.exists()).toBe(true);

    const icon = dialogDom.find(".v-icon");
    expect(icon.exists()).toBe(true);
    expect(icon.html()).toContain("mdi-key");
  });

  it("Hides the toolbar close button when showCloseButton is false", async () => {
    wrapper.unmount();
    document.body.innerHTML = "";

    wrapper = mountWrapper({ showCloseButton: false });
    dialogDom = new DOMWrapper(document.body);
    await flushPromises();

    expect(dialogDom.find('[data-test="close-btn-toolbar"]').exists()).toBe(false);
  });

  it("Emits close when the toolbar close button is clicked", async () => {
    wrapper.unmount();
    document.body.innerHTML = "";

    wrapper = mountWrapper({ showCloseButton: true });
    dialogDom = new DOMWrapper(document.body);
    await flushPromises();

    const btn = dialogDom.get('[data-test="close-btn-toolbar"]');
    await btn.trigger("click");
    await flushPromises();

    expect(wrapper.emitted("close")).toBeTruthy();
    expect(wrapper.emitted("close")).toHaveLength(1);
  });

  it("Emits close when BaseDialog emits close", async () => {
    wrapper.unmount();
    document.body.innerHTML = "";

    wrapper = mountWrapper();
    await flushPromises();

    const base = wrapper.findComponent(BaseDialog);
    expect(base.exists()).toBe(true);

    base.vm.$emit("close");
    await flushPromises();

    expect(wrapper.emitted("close")).toBeTruthy();
  });

  it("Renders default slot content", async () => {
    wrapper.unmount();
    document.body.innerHTML = "";

    wrapper = mountWrapper();
    dialogDom = new DOMWrapper(document.body);
    await flushPromises();

    const slotNode = dialogDom.find('[data-test="default-slot"]');
    expect(slotNode.exists()).toBe(true);
    expect(slotNode.text()).toBe("Default content");
  });

  it("Renders titlebar-content and titlebar-actions slots", async () => {
    wrapper.unmount();
    document.body.innerHTML = "";

    wrapper = mountWrapper(
      { title: "With Slots" },
      {
        "titlebar-content": "<div data-test='titlebar-content-slot'>Extra Titlebar</div>",
        "titlebar-actions": "<button data-test='titlebar-actions-slot'>Action</button>",
      },
    );
    dialogDom = new DOMWrapper(document.body);
    await flushPromises();

    expect(dialogDom.find('[data-test="titlebar-content-slot"]').exists()).toBe(true);
    expect(dialogDom.find('[data-test="titlebar-actions-slot"]').exists()).toBe(true);
  });

  it("Shows footer (and its slot) when showFooter is true", async () => {
    wrapper.unmount();
    document.body.innerHTML = "";

    wrapper = mountWrapper(
      { showFooter: true },
      { footer: "<div data-test='footer-slot'>Footer Area</div>" },
    );
    dialogDom = new DOMWrapper(document.body);
    await flushPromises();

    const toolbars = dialogDom.findAll(".v-toolbar");
    expect(toolbars.length).toBeGreaterThanOrEqual(2);
    expect(dialogDom.find('[data-test="footer-slot"]').exists()).toBe(true);
  });

  it("Hides footer when showFooter is false", async () => {
    wrapper.unmount();
    document.body.innerHTML = "";

    wrapper = mountWrapper(
      { showFooter: false },
      { footer: "<div data-test='footer-slot'>Footer Area</div>" },
    );
    dialogDom = new DOMWrapper(document.body);
    await flushPromises();

    const toolbars = dialogDom.findAll(".v-toolbar");
    expect(toolbars.length).toBe(1);
    expect(dialogDom.find('[data-test="footer-slot"]').exists()).toBe(false);
  });

  it("Passes threshold and forceFullscreen props to BaseDialog", async () => {
    wrapper.unmount();
    document.body.innerHTML = "";

    wrapper = mountWrapper({ threshold: "md", forceFullscreen: true });
    await flushPromises();

    const base = wrapper.findComponent(BaseDialog);
    expect(base.exists()).toBe(true);

    const props = base.props();
    expect(props.threshold).toBe("md");
    expect(props.forceFullscreen).toBe(true);
  });
});
