import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import SettingSecurity from "../../../src/components/Setting/SettingSecurity.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("SettingSecurity", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingSecurity>>;
  const vuetify = createVuetify();

  const role = ["owner", "operator"];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

  const tests = [
    {
      description: "SettingSecurity",
      variables: {
        sessionRecord: false,
      },
      props: {
        hasTenant: true,
      },
    },
  ];

  const store = (sessionRecord: boolean, currentrole: string) => createStore({
    state: {
      sessionRecord,
      currentrole,
    },
    getters: {
      "security/get": (state) => state.sessionRecord,
      "auth/role": (state) => state.currentrole,
    },
    actions: {
      "security/set": vi.fn(),
      "security/get": vi.fn(),
      "snackbar/showSnackbarErrorDefault": vi.fn(),
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(SettingSecurity, {
            global: {
              plugins: [
                [store(test.variables.sessionRecord, currentrole), key],
                routes,
                vuetify,
              ],
            },
            props: {
              hasTenant: test.props.hasTenant,
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

        /// ////s
        // Data checking
        //////
        it("Data is defined", () => {
          expect(wrapper.vm.$data).toBeDefined();
        });
        it("Receive data in props", () => {
          expect(wrapper.vm.hasTenant).toBe(test.props.hasTenant);
        });
        it("Process data in the computed", () => {
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });
      });
    });
  });
});
