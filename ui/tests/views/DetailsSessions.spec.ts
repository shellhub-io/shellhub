import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import DetailsSessions from "../../src/views/DetailsSessions.vue";
import { key } from "../../src/store";
import routes from "../../src/router";

describe("DetailsSessions", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const sessionGlobal = {
    uid: "8c354a00",
    device_uid: "a582b47a",
    device: {
      uid: "a582b47a",
      name: "39-5e-2a",
      identity: {
        mac: "00:00:00",
      },
      info: {
        id: "debian",
        pretty_name: "Debian",
        version: "v0.2.5",
      },
      public_key: "xxxxxxxx",
      tenant_id: "00000000",
      last_seen: "2020-05-18T13:27:02.498Z",
      online: true,
      namespace: "user",
    },
    tenant_id: "00000000",
    username: "user",
    ip_address: "00.00.00",
    started_at: "2020-05-18T12:30:28.824Z",
    last_seen: "2020-05-18T12:30:30.205Z",
    active: true,
    authenticated: true,
    recorded: true,
  };

  const tests = [
    {
      description:
        "Session recorded is true and device is online when user has owner role",
      role: {
        type: "owner",
        permission: true,
      },
      variables: {
        session: sessionGlobal,
        enterprise: true,
      },
      data: {
        uid: sessionGlobal.uid,
        session: sessionGlobal,
        dialog: false,
        hide: true,
      },
      computed: {
        isEnterprise: true,
      },
      template: {
        "sessionUid-field": true,
        "sessionUser-field": true,
        "sessionIpAddress-field": true,
        "sessionStartedAt-field": true,
        "sessionLastSeen-field": true,
      },
      templateText: {
        "sessionUid-field": sessionGlobal.uid,
        "sessionUser-field": sessionGlobal.username,
        "sessionIpAddress-field": sessionGlobal.ip_address,
        "sessionStartedAt-field": "Monday, May 18th 2020, 12:30:28 pm",
        "sessionLastSeen-field": "Monday, May 18th 2020, 12:30:30 pm",
      },
    },
    {
      description:
        "Session recorded is false and device is online when user has owner role",
      role: {
        type: "owner",
        permission: true,
      },
      variables: {
        session: { ...sessionGlobal, recorded: false },
        enterprise: true,
      },
      data: {
        uid: sessionGlobal.uid,
        session: { ...sessionGlobal, recorded: false },
        dialog: false,
        hide: true,
      },
      computed: {
        isEnterprise: true,
      },
      components: {
        "sessionPlay-component": false,
        "sessionClose-component": false,
        "sessionDeleteRecord-component": false,
      },
      template: {
        "sessionUid-field": true,
        "sessionUser-field": true,
        "sessionIpAddress-field": true,
        "sessionStartedAt-field": true,
        "sessionLastSeen-field": true,
      },
      templateText: {
        "sessionUid-field": sessionGlobal.uid,
        "sessionUser-field": sessionGlobal.username,
        "sessionIpAddress-field": sessionGlobal.ip_address,
        "sessionStartedAt-field": "Monday, May 18th 2020, 12:30:28 pm",
        "sessionLastSeen-field": "Monday, May 18th 2020, 12:30:30 pm",
      },
    },
    {
      description:
        "Session recorded is false and device is offline when user has owner role",
      role: {
        type: "owner",
        permission: true,
      },
      variables: {
        session: {
          ...sessionGlobal,
          device: { online: false },
          active: false,
          recorded: false,
        },
        enterprise: true,
      },
      data: {
        uid: sessionGlobal.uid,
        session: {
          ...sessionGlobal,
          device: { online: false },
          active: false,
          recorded: false,
        },
        dialog: false,
        hide: true,
      },
      computed: {
        isEnterprise: true,
      },
      components: {
        "sessionPlay-component": false,
        "sessionClose-component": false,
        "sessionDeleteRecord-component": false,
      },
      template: {
        "sessionUid-field": true,
        "sessionUser-field": true,
        "sessionIpAddress-field": true,
        "sessionStartedAt-field": true,
        "sessionLastSeen-field": true,
      },
      templateText: {
        "sessionUid-field": sessionGlobal.uid,
        "sessionUser-field": sessionGlobal.username,
        "sessionIpAddress-field": sessionGlobal.ip_address,
        "sessionStartedAt-field": "Monday, May 18th 2020, 12:30:28 pm",
        "sessionLastSeen-field": "Monday, May 18th 2020, 12:30:30 pm",
      },
    },
    {
      description:
        "Session recorded is true and device is online, but not enterprise",
      role: {
        type: "owner",
        permission: true,
      },
      variables: {
        session: sessionGlobal,
        enterprise: false,
      },
      data: {
        uid: sessionGlobal.uid,
        session: sessionGlobal,
        dialog: false,
        hide: true,
      },
      computed: {
        isEnterprise: false,
      },
      components: {
        "sessionPlay-component": false,
        "sessionClose-component": true,
      },
      template: {
        "sessionUid-field": true,
        "sessionUser-field": true,
        "sessionIpAddress-field": true,
        "sessionStartedAt-field": true,
        "sessionLastSeen-field": true,
      },
      templateText: {
        "sessionUid-field": sessionGlobal.uid,
        "sessionUser-field": sessionGlobal.username,
        "sessionIpAddress-field": sessionGlobal.ip_address,
        "sessionStartedAt-field": "Monday, May 18th 2020, 12:30:28 pm",
        "sessionLastSeen-field": "Monday, May 18th 2020, 12:30:30 pm",
      },
    },
    {
      description:
        "Session recorded is true and device is online when user has observer role",
      role: {
        type: "observer",
        permission: false,
      },
      variables: {
        session: sessionGlobal,
        enterprise: true,
      },
      data: {
        uid: sessionGlobal.uid,
        session: sessionGlobal,
        dialog: false,
        hide: true,
      },
      computed: {
        isEnterprise: true,
      },
      template: {
        "sessionUid-field": true,
        "sessionUser-field": true,
        "sessionIpAddress-field": true,
        "sessionStartedAt-field": true,
        "sessionLastSeen-field": true,
      },
      templateText: {
        "sessionUid-field": sessionGlobal.uid,
        "sessionUser-field": sessionGlobal.username,
        "sessionIpAddress-field": sessionGlobal.ip_address,
        "sessionStartedAt-field": "Monday, May 18th 2020, 12:30:28 pm",
        "sessionLastSeen-field": "Monday, May 18th 2020, 12:30:30 pm",
      },
    },
  ];

  const store = (session: any, currentRole: any) => createStore({
    state: {
      session,
      currentRole,
    },
    getters: {
      "sessions/get": (state) => state.session,
      "auth/role": (state) => state.currentRole,
    },
    actions: {
      "sessions/get": vi.fn(),
      "sessions/close": vi.fn(),
      "snackbar/showSnackbarErrorLoading": vi.fn(),
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        wrapper = mount(DetailsSessions, {
          global: {
            plugins: [
              [store(test.variables.session, test.role.type), key],
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
        expect(wrapper.vm.session).toEqual(test.data.session);
      });

      //////
      // HTML validation
      //////

      it("Renders the template with data", () => {
        // eslint-disable-next-line no-restricted-syntax
        for (const [key, value] of Object.entries(test.template)) {
          expect(wrapper.find(`[data-test="${key}"]`).exists()).toBe(value);
        }
      });
      it("Renders template with expected text", () => {
        // eslint-disable-next-line no-restricted-syntax
        for (const [key, value] of Object.entries(test.templateText)) {
          expect(wrapper.find(`[data-test="${key}"]`).text()).toBe(value);
        }
      });
    });
  });
});
