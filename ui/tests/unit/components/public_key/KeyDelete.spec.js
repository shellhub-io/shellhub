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

  const role = ['owner', 'operator'];

  const hasAuthorization = {
    owner: true,
    operator: false,
  };

  const tests = [
    {
      description: 'Icon public',
      props: {
        fingerprint: 'b7:25:f8',
        action: 'public',
        show: false,
      },
      template: {
        'keyDelete-card': false,
        'close-btn': false,
        'remove-btn': false,
      },
    },
    {
      description: 'Icon private',
      props: {
        fingerprint: 'b7:25:f8',
        action: 'private',
        show: false,
      },
      template: {
        'keyDelete-card': false,
        'close-btn': false,
        'remove-btn': false,
      },
    },
    {
      description: 'Dialog public',
      props: {
        fingerprint: 'b7:25:f8',
        action: 'public',
        show: true,
      },
      template: {
        'keyDelete-card': true,
        'close-btn': true,
        'remove-btn': true,
      },
    },
    {
      description: 'Dialog private',
      props: {
        fingerprint: 'b7:25:f8',
        action: 'private',
        show: true,
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
        it('Process data in the computed', () => {
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });

        // //////
        // // HTML validation
        // //////

        it('Renders the template with data', () => {
          Object.keys(test.template).forEach((item) => {
            if (!hasAuthorization[currentrole] && currentrole === 'operator' && test.props.show) {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(!test.template[item]);
            } else {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
            }
          });
        });
      });
    });
  });
});
