import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper } from "@vue/test-utils";
import TerminalFontPicker from "@/components/Terminal/TerminalFontPicker.vue";
import { mountComponent } from "@tests/utils/mount";
import useTerminalThemeStore from "@/store/modules/terminal_theme";

describe("TerminalFontPicker", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalFontPicker>>;
  let terminalThemeStore: ReturnType<typeof useTerminalThemeStore>;

  const mountWrapper = () => {
    wrapper = mountComponent(TerminalFontPicker, {
      piniaOptions: {
        initialState: {
          terminalTheme: {
            currentFontFamily: "Monospace",
            currentFontSize: 15,
            availableFonts: [
              "Monospace",
              "Source Code Pro",
              "Inconsolata",
              "Ubuntu Mono",
              "Fira Code",
              "Anonymous Pro",
              "JetBrains Mono",
              "Noto Mono",
            ],
          },
        },
      },
    });
    terminalThemeStore = useTerminalThemeStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  it("renders font family select", () => {
    const fontFamilySelect = wrapper.findComponent({ name: "VSelect" });
    expect(fontFamilySelect.exists()).toBe(true);
    expect(fontFamilySelect.props("label")).toBe("Font Family");
  });

  it("renders font size number input with min and max constraints", () => {
    const fontSizeInput = wrapper.findComponent({ name: "VNumberInput" });
    expect(fontSizeInput.exists()).toBe(true);
    expect(fontSizeInput.props("label")).toBe("Font Size");
    expect(fontSizeInput.props("min")).toBe(8);
    expect(fontSizeInput.props("max")).toBe(32);
  });

  it("displays current font family from store", () => {
    const fontFamilySelect = wrapper.findComponent({ name: "VSelect" });
    expect(fontFamilySelect.props("modelValue")).toBe("Monospace");
  });

  it("displays current font size from store", () => {
    const fontSizeInput = wrapper.findComponent({ name: "VNumberInput" });
    expect(fontSizeInput.props("modelValue")).toBe(15);
  });

  it("calls setFontSettings and emits update:fontSettings when font family changes", async () => {
    const fontFamilySelect = wrapper.findComponent({ name: "VSelect" });
    await fontFamilySelect.setValue("Fira Code");
    await flushPromises();

    expect(terminalThemeStore.setFontSettings).toHaveBeenCalledWith("Fira Code", 15);
    expect(wrapper.emitted("update:fontSettings")).toBeTruthy();
    expect(wrapper.emitted("update:fontSettings")?.[0]).toEqual([
      { fontFamily: "Fira Code", fontSize: 15 },
    ]);
  });

  it("calls setFontSettings and emits update:fontSettings when font size changes", async () => {
    const fontSizeInput = wrapper.findComponent({ name: "VNumberInput" });
    await fontSizeInput.setValue(20);
    await flushPromises();

    expect(terminalThemeStore.setFontSettings).toHaveBeenCalledWith("Monospace", 20);
    expect(wrapper.emitted("update:fontSettings")).toBeTruthy();
    expect(wrapper.emitted("update:fontSettings")?.[0]).toEqual([
      { fontFamily: "Monospace", fontSize: 20 },
    ]);
  });

  it("displays all available font families", () => {
    const fontFamilySelect = wrapper.findComponent({ name: "VSelect" });
    const items = fontFamilySelect.props("items");

    expect(items).toHaveLength(8);
    expect(items).toContain("Monospace");
    expect(items).toContain("Fira Code");
    expect(items).toContain("JetBrains Mono");
  });
});
