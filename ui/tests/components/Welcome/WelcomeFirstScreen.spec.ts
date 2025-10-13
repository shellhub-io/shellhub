import { createPinia, setActivePinia } from "pinia";
import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it } from "vitest";
import WelcomeFirstScreen from "@/components/Welcome/WelcomeFirstScreen.vue";
import useAuthStore from "@/store/modules/auth";

const expectedFeatures = [
  { title: "Remote Access", icon: "mdi-monitor" },
  { title: "Secure Connection", icon: "mdi-shield-check" },
  { title: "Easy Setup", icon: "mdi-cogs" },
];

const authData = {
  token: "",
  username: "test",
  name: "test",
  tenantId: "fake-tenant-data",
  email: "test@test.com",
  id: "xxxxxxxx",
  role: "owner",
};

describe("Welcome First Screen", () => {
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  authStore.$patch(authData);

  const wrapper = mount(WelcomeFirstScreen, { global: { plugins: [createVuetify()] } });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the right username", () => {
    expect(wrapper.find("[data-test='welcome-name']").text()).toEqual("Welcome, test!");
  });

  it("Renders all feature cards", () => {
    const featureCards = wrapper.findAll(".v-card");
    expect(featureCards).toHaveLength(3);
  });

  it("Renders feature cards with correct content", () => {
    expectedFeatures.forEach((feature, index) => {
      const cardTitle = wrapper.findAll(".v-card-title")[index];
      const cardIcon = wrapper.findAll(".v-icon")[index + 1]; // +1 to skip rocket icon

      expect(cardTitle.text()).toBe(feature.title);
      expect(cardIcon.classes()).toContain(feature.icon);
    });
  });
});
