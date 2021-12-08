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

  const accessType = ['owner', 'operator'];

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
        dialog: false,
      },
      props: {
        uid: session.uid,
        device: session.device_uid,
      },
      data: {
        session,
        dialog: false,
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
        dialog: true,
      },
      props: {
        uid: session.uid,
        device: session.device_uid,
      },
      data: {
        session,
        dialog: true,
        action: 'close',
      },
      template: {
        'sessionClose-card': true,
        'cancel-btn': true,
        'close-btn': true,
      },
    },
  ];

  const storeVuex = (currentAccessType) => new Vuex.Store({
    namespaced: true,
    state: {
      currentAccessType,
    },
    getters: {
      'auth/accessType': (state) => state.currentAccessType,
    },
    actions: {
      'sessions/close': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    accessType.forEach((currentAccessType) => {
      describe(`${test.description} ${currentAccessType}`, () => {
        beforeEach(() => {
          wrapper = mount(SessionClose, {
            store: storeVuex(currentAccessType),
            localVue,
            stubs: ['fragment'],
            propsData: { uid: test.props.uid, device: test.props.device },
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
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentAccessType]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        });

        if (!test.data.dialog) {
          if (hasAuthorization[currentAccessType]) {
            it('Show message tooltip user has permission', async (done) => {
              const icons = wrapper.findAll('.v-icon');
              const helpIcon = icons.at(0);
              helpIcon.trigger('mouseenter');
              await wrapper.vm.$nextTick();

              expect(icons.length).toBe(1);
              requestAnimationFrame(() => {
                expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Close');
                done();
              });
            });
          }
        }
      });
    });
  });
});
