import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MemberInvite from "@/components/Team/Member/MemberInvite.vue";
import { namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import { envVariables } from "@/envVariables";
import useNamespacesStore from "@/store/modules/namespaces";

type MemberInviteWrapper = VueWrapper<InstanceType<typeof MemberInvite>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
  showInfo: vi.fn(),
};

describe("Member Invite", () => {
  let wrapper: MemberInviteWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  beforeEach(async () => {
    envVariables.isCloud = true;
    localStorage.setItem("tenant", "fake-tenant-data");
    authStore.$patch({
      role: "owner",
      tenantId: "fake-tenant-data",
    });
    wrapper = mount(MemberInvite, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components", async () => {
    const dialog = new DOMWrapper(document.body);

    expect(wrapper.findComponent('[data-test="invite-dialog-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="invite-dialog-btn"]').trigger("click");

    expect(dialog.find('[data-test="namespace-new-member-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="email-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="role-select"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="invite-btn"]').exists()).toBe(true);
  });

  it("Invite Member Email - Error Validation", async () => {
    mockNamespacesApi.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/members").reply(409);

    const storeSpy = vi.spyOn(namespacesStore, "sendEmailInvitation");

    await wrapper.findComponent('[data-test="invite-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="email-text"]').setValue("not-working-mail");

    await wrapper.findComponent('[data-test="role-select"]').setValue("not-right-role");

    await wrapper.findComponent('[data-test="invite-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      email: "not-working-mail",
      tenant_id: "fake-tenant-data",
      role: "not-right-role",
    });

    expect(wrapper.vm.emailError).toEqual("This user is already a member of this namespace.");
  });

  it("Invite Member Email - Success Validation", async () => {
    mockNamespacesApi.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/members").reply(200);

    const storeSpy = vi.spyOn(namespacesStore, "sendEmailInvitation");

    await wrapper.findComponent('[data-test="invite-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="email-text"]').setValue("workingmail@mail.com");

    await wrapper.findComponent('[data-test="role-select"]').setValue("administrator");

    await wrapper.findComponent('[data-test="invite-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      email: "workingmail@mail.com",
      tenant_id: "fake-tenant-data",
      role: "administrator",
    });

    expect(mockSnackbar.showSuccess).toBeCalledWith("Invitation email sent successfully.");

    expect(wrapper.vm.emailError).toBeUndefined();
  });

  it("Generates Invitation Link - Failure", async () => {
    mockNamespacesApi.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/members/invites").reply(404);

    const storeSpy = vi.spyOn(namespacesStore, "generateInvitationLink");

    await wrapper.findComponent('[data-test="invite-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="email-text"]').setValue("valid@mail.com");

    await wrapper.findComponent('[data-test="role-select"]').setValue("administrator");

    await wrapper.findComponent('[data-test="link-request-checkbox"]').setValue(true); // Check the checkbox

    await wrapper.findComponent('[data-test="invite-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      email: "valid@mail.com",
      tenant_id: "fake-tenant-data",
      role: "administrator",
    });

    expect(wrapper.vm.invitationLink).toBe("");
  });

  it("Generates Invitation Link - Success", async () => {
    mockNamespacesApi.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/members/invites").reply(200, {
      link: "http://localhost/invite-link",
    });

    const storeSpy = vi.spyOn(namespacesStore, "generateInvitationLink");

    await wrapper.findComponent('[data-test="invite-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="email-text"]').setValue("valid@mail.com");

    await wrapper.findComponent('[data-test="role-select"]').setValue("administrator");

    await wrapper.findComponent('[data-test="link-request-checkbox"]').setValue(true);

    await wrapper.findComponent('[data-test="invite-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith({
      email: "valid@mail.com",
      tenant_id: "fake-tenant-data",
      role: "administrator",
    });

    expect(wrapper.vm.formWindow).toEqual("form-2");

    expect(wrapper.vm.invitationLink).toEqual("http://localhost/invite-link");
  });
});
