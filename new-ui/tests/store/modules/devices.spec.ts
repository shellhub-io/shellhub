import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe('Devices', () => {
  const numberDevices = 2;

  const devices = [
    {
      uid: 'a582b47a42d',
      name: '49-5e-2a',
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
    {
      uid: 'a582b47a42f',
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
  ];

  const device = {
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
  };

  // filter
  const searchString = '4';
  const filter = [{ type: 'property', params: { name: 'name', operator: 'like', value: searchString } }];
  const encodedFilter = btoa(JSON.stringify(filter));

  const data = {
    page: 1,
    perPage: 10,
    filter: null,
    status: 'accepted',
    name: 'newDeviceName',
    uid: 'a582b47a42f',
  };

  it('Return device default variables', () => {
    expect(store.getters['devices/list']).toEqual([]);
    expect(store.getters['devices/get']).toEqual({});
    expect(store.getters['devices/getNumberDevices']).toEqual(0);
    expect(store.getters['devices/getPage']).toEqual(1);
    expect(store.getters['devices/getPerPage']).toEqual(10);
    expect(store.getters['devices/getFilter']).toEqual("");
    expect(store.getters['devices/getStatus']).toEqual('');
    expect(store.getters['devices/getFirstPending']).toEqual({});
  });
  it('Verify initial states change for mutation setDevices', () => {
    store.commit('devices/setDevices', { data: devices, headers: { 'x-total-count': numberDevices } });
    expect(store.getters['devices/list']).toEqual(devices);
    expect(store.getters['devices/getNumberDevices']).toEqual(numberDevices);
  });
  it('Verify initial states change for mutation serPagePerpageFilter', () => {
    store.commit('devices/setPagePerpageFilter', data);
    expect(store.getters['devices/getPage']).toEqual(1);
    expect(store.getters['devices/getPerPage']).toEqual(10);
    expect(store.getters['devices/getFilter']).toEqual(null);
    expect(store.getters['devices/getStatus']).toEqual('accepted');
  });
  it('Verify initial state change for mutation setDevice', () => {
    store.commit('devices/setDevice', device);
    expect(store.getters['devices/get']).toEqual(device);
  });
  it('Verify changed filter state in setFilter mutation', () => {
    // converted to base64 filter
    store.commit('devices/setFilter', encodedFilter);
    expect(store.getters['devices/getFilter']).toEqual('W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJsaWtlIiwidmFsdWUiOiI0In19XQ==');
  });
  it('Verify changed name state in renameDevice mutation', () => {
    store.commit('devices/renameDevice', data);
    expect(store.getters['devices/get'].name).toEqual(data.name);
  });
  it('Verify empty devices state for clearListDevices mutation', () => {
    store.commit('devices/clearListDevices');
    expect(store.getters['devices/list']).toEqual([]);
  });
});
