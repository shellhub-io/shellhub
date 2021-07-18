import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import TokenList from '@/components/setting/token/TokenList';

describe('TokenList', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const numberTokens = 2;
  const tokens = [
    {
      id: 'a582b47a42d',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      read_only: true,
    },
    {
      id: 'a582b47a42e',
      tenant_id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
      read_only: false,
    },
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      tokens,
      numberTokens,
    },
    getters: {
      'tokens/list': (state) => state.tokens,
      'tokens/getNumberTokens': (state) => state.numberTokens,
    },
    actions: {
      'tokens/fetch': () => {
      },
      'snackbar/showSnackbarErrorLoading': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(TokenList, {
      store,
      localVue,
      propsData: { show: true },
      stubs: ['fragment'],
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getListTokens).toEqual(tokens);
    expect(wrapper.vm.getNumberTokens).toEqual(numberTokens);
  });
  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;
    expect(dataTableProps.items).toHaveLength(numberTokens);
  });
});
