import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import DeviceActionButton from '@/components/device/DeviceActionButton';

describe('DeviceActionButton', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const uid = '';
  const notificationStatus = false;
  const action = 'accept';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'devices/refresh': () => {
      },
      'devices/accept': () => {
      },
      'devices/reject': () => {
      },
      'devices/remove': () => {
      },
      'notifications/fetch': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(DeviceActionButton, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, notificationStatus, action },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
