import store from '@/store';

describe('Sessions', () => {
  const numberSessions = 2;

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
    recorded: true,
  };

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
      active: false,
      authenticated: false,
    },
    {
      uid: '8c354a00f50',
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

  const pagePerpageInitialValue = {
    page: 0,
    perPage: 10,
  };

  const pagePerpageValue = {
    page: 1,
    perPage: 50,
  };

  it('Return device default variables', () => {
    expect(store.getters['sessions/list']).toEqual([]);
    expect(store.getters['sessions/get']).toEqual({});
    expect(store.getters['sessions/getNumberSessions']).toEqual(0);
    expect(store.getters['sessions/getPage']).toEqual(pagePerpageInitialValue.page);
    expect(store.getters['sessions/getPerPage']).toEqual(pagePerpageInitialValue.perPage);
  });
  it('Verify initial state change for setSessions mutation', () => {
    store.commit('sessions/setSessions', { data: sessions, headers: { 'x-total-count': numberSessions } });
    expect(store.getters['sessions/list']).toEqual(sessions);
    expect(store.getters['sessions/getNumberSessions']).toEqual(numberSessions);
  });
  it('Verify inital state change for setSession mutation', () => {
    store.commit('sessions/setSession', { data: session });
    expect(store.getters['sessions/get']).toEqual(session);
  });
  it('Verify changed session object state for removeRecordedSession mutation', () => {
    store.commit('sessions/removeRecordedSession');
    expect(store.getters['sessions/get']).toEqual({ ...session, recorded: false });
  });
  it('Verify inital state change for setPagePerpage mutation', () => {
    store.commit('sessions/setPagePerpage', pagePerpageValue);
    expect(store.getters['sessions/getPage']).toEqual(pagePerpageValue.page);
    expect(store.getters['sessions/getPerPage']).toEqual(pagePerpageValue.perPage);
  });
  it('Verify inital state change for resetPagePerpage mutation', () => {
    store.commit('sessions/resetPagePerpage');
    expect(store.getters['sessions/getPage']).toEqual(pagePerpageInitialValue.page);
    expect(store.getters['sessions/getPerPage']).toEqual(pagePerpageInitialValue.perPage);
  });
  it('Verify changed session object state for clearObjectSession mutation', () => {
    store.commit('sessions/clearObjectSession');
    expect(store.getters['sessions/get']).toEqual({});
  });
  it('Verify changed session list state for clearListSession mutation', () => {
    store.commit('sessions/clearListSessions');
    expect(store.getters['sessions/list']).toEqual([]);
    expect(store.getters['sessions/getNumberSessions']).toEqual(0);
  });
});
