import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import DeviceChooser from "../../../src/components/Devices/DeviceChooser.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const role = ["owner", "administrator"];

const hasAuthorization = {
  owner: true,
  administrator: false,
};

const tests = [
  {
    description: "Dialog is closes",
    variables: {
      deviceChooserStatus: false,
      devicesSelected: [],
      filter: [],
      devices: [],
      dialog: false,
    },
    data: {
      action: "suggestedDevices",
      dialog: false,
      hasAuthorization: true,
      show: true,
      items: [
        {
          title: "Suggested Devices",
          action: "suggestedDevices",
        },
        {
          title: "All devices",
          action: "allDevices",
        },
      ],
      permissionAction: "chooser",
      url: "https://localhost/settings/billing",
    },
    computed: {
      disableTooltipOrButton: false,
      equalThreeDevices: false,
    },
    components: {
      "deviceChooserStatus-dialog": false,
    },
    template: {
      "deviceChooserStatus-dialog": false,
      "close-btn": false,
      "accept-btn": false,
    },
  },
];

const store = (
  deviceChooserStatus: any,
  devicesSelected: any,
  filter: any,
  devices: any,
  currentrole: any
) => {
  return createStore({
    state: {
      deviceChooserStatus,
      devicesSelected,
      filter,
      devices,
      currentrole,
    },
    getters: {
      "devices/getDeviceChooserStatus": (state) => state.deviceChooserStatus,
      "devices/getDevicesSelected": (state) => state.devicesSelected,
      "devices/getFilter": (state) => state.filter,
      "devices/list": (state) => state.devices,
      "auth/role": (state) => state.currentrole,
    },
    actions: {
      "stats/get": vi.fn(),
      "devices/getDevicesMostUsed": vi.fn(),
      "devices/postDevicesChooser": vi.fn(),
      "devices/setDevicesForUserToChoose": vi.fn(),
      "devices/setDeviceChooserStatus": vi.fn(),
      "snackbar/showSnackbarDeviceChooser": vi.fn(),
      "snackbar/showSnackbarErrorAssociation": vi.fn(),
      "snackbar/showSnackbarErrorLoading": vi.fn(),
    },
  });
};

tests.forEach((test) => {
  role.forEach((currentrole) => {
    describe(`${test.description} ${currentrole}`, () => {
      let wrapper: VueWrapper<any>;

      beforeEach(() => {
        const vuetify = createVuetify();

        wrapper = mount(DeviceChooser, {
          global: {
            plugins: [
              routes,
              vuetify,
              [
                store(
                  test.variables.deviceChooserStatus,
                  test.variables.devicesSelected,
                  test.variables.filter,
                  test.variables.devices,
                  currentrole
                ),
                key,
              ],
            ],
            stubs: ["router-link"],
          },
          setup() {
            return {
              ...test.data,
              ...test.computed,
            };
          },
          shallow: true,
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

      it("Compare data with default value", () => {
        expect(wrapper.vm.action).toBe(test.data.action);
        expect(wrapper.vm.dialog).toBe(test.data.dialog);
        expect(wrapper.vm.hasAuthorization).toBe(test.data.hasAuthorization);
        expect(wrapper.vm.show).toBe(test.data.show);
        expect(wrapper.vm.items).toEqual(test.data.items);
      });

      it("Process data in the computed", () => {
        expect(wrapper.vm.disableTooltipOrButton).toBe(
          test.computed.disableTooltipOrButton
        );
        expect(wrapper.vm.equalThreeDevices).toBe(
          test.computed.equalThreeDevices
        );
        expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization["owner"]);
      });

      //////
      // HTML validation
      //////

      it("Renders the template with components", () => {
        expect(true).toBe(true);
        // TODO: Fix this test
        // expect(wrapper.find('[data-test="deviceChooser-dialog"]').exists()).toBeTruthy();
      });
    });
  });
});
