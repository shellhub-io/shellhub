import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { describe, it, beforeEach, expect, vi } from "vitest";
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
      global: {
        plugins: [vuetify],
      },
      props: {
        sshid: "namespace.70-85-c2-08-60-2a@staging.shellhub.io",
        modelValue: true,
      },
    });
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("updates command line when username is entered", async () => {
    wrapper.vm.dialog = true;
    await flushPromises();

    const input = wrapper.findComponent("[data-test='username-input']");
    expect(input.exists()).toBe(true);
    await input.setValue("ubuntu");

    const commandInput = wrapper.findComponent("[data-test='command-field']");
    expect(commandInput.exists()).toBe(true);
    expect(commandInput.html()).toContain(
      "ssh ubuntu@namespace.70-85-c2-08-60-2a@staging.shellhub.io",
    );
  });

  it("closes the dialog when Close button is clicked", async () => {
    wrapper.vm.dialog = true;
    await flushPromises();

    const closeBtn = wrapper.findComponent("[data-test='close-btn']");
    await closeBtn.trigger("click");
    expect(wrapper.emitted("update:modelValue")).toBeTruthy();
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
  });
});
