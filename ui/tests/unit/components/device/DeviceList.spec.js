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

  const numberDevicesGlobal = 2;

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
      text: 'SSHID',
      value: 'namespace',
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
      text: 'Actions',
      value: 'actions',
      align: 'center',
      sortable: false,
    },
  ];

  const devicesGlobal = [
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

  const tests = [
    {
      description: 'List data when user has owner role',
      role: {
        type: 'owner',
        permission: true,
      },
      variables: {
        devices: devicesGlobal,
        numberDevices: numberDevicesGlobal,
      },
      data: {
        hostname: 'localhost',
        pagination,
        tags: [],
        tagDialogShow: [],
        deviceDeleteShow: [],
        selectedTags: [],
        updateAction: 'deviceUpdate',
        headers,
      },
      computed: {
        getListDevices: devicesGlobal,
        getNumberDevices: numberDevicesGlobal,
        hasAuthorizationFormUpdate: true,
      },
    },
    {
      description: 'List data when user has observer role',
      role: {
        type: 'observer',
        permission: false,
      },
      variables: {
        devices: devicesGlobal,
        numberDevices: numberDevicesGlobal,
      },
      data: {
        hostname: 'localhost',
        pagination,
        tags: [],
        tagDialogShow: [],
        deviceDeleteShow: [],
        selectedTags: [],
        updateAction: 'deviceUpdate',
        headers,
      },
      computed: {
        getListDevices: devicesGlobal,
        getNumberDevices: numberDevicesGlobal,
        hasAuthorizationFormUpdate: false,
      },
    },
  ];

  const storeVuex = (devices, numberDevices, currentRole) => new Vuex.Store({
    namespaced: true,
    state: {
      devices,
      numberDevices,
      currentRole,
    },
    getters: {
      'devices/list': (state) => state.devices,
      'devices/getNumberDevices': (state) => state.numberDevices,
      'auth/role': (state) => state.currentRole,
    },
    actions: {
      'modals/showAddDevice': () => {},
      'devices/fetch': () => {},
      'devices/rename': () => {},
      'tags/clearSelectedTags': () => {},
      'devices/resetListDevices': () => {},
      'tags/setSelected': () => {},
      'devices/setFilter': () => {},
      'stats/get': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(DeviceList, {
          store: storeVuex(
            test.variables.devices,
            test.variables.numberDevices,
            test.role.type,
          ),
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
      it('Process data in methods', () => {
        Object.keys(test.variables.devices).forEach((device) => {
          const address = `${device.namespace}.${device.name}@localhost`;
          expect(wrapper.vm.address(device)).toEqual(address);
        });
      });

      //////
      // HTML validation
      //////

      it('Renders the template with data', () => {
        const dt = wrapper.find('[data-test="dataTable-field"]');
        const dataTableProps = dt.vm.$options.propsData;

        expect(dataTableProps.items).toHaveLength(test.variables.numberDevices);
      });
    });
  });
});
