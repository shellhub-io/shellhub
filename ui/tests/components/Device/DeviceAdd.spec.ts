import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import DeviceAdd from "../../../src/components/Devices/DeviceAdd.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const role = ["owner", "observer"];

const hasAuthorization = {
  owner: true,
  observer: false,
};

const tests = [
  {
    description: "Button",
    variables: {
      addDevice: false,
      tenant: "xxxxxxxx",
      dialog: false,
    },
    props: {
      size: "default",
    },
    data: {
      tenant: "xxxxxxxx",
      hostname: "localhost",
      port: "",
      dialog: false,
      action: "add",
    },
    computed: {
      tenant: "xxxxxxxx",
    },
    method: {
      command:
        'curl -sSf "http://localhost/install.sh?tenant_id=xxxxxxxx" | sh',
    },
    template: {
      "add-btn": true,
      "deviceAdd-dialog": false,
      "close-btn": false,
    },
  },
  {
    description: "Dialog",
    variables: {
      addDevice: true,
      tenant: "xxxxxxxx",
      dialog: true,
    },
    props: {
      size: "default",
    },
    data: {
      tenant: "xxxxxxxx",
      hostname: "localhost",
      port: "",
      dialog: true,
      action: "add",
    },
    computed: {
      tenant: "xxxxxxxx",
    },
    method: {
      command:
        'curl -sSf "http://localhost/install.sh?tenant_id=xxxxxxxx" | sh',
    },
    template: {
      "add-btn": true,
      "deviceAdd-dialog": true,
      "close-btn": true,
    },
  },
];

const store = (addDevice: boolean, tenant: string, currentrole: string) => createStore({
  state: {
    tenant,
    addDevice,
    currentrole,
  },
  getters: {
    "auth/tenant": (state) => state.tenant,
    "modals/addDevice": (state) => state.addDevice,
    "auth/role": (state) => state.currentrole,
  },
  actions: {
    "modals/showAddDevice": vi.fn(),
    "snackbar/showSnackbarCopy": vi.fn(),
  },
});

describe("DeviceAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceAdd>>;

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          const vuetify = createVuetify();
          wrapper = mount(DeviceAdd, {
            global: {
              plugins: [
                routes,
                vuetify,
                [
                  store(
                    test.variables.addDevice,
                    test.variables.tenant,
                    currentrole,
                  ),
                  key,
                ],
              ],
              stubs: ["router-link"],
            },
            props: {
              ...test.props,
            },
            setup() {
              return {
                ...test.data,
                ...test.method,
              };
            },
            shallow: true,
          });

          // wrapper.setData({ dialog: test.variables.dialog });
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
        it("Has the correct props", () => {
          expect(wrapper.vm.size).toBe(test.props.size);
        });

        it("Has the correct data", () => {
          expect(wrapper.vm.size).toBe(test.props.size);
          expect(wrapper.vm.dialog).toBe(test.data.dialog);
          expect(wrapper.vm.action).toBe(test.data.action);
          expect(wrapper.vm.hostname).toBe(test.data.hostname);
        });

        it("Has the correct computed", () => {
          expect(wrapper.vm.tenant).toBe(test.variables.tenant);
        });

        it("Process data in methods", () => {
          expect(wrapper.vm.command).toBe(test.method.command);
        });

        //////
        // HTML validation
        //////

        it("Has the correct HTML", () => {
          expect(
            wrapper.find('[data-test="device-add-btn"]').exists(),
          ).toBeTruthy();
        });
      });
    });
  });
});
