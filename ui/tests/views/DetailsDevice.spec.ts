import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import DetailsDevice from "../../src/views/DetailsDevice.vue";
import { key } from "../../src/store";
import routes from "../../src/router";

describe("DetailsDevice", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const deviceOnline = {
    uid: "a582b47a",
    name: "39-5e-2a",
    identity: {
      mac: "00:00:00",
    },
    info: {
      id: "arch",
      pretty_name: "Linux",
      version: "latest",
    },
    public_key: "xxxxxxxx",
    tenant_id: "00000000",
    last_seen: "2020-05-20T18:58:53.276Z",
    online: true,
    namespace: "user",
    status: "accepted",
    tags: ["device1", "device2"],
  };

  const deviceOffline = { ...deviceOnline, online: false, status: "pending" };

  const tests = [
    {
      description: "Online Device and owner",
      role: {
        type: "owner",
        permission: true,
      },
      variables: {
        device: deviceOnline,
      },
      data: {
        uid: deviceOnline.uid,
        hostname: "localhost",
        hide: true,
        device: deviceOnline,
        dialogDelete: false,
        dialogError: false,
        deviceDeleteShow: false,
      },
      computed: {
        hasAuthorizationRename: true,
        hasAuthorizationFormUpdate: true,
      },
      components: {
        "deviceRename-component": true,
        "tagFormUpdate-component": true,
        "terminalDialog-component": true,
        "deviceDelete-component": true,
      },
      template: {
        "deviceUid-field": true,
        "deviceMac-field": true,
        "devicePrettyName-field": true,
        "deviceVersion-field": true,
        "deviceConvertDate-field": true,
      },
      templateText: {
        "deviceUid-field": deviceOnline.uid,
        "deviceMac-field": deviceOnline.identity.mac,
        "devicePrettyName-field": deviceOnline.info.pretty_name,
        "deviceVersion-field": deviceOnline.info.version,
        "deviceConvertDate-field": "Wednesday, May 20th 2020, 6:58:53 pm",
      },
    },
    {
      description: "Offline Device",
      role: {
        type: "observer",
        permission: false,
      },
      variables: {
        device: deviceOffline,
      },
      data: {
        uid: deviceOffline.uid,
        hostname: "localhost",
        hide: true,
        device: deviceOffline,
        dialogDelete: false,
        dialogError: false,
        deviceDeleteShow: false,
      },
      computed: {
        hasAuthorizationRename: false,
        hasAuthorizationFormUpdate: false,
      },
      components: {
        "deviceRename-component": true,
        "tagFormUpdate-component": true,
        "terminalDialog-component": false,
        "deviceDelete-component": true,
      },
      template: {
        "deviceUid-field": true,
        "deviceMac-field": true,
        "devicePrettyName-field": true,
        "deviceVersion-field": true,
        "deviceConvertDate-field": true,
      },
      templateText: {
        "deviceUid-field": deviceOffline.uid,
        "deviceMac-field": deviceOffline.identity.mac,
        "devicePrettyName-field": deviceOffline.info.pretty_name,
        "deviceVersion-field": deviceOffline.info.version,
        "deviceConvertDate-field": "Wednesday, May 20th 2020, 6:58:53 pm",
      },
    },
  ];

  const store = (device: any, currentRole: any) => createStore({
    state: {
      device,
      currentRole,
    },
    getters: {
      "devices/get": (state) => state.device,
      "auth/role": (state) => state.currentRole,
    },
    actions: {
      "devices/get": vi.fn(),
      "devices/updateTag": vi.fn(),
      "snackbar/showSnackbarSuccessAction": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
      "snackbar/showSnackbarErrorLoading": vi.fn(),
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        wrapper = mount(DetailsDevice, {
          global: {
            plugins: [
              [store(test.variables.device, test.role), key],
              vuetify,
              routes,
            ],
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

      it("Process data in the computed", () => {
        expect(wrapper.vm.device).toStrictEqual(test.data.device);
      });

      //////
      // HTML validation
      //////

      it("Renders the template with data", () => {
        Object.keys(test.template).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(
            test.template[item],
          );
        });
      });

      it("Renders template with expected text", () => {
        Object.keys(test.templateText).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).text()).toContain(
            test.templateText[item],
          );
        });
      });
    });
  });
});
