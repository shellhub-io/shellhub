import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TagList from '@/components/tag/TagList';
import { actions, authorizer } from '../../../../src/authorizer';

describe('TagList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();

  localVue.use(Vuex);

  let wrapper;

  const numberTagsGlobal = 3;
  const tagsGlobal = ['tag1', 'tag2', 'tag3'];
  const tagsObject = tagsGlobal.map((str) => ({ name: str }));

  const headers = [
    {
      text: 'Name',
      value: 'name',
      align: 'center',
    },
    {
      text: 'Actions',
      value: 'actions',
      align: 'center',
      sortable: false,
    },
  ];

  const tests = [
    {
      description: 'List data when user has owner role',
      role: {
        type: 'owner',
        permission: true,
      },
      variables: {
        tagsObject,
        numberTagsGlobal,
      },
      data: {
        tagDialogShow: [false, false, false],
        tagDeleteShow: [false, false, false],
        removeAction: 'remove',
        headers,
      },
      computed: {
        getListTags: tagsObject,
        getNumberTags: numberTagsGlobal,
        hasAuthorizationRemove: true,
      },
    },
    {
      description: 'List data when user has operator role',
      role: {
        type: 'operator',
        permission: false,
      },
      variables: {
        tagsObject,
        numberTagsGlobal,
      },
      data: {
        tagDialogShow: [false, false, false],
        tagDeleteShow: [false, false, false],
        removeAction: 'remove',
        headers,
      },
      computed: {
        getListTags: tagsObject,
        getNumberTags: numberTagsGlobal,
        hasAuthorizationRemove: false,
      },
    },
  ];

  const storeVuex = (tags, numberTags, hasAuthorizationRemove) => new Vuex.Store({
    namespaced: true,
    state: {
      tags,
      numberTags,
      hasAuthorizationRemove,
    },
    getters: {
      'tags/list': (state) => state.tags,
      'tags/getNumberTags': (state) => state.numberTags,
      'tags/hasAuthorizationRemove': (state) => state.hasAuthorizationRemove,
    },
    actions: {
      'tags/fetch': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(TagList, {
          store: storeVuex(
            test.variables.tagsObject,
            test.variables.numberTagsGlobal,
            test.role.type,
          ),
          localVue,
          stubs: ['fragment', 'router-link'],
          mocks: {
            $authorizer: authorizer,
            $actions: actions,
            $env: (isEnterprise) => isEnterprise,
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
      // Data and Props checking
      //////

      it('Compare data with default value', () => {
        Object.keys(test.data).forEach((item) => {
          expect(wrapper.vm[item]).toEqual(test.data[item]);
        });
      });

      //////
      // HTML validation
      //////

      it('Renders the template with data', () => {
        const dt = wrapper.find('[data-test="tagListList-dataTable"]');
        const dataTableProps = dt.vm.$options.propsData;

        expect(dataTableProps.items).toHaveLength(test.variables.numberTagsGlobal);
      });
    });
  });
});
