import Vuex from 'vuex';
import { mount, createLocalVue, config } from '@vue/test-utils';
import Vuetify from 'vuetify';
import Router from 'vue-router';
import publicKey from '@/components/public_key/PublicKey';

config.mocks = {
  $env: {
    isCloud: true,
  },
};

describe('PublicKey', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  const router = new Router();
  localVue.use(Vuex);
  localVue.use(Router);

  let wrapper;

  const numberPublickeysEqualZero = 0;
  const numberPublickeysGreaterThanZero = 1;

  const storeWithoutPublickeys = new Vuex.Store({
    namespaced: true,
    state: {
      numberPublickeys: numberPublickeysEqualZero,
    },
    getters: {
      'publickeys/getNumberPublicKeys': (state) => state.numberPublickeys,
    },
    actions: {
      'publickeys/refresh': () => {},
      'boxs/setStatus': () => {},
      'publickeys/resetPagePerpage': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  const storeWithPublickeys = new Vuex.Store({
    namespaced: true,
    state: {
      numberPublickeys: numberPublickeysGreaterThanZero,
    },
    getters: {
      'publickeys/getNumberPublicKeys': (state) => state.numberPublickeys,
    },
    actions: {
      'publickeys/refresh': () => {},
      'boxs/setStatus': () => {},
      'publickeys/resetPagePerpage': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  ///////
  // In this case, the rendering of the component that shows the
  // message when it does not have public key is tested.
  ///////

  describe('Without public key', () => {
    beforeEach(() => {
      wrapper = mount(publicKey, {
        store: storeWithoutPublickeys,
        localVue,
        stubs: ['fragment'],
        vuetify,
        router,
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
      expect(wrapper.vm.hasPublickey).toEqual(false);
      expect(wrapper.vm.showBoxMessage).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="publicKeyCreate-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="boxMessagePublicKey-component"]').exists()).toBe(true);
    });
  });

  ///////
  // In this case, it is tested when there is already a registered
  // public key.
  ///////

  describe('With public key', () => {
    beforeEach(() => {
      wrapper = mount(publicKey, {
        store: storeWithPublickeys,
        localVue,
        stubs: ['fragment'],
        vuetify,
        router,
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
      expect(wrapper.vm.hasPublickey).toEqual(true);
      expect(wrapper.vm.showBoxMessage).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="publicKeyCreate-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="boxMessagePublicKey-component"]').exists()).toBe(false);
    });
  });
});
