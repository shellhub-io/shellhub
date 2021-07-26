import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Session from '@/components/session/Session';
import Vuetify from 'vuetify';

describe('Session', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  const numberSessionsEqualZero = 0;
  const numberSessionsGreaterThanZero = 1;

  const storeWithoutSessions = new Vuex.Store({
    namespaced: true,
    state: {
      numberSessions: numberSessionsEqualZero,
    },
    getters: {
      'sessions/getNumberSessions': (state) => state.numberSessions,
    },
    actions: {
      'sessions/refresh': () => {},
      'boxs/setStatus': () => {},
      'sessions/resetPagePerpage': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  const storeWithSessions = new Vuex.Store({
    namespaced: true,
    state: {
      numberSessions: numberSessionsGreaterThanZero,
    },
    getters: {
      'sessions/getNumberSessions': (state) => state.numberSessions,
    },
    actions: {
      'sessions/refresh': () => {},
      'boxs/setStatus': () => {},
      'sessions/resetPagePerpage': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  ///////
  // In this case, the rendering of the component that shows the
  // message when it does not have access to the device is tested.
  ///////

  describe('Without sessions', () => {
    beforeEach(() => {
      wrapper = mount(Session, {
        store: storeWithoutSessions,
        localVue,
        stubs: ['fragment'],
        vuetify,
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

    it('Compare data with the default and defined value', () => {
      expect(wrapper.vm.show).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.hasSession).toEqual(false);
      expect(wrapper.vm.showBoxMessage).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="BoxMessageSession-component"]').exists()).toBe(true);
    });
  });

  ///////
  // In this case, it is tested when it has already accessed a
  // device.
  ///////

  describe('With sessions', () => {
    beforeEach(() => {
      wrapper = mount(Session, {
        store: storeWithSessions,
        localVue,
        stubs: ['fragment'],
        vuetify,
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

    it('Compare data with the default and defined value', () => {
      expect(wrapper.vm.show).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.hasSession).toEqual(true);
      expect(wrapper.vm.showBoxMessage).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="BoxMessageSession-component"]').exists()).toBe(false);
    });
  });
});
