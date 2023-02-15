import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import WelcomeSecondScreen from "../../../src/components/Welcome/WelcomeSecondScreen.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("WelcomeSecondScreen", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const command =
    'curl -sSf "http://localhost/install.sh?tenant_id=a582b47a42e" | sh';

  const store = createStore({
    state: {},
    getters: {},
    actions: {
      "snackbar/showSnackbarCopy": vi.fn(),
    },
  });

  beforeEach(() => {
    wrapper = mount(WelcomeSecondScreen, {
      global: {
        plugins: [[store, key], routes, vuetify],
      },
      props: {
        command,
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
  it("Receive data in props", () => {
    expect(wrapper.vm.command).toEqual(command);
  });


  //////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    const commandText = wrapper.find('[data-test="command-field"]');
    expect(commandText.exists()).toBe(true);
  });
});
