import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import DeviceRename from '@/components/device/DeviceRename';
import '@/vee-validate';

describe('DeviceRename', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  let wrapper;

  const isOwner = true;
  const uid = 'a582b47a42d';
  const name = '39-5e-2a';

  const invalidNames = [
    '\'', '"', '!', '@', '#', '$', '%', '¨', '&', '*', '(', ')', '-', '_', '=', '+', '´', '`', '[',
    '{', '~', '^', ']', ',', '<', '..', '>', ';', ':', '/', '?',
  ];

  const invalidMinAndMaxCharacters = [
    'xx', 'xx', 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx', 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
      'devices/get': (state) => state.device,
    },
    actions: {
      'devices/rename': () => {
      },
      'snackbar/showSnackbarSuccessAction': () => {
      },
      'snackbar/showSnackbarErrorAction': () => {
      },
    },
  });

  describe('Icon', () => {
    beforeEach(() => {
      wrapper = mount(DeviceRename, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { name, uid },
        vuetify,
      });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      document.body.setAttribute('data-app', true);
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receive data in props', () => {
      expect(wrapper.vm.name).toEqual(name);
      expect(wrapper.vm.uid).toEqual(uid);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.invalid).toEqual(false);
      expect(wrapper.vm.editName).toEqual('39-5e-2a');
    });

    //////
    // HTML validation
    //////

    it('Show message tooltip to user owner', async (done) => {
      const icons = wrapper.findAll('.v-icon');
      const helpIcon = icons.at(0);
      helpIcon.trigger('mouseenter');
      await wrapper.vm.$nextTick();

      expect(icons.length).toBe(1);
      requestAnimationFrame(() => {
        expect(wrapper.find('[data-test="tooltipOwner-text"]').text()).toEqual('Edit');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="deviceRename-dialog"]').exists()).toEqual(false);
    });
  });

  describe('Dialog', () => {
    beforeEach(() => {
      wrapper = mount(DeviceRename, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { name, uid },
        vuetify,
      });

      wrapper.setData({ dialog: true });
    });

    ///////
    // Component Rendering
    //////

    it('Is a Vue instance', () => {
      document.body.setAttribute('data-app', true);
      expect(wrapper).toBeTruthy();
    });
    it('Renders the component', () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data and Props checking
    //////

    it('Receive data in props', () => {
      expect(wrapper.vm.name).toEqual(name);
      expect(wrapper.vm.uid).toEqual(uid);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.invalid).toEqual(false);
      expect(wrapper.vm.editName).toEqual('39-5e-2a');
    });

    //////
    // HTML validation
    //////

    //////
    // In this case, the empty fields are validated.
    //////

    it('Show validation messages', async () => {
      wrapper.setData({ editName: '' });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerHostname;

      await validator.validate();
      expect(validator.errors[0]).toBe('This field is required');
    });

    //////
    // In this case, must not contain dots are validated.
    //////

    it('Show validation messages', async () => {
      wrapper.setData({ editName: 'ShelHub.' });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerHostname;

      await validator.validate();
      expect(validator.errors[0]).toBe('The name must not contain dots');
    });

    //////
    // In this case, RFC1123 rules are validated.
    //////

    it('Show validation messages', async (done) => {
      invalidNames.forEach(async (invalidName) => {
        wrapper.setData({ editName: invalidName });
        await flushPromises();

        const validator = wrapper.vm.$refs.providerHostname;

        await validator.validate();
        expect(validator.errors[0]).toBe('You entered an invalid RFC1123 name');

        await flushPromises();
        done();
      });
    });

    //////
    // In this case, min and max characters are validated.
    //////

    it('Show validation messages', async (done) => {
      invalidMinAndMaxCharacters.forEach(async (character) => {
        wrapper.setData({ editName: character });
        await flushPromises();

        const validator = wrapper.vm.$refs.providerHostname;

        await validator.validate();
        expect(validator.errors[0]).toBe('Your hostname should be 3-30 characters long');

        await flushPromises();
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="deviceRename-dialog"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="rename-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
    });
  });
});
