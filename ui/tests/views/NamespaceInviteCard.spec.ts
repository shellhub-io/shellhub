import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, shallowMount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { nextTick } from "vue";
import NamespaceInviteCard from "@/views/NamespaceInviteCard.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type NamespaceInviteCardWrapper = VueWrapper<InstanceType<typeof NamespaceInviteCard>>;

let wrapper: NamespaceInviteCardWrapper;
setActivePinia(createPinia());
const vuetify = createVuetify();

describe("Namespace Invite Card (Invalid User)", () => {
  beforeEach(async () => {
    localStorage.removeItem("id");
    await router.push({ query: { "user-id": "507f1f77bcf86cd799439011", "tenant-id": "fake-tenant" } });
    await router.isReady();

    wrapper = shallowMount(NamespaceInviteCard, {
      global: { plugins: [vuetify, router, SnackbarPlugin] },
    });

    await nextTick();
  });

  it("Displays appropriate error alert when user is not valid", () => {
    expect(wrapper.vm.errorAlert).toBe("You aren't logged in the account meant for this invitation.");
  });
});

describe("Namespace Invite Card", () => {
  beforeEach(async () => {
    await router.push({ query: { "user-id": "507f1f77bcf86cd799439011", "tenant-id": "fake-tenant" } });
    localStorage.setItem("tenant", "fake-tenant");
    localStorage.setItem("id", "507f1f77bcf86cd799439011");

    wrapper = mount(NamespaceInviteCard, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });

    await nextTick();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders dialog elements with correct data-test attributes", () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="message"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="actions"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="decline-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="spacer"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="accept-btn"]').exists()).toBe(true);
  });

  it("Displays the correct title and message", () => {
    expect(wrapper.find('[data-test="title"]').text()).toBe("Namespace Invitation");
    expect(wrapper.find('[data-test="message"]').text()).toBe(wrapper.vm.message);
  });
});
