import { describe, it, expect, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import CopyCommandField from "@/components/Fields/CopyCommandField.vue";

vi.mock("@/components/User/CopyWarning.vue", () => ({
  default: {
    template: "<div><slot :copyText=\"copyText\" /></div>",
    methods: { copyText: vi.fn() },
  },
}));

describe("CopyCommandField", () => {
  let wrapper: VueWrapper<InstanceType<typeof CopyCommandField>>;

  const mountWrapper = async (props: Partial<InstanceType<typeof CopyCommandField>["$props"]> = {}) => {
    wrapper = mountComponent(CopyCommandField, {
      props: {
        command: "curl -sSf https://example.com/install.sh | sh",
        ...props,
      },
    });
    await flushPromises();
  };

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  it("Displays the command in the text field", async () => {
    await mountWrapper({
      command: "echo 'test command'",
    });

    const input = wrapper.find('[data-test="copy-command-field"] input').element as HTMLInputElement;
    expect(input.value).toBe("echo 'test command'");
  });

  it("Renders with label when provided", async () => {
    await mountWrapper({
      command: "test",
      label: "Installation Command",
    });

    const field = wrapper.find('[data-test="copy-command-field"]');
    expect(field.text()).toContain("Installation Command");
  });

  it("Renders with hint when provided", async () => {
    await mountWrapper({
      command: "test",
      hint: "Run this on your terminal",
      persistentHint: true,
      hideDetails: false,
    });

    const field = wrapper.find('[data-test="copy-command-field"] .v-input__details');
    expect(field.text()).toContain("Run this on your terminal");
  });

  it("Shows $ prefix in the text field", async () => {
    await mountWrapper();

    const field = wrapper.find('[data-test="copy-command-field"] .v-text-field__prefix__text');
    expect(field.text()).toContain("$");
  });

  it("Renders copy button", async () => {
    await mountWrapper();

    const copyButton = wrapper.find('[data-test="copy-button"]');
    expect(copyButton.exists()).toBe(true);
    expect(copyButton.find("i").classes()).toContain("mdi-content-copy");
  });

  it("Text field is readonly", async () => {
    await mountWrapper();

    const input = wrapper.find('[data-test="copy-command-field"] input');
    expect(input.attributes("readonly")).toBeDefined();
  });

  it("Hides details by default", async () => {
    await mountWrapper();

    const fieldDetails = wrapper.find('[data-test="copy-command-field"] .v-input__details');
    expect(fieldDetails.exists()).toBe(false);
  });

  it("Shows details when hideDetails is false", async () => {
    await mountWrapper({ hideDetails: false });

    const fieldDetails = wrapper.find('[data-test="copy-command-field"] .v-input__details');
    expect(fieldDetails.exists()).toBe(true);
  });
});
