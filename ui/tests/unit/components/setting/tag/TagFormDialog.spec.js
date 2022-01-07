import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import TagFormDialog from '@/components/setting/tag/TagFormDialog';
import { actions, authorizer } from '../../../../../src/authorizer';
import '@/vee-validate';

describe('TagFormDialog', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'observer'];

  const hasAuthorization = {
    owner: true,
    observer: false,
  };

  // vee-validate variables bellow
  const invalidName = ['xxx/', 'xxx@', 'xxx&', 'xxx:'];
  const invalidMinAndMaxCharacters = [
    'x', 'xx',
  ];

  const tests = [
    {
      description: 'Icon create',
      props: {
        action: 'create',
        uid: '',
        tagName: '',
        show: false,
      },
      data: {
        dialog: false,
        tagLocal: '',
      },
      template: {
        'create-item': true,
        'create-icon': true,
        'edit-item': false,
        'edit-icon': false,
        'tagForm-card': false,
      },
    },
    {
      description: 'Icon edit',
      props: {
        action: 'edit',
        uid: '',
        tagName: '',
        show: false,
      },
      data: {
        dialog: false,
        tagLocal: '',
      },
      template: {
        'create-item': false,
        'create-icon': false,
        'edit-item': true,
        'edit-icon': true,
        'tagForm-card': false,
      },
    },
    {
      description: 'Dialog create',
      props: {
        action: 'create',
        uid: '',
        tagName: '',
        show: true,
      },
      data: {
        dialog: false,
        tagLocal: '',
      },
      template: {
        'create-item': true,
        'create-icon': true,
        'edit-item': false,
        'edit-icon': false,
        'tagForm-card': true,
        'cancel-btn': true,
        'doAction-btn': true,
      },
    },
    {
      description: 'Dialog edit',
      props: {
        action: 'edit',
        uid: '',
        tagName: '',
        show: true,
      },
      data: {
        dialog: false,
        tagLocal: '',
      },
      template: {
        'create-item': false,
        'create-icon': false,
        'edit-item': true,
        'edit-icon': true,
        'tagForm-card': true,
        'cancel-btn': true,
        'doAction-btn': true,
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
      'tags/post': () => {},
      'tags/edit': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(async () => {
          wrapper = mount(TagFormDialog, {
            store: storeVuex(currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: {
              action: test.props.action,
              uid: test.props.ui,
              tagName: test.props.tagName,
              show: test.props.show,
            },
            vuetify,
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
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
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          if (hasAuthorization[currentrole]) {
            Object.keys(test.template).forEach((item) => {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
            });
          } else if (!test.props.show) {
            Object.keys(test.template).forEach((item) => {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
            });
          }
        });

        // Here are two condictions:
        // - The first, when the icon is tested;
        // - And the second, when the dialog is tested;
        /// ///

        if (hasAuthorization[currentrole] && test.props.show) {
          //////
          // In this case, the route identifier are validated.
          //////

          it('Show validation messages', async (done) => {
            invalidName.forEach(async (name) => {
              wrapper.setData({ tagLocal: name });
              await flushPromises();

              const validator = wrapper.vm.$refs.providerTag;

              await validator.validate();
              expect(validator.errors[0]).toBe('The name must not contain /, @, &, and :.');

              await flushPromises();
              done();
            });
          });

          //////
          // In this case, the min and max characters are validated.
          //////

          it('Show validation messages', async (done) => {
            let invalidMaxCharacter = '';

            for (let x = 0; x < 256; x += 1) {
              invalidMaxCharacter = invalidMaxCharacter.concat('x');
            }
            invalidMinAndMaxCharacters.push(invalidMaxCharacter);
            invalidMinAndMaxCharacters.push(invalidMaxCharacter.concat('x'));

            invalidMinAndMaxCharacters.forEach(async (character) => {
              wrapper.setData({ tagLocal: character });
              await flushPromises();

              const validator = wrapper.vm.$refs.providerTag;

              await validator.validate();
              expect(validator.errors[0]).toBe('Your tag should be 3-255 characters long');

              await flushPromises();
              done();
            });
          });
        }
      });
    });
  });
});
