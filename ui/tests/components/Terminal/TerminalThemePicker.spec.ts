import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { flushPromises, VueWrapper } from "@vue/test-utils";
import TerminalThemePicker from "@/components/Terminal/TerminalThemePicker.vue";
import { mountComponent } from "@tests/utils/mount";
import { mockTerminalThemes } from "@tests/mocks/terminalTheme";

describe("TerminalThemePicker", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalThemePicker>>;

  const mountWrapper = ({ modelValue = "ShellHub Dark", themes = mockTerminalThemes } = {}) => {
    wrapper = mountComponent(TerminalThemePicker, {
      props: { modelValue },
      piniaOptions: {
        initialState: {
          terminalTheme: {
            terminalThemes: themes,
            currentThemeName: modelValue,
          },
        },
      },
    });
  };

  beforeEach(() => mountWrapper());

  afterEach(() => wrapper?.unmount());

  it("displays list of available themes", () => {
    const themeItems = wrapper.findAll('[data-test="theme-item"]');
    expect(themeItems).toHaveLength(mockTerminalThemes.length);
  });

  it("displays theme names and descriptions", () => {
    expect(wrapper.text()).toContain("ShellHub Dark");
    expect(wrapper.text()).toContain("ShellHub Light");
    expect(wrapper.text()).toContain("Dracula");
    expect(wrapper.text()).toContain("Dark theme");
    expect(wrapper.text()).toContain("Light theme");
  });

  it("shows preview for each theme with correct background and foreground colors", () => {
    const themePreviews = wrapper.findAll(".theme-preview");
    expect(themePreviews).toHaveLength(mockTerminalThemes.length);

    const firstPreview = themePreviews[0];
    const style = firstPreview.attributes("style");
    expect(style).toContain("background-color: rgb(15, 21, 38)"); // #0f1526
    expect(style).toContain("color: rgb(255, 255, 255)"); // #ffffff
  });

  it("marks the selected theme as active", () => {
    const themeItems = wrapper.findAll('[data-test="theme-item"]');
    const activeTheme = themeItems.find((item) => item.classes().includes("v-list-item--active"));

    expect(activeTheme).toBeTruthy();
    expect(activeTheme?.text()).toContain("ShellHub Dark");
  });

  it("shows check icon only for selected theme", () => {
    const checkIcons = wrapper.findAll(".mdi-check");
    expect(checkIcons).toHaveLength(1);

    const activeThemeItem = wrapper.find(".v-list-item--active");
    const checkIcon = activeThemeItem.find(".mdi-check");
    expect(checkIcon.exists()).toBe(true);
  });

  it("updates modelValue and emits update:selectedTheme when theme is clicked", async () => {
    const themeItems = wrapper.findAll('[data-test="theme-item"]');
    const draculaTheme = themeItems.find((item) => item.text().includes("Dracula"));

    await draculaTheme?.trigger("click");
    await flushPromises();

    expect(wrapper.emitted("update:modelValue")).toBeTruthy();
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["Dracula"]);

    expect(wrapper.emitted("update:selectedTheme")).toBeTruthy();
    expect(wrapper.emitted("update:selectedTheme")?.[0]).toEqual([mockTerminalThemes[2]]);
  });

  it.each([
    { index: 0, bg: "rgb(15, 21, 38)", fg: "rgb(255, 255, 255)", border: "rgb(38, 79, 120)" },
    { index: 1, bg: "rgb(255, 255, 255)", fg: "rgb(0, 0, 0)", border: "rgb(173, 214, 255)" },
    { index: 2, bg: "rgb(40, 42, 54)", fg: "rgb(248, 248, 242)", border: "rgb(68, 71, 90)" },
  ])("generates correct preview style for theme at index $index", ({ index, bg, fg, border }) => {
    const themePreviews = wrapper.findAll(".theme-preview");
    const preview = themePreviews[index];
    const style = preview.attributes("style");

    expect(style).toContain(`background-color: ${bg}`);
    expect(style).toContain(`color: ${fg}`);
    expect(style).toContain(`border: 1px solid ${border}`);
  });

  it("shows 'No themes available' when themes list is empty", () => {
    mountWrapper({ themes: [] });

    expect(wrapper.text()).toContain("No themes available");
    expect(wrapper.findAll('[data-test="theme-item"]')).toHaveLength(0);
  });

  it("displays preview code samples for each theme", () => {
    const themePreviews = wrapper.findAll(".theme-preview");

    themePreviews.forEach((preview) => {
      expect(preview.text()).toContain("$ ls");
      expect(preview.text()).toContain("file.txt");
      expect(preview.text()).toContain("home");
    });
  });

  it("applies primary color to selected theme title", () => {
    const themeItems = wrapper.findAll('[data-test="theme-item"]');
    const selectedItem = themeItems[0];
    const title = selectedItem.find(".v-list-item-title");

    expect(title.classes()).toContain("text-primary");
    expect(title.classes()).toContain("font-weight-medium");
  });
});
