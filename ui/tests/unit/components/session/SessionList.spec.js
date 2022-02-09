import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import timezoneMock from 'timezone-mock';
import SessionList from '@/components/session/SessionList';
import { actions, authorizer } from '../../../../src/authorizer';

describe('SessionList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();

  localVue.use(Vuex);

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
    sortBy: [
      'started_at',
    ],
    sortDesc: [
      true,
    ],
  };

  const sessionsGlobal = [
    {
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
        online: false,
        namespace: 'user',
      },
      tenant_id: '00000000',
      username: 'user',
      ip_address: '00.00.00',
      started_at: '2020-05-18T12:30:28.824Z',
      last_seen: '2020-05-18T12:30:30.205Z',
      active: true,
      authenticated: false,
    },
    {
      uid: '8c354a01',
      device_uid: 'a582b47a',
      device: {
        uid: 'a582b47a',
        name: 'b4-2e-99',
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
        online: false,
        namespace: 'user',
      },
      tenant_id: '00000000',
      username: 'user',
      ip_address: '00.00.00',
      started_at: '2020-05-18T12:30:28.824Z',
      last_seen: '2020-05-18T12:30:30.205Z',
      active: false,
      authenticated: false,
    },
  ];

  const headers = [
    {
      text: 'Active',
      value: 'active',
      align: 'center',
    },
    {
      text: 'Device',
      value: 'device',
      align: 'center',
    },
    {
      text: 'Username',
      value: 'username',
      align: 'center',
    },
    {
      text: 'Authenticated',
      value: 'authenticated',
      align: 'center',
    },
    {
      text: 'IP Address',
      value: 'ip_address',
      align: 'center',
    },
    {
      text: 'Started',
      value: 'started',
      align: 'center',
    },
    {
      text: 'Last Seen',
      value: 'last_seen',
      align: 'center',
    },
    {
      text: 'Actions',
      value: 'actions',
      align: 'center',
    },
  ];

  const tests = [
    {
      description: 'Sessions has enterprise version',
      variables: {
        sessions: sessionsGlobal,
        numberSessions: numberSessionsGlobal,
        enterprise: true,
      },
      data: {
        menu: false,
        pagination,
        sessionPlayShow,
        sessionCloseShow,
        headers,
      },
      computed: {
        getListSessions: sessionsGlobal,
        getNumberSessions: numberSessionsGlobal,
        isEnterprise: true,
      },
    },
    {
      description: 'Sessions has no enterprise version',
      variables: {
        sessions: sessionsGlobal,
        numberSessions: numberSessionsGlobal,
        enterprise: false,
      },
      data: {
        menu: false,
        pagination,
        sessionPlayShow,
        sessionCloseShow,
        headers,
      },
      computed: {
        getListSessions: sessionsGlobal,
        getNumberSessions: numberSessionsGlobal,
        isEnterprise: false,
      },
    },
  ];

  const storeVuex = (sessions, numberSessions) => new Vuex.Store({
    namespaced: true,
    state: {
      sessions,
      numberSessions,
    },
    getters: {
      'sessions/list': (state) => state.sessions,
      'sessions/getNumberSessions': (state) => state.numberSessions,
    },
    actions: {
      'sessions/fetch': () => {},
      'sessions/close': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      timezoneMock.register('UTC');

      const wrapper = mount(SessionList, {
        store: storeVuex(
          test.variables.sessions,
          test.variables.numberSessions,
        ),
        localVue,
        stubs: ['fragment', 'router-link'],
        vuetify,
        mocks: {
          $authorizer: authorizer,
          $actions: actions,
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

      it('Renders the template with data', () => {
        const dt = wrapper.find('[data-test="dataTable-field"]');
        const dataTableProps = dt.vm.$options.propsData;

        expect(dataTableProps.items).toHaveLength(numberSessionsGlobal);
      });
    });
  });
});
