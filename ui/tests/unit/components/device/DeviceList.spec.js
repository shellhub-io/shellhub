import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceList from '@/components/device/DeviceList';

describe('DeviceList', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const numberDevices = 2;
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
    },
  ];
  const owner = true;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      devices,
      numberDevices,
      owner,
    },
    getters: {
      'devices/list': (state) => state.devices,
      'devices/getNumberDevices': (state) => state.numberDevices,
      'namespaces/owner': (state) => state.owner,
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

  beforeEach(() => {
    wrapper = shallowMount(DeviceList, {
      store,
      localVue,
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;
    expect(dataTableProps.items).toHaveLength(numberDevices);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getListDevices).toEqual(devices);
    expect(wrapper.vm.getNumberDevices).toEqual(numberDevices);
  });
});
