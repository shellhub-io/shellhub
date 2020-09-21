import Vuex from 'vuex';
import VueRouter from 'vue-router';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Device from '@/components/device/Device';

describe('Device', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  localVue.use(VueRouter);

  let wrapper;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      stats: {
        registered_devices: 0,
        online_devices: 0,
        active_sessions: 0,
        pending_devices: 0,
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
  it('Compare data with the default and defined value', () => {
    expect(wrapper.vm.search).toEqual('');

    wrapper.setData({ search: 'ShellHub' });
    expect(wrapper.vm.search).toEqual('ShellHub');
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="search-text"]').exists()).toBe(true);

    const textInputSearch = wrapper.find('[data-test="search-text"]');
    textInputSearch.element.value = 'ShellHub';
    expect(wrapper.find('[data-test="search-text"]').element.value).toEqual('ShellHub');
  });
});
