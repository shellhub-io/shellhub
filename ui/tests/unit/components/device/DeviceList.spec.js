import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import DeviceList from '@/components/device/DeviceList';
import { actions, authorizer } from '../../../../src/authorizer';

describe('DeviceList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const numberDevices = 2;

  const pagination = {
    groupBy: [],
    groupDesc: [],
    itemsPerPage: 10,
    multiSort: false,
    mustSort: false,
    page: 1,
    sortBy: [],
    sortDesc: [],
  };

  const headers = [
    {
      text: 'Online',
      value: 'online',
      align: 'center',
    },
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
      text: 'Tags',
      value: 'tags',
      align: 'center',
      sortable: false,
    },
    {
      text: 'SSHID',
      value: 'namespace',
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
      uid: 'a582b47a42d',
      name: '39-5e-2a',
      identity: {
        mac: '00:00:00:00:00:00',
      },
      info: {
        id: 'linuxmint',
        pretty_name: 'Linux Mint 19.3',
        version: '',
      },
      public_key: '----- PUBLIC KEY -----',
      tenant_id: '00000000',
      last_seen: '2020-05-20T18:58:53.276Z',
      online: false,
      namespace: 'user',
      status: 'accepted',
      tags: ['device1', 'device2'],
    },
    {
      uid: 'a582b47a42e',
      name: '39-5e-2b',
      identity: {
        mac: '00:00:00:00:00:00',
      },
      info: {
        id: 'linuxmint',
        pretty_name: 'Linux Mint 19.3',
        version: '',
      },
      public_key: '----- PUBLIC KEY -----',
      tenant_id: '00000001',
      last_seen: '2020-05-20T19:58:53.276Z',
      online: true,
      namespace: 'user',
      status: 'accepted',
      tags: ['device1', 'device2'],
    },
  ];

  const devicesOffline = JSON.parse(JSON.stringify(devices));
  devicesOffline[1].online = false;

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
      'modals/showAddDevice': () => {
      },
      'devices/fetch': () => {
      },
      'devices/rename': () => {
      },
      'devices/resetListDevices': () => {
      },
      'stats/get': () => {
      },
    },
  });

  const storeDevicesOffline = new Vuex.Store({
    namespaced: true,
    state: {
      devices: devicesOffline,
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
      'devices/resetListDevices': () => {},
      'stats/get': () => {},
    },
  });

  ///////
  // In this case, it is tested when device is online.
  ///////

  describe('Device online', () => {
    beforeEach(() => {
      wrapper = mount(DeviceList, {
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
      expect(wrapper.vm.hostname).toEqual('localhost');
      expect(wrapper.vm.pagination).toEqual(pagination);
      expect(wrapper.vm.headers).toEqual(headers);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.getListDevices).toEqual(devices);
      expect(wrapper.vm.getNumberDevices).toEqual(numberDevices);
    });
    it('Process data in methods', () => {
      Object.keys(devices).forEach((device) => {
        const address = `${device.namespace}.${device.name}@localhost`;
        expect(wrapper.vm.address(device)).toEqual(address);
      });
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

  ///////
  // In this case, it is tested when device is offline.
  ///////

  describe('Device online', () => {
    beforeEach(() => {
      wrapper = mount(DeviceList, {
        store: storeDevicesOffline,
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
      expect(wrapper.vm.hostname).toEqual('localhost');
      expect(wrapper.vm.pagination).toEqual(pagination);
      expect(wrapper.vm.headers).toEqual(headers);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.getListDevices).toEqual(devicesOffline);
      expect(wrapper.vm.getNumberDevices).toEqual(numberDevices);
    });
    it('Process data in methods', () => {
      Object.keys(devices).forEach((device) => {
        const address = `${device.namespace}.${device.name}@localhost`;
        expect(wrapper.vm.address(device)).toEqual(address);
      });
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
});
