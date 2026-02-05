import { describe, expect, it, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import SettingsRow from "@/components/Setting/SettingsRow.vue";

describe("SettingsRow", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingsRow>>;

  const defaultIcon = "mdi-test-icon";
  const defaultIconTestId = "row-icon";
  const defaultTitle = "Row Title";
  const defaultTitleTestId = "row-title";
  const defaultSubtitle = "Row subtitle";
  const defaultSubtitleTestId = "row-subtitle";
  const defaultTitleSlot = "";
  const defaultSubtitleSlot = "";
  const defaultSlotContent = '<div data-test="row-field">Field</div>';

  const mountWrapper = ({
    icon = defaultIcon,
    iconTestId = defaultIconTestId,
    title = defaultTitle,
    titleTestId = defaultTitleTestId,
    subtitle = defaultSubtitle,
    subtitleTestId = defaultSubtitleTestId,
    titleSlot = defaultTitleSlot,
    subtitleSlot = defaultSubtitleSlot,
    defaultSlot = defaultSlotContent,
  } = {}) => {
    wrapper = mountComponent(SettingsRow, {
      props: {
        icon,
        iconTestId,
        title,
        titleTestId,
        subtitle,
        subtitleTestId,
      },
      slots: {
        ...(titleSlot && { title: titleSlot }),
        ...(subtitleSlot && { subtitle: subtitleSlot }),
        default: defaultSlot,
      },
    });
  };

  afterEach(() => wrapper?.unmount());

  describe("Icon rendering", () => {
    it("Renders icon when provided", () => {
      mountWrapper();
      const icon = wrapper.find('[data-test="row-icon"]');
      expect(icon.exists()).toBe(true);
      expect(icon.classes().join(" ")).toContain("mdi-test-icon");
    });

    it("Does not render icon when not provided", () => {
      mountWrapper({ icon: undefined, iconTestId: "row-icon" });
      const icon = wrapper.find('[data-test="row-icon"]');
      expect(icon.attributes("aria-hidden")).toBe("true");
    });
  });

  describe("Title rendering", () => {
    it("Renders title from prop", () => {
      mountWrapper();
      const title = wrapper.find('[data-test="row-title"]');
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Row Title");
    });

    it("Renders title from slot", () => {
      mountWrapper({ title: "", titleSlot: '<span data-test="custom-title">Custom Title</span>' });

      const customTitle = wrapper.find('[data-test="custom-title"]');
      expect(customTitle.exists()).toBe(true);
      expect(customTitle.text()).toBe("Custom Title");
    });

    it("Does not render title container when neither prop nor slot provided", () => {
      mountWrapper({ title: "", titleTestId: "row-title" });

      const title = wrapper.find('[data-test="row-title"]');
      expect(title.exists()).toBe(false);
    });
  });

  describe("Subtitle rendering", () => {
    it("Renders subtitle from prop", () => {
      mountWrapper();
      const subtitle = wrapper.find('[data-test="row-subtitle"]');
      expect(subtitle.exists()).toBe(true);
      expect(subtitle.text()).toBe("Row subtitle");
    });

    it("Renders subtitle from slot", () => {
      mountWrapper({ subtitle: "", subtitleSlot: '<span data-test="custom-subtitle">Custom Subtitle</span>' });

      const customSubtitle = wrapper.find('[data-test="custom-subtitle"]');
      expect(customSubtitle.exists()).toBe(true);
      expect(customSubtitle.text()).toBe("Custom Subtitle");
    });

    it("Does not render subtitle container when neither prop nor slot provided", () => {
      mountWrapper({ subtitle: "", subtitleTestId: "row-subtitle" });

      const subtitle = wrapper.find('[data-test="row-subtitle"]');
      expect(subtitle.exists()).toBe(false);
    });
  });

  describe("Default slot rendering", () => {
    it("Renders default slot content", () => {
      mountWrapper();

      const field = wrapper.find('[data-test="row-field"]');
      expect(field.exists()).toBe(true);
      expect(field.text()).toBe("Field");
    });
  });
});
