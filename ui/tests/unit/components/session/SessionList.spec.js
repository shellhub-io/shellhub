import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import SessionList from '@/components/session/SessionList';
import Vuetify from 'vuetify';

describe('DeviceAdd', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();

  localVue.use(Vuex);

  let wrapper;
  let wrapper2;

  const owner = true;
  const status = true;
  const numberSessions = 2;
  const sessions = [
    {
      uid: '8c354a00f50',
      device_uid: 'a582b47a42d',
      device: {
        uid: 'a582b47a42d',
        name: '39-5e-2a',
        identity: {
          mac: '00:00:00:00:00:00',
        },
        info: {
          id: 'debian',
          pretty_name: 'Debian GNU/Linux 10 (buster)',
          version: 'v0.2.5',
        },
        public_key: '----- PUBLIC KEY -----',
        tenant_id: '00000000',
        last_seen: '2020-05-18T13:27:02.498Z',
        online: false,
        namespace: 'user',
      },
      tenant_id: '00000000',
      username: 'user',
      ip_address: '000.000.000.000',
      started_at: '2020-05-18T12:30:28.824Z',
      last_seen: '2020-05-18T12:30:30.205Z',
      active: true,
      authenticated: false,
    },
    {
      uid: '8c354a00f51',
      device_uid: 'a582b47a42d',
      device: {
        uid: 'a582b47a42d',
        name: 'b4-2e-99',
        identity: {
          mac: '00:00:00:00:00:00',
        },
        info: {
          id: 'debian',
          pretty_name: 'Debian GNU/Linux 10 (buster)',
          version: 'v0.2.5',
        },
        public_key: '----- PUBLIC KEY -----',
        tenant_id: '00000000',
        last_seen: '2020-05-18T13:27:02.498Z',
        online: false,
        namespace: 'user',
      },
      tenant_id: '00000000',
      username: 'user',
      ip_address: '000.000.000.000',
      started_at: '2020-05-18T12:30:28.824Z',
      last_seen: '2020-05-18T12:30:30.205Z',
      active: false,
      authenticated: false,
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      sessions,
      numberSessions,
      owner,
      status,
    },
    getters: {
      'sessions/list': (state) => state.sessions,
      'sessions/getNumberSessions': (state) => state.numberSessions,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'sessions/fetch': () => {
      },
      'sessions/close': () => {
      },
    },
  });

  const store2 = new Vuex.Store({
    namespaced: true,
    state: {
      sessions,
      numberSessions,
      owner: false,
      status,
    },
    getters: {
      'sessions/list': (state) => state.sessions,
      'sessions/getNumberSessions': (state) => state.numberSessions,
      'namespaces/owner': (state) => state.owner,
      'boxs/getStatus': (state) => state.status,
    },
    actions: {
      'sessions/fetch': () => {
      },
      'sessions/close': () => {
      },
      'boxs/setStatus': () => {
      },
    },
  });

  beforeEach(() => {
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

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getListSessions).toEqual(sessions);
    expect(wrapper.vm.getNumberSessions).toEqual(numberSessions);
  });
  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;
    expect(dataTableProps.items).toHaveLength(numberSessions);
    expect(wrapper.find('[data-test="close-field"]').exists()).toBe(true);
  });
  it('Hides the close field when the user is not the owner', () => {
    wrapper2 = mount(SessionList, {
      store: store2,
      localVue,
      stubs: ['fragment', 'router-link'],
      mocks: {
        $env: (isEnterprise) => isEnterprise,
      },
      vuetify,
    });
    expect(wrapper2.find('[data-test="close-field"]').exists()).toBe(false);
  });
});
