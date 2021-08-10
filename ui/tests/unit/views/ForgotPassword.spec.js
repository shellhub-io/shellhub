import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import Vuetify from 'vuetify';
import flushPromises from 'flush-promises';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import router from '@/router/index';
import ForgotPassword from '@/views/ForgotPassword';
import '@/vee-validate';

describe('ForgotPassword', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();

  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  let wrapper;

  // vee-validate variables bellow
  const invalidEmails = ['notemail', 'missing@dot', 'with.only.dots', 'r4ndomCH@r5'];
  const validEmails = ['new@email.com', 'another@email.org'];

  const store = new Vuex.Store({
    state: {
    },
    getters: {
    },
    actions: {
      'users/recoverPassword': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  beforeEach(() => {
    wrapper = mount(ForgotPassword, {
      store,
      stubs: ['fragment'],
      localVue,
      vuetify,
      router,
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
    expect(wrapper.vm.email).toEqual('');
  });

  //////
  // HTML validation
  //////

  //////
  // In this case, the empty fields are validated.
  //////

  it('Show validation messages', async () => {
    const validatorEmail = wrapper.vm.$refs.providerEmail;

    await validatorEmail.validate();
    expect(validatorEmail.errors[0]).toBe('This field is required');
  });

  //////
  // In this case, invalid email error are validated.
  //////

  it('Show validation messages', async (done) => {
    await invalidEmails.forEach(async (iemail) => {
      wrapper.setData({ email: iemail });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerEmail;

      await validator.validate();
      expect(validator.errors[0]).toBe('This field must be a valid email');

      await flushPromises();
      done();
    });
  });

  //////
  // In this case, valid email are validated.
  //////

  it('Show validation messages', async (done) => {
    validEmails.forEach(async (vemail) => {
      wrapper.setData({ email: vemail });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerEmail;

      await validator.validate();
      expect(validator.errors).toHaveLength(0);

      await flushPromises();
      done();
    });
  });

  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="email-text"]').element.value).toEqual('');
  });
});
