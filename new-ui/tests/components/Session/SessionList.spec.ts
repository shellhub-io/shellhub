import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SessionList from "../../../src/components/Sessions/SessionList.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const numberSessionsGlobal = 2;
const sessionPlayShow = [false, false];
const sessionCloseShow = [false, false];

const pagination = {
  groupBy: [],
  groupDesc: [],
  itemsPerPage: 10,
  multiSort: false,
  mustSort: false,
  page: 1,
  sortBy: ["started_at"],
  sortDesc: [true],
};

const sessionsGlobal = [
  {
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
      online: false,
      namespace: "user",
    },
    tenant_id: "00000000",
    username: "user",
    ip_address: "00.00.00",
    started_at: "2020-05-18T12:30:28.824Z",
    last_seen: "2020-05-18T12:30:30.205Z",
    active: true,
    authenticated: false,
  },
  {
    uid: "8c354a01",
    device_uid: "a582b47a",
    device: {
      uid: "a582b47a",
      name: "b4-2e-99",
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
      online: false,
      namespace: "user",
    },
    tenant_id: "00000000",
    username: "user",
    ip_address: "00.00.00",
    started_at: "2020-05-18T12:30:28.824Z",
    last_seen: "2020-05-18T12:30:30.205Z",
    active: false,
    authenticated: false,
  },
];

const headers = [
  {
    text: "Active",
    value: "active",
  },
  {
    text: "Device",
    value: "device",
  },
  {
    text: "Username",
    value: "username",
  },
  {
    text: "Authenticated",
    value: "authenticated",
  },
  {
    text: "IP Address",
    value: "ip_address",
  },
  {
    text: "Started",
    value: "started",
  },
  {
    text: "Last Seen",
    value: "last_seen",
  },
  {
    text: "Actions",
    value: "actions",
  },
];

const tests = [
  {
    description: "Sessions has enterprise version when user has owner role",
    role: {
      type: "owner",
      permission: true,
    },
    variables: {
      sessions: sessionsGlobal,
      numberSessions: numberSessionsGlobal,
      enterprise: true,
    },
    data: {
      menu: false,
      loading: false,
      itemsPerPage: 10,
      page: 1,
      pagination,
      headers,
    },
    computed: {
      getListSessions: sessionsGlobal,
      getNumberSessions: numberSessionsGlobal,
      isEnterprise: true,
      hasAuthorizationPlay: true,
    },
  },
  {
    description: "Sessions has enterprise version when user has observer role",
    role: {
      type: "observer",
      permission: false,
    },
    variables: {
      sessions: sessionsGlobal,
      numberSessions: numberSessionsGlobal,
      enterprise: true,
    },
    data: {
      menu: false,
      loading: false,
      itemsPerPage: 10,
      page: 1,
      pagination,
      headers,
    },
    computed: {
      getListSessions: sessionsGlobal,
      getNumberSessions: numberSessionsGlobal,
      isEnterprise: true,
      hasAuthorizationPlay: false,
    },
  },
  {
    description: "Sessions has no enterprise version",
    role: {
      type: "owner",
      permission: true,
    },
    variables: {
      sessions: sessionsGlobal,
      numberSessions: numberSessionsGlobal,
      enterprise: false,
    },
    data: {
      pagination,
      menu: false,
      loading: false,
      itemsPerPage: 10,
      page: 1,
      headers,
    },
    computed: {
      getListSessions: sessionsGlobal,
      getNumberSessions: numberSessionsGlobal,
      isEnterprise: false,
      hasAuthorizationPlay: true,
    },
  },
];

const store = (sessions: any, numberSessions: any, currentRole: any) => {
  return createStore({
    state: {
      sessions,
      numberSessions,
      currentRole,
    },
    getters: {
      "sessions/list": (state) => state.sessions,
      "sessions/getNumberSessions": (state) => state.numberSessions,
      "box/getStatus": () => true,
      "auth/role": (state) => state.currentRole,
    },
    actions: {
      "sessions/fetch": vi.fn(),
      "sessions/close": vi.fn(),
      "snackbar/showSnackbarErrorAssociation": vi.fn(),
      "snackbar/showSnackbarErrorLoading": vi.fn(),
      "snackbar/setSnackbarErrorDefault": vi.fn(),
      "box/setStatus": vi.fn(),
    },
  });
};
describe("SessionList", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(SessionList, {
          global: {
            plugins: [
              [
                store(
                  test.variables.sessions,
                  test.variables.numberSessions,
                  test.role.type
                ),
                key,
              ],
              routes,
              vuetify,
            ],
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
      it("Data is defined", () => {
        expect(wrapper.vm.$data).toBeDefined();
      });
      it("Receive props in props", () => {
        expect(wrapper.vm.loading).toBe(test.data.loading);
        expect(wrapper.vm.itemsPerPage).toBe(test.data.itemsPerPage);
        expect(wrapper.vm.page).toBe(test.data.page);
        expect(wrapper.vm.headers).toEqual(test.data.headers);
      });
      it("Check the computed", () => {
        expect(wrapper.vm.sessions).toEqual(test.computed.getListSessions);
        expect(wrapper.vm.numberSessions).toEqual(
          test.computed.getNumberSessions
        );
      });

      //////
      // HTML validation
      //////

      it("Renders the template with data", () => {
        expect(
          wrapper.find('[data-test="sessions-list"]').exists()
        ).toBeTruthy();
      });
    });
  });
});
