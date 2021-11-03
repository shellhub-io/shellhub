import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import NamespaceAdd from '@/components/app_bar/namespace/NamespaceAdd';
import '@/vee-validate';

describe('NamespaceAdd', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const show = true;
  const firstNamespace = true;

  const invalidNamespaces = [
    '\'', '"', '!', '@', '#', '$', '%', '¨', '&', '*', '(', ')', '-', '_', '=', '+', '´', '`', '[',
    '{', '~', '^', ']', ',', '<', '..', '>', ';', ':', '/', '?',
  ];

  const invalidMinAndMaxCharacters = [
    's', 'sh', 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx', 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'namespaces/switchNamespace': () => {},
      'namespaces/post': () => {},
      'namespaces/fetch': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  ///////
  // In this case, the rendering of the dialog is checked. In which
  // case with the input data it cannot take place.
  ///////

  describe('Doesn\'t render the dialog', () => {
    beforeEach(() => {
      wrapper = mount(NamespaceAdd, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { show: !show, firstNamespace },
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

    it('Receive data in props', () => {
      expect(wrapper.vm.show).toEqual(!show);
      expect(wrapper.vm.firstNamespace).toEqual(firstNamespace);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.namespaceName).toEqual('');
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.showAddNamespace).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="namespaceAdd-card"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="namespace-text"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="close-btn"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="add-btn"]').exists()).toBe(false);
    });
  });

  ///////
  // In this case, the rendering of the dialog is checked. In which
  // case with the input data it cannot take place.
  ///////

  describe('Dialog', () => {
    beforeEach(() => {
      wrapper = mount(NamespaceAdd, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { show, firstNamespace },
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

    it('Receive data in props', () => {
      expect(wrapper.vm.show).toEqual(show);
      expect(wrapper.vm.firstNamespace).toEqual(firstNamespace);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.namespaceName).toEqual('');
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.showAddNamespace).toEqual(true);
    });

    //////
    // HTML validation
    //////

    //////
    // In this case, the empty fields are validated.
    //////

    it('Show validation messages', async () => {
      wrapper.setData({ namespaceName: '' });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerNamespace;

      await validator.validate();
      expect(validator.errors[0]).toBe('This field is required');
    });

    //////
    // In this case, the error for dot are validated.
    //////

    it('Show validation messages', async () => {
      wrapper.setData({ namespaceName: 'ShelHub.' });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerNamespace;

      await validator.validate();
      expect(validator.errors[0]).toBe('The name must not contain dots');
    });

    //////
    // In this case, the RFC1223 are validated.
    //////

    it('Show validation messages', async (done) => {
      invalidNamespaces.forEach(async (inamespace) => {
        wrapper.setData({ namespaceName: inamespace });
        await flushPromises();

        const validator = wrapper.vm.$refs.providerNamespace;

        await validator.validate();
        expect(validator.errors[0]).toBe('You entered an invalid RFC1123 name');

        await flushPromises();
        done();
      });
    });

    //////
    // In this case, the min and max characters are validated.
    //////

    it('Show validation messages', async (done) => {
      invalidMinAndMaxCharacters.forEach(async (character) => {
        wrapper.setData({ namespaceName: character });
        await flushPromises();

        const validator = wrapper.vm.$refs.providerNamespace;

        await validator.validate();
        expect(validator.errors[0]).toBe('Your namespace should be 3-30 characters long');

        await flushPromises();
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="namespaceAdd-card"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="namespace-text"]').element.value).toEqual('');
      expect(wrapper.find('[data-test="close-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="add-btn"]').exists()).toBe(true);
    });
  });
});
