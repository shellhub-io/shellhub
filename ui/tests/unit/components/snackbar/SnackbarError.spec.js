import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SnackbarError from '@/components/snackbar/SnackbarError';

describe('SnackbarError', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const snackbarError = true;
  let typeMessage = 'loading';
  let mainContent = 'dashboard';
  const loadingMessage = `Loading the ${mainContent} has failed, please try again.`;
  let actionMessage = `The ${mainContent} request has failed, please try again.`;
  const defaultMessage = 'The request has failed, please try again.';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      snackbarError,
    },
    getters: {
      'snackbar/snackbarError': (state) => state.snackbarError,
    },
    actions: {
      'snackbar/unsetShowStatusSnackbarError': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SnackbarError, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { typeMessage, mainContent },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in the computed', async () => {
    expect(wrapper.vm.snackbar).toEqual(snackbarError);
    expect(wrapper.vm.message).toEqual(loadingMessage);

    typeMessage = 'action';
    mainContent = 'deviceDelete';
    actionMessage = `The ${mainContent} request has failed, please try again.`;
    wrapper = shallowMount(SnackbarError, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { typeMessage, mainContent },
    });
    expect(wrapper.vm.message).toEqual(actionMessage);

    typeMessage = 'default';
    wrapper = shallowMount(SnackbarError, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { typeMessage, mainContent },
    });
    expect(wrapper.vm.message).toEqual(defaultMessage);
  });
});
