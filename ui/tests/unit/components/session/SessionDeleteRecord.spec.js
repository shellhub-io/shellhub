import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SessionDeleteRecord from '@/components/session/SessionDeleteRecord';

describe('SessionDeleteRecord', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const uid = '8c354a00f50';

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
    active: true,
    authenticated: false,
    recorded: true,
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      session,
    },
    actions: {
      'sessions/deleteSessionLogs': () => {},
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SessionDeleteRecord, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.uid).toEqual(uid);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
  });
});
