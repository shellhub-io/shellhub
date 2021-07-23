import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SnackbarSuccess from '@/components/snackbar/SnackbarSuccess';

describe('SnackbarSuccess', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const snackbarSuccess = true;
  let typeMessage = '';
  let mainContent = '';
  let message = '';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      snackbarSuccess,
    },
    getters: {
      'snackbar/snackbarSuccess': (state) => state.snackbarSuccess,
    },
    actions: {
      'snackbar/unsetShowStatusSnackbarSuccess': () => {},
    },
  });

  ///////
  // In this case, the main objective is to change the message.
  // For this test to work, the message type is changed to action.
  ///////

  describe('Action message type', () => {
    typeMessage = 'action';
    mainContent = 'renaming device';
    message = `The ${mainContent} has succeeded.`;

    beforeEach(() => {
      wrapper = shallowMount(SnackbarSuccess, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { typeMessage, mainContent },
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

    it('Process data in the computed - action message type', async () => {
      expect(wrapper.vm.snackbar).toEqual(snackbarSuccess);
      expect(wrapper.vm.message).toEqual(message);
    });
  });

  ///////
  // In this case, the main objective is to change the message.
  // For this test to work, the message type is changed to default.
  ///////

  describe('Default message type', () => {
    typeMessage = 'default';
    message = 'The request has succeeded.';

    beforeEach(() => {
      wrapper = shallowMount(SnackbarSuccess, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { typeMessage, mainContent },
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

    it('Process data in the computed - action message type', async () => {
      expect(wrapper.vm.snackbar).toEqual(snackbarSuccess);
      expect(wrapper.vm.message).toEqual(message);
    });
  });
});
