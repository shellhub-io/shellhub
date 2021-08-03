import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import SettingProfile from '@/components/setting/SettingProfile';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import '@/vee-validate';

describe('SettingProfile', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  let wrapper;

  const username = 'ShellHub';
  const emailUser = 'shellhub@shellhub.com';
  const tenant = 'xxxxxxxx';

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
      username,
      email: emailUser,
      tenant,
    },
    getters: {
      'auth/currentUser': (state) => state.username,
      'auth/email': (state) => state.email,
      'auth/tenant': (state) => state.tenant,
    },
    actions: {
      'users/put': () => {},
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
    expect(wrapper.vm.username).toEqual(username);
    expect(wrapper.vm.email).toEqual(emailUser);
    expect(wrapper.vm.currentPassword).toEqual('');
    expect(wrapper.vm.newPassword).toEqual('');
    expect(wrapper.vm.newPasswordConfirm).toEqual('');
    expect(wrapper.vm.editDataStatus).toEqual(false);
    expect(wrapper.vm.editPasswordStatus).toEqual(false);
    expect(wrapper.vm.show).toEqual(false);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.tenant).toEqual(tenant);
  });

  //////
  // HTML validation
  //////

  //////
  // In this case, the empty fields are validated.
  //////

  it('Show validation messages', async () => {
    wrapper.setData({ username: '', email: '', currentPassword: '' });
    await flushPromises();

    const validatorUser = wrapper.vm.$refs.providerName;
    const validatorEmail = wrapper.vm.$refs.providerEmail;
    const validatorCurrentPass = wrapper.vm.$refs.providerCurrentPassword;
    const validatorNewPass = wrapper.vm.$refs.providerNewPassword;
    const validatorConfirmPass = wrapper.vm.$refs.providerConfirmPassword;

    await validatorUser.validate();
    await validatorEmail.validate();
    await validatorCurrentPass.validate();
    await validatorNewPass.validate();
    await validatorConfirmPass.validate();

    expect(validatorUser.errors[0]).toBe('This field is required');
    expect(validatorEmail.errors[0]).toBe('This field is required');
    expect(validatorCurrentPass.errors[0]).toBe('This field is required');
    expect(validatorNewPass.errors[0]).toBe('This field is required');
    expect(validatorConfirmPass.errors[0]).toBe('This field is required');
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
  // In this case, error for switching to the same password are validated.
  //////

  it('Show validation messages', async (done) => {
    compareOldNewError.forEach(async (item) => {
      wrapper.setData({ currentPassword: item.old, newPassword: item.new });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerNewPassword;

      await validator.validate();
      expect(validator.errors[0]).toBe('The passwords are the same');

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

  //////
  // In this case, valid password change.
  //////

  it('Show validation messages', async (done) => {
    compareOldNewSuccess.forEach(async (item) => {
      wrapper.setData({ currentPassword: item.old, newPassword: item.new });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerNewPassword;

      await validator.validate();
      expect(validator.errors).toHaveLength(0);

      await flushPromises();
      done();
    });
  });
  it('Renders the template with data', async () => {
    expect(wrapper.find('[data-test="username-text"]').element.value).toEqual(username);
    expect(wrapper.find('[data-test="email-text"]').element.value).toEqual(emailUser);
    expect(wrapper.find('[data-test="password-text"]').element.value).toEqual('');
    expect(wrapper.find('[data-test="newPassword-text"]').element.value).toEqual('');
    expect(wrapper.find('[data-test="confirmNewPassword-text"]').element.value).toEqual('');
  });
});
