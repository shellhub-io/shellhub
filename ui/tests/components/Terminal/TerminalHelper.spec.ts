import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { describe, it, beforeEach, expect, vi, afterEach } from "vitest";
import TerminalHelper from "@/components/Terminal/TerminalHelper.vue";

vi.mock("@/components/User/CopyWarning.vue", () => ({
  default: {
    template: "<div><slot :copyText=\"copyText\" /></div>",
    methods: {
      copyText: vi.fn(),
    },
  },
}));

describe("TerminalHelper", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalHelper>>;

  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(TerminalHelper, {
      global: { plugins: [vuetify] },
      props: {
        sshid: "namespace.70-85-c2-08-60-2a@staging.shellhub.io",
        modelValue: true,
        showCheckbox: true,
        userId: "test-user-id",
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders the component", () => {
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("updates command line when username is entered", async () => {
    const input = wrapper.findComponent("[data-test='username-input']");
    expect(input.exists()).toBe(true);
    await input.setValue("ubuntu");

    const commandInput = wrapper.findComponent("[data-test='copy-command-field']");
    expect(commandInput.exists()).toBe(true);
    expect(commandInput.html()).toContain(
      "ssh ubuntu@namespace.70-85-c2-08-60-2a@staging.shellhub.io",
    );
  });

  it("closes the dialog when Close button is clicked", async () => {
    const closeBtn = wrapper.findComponent("[data-test='close-btn']");
    await closeBtn.trigger("click");
    expect(wrapper.emitted("update:modelValue")).toBeTruthy();
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
  });

  it("sets user ID in localStorage when checkbox is checked", async () => {
    localStorage.clear();

    const checkbox = wrapper.findComponent("[data-test='dispense-checkbox']");
    await checkbox.setValue(true);

    const stored = JSON.parse(localStorage.getItem("dispenseTerminalHelper") || "[]");
    expect(stored).toContain("test-user-id");
  });

  it("removes user ID from localStorage when checkbox is unchecked", async () => {
    localStorage.setItem("dispenseTerminalHelper", JSON.stringify(["test-user-id"]));

    // This code is to ensure the checkbox it is checked first
    // (true then false will trigger the watcher because the checkbox opens as false always)
    const checkbox = wrapper.findComponent("[data-test='dispense-checkbox']");
    await checkbox.setValue(true);
    await checkbox.setValue(false);

    await flushPromises();

    const stored = JSON.parse(localStorage.getItem("dispenseTerminalHelper") || "[]");
    expect(stored).not.toContain("test-user-id");
  });
});
