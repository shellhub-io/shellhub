import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SessionClose from '@/components/session/SessionClose';

describe('SessionClose', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const uid = '8c354a00f51';
  const device = 'a582b47a42d';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'sessions/close': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SessionClose, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { uid, device },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.uid).toEqual(uid);
    expect(wrapper.vm.device).toEqual(device);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
    expect(wrapper.vm.session).toEqual({ uid, device_uid: device });
  });
});
