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

  const role = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
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
        action: 'remove',
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
        action: 'remove',
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
        action: 'remove',
      },
      template: {
        'deviceDelete-card': true,
        'close-btn': true,
        'remove-btn': true,
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
      'devices/remove': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(DeviceDelete, {
            store: storeVuex(currentrole),
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
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
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
          if (hasAuthorization[currentrole]) {
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
