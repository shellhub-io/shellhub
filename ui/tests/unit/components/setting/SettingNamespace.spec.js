import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import SettingNamespace from '@/components/setting/SettingNamespace';
import '@/vee-validate';

describe('SettingNamespace', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

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

  const tests = [
    {
      description: 'Open version',
      variables: {
        namespace: openNamespace,
        authID: 'xxxxxxxx',
        tenant: 'xxxxxxxx',
        hasTenant: true,
        isEnterprise: false,
      },
      data: {
        namespaceMemberFormShow: false,
      },
      computed: {
        namespace: openNamespace,
        tenant: 'xxxxxxxx',
        isEnterprise: false,
      },
      components: {
        'namespaceRename-component': true,
        'namespaceMemberFormDialogAdd-component': true,
        'namespaceDelete-component': true,
      },
      template: {
        'tenant-div': true,
        'editOperation-div': true,
        'userOperation-div': true,
        'securityOperation-div': false,
        'deleteOperation-div': true,
        'speed-select': false,
      },
    },
    {
      description: 'Hosted version',
      variables: {
        namespace: hostedNamespace,
        authID: 'xxxxxxxx',
        tenant: 'xxxxxxxx',
        hasTenant: true,
        isEnterprise: true,
      },
      data: {
        namespaceMemberFormShow: false,
      },
      computed: {
        namespace: hostedNamespace,
        tenant: 'xxxxxxxx',
        isEnterprise: true,
      },
      components: {
        'namespaceRename-component': true,
        'namespaceMemberFormDialogAdd-component': true,
        'namespaceDelete-component': true,
      },
      template: {
        'tenant-div': true,
        'editOperation-div': true,
        'userOperation-div': true,
        'securityOperation-div': true,
        'deleteOperation-div': true,
        'speed-select': false,
      },
    },
  ];

  const storeVuex = (namespace, authID, tenant) => new Vuex.Store({
    namespaced: true,
    state: {
      namespace,
      authID,
      tenant,
    },
    getters: {
      'namespaces/get': (state) => state.namespace,
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'namespaces/get': () => {},
      'namespaces/removeUser': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorAssociation': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        jest.spyOn(Storage.prototype, 'getItem').mockReturnValue('e359bf484715');

        wrapper = shallowMount(SettingNamespace, {
          store: storeVuex(
            test.variables.namespace,
            test.variables.authID,
            test.variables.tenant,
          ),
          localVue,
          stubs: ['fragment'],
          vuetify,
          mocks: {
            $env: {
              isEnterprise: test.variables.isEnterprise,
            },
            $stripe: {
              elements: () => ({
                create: () => ({
                  mount: () => null,
                }),
              }),
            },
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
      });
      it('Process data in methods', () => {
        let percent = 0;
        if (test.variables.namespace.max_devices >= 0) {
          percent = (wrapper.vm.countDevicesHasNamespace()
            / test.variables.namespace.max_devices) * 100;
        }

        expect(wrapper.vm.hasTenant()).toEqual(test.variables.hasTenant);
        expect(wrapper.vm.countDevicesHasNamespace())
          .toEqual(test.variables.namespace.devices_count);
        expect(wrapper.vm.countDevicesHasNamespacePercent())
          .toEqual({ maxDevices: test.variables.namespace.max_devices, percent });
      });

      //////
      // HTML validation
      //////

      it('Renders the template with components', () => {
        Object.keys(test.components).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.components[item]);
        });
      });
      it('Renders the template with data', () => {
        Object.keys(test.template).forEach((item) => {
          expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
        });
      });
    });
  });
});
