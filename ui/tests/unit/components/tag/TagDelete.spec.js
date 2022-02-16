import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TagDelete from '@/components/tag/TagDelete';

describe('TagDelete', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const tests = [
    {
      description: 'Dialog Closed',
      props: {
        tagName: 'tag',
        show: false,
      },
      template: {
        'remove-icon': true,
        'remove-title': true,
        'tagDelete-card': false,
      },
      templateText: {
        'remove-title': 'Remove',
      },
    },
    {
      description: 'Dialog opened',
      props: {
        tagName: 'tag',
        show: true,
      },
      template: {
        'remove-icon': true,
        'remove-title': true,
        'tagDelete-card': true,
        'text-title': true,
        'text-text': true,
        'close-btn': true,
        'remove-btn': true,
      },
      templateText: {
        'remove-title': 'Remove',
        'text-title': 'Are you sure?',
        'text-text': 'You are about to remove this tag.',
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
      'tags/remove': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        wrapper = mount(TagDelete, {
          store: storeVuex(),
          localVue,
          stubs: ['fragment'],
          propsData: {
            tagName: test.props.tagName,
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
