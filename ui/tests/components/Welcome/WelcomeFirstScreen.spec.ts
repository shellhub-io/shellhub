import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import WelcomeFirstScreen from "@/components/Welcome/WelcomeFirstScreen.vue";
import { mountComponent } from "@tests/utils/mount";

describe("WelcomeFirstScreen", () => {
  let wrapper: VueWrapper<InstanceType<typeof WelcomeFirstScreen>>;

  const mountWrapper = (name = "John Doe") => {
    wrapper = mountComponent(WelcomeFirstScreen, {
      piniaOptions: { initialState: { auth: { name, username: "johndoe" } } },
    });
  };

  beforeEach(() => mountWrapper());

  afterEach(() => wrapper?.unmount());

  describe("Component rendering", () => {
    it("renders welcome message with user name", () => {
      const welcomeName = wrapper.find('[data-test="welcome-name"]');
      expect(welcomeName.exists()).toBe(true);
      expect(welcomeName.text()).toBe("Welcome, John Doe!");
    });

    it("renders welcome message with username when name is not available", () => {
      wrapper.unmount();
      mountWrapper("");

      const welcomeName = wrapper.find('[data-test="welcome-name"]');
      expect(welcomeName.text()).toBe("Welcome, johndoe!");
    });

    it("renders rocket icon", () => {
      const avatar = wrapper.find(".v-avatar");
      expect(avatar.exists()).toBe(true);
      expect(wrapper.html()).toContain("mdi-rocket-launch");
    });

    it("renders introductory text", () => {
      expect(wrapper.text()).toContain("Let's get you started with ShellHub");
      expect(wrapper.text()).toContain("ShellHub is a modern SSH server");
    });
  });

  describe("Feature cards", () => {
    it("renders all three feature cards", () => {
      const featureCards = wrapper.findAll(".v-card");
      expect(featureCards).toHaveLength(3);
    });

    it("renders Remote Access feature card", () => {
      const cards = wrapper.findAll(".v-card");
      const remoteAccessCard = cards[0];

      expect(remoteAccessCard.text()).toContain("Remote Access");
      expect(remoteAccessCard.text()).toContain("Access your Linux devices from anywhere");
      expect(remoteAccessCard.html()).toContain("mdi-monitor");
    });

    it("renders Secure Connection feature card", () => {
      const cards = wrapper.findAll(".v-card");
      const secureCard = cards[1];

      expect(secureCard.text()).toContain("Secure Connection");
      expect(secureCard.text()).toContain("Bypass firewalls and NAT");
      expect(secureCard.html()).toContain("mdi-shield-check");
    });

    it("renders Easy Setup feature card", () => {
      const cards = wrapper.findAll(".v-card");
      const easySetupCard = cards[2];

      expect(easySetupCard.text()).toContain("Easy Setup");
      expect(easySetupCard.text()).toContain("Automated access process");
      expect(easySetupCard.html()).toContain("mdi-cogs");
    });
  });
});
