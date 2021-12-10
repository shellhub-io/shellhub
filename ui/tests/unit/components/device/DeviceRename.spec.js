import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import DeviceRename from '@/components/device/DeviceRename';
import { actions, authorizer } from '../../../../src/authorizer';
import '@/vee-validate';

describe('DeviceRename', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: true,
  };

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
      description: 'Icon',
      variables: {
        dialog: false,
      },
      props: {
        name,
        uid,
      },
      data: {
        dialog: false,
        invalid: false,
        editName: name,
        messages: 'Examples: (foobar, foo-bar-ba-z-qux, foo-example, 127-0-0-1)',
        action: 'rename',
      },
      computed: {
        device: {
          name,
          uid,
        },
      },
      template: {
        'deviceRename-card': false,
        'rename-btn': false,
        'cancel-btn': false,
      },
    },
    {
      description: 'Dialog',
      variables: {
        dialog: true,
      },
      props: {
        name,
        uid,
      },
      data: {
        dialog: true,
        invalid: false,
        editName: name,
        messages: 'Examples: (foobar, foo-bar-ba-z-qux, foo-example, 127-0-0-1)',
        action: 'rename',
      },
      computed: {
        device: {
          name,
          uid,
        },
      },
      template: {
        'deviceRename-card': true,
        'rename-btn': true,
        'cancel-btn': true,
      },
    },
  ];

  const storeVuex = (currentrole) => new Vuex.Store({
    namespaced: true,
    state: {
      currentrole,
    },
    getters: {
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'devices/rename': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(DeviceRename, {
            store: storeVuex(currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: { name: test.props.name, uid: test.props.uid },
            vuetify,
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
            },
          });

          wrapper.setData({ dialog: test.variables.dialog });
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
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        });

        // Here are two condictions:
        // - The first, when the icon is tested;
        // - And the second, when the dialog is tested;
        /// ///

        if (!test.data.dialog) {
          if (hasAuthorization[currentrole]) {
            it('Show message tooltip user has permission', async (done) => {
              const icons = wrapper.findAll('.v-icon');
              const helpIcon = icons.at(0);
              helpIcon.trigger('mouseenter');
              await wrapper.vm.$nextTick();

              expect(icons.length).toBe(1);
              requestAnimationFrame(() => {
                expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Edit');
                done();
              });
            });
          }
        } else if (hasAuthorization[currentrole]) {
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
});
