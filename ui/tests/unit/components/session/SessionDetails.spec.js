import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import timezoneMock from 'timezone-mock';
import SessionDetails from '@/components/session/SessionDetails';

describe('SessionDetails', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  const sessionGlobal = {
    uid: '8c354a00',
    device_uid: 'a582b47a',
    device: {
      uid: 'a582b47a',
      name: '39-5e-2a',
      identity: {
        mac: '00:00:00',
      },
      info: {
        id: 'debian',
        pretty_name: 'Debian',
        version: 'v0.2.5',
      },
      public_key: 'xxxxxxxx',
      tenant_id: '00000000',
      last_seen: '2020-05-18T13:27:02.498Z',
      online: true,
      namespace: 'user',
    },
    tenant_id: '00000000',
    username: 'user',
    ip_address: '00.00.00',
    started_at: '2020-05-18T12:30:28.824Z',
    last_seen: '2020-05-18T12:30:30.205Z',
    active: true,
    authenticated: true,
    recorded: true,
  };

  const tests = [
    {
      description: 'Session recorded is true and device is online',
      variables: {
        session: sessionGlobal,
        enterprise: true,
      },
      data: {
        uid: sessionGlobal.uid,
        session: sessionGlobal,
        dialog: false,
        sessionPlayDialog: false,
        sessionCloseDialog: false,
        hide: true,
      },
      computed: {
        isEnterprise: true,
      },
      components: {
        'sessionPlay-component': true,
        'sessionClose-component': true,
        'sessionDeleteRecord-component': true,
      },
      template: {
        'sessionUid-field': true,
        'sessionUser-field': true,
        'sessionIpAddress-field': true,
        'sessionStartedAt-field': true,
        'sessionLastSeen-field': true,
      },
      templateText: {
        'sessionUid-field': sessionGlobal.uid,
        'sessionUser-field': sessionGlobal.username,
        'sessionIpAddress-field': sessionGlobal.ip_address,
        'sessionStartedAt-field': 'Monday, May 18th 2020, 12:30:28 pm',
        'sessionLastSeen-field': 'Monday, May 18th 2020, 12:30:30 pm',
      },
    },
    {
      description: 'Session recorded is false and device is online',
      variables: {
        session: { ...sessionGlobal, recorded: false },
        enterprise: true,
      },
      data: {
        uid: sessionGlobal.uid,
        session: { ...sessionGlobal, recorded: false },
        dialog: false,
        sessionPlayDialog: false,
        sessionCloseDialog: false,
        hide: true,
      },
      computed: {
        isEnterprise: true,
      },
      components: {
        'sessionPlay-component': false,
        'sessionClose-component': false,
        'sessionDeleteRecord-component': false,
      },
      template: {
        'sessionUid-field': true,
        'sessionUser-field': true,
        'sessionIpAddress-field': true,
        'sessionStartedAt-field': true,
        'sessionLastSeen-field': true,
      },
      templateText: {
        'sessionUid-field': sessionGlobal.uid,
        'sessionUser-field': sessionGlobal.username,
        'sessionIpAddress-field': sessionGlobal.ip_address,
        'sessionStartedAt-field': 'Monday, May 18th 2020, 12:30:28 pm',
        'sessionLastSeen-field': 'Monday, May 18th 2020, 12:30:30 pm',
      },
    },
    {
      description: 'Session recorded is false and device is offline',
      variables: {
        session: {
          ...sessionGlobal, device: { online: false }, active: false, recorded: false,
        },
        enterprise: true,
      },
      data: {
        uid: sessionGlobal.uid,
        session: {
          ...sessionGlobal, device: { online: false }, active: false, recorded: false,
        },
        dialog: false,
        sessionPlayDialog: false,
        sessionCloseDialog: false,
        hide: true,
      },
      computed: {
        isEnterprise: true,
      },
      components: {
        'sessionPlay-component': false,
        'sessionClose-component': false,
        'sessionDeleteRecord-component': false,
      },
      template: {
        'sessionUid-field': true,
        'sessionUser-field': true,
        'sessionIpAddress-field': true,
        'sessionStartedAt-field': true,
        'sessionLastSeen-field': true,
      },
      templateText: {
        'sessionUid-field': sessionGlobal.uid,
        'sessionUser-field': sessionGlobal.username,
        'sessionIpAddress-field': sessionGlobal.ip_address,
        'sessionStartedAt-field': 'Monday, May 18th 2020, 12:30:28 pm',
        'sessionLastSeen-field': 'Monday, May 18th 2020, 12:30:30 pm',
      },
    },
    {
      description: 'Session recorded is true and device is online, but not enterprise',
      variables: {
        session: sessionGlobal,
        enterprise: false,
      },
      data: {
        uid: sessionGlobal.uid,
        session: sessionGlobal,
        dialog: false,
        sessionPlayDialog: false,
        sessionCloseDialog: false,
        hide: true,
      },
      computed: {
        isEnterprise: false,
      },
      components: {
        'sessionPlay-component': false,
        'sessionClose-component': true,
        'sessionDeleteRecord-component': true,
      },
      template: {
        'sessionUid-field': true,
        'sessionUser-field': true,
        'sessionIpAddress-field': true,
        'sessionStartedAt-field': true,
        'sessionLastSeen-field': true,
      },
      templateText: {
        'sessionUid-field': sessionGlobal.uid,
        'sessionUser-field': sessionGlobal.username,
        'sessionIpAddress-field': sessionGlobal.ip_address,
        'sessionStartedAt-field': 'Monday, May 18th 2020, 12:30:28 pm',
        'sessionLastSeen-field': 'Monday, May 18th 2020, 12:30:30 pm',
      },
    },
  ];

  const storeVuex = (session) => new Vuex.Store({
    namespaced: true,
    state: {
      session,
    },
    getters: {
      'sessions/get': (state) => state.session,
    },
    actions: {
      'sessions/get': () => {},
      'sessions/close': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      timezoneMock.register('UTC');

      const wrapper = shallowMount(SessionDetails, {
        store: storeVuex(
          test.variables.session,
        ),
        localVue,
        stubs: ['fragment'],
        mocks: {
          $route: {
            params: {
              id: test.variables.session.uid,
            },
          },
          $env: {
            isEnterprise: test.variables.enterprise,
          },
        },
      });

      ///////
      // Component Rendering
      //////

      it('Is a Vue instance', () => {
        expect(wrapper).toBeTruthy();
      });
      it('Renders the component', () => {
        expect(wrapper.html()).toMatchSnapshot();
      });

      ///////
      // Data checking
      //////

      it('Compare data with default value', () => {
        Object.keys(test.data).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.data[item]);
        });
      });
      it('Process data in the computed', () => {
        Object.keys(test.computed).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.computed[item]);
        });
      });

      //////
      // HTML validation
      //////

      it('Renders the template with components', () => {
        Object.keys(test.components).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.components[item]);
        });
      });
      it('Renders the template with data', () => {
        Object.keys(test.template).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
        });
      });
      it('Renders template with expected text', () => {
        Object.keys(test.templateText).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).text()).toContain(test.templateText[item]);
        });
      });
    });
  });
});
