import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";
import NamespaceInviteCard from "@/views/NamespaceInviteCard.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { namespacesApi } from "@/api/http";

type NamespaceInviteCardWrapper = VueWrapper<InstanceType<typeof NamespaceInviteCard>>;

let wrapper: NamespaceInviteCardWrapper;
setActivePinia(createPinia());
const vuetify = createVuetify();
const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

describe("Namespace Invite Dialog (Invalid User)", () => {
  beforeEach(async () => {
    wrapper = mount(NamespaceInviteCard, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Displays appropriate error alert when user is not valid", async () => {
    localStorage.removeItem("id");

    await nextTick();

    expect(wrapper.vm.errorAlert).toBe("You aren't logged in the account meant for this invitation.");
    expect(wrapper.find('[data-test="decline-btn"]').exists()).toBe(true);
  });
});

describe("Namespace Invite Dialog", () => {
  beforeEach(async () => {
    await router.push({ query: { "user-id": "507f1f77bcf86cd799439011", "tenant-id": "fake-tenant" } });
    localStorage.setItem("tenant", "fake-tenant");
    localStorage.setItem("id", "507f1f77bcf86cd799439011");

    wrapper = mount(NamespaceInviteCard, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders dialog elements with correct data-test attributes", async () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="message"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="actions"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="decline-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="spacer"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="accept-btn"]').exists()).toBe(true);
  });

  it("Displays the correct title and message", () => {
    expect(wrapper.find('[data-test="title"]').text()).toBe(wrapper.vm.title);
    expect(wrapper.find('[data-test="message"]').text()).toBe(wrapper.vm.message);
  });

  it("Calls close method when decline button is clicked", async () => {
    const closeSpy = vi.spyOn(wrapper.vm, "close");
    await wrapper.findComponent('[data-test="decline-btn"]').trigger("click");

    await flushPromises();
    expect(closeSpy).toHaveBeenCalled();
  });

  it("Calls acceptInvite method when Accept Invitation button is clicked", async () => {
    mockNamespacesApi.onPatch("http://localhost:3000/api/namespaces/fake-tenant/members/accept-invite").reply(200);
    const acceptSpy = vi.spyOn(wrapper.vm, "acceptInvite");
    await flushPromises();
    await wrapper.findComponent('[data-test="accept-btn"]').trigger("click");

    await flushPromises();
    expect(acceptSpy).toHaveBeenCalled();
  });

  it("Handles error state correctly", async () => {
    wrapper.vm.handleInviteError({ response: { status: 400 } });
    await nextTick();
    expect(wrapper.find('[data-test="title"]').text()).toBe("Invite Accept Error");
    expect(wrapper.find('[data-test="message"]').text()).toBe("An unexpected error occurred. Please try again later.");
    expect(wrapper.find('[data-test="accept-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="decline-btn"]').exists()).toBe(true);
  });
});
