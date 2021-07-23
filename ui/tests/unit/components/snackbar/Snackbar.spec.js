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

  it('Process data in the computed', () => {
    expect(wrapper.vm.message).toEqual(snackbarMessageAndContentType);
  });

  //////
  // HTML validation
  //////

  it('Renders the template with components', () => {
    expect(wrapper.find('[data-test="snackbarSuccess-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="snackbarError-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="snackbarCopy-component"]').exists()).toBe(true);
  });
});
