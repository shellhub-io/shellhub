import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
// import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import DeviceActionButton from '@/components/device/DeviceActionButton';
import { actions, authorizer } from '../../../../src/authorizer';

describe('DeviceActionButton', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const accessType = ['owner', 'administrator', 'operator', 'observer'];

  const hasAuthorization = {
    owner: true,
    administrator: true,
    operator: true,
    observer: false,
  };

  const tests = [
    {
      description: 'Create button in the notification',
      variables: {
        isActive: true,
        dialog: false,
      },
      props: {
        uid: 'xxxxxxxx',
        notificationStatus: true,
        action: 'accept',
      },
      data: {
        dialog: false,
      },
      template: {
        'notification-btn': true,
        'tooltip-text': false,
        'deviceActionButton-card': false,
        'cancel-btn': false,
        'dialog-btn': false,
      },
    },
    {
      description: 'Create button in the list',
      variables: {
        isActive: true,
        dialog: false,
      },
      props: {
        uid: 'xxxxxxxx',
        notificationStatus: false,
        action: 'accept',
      },
      data: {
        dialog: false,
      },
      template: {
        'notification-btn': false,
        'tooltip-text': true,
        'deviceActionButton-card': false,
        'cancel-btn': false,
        'dialog-btn': false,
      },
    },
    {
      description: 'Reject button in the list',
      variables: {
        isActive: true,
        dialog: false,
      },
      props: {
        uid: 'xxxxxxxx',
        notificationStatus: false,
        action: 'reject',
      },
      data: {
        dialog: false,
      },
      template: {
        'notification-btn': false,
        'tooltip-text': true,
        'deviceActionButton-card': false,
        'cancel-btn': false,
        'dialog-btn': false,
      },
    },
    {
      description: 'Remove button in the list',
      variables: {
        isActive: true,
        dialog: false,
      },
      props: {
        uid: 'xxxxxxxx',
        notificationStatus: false,
        action: 'remove',
      },
      data: {
        dialog: false,
      },
      template: {
        'notification-btn': false,
        'tooltip-text': true,
        'deviceActionButton-card': false,
        'cancel-btn': false,
        'dialog-btn': false,
      },
    },
    {
      description: 'Dialog',
      variables: {
        isActive: true,
        dialog: true,
      },
      props: {
        uid: 'xxxxxxxx',
        notificationStatus: false,
        action: 'accept',
      },
      data: {
        dialog: true,
      },
      template: {
        'notification-btn': false,
        'tooltip-text': true,
        'deviceActionButton-card': true,
        'cancel-btn': true,
        'dialog-btn': true,
      },
    },
  ];

  const storeVuex = (isActive, currentAccessType) => new Vuex.Store({
    namespaced: true,
    state: {
      isActive,
      currentAccessType,
    },
    getters: {
      isActive: (state) => state.isActive,
      'auth/accessType': (state) => state.currentAccessType,
    },
    actions: {
      'devices/refresh': () => {},
      'devices/accept': () => {},
      'users/setStatusUpdateAccountDialog': () => {},
      'devices/reject': () => {},
      'devices/remove': () => {},
      'notifications/fetch': () => {},
      'stats/get': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  tests.forEach((test) => {
    accessType.forEach((currentAccessType) => {
      describe(`${test.description} ${currentAccessType}`, () => {
        beforeEach(() => {
          wrapper = mount(DeviceActionButton, {
            store: storeVuex(test.variables.isActive, currentAccessType),
            localVue,
            stubs: ['fragment'],
            propsData: {
              uid: test.props.uid,
              notificationStatus: test.props.notificationStatus,
              action: test.props.action,
            },
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
          if (!(test.props.action === 'remove' && currentAccessType === 'operator')) {
            expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentAccessType]);
          } else {
            expect(wrapper.vm.hasAuthorization).toEqual(false);
          }
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
