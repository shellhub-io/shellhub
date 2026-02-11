import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import WelcomeFourthScreen from "@/components/Welcome/WelcomeFourthScreen.vue";
import { mountComponent } from "@tests/utils/mount";

describe("WelcomeFourthScreen", () => {
  let wrapper: VueWrapper<InstanceType<typeof WelcomeFourthScreen>>;

  beforeEach(() => wrapper = mountComponent(WelcomeFourthScreen));

  afterEach(() => wrapper?.unmount());

  describe("Component rendering", () => {
    it("renders success icon", () => {
      const avatar = wrapper.find(".v-avatar");
      expect(avatar.exists()).toBe(true);
      expect(wrapper.html()).toContain("mdi-check-circle");
    });

    it("renders all set heading", () => {
      expect(wrapper.text()).toContain("All Set!");
      expect(wrapper.text()).toContain("Your device has been successfully added");
    });

    it("renders welcome message", () => {
      expect(wrapper.text()).toContain("Welcome to ShellHub!");
      expect(wrapper.html()).toContain("mdi-party-popper");
    });

    it("renders thank you message", () => {
      const thanksElement = wrapper.find('[data-test="welcome-fourth-thanks"]');
      expect(thanksElement.exists()).toBe(true);
      expect(thanksElement.text()).toBe("Thank you for choosing ShellHub!");
      expect(wrapper.text()).toContain("We're excited to have you on board");
    });
  });

  describe("Helpful resources", () => {
    it("renders helpful resources section", () => {
      expect(wrapper.text()).toContain("Helpful Resources");
      expect(wrapper.html()).toContain("mdi-compass");
    });

    it("renders documentation link", () => {
      const buttons = wrapper.findAll('[data-test="welcome-fourth-links"]');
      const docLink = buttons[0];

      expect(docLink.exists()).toBe(true);
      expect(docLink.text()).toContain("Documentation");
      expect(docLink.attributes("href")).toBe("http://docs.shellhub.io/");
      expect(docLink.attributes("target")).toBe("_blank");
      expect(docLink.attributes("rel")).toBe("noopener noreferrer");
    });

    it("renders community chat link", () => {
      const buttons = wrapper.findAll("a.v-btn");
      const communityLink = buttons[1];

      expect(communityLink.exists()).toBe(true);
      expect(communityLink.text()).toContain("Community Chat");
      expect(communityLink.attributes("href")).toBe("https://gitter.im/shellhub-io/community");
      expect(communityLink.attributes("target")).toBe("_blank");
      expect(communityLink.attributes("rel")).toBe("noopener noreferrer");
    });

    it("renders external link icons", () => {
      expect(wrapper.html()).toContain("mdi-open-in-new");
    });
  });
});
