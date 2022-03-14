import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import PublicKeyFormDialogEdit from '@/components/public_key/PublicKeyFormDialogEdit';
import '@/vee-validate';

describe('PublicKeyFormDialogEdit', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const supportedKeys = 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.';

  const keyObject = {
    name: 'ShellHub',
    username: 'ShellHub',
    data: '',
    filter: {
      hostname: '.*',
    },
  };

  const keyObject2 = {
    name: 'ShellHub',
    username: 'ShellHub',
    data: '',
    filter: {
      tags: ['tag1', 'tag2'],
    },
  };

  const tests = [
    {
      description: 'Dialog closed',
      props: {
        keyObject,
        show: false,
      },
      data: {
        keyObject,
        supportedKeys,
      },
      template: {
        'edit-icon': true,
        'edit-title': true,
        'namespaceNewMember-dialog': false,
      },
      templateText: {
        'edit-title': 'Edit',
      },
    },
    {
      description: 'Dialog opened',
      props: {
        keyObject,
        show: true,
      },
      data: {
        keyObject,
        supportedKeys,
      },
      template: {
        'edit-icon': true,
        'edit-title': true,
        'publicKeyFormDialog-card': true,
        'text-title': true,
        'name-field': true,
        'hostname-field': true,
        'tags-field': false,
        'username-field': true,
        'data-field': true,
        'cancel-btn': true,
        'edit-btn': true,
        'access-restriction-field': true,
      },
      templateText: {
        'edit-title': 'Edit',
        'text-title': 'Edit Public Key',
        'cancel-btn': 'Cancel',
        'edit-btn': 'Edit',
      },
    },
    {
      description: 'Dialog with tags',
      props: {
        keyObject: keyObject2,
        show: true,
      },
      data: {
        keyObject: keyObject2,
        supportedKeys,
      },
      template: {
        'edit-icon': true,
        'edit-title': true,
        'publicKeyFormDialog-card': true,
        'text-title': true,
        'name-field': true,
        'username-field': true,
        'data-field': true,
        'cancel-btn': true,
        'edit-btn': true,
        'access-restriction-field': true,
      },
      templateText: {
        'edit-title': 'Edit',
        'text-title': 'Edit Public Key',
        'cancel-btn': 'Cancel',
        'edit-btn': 'Edit',
      },
    },
  ];

  const storeVuex = () => new Vuex.Store({
    namespaced: true,
    state: {},
    getters: {},
    actions: {
      'publickeys/put': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(PublicKeyFormDialogEdit, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: {
            keyObject: test.props.keyObject,
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

      Object.keys(test.template).forEach((item) => {
        it(`Renders the template ${item} with data`, () => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
        });
      });
      it('Renders template with expected text', () => {
        Object.keys(test.templateText).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).text()).toContain(test.templateText[item]);
        });
      });
      if (test.props.show) {
        it('Show validation messages', async () => {
          //////
          // In this case, the empty fields are validated.
          //////

          wrapper.setData({ keyLocal: { name: '', data: '' } });
          await flushPromises();

          const validatorName = wrapper.vm.$refs.providerName;
          let validatorData = wrapper.vm.$refs.providerData;

          await validatorName.validate();
          await validatorData.validate();
          expect(validatorName.errors[0]).toBe('This field is required');
          expect(validatorData.errors[0]).toBe('This field is required');

          //////
          // In this case, any string is validated in the data.
          //////

          wrapper.setData({ keyLocal: { data: 'xxxxxxxx' } });
          await flushPromises();

          validatorData = wrapper.vm.$refs.providerData;

          await validatorData.validate();
          expect(validatorData.errors[0]).toBe('Not valid key');
        });
        it('Populate filter according to key', async () => {
          wrapper.setData({ showDialog: true });
          await flushPromises();

          const { tags } = wrapper.vm.keyObject.filter;
          const { hasTags } = wrapper.vm;
          expect(hasTags).toBe(!!tags);

          // call update hook
          wrapper.vm.handleUpdate();
          await flushPromises();

          if (hasTags) {
            // compare data
            expect(wrapper.vm.tagChoices).toStrictEqual(['tag1', 'tag2']);
            expect(wrapper.vm.choiceFilter).toStrictEqual('tags');

            // render template
            expect(wrapper.find('[data-test="hostname-field"]').exists()).toBe(false);
            expect(wrapper.find('[data-test="tags-field"]').exists()).toBe(true);
          }
        });
      }
    });
  });

  describe('Update data checks', () => {
    it('Should filter selection show field with data', async () => {
      const pubKey = {
        name: 'ShellHub',
        username: 'ShellHub',
        data: '',
        filter: {
          hostname: '.*',
        },
      };

      await wrapper.setProps({ keyObject: pubKey });
      await flushPromises();

      wrapper.vm.handleUpdate();

      await wrapper.setData({ choiceFilter: 'hostname' });
      await flushPromises();

      wrapper.vm.chooseFilter();
      expect(wrapper.vm.hostname).toBe('');
    });
  });
});
