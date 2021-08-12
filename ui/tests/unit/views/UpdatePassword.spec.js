import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import UpdatePassword from '@/views/UpdatePassword';
import router from '@/router/index';
import '@/vee-validate';

describe('UpdatePassword', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  let wrapper;

  // vee-validate variables bellow
  const invalidPasswords = ['aPasswordBiggerThanExpectedBecauseHasMoreThan30chars', 'shor'];
  const validPasswords = ['newPassword', 'password123'];
  const confirmPasswordsMatchError = [{ new: 'newpass', confirmNew: 'newpas' }, { new: 'Newpass', confirmNew: 'newpass' }];
  const confirmPasswordsMatchSuccess = [{ new: 'newpass', confirmNew: 'newpass' }, { new: 'changedpassword', confirmNew: 'changedpassword' }];

  const store = new Vuex.Store({
    state: {
    },
    getters: {
    },
    actions: {
      'users/updatePassword': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  beforeEach(() => {
    wrapper = mount(UpdatePassword, {
      store,
      stubs: ['fragment'],
      localVue,
      router,
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
    expect(wrapper.vm.newPassword).toEqual('');
    expect(wrapper.vm.newPasswordConfirm).toEqual('');
    expect(wrapper.vm.data).toEqual({});
  });

  //////
  // HTML validation
  //////

  //////
  // In this case, the empty fields are validated.
  //////

  it('Show validation messages', async () => {
    wrapper.setData({ newPassword: '', newPasswordConfirm: '' });
    await flushPromises();

    const validatorNewPass = wrapper.vm.$refs.providerNewPassword;
    const validatorConfirmPass = wrapper.vm.$refs.providerConfirmPassword;

    await validatorNewPass.validate();
    await validatorConfirmPass.validate();

    expect(validatorNewPass.errors[0]).toBe('This field is required');
    expect(validatorConfirmPass.errors[0]).toBe('This field is required');
  });

  //////
  // In this case, invalid password length are validated.
  //////

  it('Show validation messages', async (done) => {
    invalidPasswords.forEach(async (ipass) => {
      wrapper.setData({ newPassword: ipass });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerNewPassword;

      await validator.validate();
      expect(validator.errors[0]).toBe('Your password should be 5-30 characters long');

      await flushPromises();
      done();
    });
  });

  //////
  // In this case, invalid password match are validated.
  //////

  it('Show validation messages', async (done) => {
    confirmPasswordsMatchError.forEach(async (item) => {
      wrapper.setData({ newPassword: item.new, newPasswordConfirm: item.confirmNew });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerConfirmPassword;

      await validator.validate();
      expect(validator.errors[0]).toBe('The passwords do not match');

      await flushPromises();
      done();
    });
  });

  //////
  // In this case, valid password length are validated.
  //////

  it('Show validation messages', async (done) => {
    validPasswords.forEach(async (vpass) => {
      wrapper.setData({ newPassword: vpass });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerNewPassword;

      await validator.validate();
      expect(validator.errors).toHaveLength(0);

      await flushPromises();
      done();
    });
  });

  //////
  // In this case, valid password match are validated.
  //////

  it('Show validation messages', async (done) => {
    confirmPasswordsMatchSuccess.forEach(async (item) => {
      wrapper.setData({ newPassword: item.new, newPasswordConfirm: item.confirmNew });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerConfirmPassword;

      await validator.validate();
      expect(validator.errors).toHaveLength(0);

      await flushPromises();
      done();
    });
  });
});
