import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SessionList from '@/components/session/SessionList.vue';

describe('DeviceAdd', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      sessions: [
        {
          'uid': '8c354a00f50',
          'device_uid': 'a582b47a42d',
          'device': {
            'uid': 'a582b47a42d',
            'name': '39-5e-2a',
            'identity': {
              'mac': '00:00:00:00:00:00'
            },
            'info': {
              'id': 'debian',
              'pretty_name': 'Debian GNU/Linux 10 (buster)',
              'version': 'v0.2.5'
            },
            'public_key': '----- PUBLIC KEY -----',
            'tenant_id': '00000000',
            'last_seen': '2020-05-18T13:27:02.498Z',
            'online': false,
            'namespace': 'user'
          },
          'tenant_id': '00000000',
          'username': 'user',
          'ip_address': '000.000.000.000',
          'started_at': '2020-05-18T12:30:28.824Z',
          'last_seen': '2020-05-18T12:30:30.205Z',
          'active': false,
          'authenticated': false
        },
        {
          'uid': '8c354a00f50',
          'device_uid': 'a582b47a42d',
          'device': {
            'uid': 'a582b47a42d',
            'name': 'b4-2e-99',
            'identity': {
              'mac': '00:00:00:00:00:00'
            },
            'info': {
              'id': 'debian',
              'pretty_name': 'Debian GNU/Linux 10 (buster)',
              'version': 'v0.2.5'
            },
            'public_key': '----- PUBLIC KEY -----',
            'tenant_id': '00000000',
            'last_seen': '2020-05-18T13:27:02.498Z',
            'online': false,
            'namespace': 'user'
          },
          'tenant_id': '00000000',
          'username': 'user',
          'ip_address': '000.000.000.000',
          'started_at': '2020-05-18T12:30:28.824Z',
          'last_seen': '2020-05-18T12:30:30.205Z',
          'active': false,
          'authenticated': false
        }
      ],
      numberSessions: 2,
    },
    getters: {
      'sessions/list': (state) => state.sessions,
      'sessions/getNumberSessions': (state) => state.numberSessions,
    },
    actions: {
      'sessions/fetch': () => {
      },
      'sessions/close': () => {
      },
    }
  });

  beforeEach(() => {

    wrapper = shallowMount(SessionList, {
      store,
      localVue,
      stubs: ['fragment']
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper.isVueInstance()).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
