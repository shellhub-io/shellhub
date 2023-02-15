import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import Card from "../../../src/components/Card/Card.vue";

describe("Card", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(Card, {
      props: {
        id: 0,
        title: "Registered Users", 
        fieldObject: "registered_users",
        content: "Registered users",
        icon: "mdi-account-group",
        stats: 10,
        buttonName: "View all Users",
        pathName: "users",
        nameUseTest: "viewUsers-btn",
      },
      global: {
        plugins: [vuetify],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="viewUsers-btn"]').exists()).toBe(true);
  });

  it("Renders the title with data", () => {
    const title = wrapper.find('[title-test="Registered Users"]');
    expect(title.exists()).toBe(true);
    expect(title.text()).toBe("Registered Users");
  });

  it("Renders the stats with data", () => {
    const stats = wrapper.find('[stats-test="10"]');
    expect(stats.exists()).toBe(true);
    expect(stats.text()).toBe("10");
  });

  it("Renders the content with data", () => {
    const content = wrapper.find('[content-test="Registered users"]');
    expect(content.exists()).toBe(true);
    expect(content.text()).toBe("Registered users");
  });

  it("Renders the icon name with data", () => {
    const icon = wrapper.find('[iconName-test="mdi-account-group"]');
    expect(icon.exists()).toBe(true);
  });

  it("Renders the button name with data", () => {
    const buttonName = wrapper.find("[buttonName-test]");
    expect(buttonName.exists()).toBe(true);
    expect(buttonName.text()).toBe("View all Users");
  });
});