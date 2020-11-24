import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import PublicKeyDelete from '@/components/public_key/PublicKeyDelete';

describe('PublicKeyDelete', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const fingerprint = 'b7:25:f8';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'publickeys/remove': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(PublicKeyDelete, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { fingerprint },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.fingerprint).toEqual(fingerprint);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
  });
});
