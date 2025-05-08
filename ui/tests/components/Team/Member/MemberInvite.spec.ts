import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MemberInvite from "@/components/Team/Member/MemberInvite.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type MemberInviteWrapper = VueWrapper<InstanceType<typeof MemberInvite>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
  showInfo: vi.fn(),
};

describe("Member Invite", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: MemberInviteWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    settings: {
      session_record: true,
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  };

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(MemberInvite, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      attachTo: el,
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

    expect(dialog.find('[data-test="namespaceNewMember-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="email-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="role-select"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="invite-btn"]').exists()).toBe(true);
  });

  it("Invite Member Email - Error Validation", async () => {
    mockNamespace.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/members").reply(409);

    const storeSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="invite-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="email-text"]').setValue("not-working-mail");

    await wrapper.findComponent('[data-test="role-select"]').setValue("not-right-role");

    await wrapper.findComponent('[data-test="invite-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith("namespaces/sendEmailInvitation", {
      email: "not-working-mail",
      tenant_id: "fake-tenant-data",
      role: "not-right-role",
    });

    expect(wrapper.vm.emailError).toEqual("This user is already a member of this namespace.");
  });

  it("Invite Member Email - Success Validation", async () => {
    mockNamespace.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/members").reply(200);

    const storeSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="invite-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="email-text"]').setValue("workingmail@mail.com");

    await wrapper.findComponent('[data-test="role-select"]').setValue("administrator");

    await wrapper.findComponent('[data-test="invite-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith("namespaces/sendEmailInvitation", {
      email: "workingmail@mail.com",
      tenant_id: "fake-tenant-data",
      role: "administrator",
    });

    expect(mockSnackbar.showSuccess).toBeCalledWith("Invitation email sent successfully.");

    expect(wrapper.vm.emailError).toBeUndefined();
  });

  it("Generates Invitation Link - Failure", async () => {
    mockNamespace.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/members/invites").reply(404);

    const storeSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="invite-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="email-text"]').setValue("valid@mail.com");

    await wrapper.findComponent('[data-test="role-select"]').setValue("administrator");

    await wrapper.findComponent('[data-test="link-request-checkbox"]').setValue(true); // Check the checkbox

    await wrapper.findComponent('[data-test="invite-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith("namespaces/generateInvitationLink", {
      email: "valid@mail.com",
      tenant_id: "fake-tenant-data",
      role: "administrator",
    });

    expect(wrapper.vm.invitationLink).toBe("");
  });

  it("Generates Invitation Link - Success", async () => {
    mockNamespace.onPost("http://localhost:3000/api/namespaces/fake-tenant-data/members/invites").reply(200, {
      link: "http://localhost/invite-link",
    });

    const storeSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="invite-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="email-text"]').setValue("valid@mail.com");

    await wrapper.findComponent('[data-test="role-select"]').setValue("administrator");

    await wrapper.findComponent('[data-test="link-request-checkbox"]').setValue(true);

    await wrapper.findComponent('[data-test="invite-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toBeCalledWith("namespaces/generateInvitationLink", {
      email: "valid@mail.com",
      tenant_id: "fake-tenant-data",
      role: "administrator",
    });

    expect(wrapper.vm.formWindow).toEqual("form-2");

    expect(wrapper.vm.invitationLink).toEqual("http://localhost/invite-link");
  });
});
