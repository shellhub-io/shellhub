import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingSecurity from '@/components/setting/SettingSecurity';

describe('SettingSecurity', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const hasTenant = true;
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
      'security/set': () => {},
      'security/get': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SettingSecurity, {
      store,
      localVue,
      propsData: { hasTenant },
      stubs: ['fragment'],
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

  it('Receive data in props', () => {
    expect(wrapper.vm.hasTenant).toEqual(hasTenant);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.sessionRecord).toEqual(sessionRecord);
  });
});
