import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import NamespaceRename from '@/components/namespace/NamespaceRename';
import '@/vee-validate';
import { actions, authorizer } from '../../../../src/authorizer';

describe('SettingNamespace', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'operator'];

  const hasAuthorizationRenameNamespace = {
    owner: true,
    operator: false,
  };

  const members = [
    {
      id: 'xxxxxxxx',
      type: 'owner',
      username: 'user1',
    },
    {
      id: 'xxxxxxxy',
      type: 'observer',
      username: 'user2',
    },
  ];

  const openNamespace = {
    name: 'namespace',
    members,
    owner: 'owner',
    tenant_id: 'xxxxxxxx',
    devices_count: 1,
    max_devices: 3,
  };

  const hostedNamespace = { ...openNamespace, max_devices: -1 };

  const invalidNamespaces = [
    '\'', '"', '!', '@', '#', '$', '%', '¨', '&', '*', '(', ')', '-', '_', '=', '+', '´', '`', '[',
    '{', '~', '^', ']', ',', '<', '..', '>', ';', ':', '/', '?',
  ];

  const invalidMinAndMaxCharacters = [
    's', 'sh', 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx', 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
  ];

  const tests = [
    {
      description: 'Open version',
      variables: {
        namespace: openNamespace,
        tenant: 'xxxxxxxx',
        hasTenant: true,
        isEnterprise: false,
      },
      data: {
        name: '',
      },
      computed: {
        namespace: openNamespace,
        tenant: 'xxxxxxxx',
      },
      template: {
        'name-text': true,
      },
    },
    {
      description: 'Hosted version',
      variables: {
        namespace: hostedNamespace,
        tenant: 'xxxxxxxx',
        hasTenant: true,
        isEnterprise: true,
      },
      data: {
        name: '',
      },
      computed: {
        namespace: hostedNamespace,
        tenant: 'xxxxxxxx',
      },
      template: {
        'name-text': true,
      },
    },
  ];

  const storeVuex = (namespace, tenant, currentrole) => new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      tenant,
      currentrole,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'auth/tenant': (state) => state.tenant,
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'namespaces/put': () => {},
      'namespaces/get': () => {},
      'namespaces/removeUser': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(async () => {
          jest.spyOn(Storage.prototype, 'getItem').mockReturnValue('e359bf484715');

          wrapper = mount(NamespaceRename, {
            store: storeVuex(
              test.variables.namespace,
              test.variables.tenant,
              currentrole,
            ),
            localVue,
            stubs: ['fragment'],
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

        it('Compare data with default value', () => {
          Object.keys(test.data).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.data[item]);
          });
        });
        it('Process data in the computed', () => {
          Object.keys(test.computed).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.computed[item]);
          });
          expect(wrapper.vm.hasAuthorizationRenameNamespace)
            .toEqual(hasAuthorizationRenameNamespace[currentrole]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        });

        //////
        // In this case, invalid RFC1123.
        //////

        invalidNamespaces.forEach((inamespace) => {
          it(`Shows invalid namespace error for ${inamespace}`, async () => {
            wrapper.setData({ name: inamespace });
            await flushPromises();

            const validator = wrapper.vm.$refs.providerName;

            await validator.validate();
            expect(validator.errors[0]).toBe('You entered an invalid RFC1123 name');
          });
        });

        //////
        // In this case, password should be 3-30 characters long.
        //////

        invalidMinAndMaxCharacters.forEach((character) => {
          it(`Shows invalid namespace error for ${character}`, async () => {
            wrapper.setData({ name: character });
            await flushPromises();

            const validator = wrapper.vm.$refs.providerName;

            await validator.validate();
            expect(validator.errors[0]).toBe('Your namespace should be 3-30 characters long');
          });
        });

        it('Show validation messages', async () => {
          //////
          // In this case, validate fields required.
          //////

          wrapper.setData({ name: '' });
          await flushPromises();

          let validator = wrapper.vm.$refs.providerName;

          await validator.validate();
          expect(validator.errors[0]).toBe('This field is required');

          //////
          // In this case, must not contain dots.
          //////

          wrapper.setData({ name: 'ShelHub.' });
          await flushPromises();

          validator = wrapper.vm.$refs.providerName;

          await validator.validate();
          expect(validator.errors[0]).toBe('The name must not contain dots');
        });
      });
    });
  });
  describe('Updating checks', () => {
    it('Should update name when namespace watcher hook gets called', () => {
      const name = 'newName';
      wrapper.vm.$options.watch.namespace.call(wrapper.vm, { name });
      expect(wrapper.vm.name).toBe('newName');
    });
  });
});
