import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import Account from "../../../src/components/Account/AccountCreated.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "users/resendEmail": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("Account Render", () => {
  let wrapper: VueWrapper<InstanceType<typeof Account>>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(Account, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: {
        show: true,
        username: "test",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it("Receive data in props", () => {
    expect(wrapper.vm.show).toEqual(true);
    expect(wrapper.vm.username).toEqual("test");
  });
  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="accountCreated-card"]').exists()).toEqual(
      true,
    );
    expect(wrapper.find('[data-test="resendEmail-btn"]').exists()).toEqual(
      true,
    );
  });
});

describe("Doesn't render component", () => {
  let wrapper: VueWrapper<InstanceType<typeof Account>>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(Account, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: {
        show: false,
        username: "test",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Receive data in props", () => {
    expect(wrapper.vm.show).toEqual(false);
    expect(wrapper.vm.username).toEqual("test");
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="accountCreated-card"]').exists()).toEqual(
      false,
    );
    expect(wrapper.find('[data-test="resendEmail-btn"]').exists()).toEqual(
      false,
    );
  });
});
