import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import SessionDeleteRecord from '@/components/session/SessionDeleteRecord';

describe('SessionDeleteRecord', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const tests = [
    {
      description: 'Dialog closed',
      props: {
        uid: '8c354a00',
        show: false,
      },
      template: {
        'removeRecord-icon': true,
        'removeRecord-title': true,
        'sessionDeleteRecord-card': false,
      },
      templateText: {
        'removeRecord-title': 'Delete Session Record',
      },
    },
    {
      description: 'Dialog opened',
      props: {
        uid: '8c354a00',
        show: true,
      },
      template: {
        'removeRecord-icon': true,
        'removeRecord-title': true,
        'sessionDeleteRecord-card': true,
        'text-title': true,
        'text-text': true,
        'cancel-btn': true,
        'delete-btn': true,
      },
      templateText: {
        'removeRecord-title': 'Delete Session Record',
        'text-title': 'Are you sure?',
        'text-text': 'You are going to delete the logs recorded for this session.',
        'cancel-btn': 'Cancel',
        'delete-btn': 'Delete',
      },
    },
  ];

  const storeVuex = () => new Vuex.Store({
    namespaced: true,
    state: { },
    getters: { },
    actions: {
      'sessions/deleteSessionLogs': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(SessionDeleteRecord, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: { uid: test.props.uid, show: test.props.show },
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
