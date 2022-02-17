import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import DeviceDelete from '@/components/device/DeviceDelete';

describe('DeviceDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const tests = [
    {
      description: 'Dialog closed',
      props: {
        uid: 'a582b47a42d',
        redirect: false,
        show: false,
      },
      template: {
        'remove-icon': true,
        'remove-title': true,
        'deviceDelete-card': false,
      },
      templateText: {
        'remove-title': 'Remove',
      },
    },
    {
      description: 'Dialog opened without redirect',
      props: {
        uid: 'a582b47a42d',
        redirect: false,
        show: true,
      },
      template: {
        'remove-icon': true,
        'remove-title': true,
        'deviceDelete-card': true,
        'text-title': true,
        'text-text': true,
        'close-btn': true,
        'remove-btn': true,
      },
      templateText: {
        'remove-title': 'Remove',
        'text-title': 'Are you sure?',
        'text-text': 'You are about to remove this device.',
        'close-btn': 'Close',
        'remove-btn': 'Remove',
      },
    },
    {
      description: 'Dialog opened with redirect',
      props: {
        uid: 'a582b47a42d',
        redirect: true,
        show: true,
      },
      template: {
        'remove-icon': true,
        'remove-title': true,
        'deviceDelete-card': true,
        'text-title': true,
        'text-text': true,
        'close-btn': true,
        'remove-btn': true,
      },
      templateText: {
        'remove-title': 'Remove',
        'text-title': 'Are you sure?',
        'text-text': 'You are about to remove this device.',
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
      'devices/remove': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(DeviceDelete, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: {
            uid: test.props.uid,
            redirect: test.props.redirect,
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
