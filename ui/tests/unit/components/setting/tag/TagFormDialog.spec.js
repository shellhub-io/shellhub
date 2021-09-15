import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import { ValidationProvider, ValidationObserver } from 'vee-validate';
import flushPromises from 'flush-promises';
import Vuetify from 'vuetify';
import TagFormDialog from '@/components/setting/tag/TagFormDialog';
import '@/vee-validate';

describe('TagFormDialog', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.use(Vuex);
  localVue.component('ValidationProvider', ValidationProvider);
  localVue.component('ValidationObserver', ValidationObserver);

  document.body.setAttribute('data-app', true);

  let wrapper;

  const isOwner = true;
  const createAction = 'create';
  const editAction = 'edit';
  const tagName = 'tag1';

  // vee-validate variables bellow
  const invalidName = ['xxx/', 'xxx@', 'xxx&', 'xxx:'];
  const invalidMinAndMaxCharacters = [
    'x', 'xx',
  ];

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      isOwner,
    },
    getters: {
      'namespaces/owner': (state) => state.isOwner,
    },
    actions: {
      'tags/post': () => {},
      'tags/edit': () => {},
      'snackbar/showSnackbarSuccessAction': () => {},
      'snackbar/showSnackbarErrorAction': () => {},
    },
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is button rendering. Add tag
  ///////

  describe('Icon create', () => {
    beforeEach(() => {
      wrapper = mount(TagFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { action: createAction },
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
      expect(wrapper.vm.action).toEqual(createAction);
      expect(wrapper.vm.tagName).toEqual('');
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.tagLocal).toEqual('');
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
      requestAnimationFrame(() => {
        expect(wrapper.find('[data-test="text-span"]').text()).toEqual('Create tag');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="tagForm-card"]').exists()).toEqual(false);
    });
  });

  //////
  // In this case, when the user owns the namespace and the focus of
  // the test is icon rendering. Editing tag
  //////

  describe('Icon edit', () => {
    beforeEach(() => {
      wrapper = mount(TagFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { action: editAction, tagName },
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
      expect(wrapper.vm.action).toEqual(editAction);
      expect(wrapper.vm.tagName).toEqual(tagName);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.tagLocal).toEqual(tagName);
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
      requestAnimationFrame(() => {
        expect(wrapper.find('[data-test="text-span"]').text()).toEqual('Edit');
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="add-btn"]').exists()).toEqual(false);
      expect(wrapper.find('[data-test="tagForm-card"]').exists()).toEqual(false);
    });
  });

  ///////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering. Creating tag
  ///////

  describe('Dialog create', () => {
    beforeEach(() => {
      wrapper = mount(TagFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { action: createAction },
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
      expect(wrapper.vm.action).toEqual(createAction);
      expect(wrapper.vm.tagName).toEqual('');
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.tagLocal).toEqual('');
    });

    //////
    // HTML validation
    //////

    //////
    // In this case, the empty fields are validated.
    //////

    it('Show validation messages', async () => {
      wrapper.setData({ tagLocal: '' });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerTag;

      await validator.validate();
      expect(validator.errors[0]).toBe('This field is required');
    });

    //////
    // In this case, the route identifier are validated.
    //////

    it('Show validation messages', async (done) => {
      invalidName.forEach(async (name) => {
        wrapper.setData({ tagLocal: name });
        await flushPromises();

        const validator = wrapper.vm.$refs.providerTag;

        await validator.validate();
        expect(validator.errors[0]).toBe('The name must not contain /, @, &, and :.');

        await flushPromises();
        done();
      });
    });

    //////
    // In this case, the min and max characters are validated.
    //////

    it('Show validation messages', async (done) => {
      let invalidMaxCharacter = '';

      for (let x = 0; x < 256; x += 1) {
        invalidMaxCharacter = invalidMaxCharacter.concat('x');
      }
      invalidMinAndMaxCharacters.push(invalidMaxCharacter);
      invalidMinAndMaxCharacters.push(invalidMaxCharacter.concat('x'));

      invalidMinAndMaxCharacters.forEach(async (character) => {
        wrapper.setData({ tagLocal: character });
        await flushPromises();

        const validator = wrapper.vm.$refs.providerTag;

        await validator.validate();
        expect(validator.errors[0]).toBe('Your tag should be 3-255 characters long');

        await flushPromises();
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="tagForm-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="doAction-btn"]').exists()).toEqual(true);
    });
  });

  //////
  // In this case, when the user owns the namespace and the focus of
  // the test is dialog rendering. Editing tag
  //////

  describe('Dialog Edit', () => {
    beforeEach(async () => {
      wrapper = mount(TagFormDialog, {
        store,
        localVue,
        stubs: ['fragment'],
        propsData: { action: editAction, tagName },
        vuetify,
      });

      wrapper.setData({ dialog: true });
      await flushPromises();
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
      expect(wrapper.vm.action).toEqual(editAction);
      expect(wrapper.vm.tagName).toEqual(tagName);
    });
    it('Compare data with default value', () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.tagLocal).toEqual(tagName);
    });

    //////
    // HTML validation
    //////

    //////
    // In this case, the empty fields are validated.
    //////

    it('Show validation messages', async () => {
      wrapper.setData({ tagLocal: '' });
      await flushPromises();

      const validator = wrapper.vm.$refs.providerTag;

      await validator.validate();
      expect(validator.errors[0]).toBe('This field is required');
    });

    //////
    // In this case, the route identifier are validated.
    //////

    it('Show validation messages', async (done) => {
      invalidName.forEach(async (name) => {
        wrapper.setData({ tagLocal: name });
        await flushPromises();

        const validator = wrapper.vm.$refs.providerTag;

        await validator.validate();
        expect(validator.errors[0]).toBe('The name must not contain /, @, &, and :.');

        await flushPromises();
        done();
      });
    });

    //////
    // In this case, the min and max characters are validated.
    //////

    it('Show validation messages', async (done) => {
      let invalidMaxCharacter = '';

      for (let x = 0; x < 256; x += 1) {
        invalidMaxCharacter = invalidMaxCharacter.concat('x');
      }
      invalidMinAndMaxCharacters.push(invalidMaxCharacter);
      invalidMinAndMaxCharacters.push(invalidMaxCharacter.concat('x'));

      invalidMinAndMaxCharacters.forEach(async (character) => {
        wrapper.setData({ tagLocal: character });
        await flushPromises();

        const validator = wrapper.vm.$refs.providerTag;

        await validator.validate();
        expect(validator.errors[0]).toBe('Your tag should be 3-255 characters long');

        await flushPromises();
        done();
      });
    });
    it('Renders the template with data', () => {
      expect(wrapper.find('[data-test="tagForm-card"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="cancel-btn"]').exists()).toEqual(true);
      expect(wrapper.find('[data-test="doAction-btn"]').exists()).toEqual(true);
    });
  });
});
