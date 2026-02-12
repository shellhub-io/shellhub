import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { flushPromises, VueWrapper } from "@vue/test-utils";
import TerminalThemeDrawer from "@/components/Terminal/TerminalThemeDrawer.vue";
import { mountComponent } from "@tests/utils/mount";
import { mockTerminalThemes } from "@tests/mocks/terminalTheme";
import { VLayout } from "vuetify/components";

const Component = {
  template: "<v-layout><TerminalThemeDrawer v-model=\"selectedTheme\" v-model:show-drawer=\"showDrawer\" /></v-layout>",
  data: () => ({
    showDrawer: true,
    selectedTheme: "ShellHub Dark",
  }),
};

describe("TerminalThemeDrawer", () => {
  let wrapper: VueWrapper;
  let drawer: VueWrapper<InstanceType<typeof TerminalThemeDrawer>>;

  const mountWrapper = () => {
    wrapper = mountComponent(Component, {
      global: {
        stubs: { teleport: true },
        components: { TerminalThemeDrawer, "v-layout": VLayout },
      },
      attachTo: document.body,
      piniaOptions: {
        initialState: {
          terminalTheme: {
            terminalThemes: mockTerminalThemes,
            currentThemeName: "ShellHub Dark",
            currentFontFamily: "Monospace",
            currentFontSize: 15,
            availableFonts: ["Monospace", "Fira Code", "Ubuntu Mono"],
          },
        },
      },
    });
    drawer = wrapper.findComponent(TerminalThemeDrawer);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  it("renders the navigation drawer when showDrawer is true", () => {
    const themeDrawer = drawer.find('[data-test="theme-drawer"]');
    expect(themeDrawer.exists()).toBe(true);
  });

  it("renders drawer on the right side", () => {
    const navigationDrawer = drawer.findComponent({ name: "VNavigationDrawer" });
    expect(navigationDrawer.props("location")).toBe("right");
  });

  it("sets drawer as temporary", () => {
    const navigationDrawer = drawer.findComponent({ name: "VNavigationDrawer" });
    expect(navigationDrawer.props("temporary")).toBe(true);
  });

  it("displays Font Settings section header", () => {
    expect(drawer.text()).toContain("Font Settings");
  });

  it("displays Color Theme section header", () => {
    expect(drawer.text()).toContain("Color Theme");
  });

  it("renders TerminalFontPicker component", () => {
    const fontPicker = drawer.findComponent({ name: "TerminalFontPicker" });
    expect(fontPicker.exists()).toBe(true);
  });

  it("renders TerminalThemePicker component", () => {
    const themePicker = drawer.findComponent({ name: "TerminalThemePicker" });
    expect(themePicker.exists()).toBe(true);
  });

  it("passes selectedTheme to TerminalThemePicker", () => {
    const themePicker = drawer.findComponent({ name: "TerminalThemePicker" });
    expect(themePicker.props("modelValue")).toBe("ShellHub Dark");
  });

  it("emits update:fontSettings when TerminalFontPicker emits", async () => {
    const fontPicker = drawer.findComponent({ name: "TerminalFontPicker" });
    const fontSettings = { fontFamily: "Fira Code", fontSize: 18 };

    await fontPicker.vm.$emit("update:fontSettings", fontSettings);
    await flushPromises();

    expect(drawer.emitted("update:fontSettings")).toBeTruthy();
    expect(drawer.emitted("update:fontSettings")?.[0]).toEqual([fontSettings]);
  });

  it("emits update:selectedTheme when TerminalThemePicker emits", async () => {
    const themePicker = drawer.findComponent({ name: "TerminalThemePicker" });
    const selectedTheme = mockTerminalThemes[1];

    await themePicker.vm.$emit("update:selectedTheme", selectedTheme);
    await flushPromises();

    expect(drawer.emitted("update:selectedTheme")).toBeTruthy();
    expect(drawer.emitted("update:selectedTheme")?.[0]).toEqual([selectedTheme]);
  });

  it("has divider between Font Settings and Color Theme sections", () => {
    const dividers = drawer.findAllComponents({ name: "VDivider" });
    expect(dividers.length).toBeGreaterThan(0);
  });

  it("sets correct drawer width", () => {
    const navigationDrawer = drawer.findComponent({ name: "VNavigationDrawer" });
    expect(navigationDrawer.props("width")).toBe("300");
  });

  it("emits update:showDrawer when drawer is closed", async () => {
    const navigationDrawer = drawer.findComponent({ name: "VNavigationDrawer" });
    await navigationDrawer.vm.$emit("update:modelValue", false);
    await flushPromises();

    expect(drawer.emitted("update:showDrawer")).toBeTruthy();
    expect(drawer.emitted("update:showDrawer")?.[0]).toEqual([false]);
  });
});
