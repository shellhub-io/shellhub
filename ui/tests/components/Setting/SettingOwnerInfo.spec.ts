import { describe, expect, it } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import SettingOwnerInfo from "@/components/Setting/SettingOwnerInfo.vue";

describe("SettingOwnerInfo", () => {
  const wrapper = mountComponent(SettingOwnerInfo);

  describe("Message rendering", () => {
    it("Renders message container", () => {
      const messageDiv = wrapper.find('[data-test="message-div"]');
      expect(messageDiv.exists()).toBe(true);
    });

    it("Displays non-owner message", () => {
      const messageDiv = wrapper.find('[data-test="message-div"]');
      expect(messageDiv.text()).toContain("You're not the owner of this namespace.");
    });

    it("Displays contact user message", () => {
      const contactP = wrapper.find('[data-test="contactUser-p"]');
      expect(contactP.exists()).toBe(true);
      expect(contactP.text()).toBe("Contact the owner for more information.");
    });
  });
});
