import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PublicKeys from "../../src/views/PublicKeys.vue";
import { createStore } from "vuex";
import { key } from "../../src/store";
import routes from "../../src/router";

describe("Publickeys", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const numberPublickeysEqualZero = 0;
  const numberPublickeysGreaterThanZero = 1;

  const actionsMock = {
    'publicKeys/refresh': vi.fn(),
    'box/setStatus': vi.fn(),
    'publicKeys/resetPagePerpage': vi.fn(),
    'snackbar/showSnackbarErrorLoading': vi.fn(),
    'tags/fetch': vi.fn(),
    'publicKeys/fetch': vi.fn(),
  };

  const storeWithoutPublickeys = createStore({
    state: {
      numberPublickeys: numberPublickeysEqualZero,
    },
    getters: {
      'publicKeys/getNumberPublicKeys': (state) => state.numberPublickeys,
    },
    actions: actionsMock,
  });

  const storeWithPublickeys = createStore({
    state: {
      numberPublickeys: numberPublickeysGreaterThanZero,
    },
    getters: {
      'publicKeys/getNumberPublicKeys': (state) => state.numberPublickeys,
    },
    actions: actionsMock,
  });

  ///////
  // In this case, the rendering of the component that shows the
  // message when it does not have public key is tested.
  ///////

  describe('Without public key', () => {
    beforeEach(async () => {
      wrapper = mount(PublicKeys, {
        global: {
          plugins: [[storeWithoutPublickeys, key], vuetify, routes],
        },
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

    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });
    it('Compare data with the default and defined value', () => {
      expect(wrapper.vm.show).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.hasPublicKey).toEqual(false);
      expect(wrapper.vm.showBoxMessage).toEqual(true);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="public-key-add-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="BoxMessagePublicKey-component"]').exists()).toBe(true);
    });
  });

   ///////
  // In this case, it is tested when there is already a registered
  // public key.
  ///////

  describe('With public key', () => {
    beforeEach(async () => {
      wrapper = mount(PublicKeys, {
        global: {
          plugins: [[storeWithPublickeys, key], vuetify, routes],
        },
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

    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });
    it('Compare data with the default and defined value', () => {
      expect(wrapper.vm.show).toEqual(true);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.hasPublicKey).toEqual(true);
      expect(wrapper.vm.showBoxMessage).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="public-key-add-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="BoxMessagePublicKey-component"]').exists()).toBe(false);
    });
  });
});
