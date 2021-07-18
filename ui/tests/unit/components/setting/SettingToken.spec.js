import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SettingToken from '@/components/setting/SettingToken';

describe('SettingToken', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'tokens/fetch': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SettingToken, {
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
});
