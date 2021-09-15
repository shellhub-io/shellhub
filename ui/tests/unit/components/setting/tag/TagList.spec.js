import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import TagList from '@/components/setting/tag/TagList';

describe('TagList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();

  localVue.use(Vuex);

  let wrapper;

  const status = true;
  const numberTags = 3;
  const tags = ['tag1', 'tag2', 'tag3'];
  const tagsObject = tags.map((str) => ({ name: str }));

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

  const storeWithoutData = new Vuex.Store({
    namespaced: true,
    state: {
      tags: [],
      numberTags: 0,
      status,
    },
    getters: {
      'tags/list': (state) => state.tags,
      'tags/getNumberTags': (state) => state.numberTags,
    },
    actions: {
      'tags/fetch': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  const storeWithData = new Vuex.Store({
    namespaced: true,
    state: {
      tags,
      numberTags,
      status,
    },
    getters: {
      'tags/list': (state) => state.tags,
      'tags/getNumberTags': (state) => state.numberTags,
    },
    actions: {
      'tags/fetch': () => {},
      'snackbar/showSnackbarErrorLoading': () => {},
    },
  });

  describe('Without Data', () => {
    beforeEach(() => {
      wrapper = mount(TagList, {
        store: storeWithoutData,
        localVue,
        stubs: ['fragment', 'router-link'],
        mocks: {
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
      expect(wrapper.vm.headers).toEqual(headers);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.getListTags).toEqual([]);
      expect(wrapper.vm.getNumberTags).toEqual(0);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="tagFormDialog-component"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="tagDelete-component"]').exists()).toBe(false);
    });
    it('Renders the template with data', () => {
      const dt = wrapper.find('[data-test="tagListList-dataTable"]');
      const dataTableProps = dt.vm.$options.propsData;

      expect(dataTableProps.items).toHaveLength(0);
    });
  });

  describe('With Data', () => {
    beforeEach(() => {
      wrapper = mount(TagList, {
        store: storeWithData,
        localVue,
        stubs: ['fragment', 'router-link'],
        mocks: {
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
      expect(wrapper.vm.headers).toEqual(headers);
    });
    it('Process data in the computed', () => {
      expect(wrapper.vm.getListTags).toEqual(tagsObject);
      expect(wrapper.vm.getNumberTags).toEqual(numberTags);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with components', () => {
      expect(wrapper.find('[data-test="tagFormDialog-component"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="tagDelete-component"]').exists()).toBe(true);
    });
    it('Renders the template with data', () => {
      const dt = wrapper.find('[data-test="tagListList-dataTable"]');
      const dataTableProps = dt.vm.$options.propsData;

      expect(dataTableProps.items).toHaveLength(numberTags);
    });
  });
});
