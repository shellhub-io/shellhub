import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import WelcomeFourthScreen from "../../../src/components/Welcome/WelcomeFourthScreen.vue";
import routes from "../../../src/router";

describe("WelcomeFourthScreen", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const uid = "a582b47a42d";
  const username = "user";
  const password = "user";

  beforeEach(() => {
    wrapper = mount(WelcomeFourthScreen, {
      global: {
        plugins: [routes, vuetify],
      },
      props: {
        uid,
        username,
        password,
      },
    });
  });

  ///////
  // Component Rendering
  //////

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  ///////
  // Data checking
  //////
  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  ///////
  // HTML checking
  //////

  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="welcome-fourth-succesfully"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="welcome-fourth-links"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="welcome-fourth-thanks"]').exists()).toBeTruthy();
  });
});
