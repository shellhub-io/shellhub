import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import timezoneMock from 'timezone-mock';
import DeviceDetails from '@/components/device/DeviceDetails';
import { actions, authorizer } from '../../../../src/authorizer';

describe('DeviceDetails', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const deviceOnline = {
    uid: 'a582b47a',
    name: '39-5e-2a',
    identity: {
      mac: '00:00:00',
    },
    info: {
      id: 'arch',
      pretty_name: 'Linux',
      version: '',
    },
    public_key: 'xxxxxxxx',
    tenant_id: '00000000',
    last_seen: '2020-05-20T18:58:53.276Z',
    online: true,
    namespace: 'user',
    status: 'accepted',
    tags: ['device1', 'device2'],
  };

  const deviceOffline = { ...deviceOnline, online: false, status: 'pending' };

  const tests = [
    {
      description: 'Online Device and owner',
      role: {
        type: 'owner',
        permission: true,
      },
      variables: {
        device: deviceOnline,
      },
      data: {
        uid: deviceOnline.uid,
        hostname: 'localhost',
        hide: true,
        device: deviceOnline,
        dialogDelete: false,
        dialogError: false,
        deviceDeleteShow: false,
      },
      computed: {
        hasAuthorizationRename: true,
        hasAuthorizationFormUpdate: true,
      },
      components: {
        'deviceRename-component': true,
        'tagFormUpdate-component': true,
        'terminalDialog-component': true,
        'deviceDelete-component': true,
      },
      template: {
        'deviceUid-field': true,
        'deviceMac-field': true,
        'devicePrettyName-field': true,
        'deviceConvertDate-field': true,
      },
      templateText: {
        'deviceUid-field': deviceOnline.uid,
        'deviceMac-field': deviceOnline.identity.mac,
        'devicePrettyName-field': deviceOnline.info.pretty_name,
        'deviceConvertDate-field': 'Wednesday, May 20th 2020, 6:58:53 pm',
      },
    },
    {
      description: 'Offline Device',
      role: {
        type: 'observer',
        permission: false,
      },
      variables: {
        device: deviceOffline,
      },
      data: {
        uid: deviceOffline.uid,
        hostname: 'localhost',
        hide: true,
        device: deviceOffline,
        dialogDelete: false,
        dialogError: false,
        deviceDeleteShow: false,
      },
      computed: {
        hasAuthorizationRename: false,
        hasAuthorizationFormUpdate: false,
      },
      components: {
        'deviceRename-component': true,
        'tagFormUpdate-component': true,
        'terminalDialog-component': false,
        'deviceDelete-component': true,
      },
      template: {
        'deviceUid-field': true,
        'deviceMac-field': true,
        'devicePrettyName-field': true,
        'deviceConvertDate-field': true,
      },
      templateText: {
        'deviceUid-field': deviceOffline.uid,
        'deviceMac-field': deviceOffline.identity.mac,
        'devicePrettyName-field': deviceOffline.info.pretty_name,
        'deviceConvertDate-field': 'Wednesday, May 20th 2020, 6:58:53 pm',
      },
    },
  ];

  const storeVuex = (device, currentRole) => new Vuex.Store({
    namespaced: true,
    state: {
      device,
      currentRole,
    },
    getters: {
      'devices/get': (state) => state.device,
      'auth/role': (state) => state.currentRole,
    },
    actions: {
      'devices/get': () => {},
      'devices/updateTag': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        timezoneMock.register('UTC');

        wrapper = shallowMount(DeviceDetails, {
          store: storeVuex(
            test.variables.device,
            test.role.type,
          ),
          localVue,
          stubs: ['fragment'],
          vuetify,
          mocks: {
            $authorizer: authorizer,
            $actions: actions,
            $route: {
              params: {
                id: test.variables.device.uid,
              },
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
    });
  });
});
