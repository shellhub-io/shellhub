import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it } from "vitest";
import WelcomeFourthScreen from "@/components/Welcome/WelcomeFourthScreen.vue";

describe("Welcome Fourth Screen", () => {
  const wrapper = mount(WelcomeFourthScreen, { global: { plugins: [createVuetify()] } });
  it("Renders the component", () => { expect(wrapper.html()).toMatchSnapshot(); });
});
