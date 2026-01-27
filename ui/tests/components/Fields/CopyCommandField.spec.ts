import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import CopyCommandField from "@/components/Fields/CopyCommandField.vue";

vi.mock("@/components/User/CopyWarning.vue", () => ({
  default: {
    template: "<div><slot :copyText=\"copyText\" /></div>",
    methods: {
      copyText: vi.fn(),
    },
  },
}));

describe("CopyCommandField", () => {
  const vuetify = createVuetify();
  const props = {
    command: "curl -sSf https://example.com/install.sh | sh",
    label: "Installation Command",
    hint: "Run this on your terminal",
    persistentHint: true,
    persistentPlaceholder: true,
    hideDetails: false,
  };

  const wrapper = mount(CopyCommandField, {
    props,
    global: { plugins: [vuetify] },
  });

  it("displays the command in the text field", () => {
    const textField = wrapper.find('[data-test="copy-command-field"] input');
    expect(textField.exists()).toBe(true);
    expect(textField.attributes("value")).toBe(props.command);
  });

  it("renders with correct label and hint", () => {
    const textFieldContent = wrapper.find('[data-test="copy-command-field"]').text();
    expect(textFieldContent).toContain("Installation Command");
    expect(textFieldContent).toContain("Run this on your terminal");
  });

  it("renders copy button in append slot", () => {
    const copyButton = wrapper.find('[data-test="copy-button"]');
    expect(copyButton.exists()).toBe(true);
    expect(copyButton.attributes("class")).toContain("bg-primary");
  });
});
