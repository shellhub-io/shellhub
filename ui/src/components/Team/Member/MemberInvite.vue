<template>
  <div>
    <v-tooltip
      location="bottom"
      class="text-center"
      :disabled="hasAuthorization()"
    >
      <template v-slot:activator="{ props }">
        <div v-bind="props">
          <v-btn
            :disabled="!hasAuthorization()"
            color="primary"
            @click="showDialog = true"
            data-test="invite-dialog-btn"
          >
            Invite Member
          </v-btn>
        </div>
      </template>
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <BaseDialog
      v-model="showDialog"
      @click:outside="close()"
    >
      <v-card
        data-test="namespace-new-member-dialog"
        class="bg-v-theme-surface"
      >
        <div class="mt-4 mb-4">
          <div class="d-flex justify-center align-center">
            <v-avatar
              class="-right-32 z-0"
              size="46"
            >
              <svg v-html="getAvatar(0)" />
            </v-avatar>

            <v-avatar
              class="-right-16 z-1"
              size="56"
            >
              <svg v-html="getAvatar(1)" />
            </v-avatar>

            <v-avatar
              class="z-2"
              size="72"
              color="primary"
            >
              <v-icon color="white">mdi-account</v-icon>
            </v-avatar>

            <v-avatar
              class="-left-16 z-1"
              size="56"
            >
              <svg v-html="getAvatar(2)" />
            </v-avatar>
            <v-avatar
              class="-left-32 z-0"
              size="46"
            >
              <svg v-html="getAvatar(3)" />
            </v-avatar>
          </div>
        </div>
        <v-card-title class="text-center">
          Invite Member
        </v-card-title>
        <v-window v-model="formWindow">
          <v-window-item value="form-1">
            <v-card-text>
              <p
                class="mb-4"
                v-if="envVariables.isCloud"
              >
                If this email isn't associated with an existing account, we'll send an email to sign-up.
              </p>

              <v-text-field
                v-model="email"
                class="mb-4"
                label="Email"
                :error-messages="emailError"
                required
                data-test="email-text"
              />
            </v-card-text>

            <v-card-text class="mt-n10">
              <RoleSelect
                v-model="selectedRole"
                data-test="role-select"
              />
              <v-checkbox
                v-if="envVariables.isCloud"
                v-model="getInvitationCheckbox"
                label="Get the invite link instead of sending an e-mail"
                hide-details
                data-test="link-request-checkbox"
              />
            </v-card-text>
          </v-window-item>
          <v-window-item value="form-2">
            <v-card-text>
              <p class="mb-4">
                Share this link with the person you want to invite. They can use it to join your namespace.
              </p>
              <p class="mb-4"><strong>Note:</strong> This link is only valid for the email address you entered earlier.
              </p>
              <CopyWarning :copied-item="'Invitation link'">
                <template #default="{ copyText }">
                  <v-text-field
                    v-model="invitationLink"
                    @click="copyText(invitationLink)"
                    @keypress="copyText(invitationLink)"
                    readonly
                    active
                    density="compact"
                    append-icon="mdi-content-copy"
                    label="Invitation Link"
                    data-test="invitation-link"
                  />
                </template>
              </CopyWarning>
              <p class="text-caption text-grey-darken-1">
                The invitation link remains valid for 7 days, if the link does not work, ensure the invite has not expired.
              </p>
            </v-card-text>
          </v-window-item>

        </v-window>

        <v-card-actions>
          <v-spacer />
          <v-btn
            data-test="close-btn"
            @click="close()"
          > Close </v-btn>

          <v-btn
            color="primary"
            data-test="invite-btn"
            @click="getInvitationCheckbox ? generateLinkInvite() : sendEmailInvite()"
            :disabled="!email || !selectedRole || !!emailError || formWindow === 'form-2'"
          >
            Invite
          </v-btn>

        </v-card-actions>
      </v-card>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import multiavatar from "@multiavatar/multiavatar";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import handleError from "@/utils/handleError";
import { envVariables } from "@/envVariables";
import useSnackbar from "@/helpers/snackbar";
import CopyWarning from "@/components/User/CopyWarning.vue";
import BaseDialog from "@/components/BaseDialog.vue";
import RoleSelect from "../RoleSelect.vue";
import { BasicRole } from "@/interfaces/INamespace";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const emit = defineEmits(["update"]);
const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const showDialog = ref(false);
const getInvitationCheckbox = ref(false);
const invitationLink = ref("");
const formWindow = ref("form-1");
const selectedRole = ref<BasicRole>("administrator");

const {
  value: email,
  errorMessage: emailError,
  setErrors: setEmailError,
  resetField: resetIdentifier,
} = useField<string>("identifier", yup.string().email().required(), {
  initialValue: "",
});

const hasAuthorization = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.namespace.addMember);
};

const getAvatar = (index: number) => multiavatar(`${Math.floor(Math.random() * (Number.MAX_SAFE_INTEGER - index + 1)) + index}`);

const resetFields = () => {
  resetIdentifier();
  selectedRole.value = "administrator";
};

const close = () => {
  showDialog.value = false;
  resetFields();
  formWindow.value = "form-1";
};

const update = () => {
  emit("update");
  close();
};

const handleInviteError = (error: unknown) => {
  snackbar.showError("Failed to send invitation.");

  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError;
    switch (axiosError.response?.status) {
      case 409:
        setEmailError("This user is already a member of this namespace.");
        break;
      case 404:
        setEmailError("This user does not exist.");
        break;
      default:
        setEmailError("An error occurred while sending the invitation.");
    }
    handleError(error);
  }
};

const getInvitePayload = () => ({
  email: email.value,
  tenant_id: authStore.tenantId,
  role: selectedRole.value,
});

const generateLinkInvite = async () => {
  try {
    invitationLink.value = await namespacesStore.generateInvitationLink(getInvitePayload());
    snackbar.showSuccess("Invitation link generated successfully.");
    formWindow.value = "form-2";
  } catch (error) {
    handleInviteError(error);
  }
};

const sendEmailInvite = async () => {
  try {
    await namespacesStore.sendEmailInvitation(getInvitePayload());
    snackbar.showSuccess("Invitation email sent successfully.");
    update();
    resetFields();
  } catch (error) {
    handleInviteError(error);
  }
};

defineExpose({ emailError, formWindow, invitationLink });
</script>

<style lang="scss" scoped>
@for $i from 0 through 32 {
  .-left-#{$i} {
    left: -#{$i}px;
  }

  .-right-#{$i} {
    right: -#{$i}px;
  }
}

@for $i from 0 through 2 {
  .z-#{$i} {
    z-index: #{$i};
  }
}
</style>
