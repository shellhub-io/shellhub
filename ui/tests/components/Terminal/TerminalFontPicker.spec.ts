import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, VueWrapper, flushPromises } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { createPinia, setActivePinia } from "pinia";
import TerminalFontPicker from "@/components/Terminal/TerminalFontPicker.vue";
import useTerminalThemeStore from "@/store/modules/terminal_theme";

vi.mock("fontfaceobserver", () => ({
  default: vi.fn().mockImplementation(() => ({
    load: vi.fn().mockResolvedValue(undefined),
  })),
}));

describe("TerminalFontPicker.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalFontPicker>>;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const terminalThemeStore = useTerminalThemeStore();

  terminalThemeStore.setFontSettings = vi.fn().mockResolvedValue(undefined);
  terminalThemeStore.$patch({
    currentFontFamily: "Monospace",
    currentFontSize: 15,
  });

  beforeEach(async () => {
    wrapper = mount(TerminalFontPicker, {
      global: {
        plugins: [vuetify],
      },
    });
    await flushPromises();
  });

  it("renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("shows current font size value", async () => {
    const fontSizeInput = wrapper.find(".v-number-input input");
    expect((fontSizeInput.element as HTMLInputElement).value).toBe("15");
  });

  it("calls setFontSettings and emits update:fontSettings when font family changes", async () => {
    const fontFamilySelect = wrapper.findComponent({ name: "VSelect" });

    await fontFamilySelect.setValue("Fira Code");
    await flushPromises();

    expect(terminalThemeStore.setFontSettings).toHaveBeenCalledWith("Fira Code", 15);

    const emitted = wrapper.emitted("update:fontSettings");
    expect(emitted).toHaveLength(1);
    expect(emitted?.[0][0]).toEqual({
      fontFamily: "Fira Code",
      fontSize: 15,
    });
  });

  it("calls setFontSettings and emits update:fontSettings when font size changes", async () => {
    const fontSizeInput = wrapper.find(".v-number-input input");

    await fontSizeInput.setValue(18);
    await flushPromises();

    expect(terminalThemeStore.setFontSettings).toHaveBeenCalledWith("Fira Code", 18);

    const emitted = wrapper.emitted("update:fontSettings");
    expect(emitted).toHaveLength(1);
    expect(emitted?.[0][0]).toEqual({
      fontFamily: "Fira Code",
      fontSize: 18,
    });
  });

  it("uses storeToRefs for reactive store properties", async () => {
    terminalThemeStore.$patch({
      currentFontFamily: "Ubuntu Mono",
      currentFontSize: 22,
    });
    await flushPromises();

    const fontSelect = wrapper.find(".v-select input");
    const fontSizeInput = wrapper.find(".v-number-input input");

    expect((fontSelect.element as HTMLInputElement).value).toBe("Ubuntu Mono");
    expect((fontSizeInput.element as HTMLInputElement).value).toBe("22");
  });
});
