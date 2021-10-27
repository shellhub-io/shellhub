import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import SnackbarError from '@/components/snackbar/SnackbarError';

describe('SnackbarError', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const snackbarError = true;
  let typeMessage = '';
  let mainContent = '';
  let message = '';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      snackbarError,
    },
    getters: {
      'snackbar/snackbarError': (state) => state.snackbarError,
    },
    actions: {
      'snackbar/unsetShowStatusSnackbarError': () => {},
    },
  });

  ///////
  // In this case, the main objective is to change the message.
  // For this test to work, the message type is changed to loading.
  ///////

  describe('Loading message type', () => {
    typeMessage = 'loading';
    mainContent = 'dashboard';
    message = `Loading the ${mainContent} has failed, please try again.`;

    beforeEach(() => {
      wrapper = shallowMount(SnackbarError, {
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.snackbar).toEqual(snackbarError);
      expect(wrapper.vm.message).toEqual(message);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.message).toEqual(message);
    });
  });

  ///////
  // In this case, the main objective is to change the message.
  // For this test to work, the message type is changed to association.
  ///////

  describe('Association message type', () => {
    typeMessage = 'association';
    message = 'There is no namespace associated with your account.';

    beforeEach(() => {
      wrapper = shallowMount(SnackbarError, {
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.snackbar).toEqual(snackbarError);
      expect(wrapper.vm.message).toEqual(message);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.message).toEqual(message);
    });
  });

  ///////
  // In this case, the main objective is to change the message.
  // For this test to work, the message type is changed to not request.
  ///////

  describe('NotRequest message type', () => {
    typeMessage = 'notRequest';
    mainContent = 'deviceDelete';
    message = `The ${mainContent} has failed, please try again.`;

    beforeEach(() => {
      wrapper = shallowMount(SnackbarError, {
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.snackbar).toEqual(snackbarError);
      expect(wrapper.vm.message).toEqual(message);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.message).toEqual(message);
    });
  });

  ///////
  // In this case, the main objective is to change the message.
  // For this test to work, the message type is changed to incorrect.
  ///////

  describe('Incorrect message type', () => {
    typeMessage = 'incorrect';
    mainContent = 'deviceDelete';
    message = `Incorrect ${mainContent} information, please try again.`;

    beforeEach(() => {
      wrapper = shallowMount(SnackbarError, {
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.snackbar).toEqual(snackbarError);
      expect(wrapper.vm.message).toEqual(message);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.message).toEqual(message);
    });
  });

  ///////
  // In this case, the main objective is to change the message.
  // For this test to work, the message type is changed to choose
  // devices.
  ///////

  describe('Incorrect message type', () => {
    typeMessage = 'deviceChooser';
    mainContent = '';
    message = 'You need to select 3 devices.';

    beforeEach(() => {
      wrapper = shallowMount(SnackbarError, {
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.snackbar).toEqual(snackbarError);
      expect(wrapper.vm.message).toEqual(message);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.message).toEqual(message);
    });
  });

  ///////
  // In this case, the main objective is to change the message.
  // For this test to work, the message type is changed to default.
  ///////

  describe('Default message type', () => {
    typeMessage = 'default';
    message = 'The request has failed, please try again.';

    beforeEach(() => {
      wrapper = shallowMount(SnackbarError, {
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

    it('Process data in the computed', () => {
      expect(wrapper.vm.snackbar).toEqual(snackbarError);
      expect(wrapper.vm.message).toEqual(message);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.message).toEqual(message);
    });
  });
});
