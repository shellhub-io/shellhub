import Vuex from 'vuex';
import { mount, shallowMount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TagSelector from '@/components/setting/tag/TagSelector';

describe('TagSelector', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const tagsGlobal = ['ShellHub', 'Shell', 'Hub'];

  const tests = [
    {
      description: 'Without tags',
      variables: {
        tags: [],
        selectedTags: [],
      },
      computed: {
        getListTags: [],
      },
      template: {
        'tags-btn': true,
      },
    },
    {
      description: 'With tags',
      variables: {
        tags: tagsGlobal,
        selectedTags: [],
      },
      computed: {
        getListTags: tagsGlobal,
      },
    },
  ];

  const storeVuex = (tags, selectedTags) => new Vuex.Store({
    namespaced: true,
    state: {
      tags,
      selectedTags,
    },
    getters: {
      'tags/list': (state) => state.tags,
      'tags/selected': (state) => state.selectedTags,
    },
    actions: {
      'tags/setSelected': () => {},
      'tags/fetch': () => {},
      'devices/setFilter': () => {},
      'devices/refresh': () => {},
      'snackbar/showSnackbarErrorAssociation': () => {},
      'snackbar/showSnackbarErrorDefault': () => {},
    },
  });

  tests.forEach((test, index) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        if (index === 0) {
          wrapper = mount(TagSelector, {
            store: storeVuex(
              test.variables.tags,
              test.variables.selectedTags,
            ),
            localVue,
            stubs: ['fragment'],
            vuetify,
          });
        } else {
          wrapper = shallowMount(TagSelector, {
            store: storeVuex(test.variables.tags),
            localVue,
            stubs: ['fragment'],
            vuetify,
          });
        }
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

      it('Process data in the computed', () => {
        Object.keys(test.computed).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.computed[item]);
        });
      });

      //////
      // HTML validation
      //////

      it('Renders the template with data', () => {
        if (index === 0) {
          Object.keys(test.template).forEach((item) => {
            expect(wrapper.find(`[data-test="${item}"]`).exists()).toBe(test.template[item]);
          });
        } else {
          Object.keys(test.variables.tags).forEach((tagIndex) => {
            expect(wrapper.find(`[data-test="${test.variables.tags[tagIndex]}-item"]`).exists()).toBe(true);
          });
        }
      });
    });
  });
});
