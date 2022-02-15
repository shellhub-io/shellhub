import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import FirewallRuleDelete from '@/components/firewall_rule/FirewallRuleDelete';

describe('FirewallRuleDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const tests = [
    {
      description: 'Dialog closed',
      props: {
        id: 'a582b47a42d',
        show: false,
      },
      data: {
        dialog: false,
      },
      template: {
        'remove-icon': true,
        'remove-title': true,
        'firewallRuleDelete-card': false,
      },
      templateText: {
        'remove-title': 'Remove',
      },
    },
    {
      description: 'Dialog opened',
      props: {
        id: 'a582b47a42d',
        show: true,
      },
      data: {
        dialog: false,
      },
      template: {
        'remove-icon': true,
        'remove-title': true,
        'firewallRuleDelete-card': true,
        'text-title': true,
        'text-text': true,
        'close-btn': true,
        'remove-btn': true,
      },
      templateText: {
        'remove-title': 'Remove',
        'text-title': 'Are you sure?',
        'text-text': 'You are about to remove this firewall rule.',
        'close-btn': 'Close',
        'remove-btn': 'Remove',
      },
    },
  ];

  const storeVuex = () => new Vuex.Store({
    namespaced: true,
    state: { },
    getters: { },
    actions: {
      'firewallrules/remove': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(FirewallRuleDelete, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: { id: test.props.id, show: test.props.show },
          vuetify,
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
