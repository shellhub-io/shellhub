import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import SessionClose from '@/components/session/SessionClose';
import { actions, authorizer } from '../../../../src/authorizer';

describe('SessionClose', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

  const session = {
    uid: '8c354a00',
    device_uid: 'a582b47a',
  };

  const tests = [
    {
      description: 'Icon',
      variables: {
        session,
      },
      props: {
        uid: session.uid,
        device: session.device_uid,
        show: false,
      },
      data: {
        session,
        action: 'close',
      },
      template: {
        'sessionClose-card': false,
        'cancel-btn': false,
        'close-btn': false,
      },
    },
    {
      description: 'Dialog',
      variables: {
        session,
      },
      props: {
        uid: session.uid,
        device: session.device_uid,
        show: true,
      },
      data: {
        session,
        action: 'close',
      },
      template: {
        'sessionClose-card': true,
        'cancel-btn': true,
        'close-btn': true,
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
      'sessions/close': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(SessionClose, {
            store: storeVuex(currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: {
              uid: test.props.uid,
              device: test.props.device,
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
      });
    });
  });
});
