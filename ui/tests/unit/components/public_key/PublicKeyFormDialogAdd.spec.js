import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import PublicKeyFormDialogAdd from '@/components/public_key/PublicKeyFormDialogAdd';
import { actions, authorizer } from '../../../../src/authorizer';
import '@/vee-validate';

describe('PublicKeyFormDialogAdd', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const tests = [
    {
      description: 'Button create publicKey has authorization',
      role: {
        type: 'owner',
        permission: true,
      },
      data: {
        dialog: false,
        action: 'create',
        keyLocal: {},
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      computed: {
        hasAuthorization: true,
      },
      template: {
        'createKey-btn': true,
        'publicKeyFormDialog-card': false,
      },
      templateText: {
        'createKey-btn': 'Add Public Key',
      },
    },
    {
      description: 'Button create publicKey has no authorization',
      role: {
        type: 'operator',
        permission: false,
      },
      data: {
        dialog: false,
        keyLocal: {},
        action: 'create',
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      computed: {
        hasAuthorization: false,
      },
      template: {
        'createKey-btn': true,
        'publicKeyFormDialog-card': false,
      },
      templateText: {
        'createKey-btn': 'Add Public Key',
      },
    },
    {
      description: 'Dialog create publicKey has authorization',
      role: {
        type: 'owner',
        permission: true,
      },
      data: {
        dialog: true,
        keyLocal: {},
        action: 'create',
        hostname: '',
        choiceFilter: 'all',
        choiceUsername: 'all',
        supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
      },
      computed: {
        hasAuthorization: true,
      },
      template: {
        'createKey-btn': true,
        'publicKeyFormDialog-card': true,
        'text-title': true,
        'name-field': true,
        'hostname-field': false,
        'username-field': false,
        'data-field': true,
        'cancel-btn': true,
        'create-btn': true,
      },
      templateText: {
        'createKey-btn': 'Add Public Key',
        'text-title': 'New Public Key',
        'name-field': '',
        'data-field': '',
        'cancel-btn': 'Cancel',
        'create-btn': 'Create',
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
      'publickeys/post': () => {},
      'publickeys/put': () => {},
      'privatekeys/set': () => {},
      'privatekeys/edit': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarSuccessNotRequest': () => {},
      'snackbar/showSnackbarErrorNotRequest': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description} - ${test.role.type}`, () => {
      beforeEach(() => {
        wrapper = mount(PublicKeyFormDialogAdd, {
          store: storeVuex(test.role.type),
          localVue,
          stubs: ['fragment'],
          vuetify,
          mocks: {
            $authorizer: authorizer,
            $actions: actions,
          },
        });

        wrapper.setData({ dialog: test.data.dialog });
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

      Object.keys(test.data).forEach((item) => {
        it(`Compare data ${item} with default value`, () => {
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
      if (test.data.dialog && test.role.permission) {
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
      }
    });
  });
});
