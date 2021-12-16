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
      props: {
        uid: 'a582b47a42d',
        redirect: false,
        show: false,
      },
      data: {
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
      props: {
        uid: 'a582b47a42d',
        redirect: false,
        show: true,
      },
      data: {
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
      props: {
        uid: 'a582b47a42d',
        redirect: true,
        show: true,
      },
      data: {
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
            propsData: {
              uid: test.props.uid,
              redirect: test.props.redirect,
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
