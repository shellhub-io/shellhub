import Vuex from 'vuex';
import VueRouter from 'vue-router';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Device from '@/components/device/Device';

describe('Device', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  localVue.use(VueRouter);

  let wrapper;

  const pendingDevices = 2;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      stats: {
        registered_devices: 0,
        online_devices: 0,
        active_sessions: 0,
        pending_devices: pendingDevices,
        rejected_devices: 0,
      },
    },
    getters: {
      'stats/stats': (state) => state.stats,
    },
    actions: {
      'stats/get': () => {
      },
      'devices/setFilter': () => {
      },
      'devices/refresh': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(Device, {
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
  it('Process data in the computed', () => {
    expect(wrapper.vm.getNumberPendingDevices).toEqual(pendingDevices);
    expect(wrapper.vm.hasDevice).toEqual(true);
    expect(wrapper.vm.showBoxMessage).toEqual(false);
  });
  it('Compare data with the default and defined value', () => {
    expect(wrapper.vm.show).toEqual(true);
    expect(wrapper.vm.search).toEqual('');

    wrapper.setData({ search: 'ShellHub' });

    expect(wrapper.vm.search).toEqual('ShellHub');
  });
  it('Process data in methods', () => {
    const inputs = [
      { field: undefined, isDesc: undefined },
      { field: 'hostname', isDesc: undefined },
      { field: 'hostname', isDesc: true },
    ];

    const output = [
      { field: null, status: false, statusString: 'asc' },
      { field: 'name', status: false, statusString: 'asc' },
      { field: 'name', status: true, statusString: 'desc' },
    ];

    Object.keys(inputs).forEach((index) => {
      expect(wrapper.vm.formatSortObject(inputs[index].field, inputs[index].isDesc))
        .toEqual(output[index]);
    });
  });
  it('Renders the template with components', () => {
    expect(wrapper.find('[data-test="boxMessageDevice-component"]').exists()).toBe(false);
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="search-text"]').exists()).toBe(true);

    const textInputSearch = wrapper.find('[data-test="search-text"]');
    textInputSearch.element.value = 'ShellHub';

    expect(wrapper.find('[data-test="search-text"]').element.value).toEqual('ShellHub');
    expect(wrapper.find('[data-test="badge-field"]').vm.$options.propsData.content).toEqual(pendingDevices);
  });
});
