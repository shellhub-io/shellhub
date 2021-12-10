import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import KeyDelete from '@/components/public_key/KeyDelete';
import { actions, authorizer } from '../../../../src/authorizer';

describe('KeyDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'administrator', 'operator', 'observer'];

  const hasAuthorization = {
    owner: true,
    administrator: true,
    operator: false,
    observer: false,
  };

  const tests = [
    {
      description: 'Icon public',
      variables: {
        dialog: false,
      },
      props: {
        fingerprint: 'b7:25:f8',
        action: 'public',
      },
      data: {
        dialog: false,
      },
      template: {
        'keyDelete-card': false,
        'close-btn': false,
        'remove-btn': false,
      },
    },
    {
      description: 'Icon private',
      variables: {
        dialog: false,
      },
      props: {
        fingerprint: 'b7:25:f8',
        action: 'private',
      },
      data: {
        dialog: false,
      },
      template: {
        'keyDelete-card': false,
        'close-btn': false,
        'remove-btn': false,
      },
    },
    {
      description: 'Dialog public',
      variables: {
        dialog: true,
      },
      props: {
        fingerprint: 'b7:25:f8',
        action: 'public',
      },
      data: {
        dialog: true,
      },
      template: {
        'keyDelete-card': true,
        'close-btn': true,
        'remove-btn': true,
      },
    },
    {
      description: 'Dialog private',
      variables: {
        dialog: true,
      },
      props: {
        fingerprint: 'b7:25:f8',
        action: 'private',
      },
      data: {
        dialog: true,
      },
      template: {
        'keyDelete-card': true,
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
      'publickeys/remove': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(KeyDelete, {
            store: storeVuex(currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: {
              fingerprint: test.props.fingerprint,
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
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });

        // //////
        // // HTML validation
        // //////

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
