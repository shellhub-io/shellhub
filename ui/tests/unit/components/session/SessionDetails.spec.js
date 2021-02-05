import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SessionDetails from '@/components/session/SessionDetails';

describe('SessionDetails', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const owner = true;

  const session = {
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
    active: false,
    authenticated: false,
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      session,
      owner,
    },
    getters: {
      'sessions/get': (state) => state.session,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'sessions/get': () => {
      },
      'sessions/close': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SessionDetails, {
      store,
      localVue,
      stubs: ['fragment'],
      mocks: {
        $route: {
          params: {
            id: '8c354a00f50',
          },
        },
      },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  Object.keys(session).forEach((field) => {
    it(`Receives the field ${field} of session state from store`, () => {
      expect(wrapper.vm.session[field]).toEqual(session[field]);
    });
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.uid).toEqual(session.uid);
    expect(wrapper.vm.session).toEqual(session);
    expect(wrapper.vm.dialog).toEqual(false);
    expect(wrapper.vm.hide).toEqual(true);
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="sessionUid-field"]').text()).toEqual(session.uid);
    expect(wrapper.find('[data-test="sessionUser-field"]').text()).toEqual(session.username);
    expect(wrapper.find('[data-test="sessionIpAddress-field"]').text()).toEqual(session.ip_address);
    expect(wrapper.find('[data-test="sessionStartedAt-field"]').text()).toEqual('Monday, May 18th 2020, 12:30:28 pm');
    expect(wrapper.find('[data-test="sessionLastSeen-field"]').text()).toEqual('Monday, May 18th 2020, 12:30:30 pm');
  });
});
