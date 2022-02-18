import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import SessionClose from '@/components/session/SessionClose';

describe('SessionClose', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const session = {
    uid: '8c354a00',
    device_uid: 'a582b47a',
  };

  const tests = [
    {
      description: 'Dialog closed',
      variables: {
        session,
      },
      props: {
        uid: session.uid,
        device: session.device_uid,
        show: false,
      },
      data: {
        session,
      },
      template: {
        'close-icon': true,
        'close-title': true,
        'sessionClose-card': false,
      },
      templateText: {
        'close-title': 'Close',
      },
    },
    {
      description: 'Dialog opened',
      variables: {
        session,
      },
      props: {
        uid: session.uid,
        device: session.device_uid,
        show: true,
      },
      data: {
        session,
      },
      template: {
        'close-icon': true,
        'close-title': true,
        'sessionClose-card': true,
        'text-title': true,
        'text-text': true,
        'cancel-btn': true,
        'close-btn': true,
      },
      templateText: {
        'close-title': 'Close',
        'text-title': 'Are you sure?',
        'text-text': 'You are going to close connection for this device.',
        'cancel-btn': 'Cancel',
        'close-btn': 'Close',
      },
    },
  ];

  const storeVuex = () => new Vuex.Store({
    namespaced: true,
    state: { },
    getters: { },
    actions: {
      'sessions/close': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(SessionClose, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: {
            uid: test.props.uid,
            device: test.props.device,
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
