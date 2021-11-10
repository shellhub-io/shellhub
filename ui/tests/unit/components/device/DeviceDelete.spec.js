import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import DeviceDelete from '@/components/device/DeviceDelete';
import { actions, authorizer } from '../../../../src/authorizer';

describe('DeviceDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const accessType = ['owner', 'administrator', 'operator', 'observer'];

  const hasAuthorization = {
    owner: true,
    administrator: true,
    operator: false,
    observer: false,
  };

  const tests = [
    {
      description: 'Icon',
      variables: {
        dialog: false,
      },
      props: {
        uid: 'a582b47a42d',
        redirect: false,
      },
      data: {
        dialog: false,
      },
      template: {
        'deviceDelete-card': false,
        'close-btn': false,
        'remove-btn': false,
      },
    },
    {
      description: 'Dialog without redirect',
      variables: {
        dialog: true,
      },
      props: {
        uid: 'a582b47a42d',
        redirect: false,
      },
      data: {
        dialog: true,
      },
      template: {
        'deviceDelete-card': true,
        'close-btn': true,
        'remove-btn': true,
      },
    },
    {
      description: 'Dialog with redirect',
      variables: {
        dialog: true,
      },
      props: {
        uid: 'a582b47a42d',
        redirect: true,
      },
      data: {
        dialog: true,
      },
      template: {
        'deviceDelete-card': true,
        'close-btn': true,
        'remove-btn': true,
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
      'devices/remove': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    accessType.forEach((currentAccessType) => {
      describe(`${test.description} ${currentAccessType}`, () => {
        beforeEach(() => {
          wrapper = mount(DeviceDelete, {
            store: storeVuex(currentAccessType),
            localVue,
            stubs: ['fragment'],
            propsData: { uid: test.props.uid, redirect: test.props.redirect },
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
                expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Remove');
                done();
              });
            });
          }
        }
      });
    });
  });
});
