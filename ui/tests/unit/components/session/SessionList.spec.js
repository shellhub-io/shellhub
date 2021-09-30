import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import timezoneMock from 'timezone-mock';
import SessionList from '@/components/session/SessionList';

describe('SessionList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();

  localVue.use(Vuex);

  let wrapper;

  const status = true;
  const numberSessions = 2;

  const sessions = [
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

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      sessions,
      numberSessions,
      status,
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

  beforeEach(() => {
    timezoneMock.register('UTC');

    wrapper = mount(SessionList, {
      store,
      localVue,
      stubs: ['fragment', 'router-link'],
      mocks: {
        $env: (isEnterprise) => isEnterprise,
      },
      vuetify,
    });
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
  // Data and Props checking
  //////

  it('Compare data with default value', () => {
    expect(wrapper.vm.headers).toEqual(headers);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getListSessions).toEqual(sessions);
    expect(wrapper.vm.getNumberSessions).toEqual(numberSessions);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;

    expect(dataTableProps.items).toHaveLength(numberSessions);
    expect(wrapper.find('[data-test="close-field"]').exists()).toBe(true);
  });
});
