import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import SettingProfile from '@/components/setting/SettingProfile';
import { ValidationProvider, ValidationObserver, extend } from 'vee-validate';
import flushPromises from 'flush-promises';
import { required, email } from 'vee-validate/dist/rules';
import Vuetify from 'vuetify';
import '@/vee-validate';

extend('required', {
  ...required,
  message: 'This field is required',
});

extend('email', {
  ...email,
  message: 'This field must be a valid email',
});

describe('SettingProfile', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);
  const vuetify = new Vuetify();
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  let wrapper;

  const user = 'ShellHub';
  const emailUser = 'shellhub@shellhub.com';
  const tenant = '';

  // vee-validate variables bellow
  const invalidEmails = ['notemail', 'missing@dot', 'with.only.dots', 'r4ndomCH@r5'];
  const validEmails = ['new@email.com', 'another@email.org'];
  const invalidPasswords = ['aPasswordBiggerThanExpectedBecauseHasMoreThan30chars', 'shor'];
  const validPasswords = ['newPassword', 'password123'];
  const confirmPasswordsMatchError = [{ new: 'newpass', confirmNew: 'newpas' }, { new: 'Newpass', confirmNew: 'newpass' }];
  const confirmPasswordsMatchSuccess = [{ new: 'newpass', confirmNew: 'newpass' }, { new: 'changedpassword', confirmNew: 'changedpassword' }];
  const compareOldNewError = [{ old: 'oldpass', new: 'oldpass' }, { old: 'currentPass', new: 'currentPass' }];
  const compareOldNewSuccess = [{ old: 'oldpass', new: 'newpass' }, { old: 'currentPass', new: 'newPassword' }];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      user,
      email: emailUser,
      tenant,
    },
    getters: {
      'auth/currentUser': (state) => state.user,
      'auth/email': (state) => state.email,
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'users/put': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = mount(SettingProfile, {
      store,
      localVue,
      stubs: ['fragment'],
      vuetify,
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.tenant).toEqual(tenant);
  });
  it('Compare data with default value', () => {
    expect(wrapper.vm.username).toEqual(user);
    expect(wrapper.vm.email).toEqual(emailUser);
    expect(wrapper.vm.currentPassword).toEqual('');
    expect(wrapper.vm.newPassword).toEqual('');
    expect(wrapper.vm.newPasswordConfirm).toEqual('');
    expect(wrapper.vm.editDataStatus).toEqual(false);
    expect(wrapper.vm.editPasswordStatus).toEqual(false);
    expect(wrapper.vm.show).toEqual(false);
  });
  it('Renders the template with data', async () => {
    expect(wrapper.find('[data-test="username-text"]').element.value).toEqual(user);
    expect(wrapper.find('[data-test="email-text"]').element.value).toEqual(emailUser);
    expect(wrapper.find('[data-test="password-text"]').element.value).toEqual('');
    expect(wrapper.find('[data-test="newPassword-text"]').element.value).toEqual('');
    expect(wrapper.find('[data-test="confirmNewPassword-text"]').element.value).toEqual('');
  });

  it('Show empty fields required in validation', async () => {
    wrapper.setData({ username: '', email: '', currentPassword: '' });
    await flushPromises();
    const validatorUser = wrapper.vm.$refs.providerName;
    const validatorEmail = wrapper.vm.$refs.providerEmail;
    const validatorCurrentPass = wrapper.vm.$refs.providerCurrentPassword;
    const validatorNewPass = wrapper.vm.$refs.providerNewPassword;
    const validatorConfirmPass = wrapper.vm.$refs.providerConfirmPassword;

    await validatorUser.validate();
    expect(validatorUser.errors[0]).toBe('This field is required');
    await validatorEmail.validate();
    expect(validatorEmail.errors[0]).toBe('This field is required');
    await validatorCurrentPass.validate();
    expect(validatorCurrentPass.errors[0]).toBe('This field is required');
    await validatorNewPass.validate();
    expect(validatorNewPass.errors[0]).toBe('This field is required');
    await validatorConfirmPass.validate();
    expect(validatorConfirmPass.errors[0]).toBe('This field is required');
  });

  invalidEmails.forEach((iemail) => {
    it(`Shows invalid email error for ${iemail}`, async () => {
      wrapper.setData({ email: iemail });
      await flushPromises();
      const validator = wrapper.vm.$refs.providerEmail;
      await validator.validate();
      expect(validator.errors[0]).toBe('This field must be a valid email');
    });
  });

  validEmails.forEach((vemail) => {
    it(`Valid email for ${vemail}`, async () => {
      wrapper.setData({ email: vemail });
      await flushPromises();
      const validator = wrapper.vm.$refs.providerEmail;
      await validator.validate();
      expect(validator.errors).toHaveLength(0);
    });
  });

  invalidPasswords.forEach((ipass) => {
    it(`Shows invalid password length for ${ipass}`, async () => {
      wrapper.setData({ newPassword: ipass });
      await flushPromises();
      const validator = wrapper.vm.$refs.providerNewPassword;
      await validator.validate();
      expect(validator.errors[0]).toBe('Your password should be 5-30 characters long');
    });
  });

  validPasswords.forEach((vpass) => {
    it(`Valid password for ${vpass}`, async () => {
      wrapper.setData({ newPassword: vpass });
      await flushPromises();
      const validator = wrapper.vm.$refs.providerNewPassword;
      await validator.validate();
      expect(validator.errors).toHaveLength(0);
    });
  });

  confirmPasswordsMatchError.forEach((item) => {
    it(`Shows invalid password match for ${item.new} and ${item.confirmNew}`, async () => {
      wrapper.setData({ newPassword: item.new, newPasswordConfirm: item.confirmNew });
      await flushPromises();
      const validator = wrapper.vm.$refs.providerConfirmPassword;
      await validator.validate();
      expect(validator.errors[0]).toBe('The passwords do not match');
    });
  });

  confirmPasswordsMatchSuccess.forEach((item) => {
    it(`Valid password match for ${item.new} and ${item.confirmNew}`, async () => {
      wrapper.setData({ newPassword: item.new, newPasswordConfirm: item.confirmNew });
      await flushPromises();
      const validator = wrapper.vm.$refs.providerConfirmPassword;
      await validator.validate();
      expect(validator.errors).toHaveLength(0);
    });
  });

  compareOldNewError.forEach((item) => {
    it(`Shows error for switching to the same password ${item.old}`, async () => {
      wrapper.setData({ currentPassword: item.old, newPassword: item.new });
      await flushPromises();
      const validator = wrapper.vm.$refs.providerNewPassword;
      await validator.validate();
      expect(validator.errors[0]).toBe('The passwords are the same');
    });
  });

  compareOldNewSuccess.forEach((item) => {
    it(`Valid password change from ${item.old} to ${item.new}`, async () => {
      wrapper.setData({ currentPassword: item.old, newPassword: item.new });
      await flushPromises();
      const validator = wrapper.vm.$refs.providerNewPassword;
      await validator.validate();
      expect(validator.errors).toHaveLength(0);
    });
  });
});
