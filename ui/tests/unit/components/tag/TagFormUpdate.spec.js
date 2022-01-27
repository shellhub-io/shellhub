import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TagFormUpdate from '@/components/tag/TagFormUpdate';
import { actions, authorizer } from '../../../../src/authorizer';

describe('TagFormUpdate', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const role = ['owner', 'observer'];

  const hasAuthorization = {
    owner: true,
    observer: false,
  };

  const tests = [
    {
      description: 'Icon add tag',
      props: {
        deviceUid: '',
        tagsList: [],
        show: false,
      },
      data: {
        dialog: false,
        listTagLocal: [],
        errorMsg: '',
        action: 'deviceUpdate',
      },
      computed: {
        hasTag: false,
      },
      template: {
        'title-item': true,
        'tag-icon': true,
        'tagForm-card': false,
        'deviceTag-combobox': false,
        'close-btn': false,
        'save-btn': false,
      },
      templateFields: {
        'title-item': 'Add tags',
      },
    },
    {
      description: 'Icon edit tag',
      props: {
        deviceUid: 'xxxxxxxx',
        tagsList: ['tag1'],
        show: false,
      },
      data: {
        dialog: false,
        listTagLocal: ['tag1'],
        errorMsg: '',
        action: 'deviceUpdate',
      },
      computed: {
        hasTag: true,
      },
      template: {
        'title-item': true,
        'tag-icon': true,
        'tagForm-card': false,
        'deviceTag-combobox': false,
        'close-btn': false,
        'save-btn': false,
      },
      templateFields: {
        'title-item': 'Edit tags',
      },
    },
    {
      description: 'Dialog',
      props: {
        deviceUid: 'xxxxxxxx',
        tagsList: ['tag1'],
        show: true,
      },
      data: {
        dialog: false,
        listTagLocal: ['tag1'],
        errorMsg: '',
        action: 'deviceUpdate',
      },
      computed: {
        hasTag: true,
      },
      template: {
        'title-item': true,
        'tag-icon': true,
        'tagForm-card': true,
        'deviceTag-combobox': true,
        'close-btn': true,
        'save-btn': true,
      },
      templateObserver: {
        'title-item': true,
        'tag-icon': true,
        'tagForm-card': false,
        'deviceTag-combobox': false,
        'close-btn': false,
        'save-btn': false,
      },
      templateFields: {
        'title-item': 'Edit tags',
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
      'devices/updateDeviceTag': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(async () => {
          wrapper = mount(TagFormUpdate, {
            store: storeVuex(currentrole),
            localVue,
            stubs: ['fragment'],
            propsData: {
              deviceUid: test.props.deviceUid,
              tagsList: test.props.tagsList,
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
          Object.keys(test.computed).forEach((item) => {
            expect(wrapper.vm[item]).toEqual(test.computed[item]);
          });
          expect(wrapper.vm.hasAuthorization).toEqual(hasAuthorization[currentrole]);
        });

        //////
        // HTML validation
        //////

        it('Renders the template with data', () => {
          Object.keys(test.template).forEach((item) => {
            if (!hasAuthorization[currentrole] && test.props.show) {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.templateObserver[item]);
            } else {
              expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
            }
          });
          Object.keys(test.templateFields).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).text()).toEqual(test.templateFields[item]);
          });
        });
      });
    });
  });
});
