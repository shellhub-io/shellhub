import { describe, expect, it } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import MfaMailRecover from "@/components/AuthMFA/MfaMailRecover.vue";
import { createCleanRouter } from "@tests/utils/router";

describe("MfaMailRecover", () => {
  const wrapper = mountComponent(MfaMailRecover, { global: { plugins: [createCleanRouter()] } });

  describe("rendering", () => {
    it("displays MFA mail recovery confirmation", () => {
      expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="title"]').text()).toContain("Multi-factor Authentication");
    });

    it("shows email sent confirmation message", () => {
      expect(wrapper.find('[data-test="sub-title"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("An email has been sent to the primary and recovery mail");
      expect(wrapper.text()).toContain("Please check your inbox and click the link we've provided to disable the MFA");
    });

    it("displays back to login link", () => {
      expect(wrapper.find('[data-test="back-to-login"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("Back to");
    });

    it("displays login navigation link", () => {
      const loginLink = wrapper.find('[data-test="login-btn"]');
      expect(loginLink.exists()).toBe(true);
      expect(loginLink.text()).toBe("Login");
    });
  });

  describe("navigation", () => {
    it("links to login page", () => {
      const loginLink = wrapper.find('[data-test="login-btn"]');
      expect(loginLink.attributes("href")).toBe("/login");
    });
  });
});
