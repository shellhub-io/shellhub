import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import NamespaceAdd from '@/components/app_bar/namespace/NamespaceAdd';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import '@/vee-validate';

describe('NamespaceAdd', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  document.body.setAttribute('data-app', true);
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  let wrapper;

  const show = true;
  const firstNamespace = true;
  const invalidNamespaces = [
    '\'', '"', '!', '@', '#', '$', '%', '¨', '&', '*', '(', ')', '-', '_', '=', '+', '´', '`', '[',
    '{', '~', '^', ']', ',', '<', '..', '>', ';', ':', '/', '?',
  ];

  beforeEach(() => {
    wrapper = mount(NamespaceAdd, {
      localVue,
      stubs: ['fragment'],
      propsData: { show, firstNamespace },
      vuetify,
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Receive data in props', () => {
    expect(wrapper.vm.show).toEqual(show);
    expect(wrapper.vm.firstNamespace).toEqual(firstNamespace);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.dialog).toEqual(false);
    expect(wrapper.vm.namespaceName).toEqual('');
  });
  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="namespace-text"]').element.value).toEqual('');
  });
  it('Show empty fields required in validation', async () => {
    wrapper.setData({ namespaceName: '' });
    await flushPromises();

    const validator = wrapper.vm.$refs.providerNamespace;

    await validator.validate();
    expect(validator.errors[0]).toBe('This field is required');
  });
  it('Shows invalid namespace error for dot', async () => {
    wrapper.setData({ namespaceName: 'ShelHub.' });
    await flushPromises();

    const validator = wrapper.vm.$refs.providerNamespace;

    await validator.validate();
    expect(validator.errors[0]).toBe('The name must not contain dots');
  });
  invalidNamespaces.forEach((inamespace) => {
    it(`Shows invalid namespace error for ${inamespace}`, async () => {
      wrapper.setData({ namespaceName: inamespace });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerNamespace;

      await validator.validate();
      expect(validator.errors[0]).toBe('You entered an invalid RFC1123 name');
    });
  });
});
