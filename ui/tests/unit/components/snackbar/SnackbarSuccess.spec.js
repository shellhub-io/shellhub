import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SnackbarSuccess from '@/components/snackbar/SnackbarSuccess';

describe('SnackbarSuccess', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const snackbarSuccess = true;
  let typeMessage = 'action';
  const mainContent = 'renaming device';
  const actionMessage = `The ${mainContent} has succeeded.`;
  const defaultMessage = 'The request has succeeded.';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      snackbarSuccess,
    },
    getters: {
      'snackbar/snackbarSuccess': (state) => state.snackbarSuccess,
    },
    actions: {
      'snackbar/unsetShowStatusSnackbarSuccess': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(SnackbarSuccess, {
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
    expect(wrapper.vm.snackbar).toEqual(snackbarSuccess);
    expect(wrapper.vm.message).toEqual(actionMessage);

    typeMessage = 'default';
    wrapper = shallowMount(SnackbarSuccess, {
      store,
      localVue,
      stubs: ['fragment'],
      propsData: { typeMessage, mainContent },
    });
    expect(wrapper.vm.message).toEqual(defaultMessage);
  });
});
