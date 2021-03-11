import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import TokenAdd from '@/components/setting/token/TokenAdd';

describe('TokenAdd', () => {
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
      'tokens/post': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(TokenAdd, {
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
