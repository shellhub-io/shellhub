import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingSecurity from '@/components/setting/SettingSecurity';

describe('SettingSecurity', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const sessionRecord = true;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      sessionRecord,
    },
    getters: {
      'security/get': (state) => state.sessionRecord,
    },
    actions: {
      'security/set': () => {
      },
      'security/get': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SettingSecurity, {
      store,
      localVue,
      propsData: { show: true },
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receives props data', () => {
    expect(wrapper.vm.show).toBe(true);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.sessionRecord).toEqual(sessionRecord);
  });
});
