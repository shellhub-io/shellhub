import Vuex from 'vuex';
import { config, mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import DeviceAdd from '@/components/device/DeviceAdd';
import { actions, authorizer } from '../../../../src/authorizer';

// mocks global vars and clipboard function
config.mocks = {
  $copy: {
    command: 'Command',
    deviceSSHID: 'Device SSHID',
    tenantId: 'Tenant ID',
  },
  $clipboard: () => {},
};

describe('DeviceAdd', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'observer'];

  const hasAuthorization = {
    owner: true,
    observer: false,
  };

  const tests = [
    {
      description: 'Button',
      variables: {
        addDevice: false,
        tenant: 'xxxxxxxx',
        dialog: false,
      },
      props: {
        smallButton: false,
      },
      data: {
        hostname: 'localhost',
        port: '',
        dialog: false,
        action: 'add',
      },
      computed: {
        tenant: 'xxxxxxxx',
      },
      method: {
        command: 'curl -sSf "http://localhost/install.sh?tenant_id=xxxxxxxx" | sh',
      },
      template: {
        'add-btn': true,
        'deviceAdd-dialog': false,
        'close-btn': false,
      },
    },
    {
      description: 'Dialog',
      variables: {
        addDevice: true,
        tenant: 'xxxxxxxx',
        dialog: true,
      },
      props: {
        smallButton: true,
      },
      data: {
        hostname: 'localhost',
        port: '',
        dialog: true,
        action: 'add',
      },
      computed: {
        tenant: 'xxxxxxxx',
      },
      method: {
        command: 'curl -sSf "http://localhost/install.sh?tenant_id=xxxxxxxx" | sh',
      },
      template: {
        'add-btn': true,
        'deviceAdd-dialog': true,
        'close-btn': true,
      },
    },
  ];

  const storeVuex = (addDevice, tenant, currentrole) => new Vuex.Store({
    namespaced: true,
    state: {
      tenant,
      addDevice,
      currentrole,
    },
    getters: {
      'auth/tenant': (state) => state.tenant,
      'modals/addDevice': (state) => state.addDevice,
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'modals/showAddDevice': () => {},
      'snackbar/showSnackbarCopy': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(DeviceAdd, {
            store: storeVuex(test.variables.addDevice, test.variables.tenant, currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: { smallButton: test.props.smallButton },
            vuetify,
            mocks: {
              $authorizer: authorizer,
              $actions: actions,
              $copy: {
                command: 'Command',
                deviceSSHID: 'Device SSHID',
                tenantId: 'Tenant ID',
              },
              $clipboard: () => {},
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
        it('Process data in methods', () => {
          jest.spyOn(wrapper.vm, 'copyCommand');
          wrapper.vm.copyCommand();

          expect(wrapper.vm.command()).toBe(test.method.command);
          expect(wrapper.vm.copyCommand).toHaveBeenCalled();
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        });
      });
    });
  });
});
