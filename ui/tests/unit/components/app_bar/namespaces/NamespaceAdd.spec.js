import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import NamespaceAdd from '@/components/app_bar/namespace/NamespaceAdd';

describe('NamespaceAdd', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;

  const show = true;

  beforeEach(() => {
    wrapper = shallowMount(NamespaceAdd, {
      localVue,
      stubs: ['fragment'],
      propsData: { show },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.show).toEqual(true);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
    expect(wrapper.vm.namespaceName).toEqual('');
  });
});
