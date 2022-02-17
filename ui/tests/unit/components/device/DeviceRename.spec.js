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

  document.body.setAttribute('data-app', true);

  let wrapper;

  const name = '39-5e-2a';
  const uid = 'a582b47a42d';

  const invalidNames = [
    '\'', '"', '!', '@', '#', '$', '%', '¨', '&', '*', '(', ')', '-', '_', '=', '+', '´', '`', '[',
    '{', '~', '^', ']', ',', '<', '..', '>', ';', ':', '/', '?',
  ];

  const invalidMinAndMaxCharacters = [
    'xx', 'xx', 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx', 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
  ];

  const tests = [
    {
      description: 'Dialog closed',
      props: {
        name,
        uid,
        show: false,
      },
      data: {
        invalid: false,
        editName: name,
        messages: 'Examples: (foobar, foo-bar-ba-z-qux, foo-example, 127-0-0-1)',
      },
      computed: {
        device: {
          name,
          uid,
        },
      },
      template: {
        'rename-icon': true,
        'rename-title': true,
        'deviceRename-card': false,
      },
      templateText: {
        'rename-title': 'Rename',
      },
    },
    {
      description: 'Dialog opened',
      props: {
        name,
        uid,
        show: true,
      },
      data: {
        invalid: false,
        editName: name,
        messages: 'Examples: (foobar, foo-bar-ba-z-qux, foo-example, 127-0-0-1)',
      },
      computed: {
        device: {
          name,
          uid,
        },
      },
      template: {
        'rename-icon': true,
        'rename-title': true,
        'deviceRename-card': true,
        'text-title': true,
        'hostname-field': true,
        'close-btn': true,
        'rename-btn': true,
      },
      templateText: {
        'rename-title': 'Rename',
        'text-title': 'Rename Device',
        'close-btn': 'Close',
        'rename-btn': 'Rename',
      },
    },
  ];

  const storeVuex = () => new Vuex.Store({
    namespaced: true,
    state: { },
    getters: { },
    actions: {
      'devices/rename': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(DeviceRename, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: {
            name: test.props.name,
            uid: test.props.uid,
            show: test.props.show,
          },
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
      // Data checking
      //////

      it('Receive data in props', () => {
        Object.keys(test.props).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.props[item]);
        });
      });
      it('Compare data with default value', () => {
        Object.keys(test.data).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.data[item]);
        });
      });
      it('Process data in the computed', () => {
        Object.keys(test.computed).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.computed[item]);
        });
      });

      //////
      // HTML validation
      //////

      it('Renders the template with data', () => {
        Object.keys(test.template).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
        });
      });
      it('Renders template with expected text', () => {
        Object.keys(test.templateText).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).text()).toContain(test.templateText[item]);
        });
      });

      if (test.props.show) {
        //////
        // In this case, the empty fields are validated.
        //////

        it('Show validation messages, empty fields', async () => {
          wrapper.setData({ editName: '' });
          await flushPromises();

          const validator = wrapper.vm.$refs.providerHostname;

          await validator.validate();
          expect(validator.errors[0]).toBe('This field is required');
        });

        //////
        // In this case, must not contain dots are validated.
        //////

        it('Show validation messages, must not contain dots', async () => {
          wrapper.setData({ editName: 'ShelHub.' });
          await flushPromises();

          const validator = wrapper.vm.$refs.providerHostname;

          await validator.validate();
          expect(validator.errors[0]).toBe('The name must not contain dots');
        });

        //////
        // In this case, RFC1123 rules are validated.
        //////

        it('Show validation messages, RFC1123 rules', async (done) => {
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

        it('Show validation messages, min and max characters are validated', async (done) => {
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
      }
    });
  });
});
