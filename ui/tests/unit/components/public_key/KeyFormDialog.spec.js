import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import KeyFormDialog from '@/components/public_key/KeyFormDialog';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import '@/vee-validate';

describe('KeyFormDialog', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const isOwner = true;
  const supportedKeys = 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.';

  const creatingKey = {
    data: '',
    name: '',
  };

  const keyObject = {
    data: 'AbGVvbmF',
    fingerprint: '00:00:00',
    created_at: '2020-11-23T20:59:13.323Z',
    tenant_id: 'xxxxxxxx',
    name: 'shellhub',
  };

  // The keys were generated from the website:
  // https://www.devglan.com/online-tools/rsa-encryption-decryption

  const publicKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCxXq0NZ
  tbRBJlyyW5LOTMuqsZj3pL+Q5UCSQxnEjMpnz6yL6ALTS+fpVLzHIQwfZ3p5kMGk
  vAwXOwLuvkFpvQvbGsj7/kBov6zDeL7exdzPVvhVclsIU//aTm2ryT1898RFgEOm
  2YDSsNteG4dYBe9SbNJIbezAg7lCAdKxsbZD05phX8NewGOcFolPk8kSuYqJ6lWB
  /WWncLT8eXgP8Ew95rwug9Am3ApijuoD1j1RIb1LirF9xkNNg13DA9QYEFOO16XV
  EIxIS1frW7Krh+3LP2W6Q5ISFRzGF7hxlWs9RRzB/SG2WxrOpeQAoDOLrt/fu3g7
  sVL9pA32YbLgyAT`;

  const privateKey = `-----BEGIN RSA PRIVATE KEY-----
  MIIEowIBAAKCAQEAsV6tDWbW0QSZcsluSzkzLqrGY96S/kOVAkkMZxIzKZ8+si+g
  C00vn6VS8xyEMH2d6eZDBpLwMFzsC7r5Bab0L2xrI+/5AaL+sw3i+3sXcz1b4VXJ
  bCFP/2k5tq8k9fPfERYBDptmA0rDbXhuHWAXvUmzSSG3swIO5QgHSsbG2Q9OaYV/
  DXsBjnBaJT5PJErmKiepVgf1lp3C0/Hl4D/BMPea8LoPQJtwKYo7qA9Y9USG9S4q
  xfcZDTYNdwwPUGBBTjtel1RCMSEtX61uyq4ftyz9lukOSEhUcxhe4cZVrPUUcwf0
  htlsazqXkAKAzi67f37t4O7FS/aQN9mGy4MgEwIDAQABAoIBACm+XnwI+AW5T2P0
  hECv9ZvGFWrrtyygOzGOf5zCa8gf8mF9U+1U/SBViHAvBe1jowapapzheCXpuKQK
  HRF3eYCvw4dxsujvs4Hwgrss/RfiGP2vcvg/3tP9r9eO4OQBwT4EL7uAV0HvFo9t
  CH1hYDTsY4WSqek3UsoUWaL/pUzwKMijUgh2Dzj5o9AlNGWANu6txI1mIgHmwUvj
  2kV7E4R1mGynSprdsW68V36viB/V9d82XGxd3tYhKojiS1Dir68mR2U8ld3728Pd
  xU7o9x6NcWOtpTY1nS9MpufaYUTlp/chOXSd2RIY6JmtgbJcVTdE4rasfIAEnlAZ
  XALqKAECgYEA4kl6ZfcwKtxebVyczMCk6QWOJtsJ6CT17w2oehQGSuSLXeidjkFe
  bm04hUcN4Rm5iipUwDlA6JT8QoUgSG7Mjf8aDLv68FjXHxHjVvQaj0pg2I+1qADZ
  bN6m5xaazqAShF5MN4zQQTnNHTp6AIXOSQhIpqKS/Bjf3FYw48pxCyMCgYEAyKjf
  GnwiFJZN/q3s2mCmlEPblJ5mbXGCmIK/wjcoDST3+YrFi5VoWsHu0hRoZHtxIiaH
  sjSj8f8hWaZJ+yTL/V6zAO93JMovmoYyClmGt5pl56pFT2B7VGDC5FU9bylzWF3g
  HDdCTXOE72c5cOHnOddxVSBdD6GLC7Qe4CUVnlECgYBYNmSskywHyVhWMaA+gWrI
  HA5KP2EhSidFRYHD9UJut6FMvn2NExaI3bMG4agbdDfMEKxxMuCGym18UQFAu1Cq
  miPBixZL05Yo2oRRRV+FNG2EfqFGGO6pbjKKK1m16tjNGSWFEjOs+adoGX+t7Ht6
  JOyNaRr7g4bhEgiFBEoFGQKBgF3XI+dl8CZCmJ0nR6JlGuIxzen2Hh7Gu/WJCBbS
  5qcnB9UrAfGiYNg44/BZXOzJEgKPlFxR4+4Ti8w6SVTrQ37tn7crRkPtTk/svFA8
  yBTrXwb1iU5y55pxWhOgjYeEEg5ccKehbB9+i8fONX3GF/Xj/Ht8FClwOe+yP9JB
  ZZfRAoGBAJb08mFdb0Csbp+ed3LFznWINpXf2vlRKqIf+w8VOsEItbiB0r08AVdA
  Tik8VkRWm9ZHnMeMRRg2sEsI8gfaEXwSfLfMi10fn9YuWC2GSt5z+lA52H/S1zU2
  sGHPNn1H/cu7eM+nr9NxzJIT2CzKMHt5w4epp/UgkYFri4n2wDNS
  -----END RSA PRIVATE KEY-----`;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
      'publickeys/post': () => {},
      'publickeys/put': () => {},
      'privatekeys/set': () => {},
      'privatekeys/edit': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
      'snackbar/showSnackbarSuccessNotRequest': () => {},
      'snackbar/showSnackbarErrorNotRequest': () => {},
    },
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is icon rendering. Creating public key.
  ///////

  describe('Icon', () => {
    const createKey = true;
    const action = 'public';

    beforeEach(() => {
      wrapper = mount(KeyFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { keyObject, createKey, action },
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

    it('Receive data in props', () => {
      expect(wrapper.vm.keyObject).toEqual(keyObject);
      expect(wrapper.vm.createKey).toEqual(createKey);
      expect(wrapper.vm.action).toEqual(action);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="createKey-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="keyFormDialog-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is icon rendering. Editing public key.
  ///////

  describe('Icon', () => {
    const createKey = false;
    const action = 'public';

    beforeEach(() => {
      wrapper = mount(KeyFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { keyObject, createKey, action },
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

    it('Receive data in props', () => {
      expect(wrapper.vm.keyObject).toEqual(keyObject);
      expect(wrapper.vm.createKey).toEqual(createKey);
      expect(wrapper.vm.action).toEqual(action);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Show message tooltip to user owner', async (done) => {
      const icons = wrapper.findAll('.v-icon');
      const helpIcon = icons.at(0);
      helpIcon.trigger('mouseenter');
      await wrapper.vm.$nextTick();

      expect(icons.length).toBe(1);
      expect(helpIcon.text()).toEqual('edit');
      requestAnimationFrame(() => {
        expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Edit');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="keyFormDialog-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is icon rendering. Creating private key.
  ///////

  describe('Icon', () => {
    const createKey = true;
    const action = 'private';

    beforeEach(() => {
      wrapper = mount(KeyFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { keyObject, createKey, action },
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

    it('Receive data in props', () => {
      expect(wrapper.vm.keyObject).toEqual(keyObject);
      expect(wrapper.vm.createKey).toEqual(createKey);
      expect(wrapper.vm.action).toEqual(action);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="createKey-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="keyFormDialog-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // in this case, when the user owns the namespace and the focus of
  // the test is icon rendering. Editing private key.
  ///////

  describe('Icon', () => {
    const createKey = false;
    const action = 'private';

    beforeEach(() => {
      wrapper = mount(KeyFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { keyObject, createKey, action },
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

    it('Receive data in props', () => {
      expect(wrapper.vm.keyObject).toEqual(keyObject);
      expect(wrapper.vm.createKey).toEqual(createKey);
      expect(wrapper.vm.action).toEqual(action);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.keyLocal).toEqual(keyObject);
      expect(wrapper.vm.supportedKeys).toEqual(supportedKeys);
    });

    //////
    // HTML validation
    //////

    it('Show message tooltip to user owner', async (done) => {
      const icons = wrapper.findAll('.v-icon');
      const helpIcon = icons.at(0);
      helpIcon.trigger('mouseenter');
      await wrapper.vm.$nextTick();

      expect(icons.length).toBe(1);
      expect(helpIcon.text()).toEqual('edit');
      requestAnimationFrame(() => {
        expect(wrapper.find('[data-test="text-tooltip"]').text()).toEqual('Edit');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="keyFormDialog-card"]').exists()).toEqual(false);
    });
  }, 5000);

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering. Creating public key.
  ///////

  describe('Dialog opened', () => {
    const createKey = true;
    const action = 'public';

    beforeEach(() => {
      wrapper = mount(KeyFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { createKey, action },
        vuetify,
      });

      wrapper.setData({ dialog: true });
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

    it('Receive data in props', () => {
      expect(wrapper.vm.keyObject).toEqual({});
      expect(wrapper.vm.createKey).toEqual(createKey);
      expect(wrapper.vm.action).toEqual(action);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.keyLocal).toEqual({ ...creatingKey, hostname: '' });
      expect(wrapper.vm.supportedKeys).toEqual(supportedKeys);
    });

    //////
    // HTML validation
    //////

    it('Show validation messages', async () => {
      //////
      // In this case, the empty fields are validated.
      //////

      wrapper.setData({ keyLocal: { name: '', data: '' } });
      await flushPromises();

      const validatorName = wrapper.vm.$refs.providerName;
      let validatorData = wrapper.vm.$refs.providerData;

      await validatorName.validate();
      await validatorData.validate();
      expect(validatorName.errors[0]).toBe('This field is required');
      expect(validatorData.errors[0]).toBe('This field is required');

      //////
      // In this case, any string is validated in the data.
      //////

      wrapper.setData({ keyLocal: { data: 'xxxxxxxx' } });
      await flushPromises();

      validatorData = wrapper.vm.$refs.providerData;

      await validatorData.validate();
      expect(validatorData.errors[0]).toBe('Not valid key');
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="keyFormDialog-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="create-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="edit-btn"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering. Editing public key.
  ///////

  describe('Dialog opened', () => {
    const createKey = false;
    const action = 'public';

    beforeEach(() => {
      wrapper = mount(KeyFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { keyObject, createKey, action },
        vuetify,
      });

      wrapper.setData({ dialog: true });
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

    it('Receive data in props', () => {
      expect(wrapper.vm.keyObject).toEqual(keyObject);
      expect(wrapper.vm.createKey).toEqual(createKey);
      expect(wrapper.vm.action).toEqual(action);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.supportedKeys).toEqual(supportedKeys);
    });

    //////
    // HTML validation
    //////

    it('Show validation messages', async () => {
      //////
      // In this case, the empty fields are validated.
      //////

      wrapper.setData({ keyLocal: { name: '', data: '' } });
      await flushPromises();

      const validatorName = wrapper.vm.$refs.providerName;
      let validatorData = wrapper.vm.$refs.providerData;

      await validatorName.validate();
      await validatorData.validate();
      expect(validatorName.errors[0]).toBe('This field is required');
      expect(validatorData.errors[0]).toBe('This field is required');

      //////
      // In this case, any string is validated in the data.
      //////

      wrapper.setData({ keyLocal: { data: 'xxxxxxxx' } });
      await flushPromises();

      validatorData = wrapper.vm.$refs.providerData;

      await validatorData.validate();
      expect(validatorData.errors[0]).toBe('Not valid key');
    });
    it('Show validation messages', async () => {
      //////
      // In this case, the empty fields are validated.
      //////

      wrapper.setData({ keyLocal: { name: '', data: '' } });
      await flushPromises();

      const validatorName = wrapper.vm.$refs.providerName;
      const validatorData = wrapper.vm.$refs.providerData;

      await validatorName.validate();
      await validatorData.validate();
      expect(validatorName.errors[0]).toBe('This field is required');
      expect(validatorData.errors[0]).toBe('This field is required');
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="keyFormDialog-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="create-btn"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="edit-btn"]').exists()).toEqual(true);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering. Creating private key.
  ///////

  describe('Dialog opened', () => {
    const createKey = true;
    const action = 'private';

    beforeEach(() => {
      wrapper = mount(KeyFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { createKey, action },
        vuetify,
      });

      wrapper.setData({ dialog: true });
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

    it('Receive data in props', () => {
      expect(wrapper.vm.keyObject).toEqual({});
      expect(wrapper.vm.createKey).toEqual(createKey);
      expect(wrapper.vm.action).toEqual(action);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.keyLocal).toEqual(creatingKey);
      expect(wrapper.vm.supportedKeys).toEqual(supportedKeys);
    });

    //////
    // HTML validation
    //////

    it('Show validation messages', async () => {
      //////
      // In this case, the empty fields are validated.
      //////

      wrapper.setData({ keyLocal: { name: '', data: '' } });
      await flushPromises();

      let validatorName = wrapper.vm.$refs.providerName;
      let validatorData = wrapper.vm.$refs.providerData;

      await validatorName.validate();
      await validatorData.validate();
      expect(validatorName.errors[0]).toBe('This field is required');
      expect(validatorData.errors[0]).toBe('This field is required');

      //////
      // In this case, any string is validated in the data.
      //////

      wrapper.setData({ keyLocal: { data: 'xxxxxxxx' } });
      await flushPromises();

      validatorData = wrapper.vm.$refs.providerData;

      await validatorData.validate();
      expect(validatorData.errors[0]).toBe('Not valid key');

      //////
      // In this case, the public key is inserted where the private
      // key should be inserted.
      //////

      wrapper.setData({ keyLocal: { data: publicKey } });
      await flushPromises();

      validatorData = wrapper.vm.$refs.providerData;

      await validatorData.validate();
      expect(validatorData.errors[0]).toBe('Not valid key');

      //////
      // In this case, the private key is inserted to validation.
      //////

      wrapper.setData({ keyLocal: { name: 'ShellHub', data: privateKey } });
      await flushPromises();

      validatorName = wrapper.vm.$refs.providerName;
      validatorData = wrapper.vm.$refs.providerData;

      await validatorName.validate();
      await validatorData.validate();
      expect(validatorName.errors[0]).toBe(undefined);
      expect(validatorData.errors[0]).toBe(undefined);
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="keyFormDialog-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="create-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="edit-btn"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering. Editing private key.
  ///////

  describe('Dialog opened', () => {
    const createKey = false;
    const action = 'private';

    beforeEach(() => {
      wrapper = mount(KeyFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { keyObject, createKey, action },
        vuetify,
      });

      wrapper.setData({ dialog: true });
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

    it('Receive data in props', () => {
      expect(wrapper.vm.keyObject).toEqual(keyObject);
      expect(wrapper.vm.createKey).toEqual(createKey);
      expect(wrapper.vm.action).toEqual(action);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.keyLocal).toEqual(keyObject);
      expect(wrapper.vm.supportedKeys).toEqual(supportedKeys);
    });

    //////
    // HTML validation
    //////

    it('Show validation messages', async () => {
      //////
      // In this case, the empty fields are validated.
      //////

      wrapper.setData({ keyLocal: { name: '', data: '' } });
      await flushPromises();

      let validatorName = wrapper.vm.$refs.providerName;
      let validatorData = wrapper.vm.$refs.providerData;

      await validatorName.validate();
      await validatorData.validate();
      expect(validatorName.errors[0]).toBe('This field is required');
      expect(validatorData.errors[0]).toBe('This field is required');

      //////
      // In this case, any string is validated in the data.
      //////

      wrapper.setData({ keyLocal: { data: 'xxxxxxxx' } });
      await flushPromises();

      validatorData = wrapper.vm.$refs.providerData;

      await validatorData.validate();
      expect(validatorData.errors[0]).toBe('Not valid key');

      //////
      // In this case, the public key is inserted where the private
      // key should be inserted.
      //////

      wrapper.setData({ keyLocal: { data: publicKey } });
      await flushPromises();

      validatorData = wrapper.vm.$refs.providerData;

      await validatorData.validate();
      expect(validatorData.errors[0]).toBe('Not valid key');

      //////
      // In this case, the private key is inserted to validation.
      //////

      wrapper.setData({ keyLocal: { name: 'ShellHub', data: privateKey } });
      await flushPromises();

      validatorName = wrapper.vm.$refs.providerName;
      validatorData = wrapper.vm.$refs.providerData;

      await validatorName.validate();
      await validatorData.validate();
      expect(validatorName.errors[0]).toBe(undefined);
      expect(validatorData.errors[0]).toBe(undefined);
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="keyFormDialog-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="create-btn"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="edit-btn"]').exists()).toEqual(true);
    });
  });
});
