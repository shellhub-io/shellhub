import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import DeviceChooser from '@/components/device/DeviceChooser';
import { actions, authorizer } from '../../../../src/authorizer';

describe('DeviceChooser', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  let wrapper;

  document.body.setAttribute('data-app', true);

  const role = ['owner', 'administrator'];

  const hasAuthorization = {
    owner: true,
    administrator: false,
  };

  const tests = [
    {
      description: 'Dialog is closes',
      variables: {
        deviceChooserStatus: false,
        devicesSelected: [],
        filter: [],
        devices: [],
        dialog: false,
      },
      data: {
        hostname: 'localhost',
        action: 'suggestedDevices',
        dialog: false,
        items: [
          {
            title: 'Suggested Devices',
            action: 'suggestedDevices',
          },
          {
            title: 'All devices',
            action: 'allDevices',
          },
        ],
        permissionAction: 'chooser',
      },
      computed: {
        disableTooltipOrButton: false,
        equalThreeDevices: false,
      },
      components: {
        'deviceChooserStatus-dialog': false,
      },
      template: {
        'deviceChooserStatus-dialog': false,
        'close-btn': false,
        'accept-btn': false,
      },
    },
  ];

  const storeVuex = (
    deviceChooserStatus,
    devicesSelected,
    filter,
    devices,
    currentrole,
  ) => new Vuex.Store({
    namespaced: true,
    state: {
      deviceChooserStatus,
      devicesSelected,
      filter,
      devices,
      currentrole,
    },
    getters: {
      'devices/getDeviceChooserStatus': (state) => state.deviceChooserStatus,
      'devices/getDevicesSelected': (state) => state.devicesSelected,
      'devices/getFilter': (state) => state.filter,
      'devices/list': (state) => state.devices,
      'auth/role': (state) => state.currentrole,
    },
    actions: {
      'stats/get': () => {},
      'devices/getDevicesMostUsed': () => {},
      'devices/postDevicesChooser': () => {},
      'devices/setDevicesForUserToChoose': () => {},
      'devices/setDeviceChooserStatus': () => {},
      'snackbar/showSnackbarDeviceChooser': () => {},
      'snackbar/showSnackbarErrorAssociation': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(DeviceChooser, {
            store: storeVuex(
              test.variables.deviceChooserStatus,
              test.variables.devicesSelected,
              test.variables.filter,
              test.variables.devices,
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
});
