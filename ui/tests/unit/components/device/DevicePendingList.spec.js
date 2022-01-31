import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import DevicePendingList from '@/components/device/DevicePendingList';
import { actions, authorizer } from '../../../../src/authorizer';

describe('DevicePendingList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.filter('moment', () => {});

  let wrapper;

  const numberDevices = 4;

  const pagination = {
    groupBy: [],
    groupDesc: [],
    itemsPerPage: 10,
    multiSort: false,
    mustSort: false,
    page: 1,
    sortBy: [],
    sortDesc: [false],
  };

  const headers = [
    {
      text: 'Hostname',
      value: 'hostname',
      align: 'center',
    },
    {
      text: 'Operating System',
      value: 'info.pretty_name',
      align: 'center',
      sortable: false,
    },
    {
      text: 'Request Time',
      value: 'request_time',
      align: 'center',
      sortable: false,
    },
    {
      text: 'Actions',
      value: 'actions',
      align: 'center',
      sortable: false,
    },
  ];

  const devices = [
    {
      uid: '1234',
      name: 'hi-23-23-54',
      identity: {
        mac: '00:00:00',
      },
      info: {
        id: 'linuxmint',
        pretty_name: 'Linux Mint 20.0',
        version: '',
      },
      public_key: 'xxxxxxxx',
      tenant_id: '0000000',
      last_seen: '2020-09-21T18:58:53.276Z',
      online: true,
      namespace: 'user',
      status: 'pending',
    },
    {
      uid: '1235',
      name: 'hi-23-23-55',
      identity: {
        mac: '00:00:00',
      },
      info: {
        id: 'linuxmint',
        pretty_name: 'Linux Mint 20.0',
        version: '',
      },
      public_key: 'xxxxxxxx',
      tenant_id: '0000000',
      last_seen: '2020-09-21T18:59:53.276Z',
      online: false,
      namespace: 'user',
      status: 'pending',
    },
    {
      uid: '1236',
      name: 'hi-23-23-56',
      identity: {
        mac: '00:00:00',
      },
      info: {
        id: 'linuxmint',
        pretty_name: 'Linux Mint 20.0',
        version: '',
      },
      public_key: 'xxxxxxxx',
      tenant_id: '0000000',
      last_seen: '2020-09-21T19:58:53.276Z',
      online: false,
      namespace: 'user',
      status: 'pending',
    },
    {
      uid: '1237',
      name: 'hi-23-23-57',
      identity: {
        mac: '00:00:00',
      },
      info: {
        id: 'linuxmint',
        pretty_name: 'Linux',
        version: '',
      },
      public_key: 'xxxxxxxx',
      tenant_id: '0000000',
      last_seen: '2020-09-21T120:58:53.276Z',
      online: true,
      namespace: 'user',
      status: 'pending',
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      devices,
      numberDevices,
    },
    getters: {
      'devices/list': (state) => state.devices,
      'devices/getNumberDevices': (state) => state.numberDevices,
    },
    actions: {
      'modals/showAddDevice': () => {},
      'devices/fetch': () => {},
      'devices/rename': () => {},
      'stats/get': () => {},
      'devices/resetListDevices': () => {},
    },
  });

  beforeEach(() => {
    wrapper = mount(DevicePendingList, {
      store,
      localVue,
      stubs: ['fragment', 'router-link'],
      vuetify,
      mocks: {
        $authorizer: authorizer,
        $actions: actions,
      },
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
    expect(wrapper.vm.pagination).toEqual(pagination);
    expect(wrapper.vm.headers).toEqual(headers);
    expect(wrapper.vm.deviceAcceptButtonShow).toEqual([false, false, false, false]);
    expect(wrapper.vm.deviceRejectButtonShow).toEqual([false, false, false, false]);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getListPendingDevices).toEqual(devices);
    expect(wrapper.vm.getNumberPendingDevices).toEqual(numberDevices);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with components', async () => {
    expect(wrapper.find('[data-test="deviceIcon-component"]').exists()).toEqual(true);
  });
  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;

    expect(dataTableProps.items).toHaveLength(numberDevices);
  });
});
