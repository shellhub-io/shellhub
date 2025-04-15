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
            @click="dialog = !dialog"
            data-test="invite-dialog-btn"
          >
            Invite Member
          </v-btn>
        </div>
      </template>
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="450"
      @click:outside="close()"
    >
      <v-card
        data-test="namespaceNewMember-dialog"
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
                label="Email"
                :error-messages="emailError"
                required
                data-test="email-text"
              />
            </v-card-text>

            <v-card-text class="mt-n10">
              <v-select
                v-model="selectedRole"
                :items="items"
                label="Role"
                :error-messages="selectedRoleError"
                required
                data-test="role-select"
              />

              <v-checkbox
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
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import multiavatar from "@multiavatar/multiavatar";
import hasPermission from "@/utils/permission";
import { useStore } from "@/store";
import { actions, authorizer } from "@/authorizer";
import {
  INotificationsCopy,
  INotificationsError,
  INotificationsSuccess,
} from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";
import { envVariables } from "@/envVariables";

const items = ["administrator", "operator", "observer"];

const emit = defineEmits(["update"]);
const store = useStore();
const dialog = ref(false);
const getInvitationCheckbox = ref(false);
const invitationLink = computed(() => store.getters["namespaces/getInvitationLink"]);
const formWindow = ref("form-1");

const {
  value: email,
  errorMessage: emailError,
  setErrors: setEmailError,
  resetField: resetIdentifier,
} = useField<string>("identifier", yup.string().email().required(), {
  initialValue: "",
});

const {
  value: selectedRole,
  errorMessage: selectedRoleError,
  resetField: resetSelectedRole,
} = useField<string>("selectedRole", yup.string().required(), {
  initialValue: "",
});

const hasAuthorization = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.namespace.addMember,
    );
  }

  return false;
};

const getAvatar = (index: number) => multiavatar(`${Math.floor(Math.random() * (Number.MAX_SAFE_INTEGER - index + 1)) + index}`);

const resetFields = () => {
  resetIdentifier();
  resetSelectedRole();
};

const close = () => {
  resetFields();
  dialog.value = false;
  formWindow.value = "form-1";
};

const update = () => {
  emit("update");
  close();
};

const copyText = (value: string | undefined) => {
  if (value) {
    navigator.clipboard.writeText(value);
    store.dispatch("snackbar/showSnackbarCopy", INotificationsCopy.invitationLink);
  }
};

const handleInviteError = (error: unknown) => {
  store.dispatch(
    "snackbar/showSnackbarErrorAction",
    INotificationsError.namespaceNewMember,
  );

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

const generateLinkInvite = async () => {
  try {
    await store.dispatch("namespaces/generateInvitationLink", {
      email: email.value,
      tenant_id: store.getters["auth/tenant"],
      role: selectedRole.value,
    });

    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.namespaceNewMember,
    );

    formWindow.value = "form-2";
  } catch (error) {
    handleInviteError(error);
  }
};

const sendEmailInvite = async () => {
  try {
    await store.dispatch("namespaces/sendEmailInvitation", {
      email: email.value,
      tenant_id: store.getters["auth/tenant"],
      role: selectedRole.value,
    });

    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.namespaceNewMember,
    );

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
