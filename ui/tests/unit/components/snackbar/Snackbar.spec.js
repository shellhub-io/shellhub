import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Snackbar from '@/components/snackbar/Snackbar';

describe('Snackbar', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const snackbarMessageAndContentType = {
    typeMessage: '',
    typeContent: '',
  };

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      snackbarMessageAndContentType,
    },
    getters: {
      'snackbar/snackbarMessageAndContentType': (state) => state.snackbarMessageAndContentType,
    },
    actions: {
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(Snackbar, {
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
    expect(wrapper.vm.message).toEqual(snackbarMessageAndContentType);
  });
});
