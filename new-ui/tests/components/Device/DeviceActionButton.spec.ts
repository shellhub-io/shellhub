import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import DeviceActionButton from "../../../src/components/Devices/DeviceActionButton.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";
import { authorizer, actions } from "../../../src/authorizer";

const role = ["owner", "observer"];

const hasAuthorization = {
  owner: true,
  observer: false,
};

const tests = [
  {
    description: "Create button in the notification",
    variables: {
      isActive: true,
    },
    props: {
      uid: "xxxxxxxx",
      notificationStatus: true,
      action: "accept",
      show: false,
    },
    data: {
      icon: "mdi-check",
    },
    template: {
      "notification-btn": true,
      "action-item": false,
      "action-icon": false,
      "deviceActionButton-card": false,
      "cancel-btn": false,
      "dialog-btn": false,
    },
  },
  {
    description: "Create button in the list",
    variables: {
      isActive: true,
    },
    props: {
      uid: "xxxxxxxx",
      notificationStatus: false,
      action: "accept",
      show: false,
    },
    data: {
      icon: "mdi-check",
    },
    template: {
      "notification-btn": false,
      "action-item": true,
      "action-icon": true,
      "deviceActionButton-card": false,
      "cancel-btn": false,
      "dialog-btn": false,
    },
  },
  {
    description: "Reject button in the list",
    variables: {
      isActive: true,
    },
    props: {
      uid: "xxxxxxxx",
      notificationStatus: false,
      action: "reject",
      show: false,
    },
    data: {
      icon: "mdi-close",
    },
    template: {
      "notification-btn": false,
      "action-item": true,
      "action-icon": true,
      "deviceActionButton-card": false,
      "cancel-btn": false,
      "dialog-btn": false,
    },
  },
  {
    description: "Remove button in the list",
    variables: {
      isActive: true,
    },
    props: {
      uid: "xxxxxxxx",
      notificationStatus: false,
      action: "remove",
      show: false,
    },
    data: {
      icon: "mdi-delete",
    },
    template: {
      "notification-btn": false,
      "action-item": true,
      "action-icon": true,
      "deviceActionButton-card": false,
      "cancel-btn": false,
      "dialog-btn": false,
    },
  },
  {
    description: "Dialog",
    variables: {
      isActive: true,
    },
    props: {
      uid: "xxxxxxxx",
      notificationStatus: false,
      action: "accept",
      show: true,
    },
    data: {
      icon: "mdi-check",
    },
    template: {
      "notification-btn": false,
      "deviceActionButton-card": true,
      "cancel-btn": true,
      "dialog-btn": true,
    },
  },
];

const store = (isActive: any, currentRole: any) => {
  return createStore({
    state: {
      isActive,
      currentRole,
    },
    getters: {
      isActive: (state) => state.isActive,
      "auth/role": (state) => state.currentRole,
    },
    actions: {
      "devices/refresh": vi.fn(),
      "devices/accept": vi.fn(),
      "users/setStatusUpdateAccountDialog": vi.fn(),
      "devices/reject": vi.fn(),
      "devices/remove": vi.fn(),
      "notifications/fetch": vi.fn(),
      "stats/get": vi.fn(),
      "snackbar/showSnackbarErrorDefault": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
      "snackbar/showSnackbarErrorLoading": vi.fn(),
    },
  });
};

tests.forEach((test) => {
  role.forEach((currentRole) => {
    describe(`DeviceActionButton: ${test.description} - ${currentRole}`, () => {
      let wrapper: VueWrapper<any>;

      beforeEach(() => {
        const vuetify = createVuetify();

        wrapper = mount(DeviceActionButton, {
          global: {
            plugins: [
              routes,
              vuetify,
              [store(test.variables.isActive, currentRole), key],
            ],
            stubs: ["router-link"],
          },
          shallow: true,
          props: {
            uid: test.props.uid,
            notificationStatus: test.props.notificationStatus,
            action: test.props.action,
            show: test.props.show,
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

      it("Renders the component with the correct data", () => {
        expect(wrapper.vm.icon).toEqual(test.data.icon);
      });

      ///////
      // Data checking
      //////
      it("Renders the component with the correct data", () => {
        expect(wrapper.vm.uid).toBe(test.props.uid);
        expect(wrapper.vm.notificationStatus).toBe(
          test.props.notificationStatus
        );
        expect(wrapper.vm.action).toBe(test.props.action);
      });

      it('Process data in the computed', () => {
        if (!(test.props.action === 'remove' && currentRole === 'operator')) {
          // @ts-ignore
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentRole]);
        } else {
          expect(wrapper.vm.hasAuthorization).toEqual(false);
        }
      });

      it("Renders the component with the correct template", () => {
        expect(wrapper.find('[data-test="notification-btn"')).toBeTruthy();
        expect(wrapper.find('[data-test="action-item"')).toBeTruthy();
      });
    });
  });
});
