import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import FirewallRuleDelete from '@/components/firewall_rule/FirewallRuleDelete';
import { actions, authorizer } from '../../../../src/authorizer';

describe('FirewallRuleDelete', () => {
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
        id: 'a582b47a42d',
        show: false,
      },
      data: {
        action: 'remove',
      },
      template: {
        'firewallRuleDelete-dialog': false,
        'close-btn': false,
        'remove-btn': false,
      },
    },
    {
      description: 'Dialog',
      props: {
        id: 'a582b47a42d',
        show: true,
      },
      data: {
        action: 'remove',
      },
      template: {
        'firewallRuleDelete-card': true,
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
      'firewallrules/remove': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(FirewallRuleDelete, {
            store: storeVuex(currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: { id: test.props.id, show: test.props.show },
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
