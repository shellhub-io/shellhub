import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TagFormUpdate from '@/components/tag/TagFormUpdate';

describe('TagFormUpdate', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const tests = [
    {
      description: 'Dialog closed with add',
      props: {
        deviceUid: '',
        tagsList: [],
        show: false,
      },
      data: {
        dialog: false,
        listTagLocal: [],
        errorMsg: '',
      },
      computed: {
        hasTag: false,
      },
      template: {
        'edit-icon': true,
        'edit-title': true,
        'tagForm-card': false,
      },
      templateText: {
        'edit-title': 'Add tags',
      },
    },
    {
      description: 'Dialog closed with edit',
      props: {
        deviceUid: 'xxxxxxx',
        tagsList: ['ShellHub'],
        show: false,
      },
      data: {
        dialog: false,
        listTagLocal: ['ShellHub'],
        errorMsg: '',
      },
      computed: {
        hasTag: true,
      },
      template: {
        'edit-icon': true,
        'edit-title': true,
        'tagForm-card': false,
      },
      templateText: {
        'edit-title': 'Edit tags',
      },
    },
    {
      description: 'Dialog opened',
      props: {
        deviceUid: 'xxxxxxx',
        tagsList: ['ShellHub'],
        show: true,
      },
      data: {
        dialog: false,
        listTagLocal: ['ShellHub'],
        errorMsg: '',
      },
      computed: {
        hasTag: true,
      },
      template: {
        'edit-icon': true,
        'edit-title': true,
        'tagForm-card': true,
      },
      templateText: {
        'edit-title': 'Edit tags',
      },
    },
  ];

  const storeVuex = () => new Vuex.Store({
    namespaced: true,
    state: { },
    getters: { },
    actions: {
      'devices/updateDeviceTag': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        wrapper = mount(TagFormUpdate, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: {
            deviceUid: test.props.deviceUid,
            tagsList: test.props.tagsList,
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
      it('Process data in the computed', () => {
        Object.keys(test.computed).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.computed[item]);
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
