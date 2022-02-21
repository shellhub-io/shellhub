import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import PrivateKeyDelete from '@/components/private_key/PrivateKeyDelete';

describe('PrivateKeyDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const tests = [
    {
      description: 'Dialog closed',
      props: {
        fingerprint: 'b7:25:f8',
        show: false,
      },
      template: {
        'remove-icon': true,
        'remove-title': true,
        'privateKeyDelete-card': false,
      },
      templateText: {
        'remove-title': 'Remove',
      },
    },
    {
      description: 'Dialog opened',
      props: {
        fingerprint: 'b7:25:f8',
        show: true,
      },
      template: {
        'remove-icon': true,
        'remove-title': true,
        'privateKeyDelete-card': true,
        'text-title': true,
        'text-text': true,
        'close-btn': true,
        'remove-btn': true,
      },
      templateText: {
        'remove-title': 'Remove',
        'text-title': 'Are you sure?',
        'text-text': 'You are about to remove this private key.',
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
      'privatekeys/remove': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(PrivateKeyDelete, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: {
            fingerprint: test.props.fingerprint,
            show: test.props.show,
          },
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
