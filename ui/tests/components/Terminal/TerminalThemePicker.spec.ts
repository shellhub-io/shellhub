import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, VueWrapper, flushPromises } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { createPinia, setActivePinia } from "pinia";
import TerminalThemePicker from "@/components/Terminal/TerminalThemePicker.vue";
import useTerminalThemeStore from "@/store/modules/terminal_theme";
import { ITerminalTheme } from "@/interfaces/ITerminal";

const mockThemes: ITerminalTheme[] = [
  {
    name: "ShellHub Dark",
    description: "Dark theme",
    colors: {
      background: "#0f1526",
      foreground: "#ffffff",
      cursor: "#ffffff",
      selection: "#264f78",
    },
  },
  {
    name: "ShellHub Henry",
    description: "Light theme",
    colors: {
      background: "#ffffff",
      foreground: "#000000",
      cursor: "#000000",
      selection: "#add6ff",
    },
  },
  {
    name: "Dracula",
    description: "Dark theme",
    colors: {
      background: "#282a36",
      foreground: "#f8f8f2",
      cursor: "#f8f8f0",
      selection: "#44475a",
    },
  },
];

describe("TerminalThemePicker.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalThemePicker>>;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const terminalThemeStore = useTerminalThemeStore();

  terminalThemeStore.loadThemes = vi.fn().mockResolvedValue(undefined);
  terminalThemeStore.setTheme = vi.fn();
  terminalThemeStore.$patch({
    terminalThemes: mockThemes,
    currentThemeName: "ShellHub Dark",
  });

  beforeEach(async () => {
    wrapper = mount(TerminalThemePicker, {
      global: {
        plugins: [vuetify],
      },
      props: {
        modelValue: "ShellHub Dark",
      },
    });
  });

  it("renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("displays list of available themes and its descriptions", async () => {
    const themeItems = wrapper.findAll('[data-test="theme-item"]');
    expect(themeItems).toHaveLength(mockThemes.length);

    expect(wrapper.text()).toContain("ShellHub Dark");
    expect(wrapper.text()).toContain("ShellHub Henry");
    expect(wrapper.text()).toContain("Dracula");
    expect(wrapper.text()).toContain("Dark theme");
    expect(wrapper.text()).toContain("Light theme");
  });

  it("shows preview for each theme with correct colors", async () => {
    const themePreviews = wrapper.findAll(".theme-preview");
    expect(themePreviews).toHaveLength(mockThemes.length);

    const firstPreview = themePreviews[0];
    const style = firstPreview.attributes("style");
    expect(style).toContain("background-color: rgb(15, 21, 38)"); // #0f1526
    expect(style).toContain("color: rgb(255, 255, 255)"); // #ffffff
  });

  it("marks the selected theme as active", async () => {
    const themeItems = wrapper.findAll('[data-test="theme-item"]');
    const activeTheme = themeItems.find((item) => item.classes().includes("v-list-item--active"));

    expect(activeTheme).toBeTruthy();
    expect(activeTheme?.text()).toContain("ShellHub Dark");
  });

  it("shows check icon for selected theme", async () => {
    const checkIcons = wrapper.findAll(".mdi-check");
    expect(checkIcons).toHaveLength(1);

    const activeThemeItem = wrapper.find(".v-list-item--active");
    const checkIcon = activeThemeItem.find(".mdi-check");
    expect(checkIcon.exists()).toBe(true);
  });

  it("updates theme when theme item is clicked", async () => {
    const themeItems = wrapper.findAll('[data-test="theme-item"]');
    const draculaTheme = themeItems.find((item) => item.text().includes("Dracula"));

    await draculaTheme?.trigger("click");
    await flushPromises();

    const emitted = wrapper.emitted("update:selectedTheme");
    expect(emitted).toHaveLength(1);
    expect(emitted?.[0][0]).toEqual(mockThemes[2]); // Dracula theme
  });

  it("generates correct theme preview styles for different color combinations", async () => {
    const themeItems = wrapper.findAll('[data-test="theme-item"]');

    // Test first theme (ShellHub Dark)
    const darkPreview = themeItems[0].find(".theme-preview");
    const darkStyle = darkPreview.attributes("style");
    expect(darkStyle).toContain("background-color: rgb(15, 21, 38)");
    expect(darkStyle).toContain("color: rgb(255, 255, 255)");
    expect(darkStyle).toContain("border: 1px solid rgb(38, 79, 120)");

    // Test second theme (ShellHub Henry)
    const lightPreview = themeItems[1].find(".theme-preview");
    const lightStyle = lightPreview.attributes("style");
    expect(lightStyle).toContain("background-color: rgb(255, 255, 255)");
    expect(lightStyle).toContain("color: rgb(0, 0, 0)");
    expect(lightStyle).toContain("border: 1px solid rgb(173, 214, 255)");
  });

  it("shows 'No themes available' when themes list is empty", async () => {
    terminalThemeStore.$patch({ terminalThemes: [] });
    await flushPromises();
    expect(wrapper.text()).toContain("No themes available");
    expect(wrapper.findAll('[data-test="theme-item"]')).toHaveLength(0);
  });
});
