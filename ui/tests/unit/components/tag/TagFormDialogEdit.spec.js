import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import TagFormDialogEdit from '@/components/tag/TagFormDialogEdit';
import '@/vee-validate';

describe('TagFormDialogEdit', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  // vee-validate variables bellow
  const invalidName = ['xxx/', 'xxx@', 'xxx&', 'xxx:'];
  const invalidMinAndMaxCharacters = [
    'x', 'xx',
  ];

  const tests = [
    {
      description: 'Dialog closed',
      props: {
        uid: 'b7:25:f8',
        tagName: 'ShellHub',
        show: false,
      },
      data: {
        dialog: false,
        tagLocal: '',
      },
      template: {
        'edit-icon': true,
        'edit-title': true,
        'edit-itemtagForm-card': false,
      },
      templateText: {
        'edit-title': 'Edit',
      },
    },
    {
      description: 'Dialog Opened',
      props: {
        uid: 'b7:25:f8',
        tagName: 'ShellHub',
        show: true,
      },
      data: {
        dialog: false,
        tagLocal: '',
      },
      template: {
        'edit-icon': true,
        'edit-title': true,
        'tagForm-card': true,
        'text-title': true,
        'name-field': true,
        'cancel-btn': true,
        'edit-btn': true,
      },
      templateText: {
        'edit-title': 'Edit',
        'text-title': 'Edit tag',
        'cancel-btn': 'Cancel',
        'edit-btn': 'Edit',
      },
    },
  ];

  const storeVuex = () => new Vuex.Store({
    namespaced: true,
    state: { },
    getters: { },
    actions: {
      'tags/edit': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(TagFormDialogEdit, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: {
            uid: test.props.uid,
            tagName: test.props.tagName,
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
          const invalidMaxCharacter = 'x'.repeat(256);

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
