import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceRejectedList from '@/components/device/DeviceRejectedList';

describe('DeviceRejectedList', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const numberDevices = 1;
  const devices = [
    {
      uid: '2378hj238',
      name: '37-23-hf-1c',
      identity: {
        mac: '00:00:00:00:00:00',
      },
      info: {
        id: 'linuxmint',
        pretty_name: 'Linux Mint 20.0',
        version: '',
      },
      public_key: '---pub_key---',
      tenant_id: '8490393000',
      last_seen: '2020-05-22T18:58:53.276Z',
      online: true,
      namespace: 'user',
      status: 'rejected',
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
    wrapper = shallowMount(DeviceRejectedList, {
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
    expect(wrapper.vm.getListRejectedDevices).toEqual(devices);
    expect(wrapper.vm.getNumberRejectedDevices).toEqual(numberDevices);
  });
});
