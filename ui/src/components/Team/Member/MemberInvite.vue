<template>
  <div>
    <v-tooltip
      location="bottom"
      class="text-center"
      :disabled="canAddMember"
    >
      <template #activator="{ props }">
        <div v-bind="props">
          <v-btn
            :disabled="!canAddMember"
            color="primary"
            data-test="invite-dialog-btn"
            @click="showDialog = true"
          >
            Invite Member
          </v-btn>
        </div>
      </template>
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <FormDialog
      v-model="showDialog"
      title="Invite Member"
      icon="mdi-account-plus"
      :confirm-text="formWindow === 'form-1' ? 'Invite' : ''"
      :confirm-disabled="!email || !selectedRole || !!emailError || formWindow === 'form-2'"
      :confirm-loading="isLoading"
      cancel-text="Close"
      :confirm-data-test="'invite-btn'"
      :cancel-data-test="'close-btn'"
      @close="close"
      @confirm="getInvitationCheckbox ? generateLinkInvite() : sendEmailInvite()"
      @cancel="close"
    >
      <v-card-text class="pa-6 text-justify">
        <v-window v-model="formWindow">
          <v-window-item value="form-1">
            <p
              v-if="envVariables.isCloud"
              class="mb-4"
            >
              If this email isn't associated with an existing account, we'll send an email to sign-up.
            </p>
            <v-text-field
              v-model="email"
              class="mb-6"
              label="Email"
              :error-messages="emailError"
              required
              hide-details="auto"
              data-test="email-text"
            />
            <RoleSelect
              v-model="selectedRole"
              data-test="role-select"
              class="mb-4"
            />
            <v-checkbox
              v-if="envVariables.isCloud"
              v-model="getInvitationCheckbox"
              label="Get the invite link instead of sending an e-mail"
              hide-details
              density="compact"
              data-test="link-request-checkbox"
            />
          </v-window-item>
          <v-window-item value="form-2">
            <div>
              <p class="mb-4">
                Share this link with the person you want to invite. They can use it to join your namespace.
              </p>
              <p class="mb-4">
                <strong>Note:</strong> This link is only valid for the email address you entered earlier.
              </p>
              <CopyWarning :copied-item="'Invitation link'">
                <template #default="{ copyText }">
                  <v-text-field
                    v-model="invitationLink"
                    readonly
                    active
                    density="compact"
                    append-icon="mdi-content-copy"
                    label="Invitation Link"
                    data-test="invitation-link"
                    @click="copyText(invitationLink)"
                    @keypress="copyText(invitationLink)"
                  />
                </template>
              </CopyWarning>
              <p class="text-caption text-grey-darken-1">
                The invitation link remains valid for 7 days, if the link does not work, ensure the invite has not expired.
              </p>
            </div>
          </v-window-item>
        </v-window>
      </v-card-text>
    </FormDialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import { envVariables } from "@/envVariables";
import useSnackbar from "@/helpers/snackbar";
import CopyWarning from "@/components/User/CopyWarning.vue";
import RoleSelect from "../RoleSelect.vue";
import { BasicRole } from "@/interfaces/INamespace";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const emit = defineEmits(["update"]);
const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const showDialog = ref(false);
const isLoading = ref(false);
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

const canAddMember = hasPermission("namespace:addMember");

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
  isLoading.value = true;
  try {
    invitationLink.value = await namespacesStore.generateInvitationLink(getInvitePayload());
    snackbar.showSuccess("Invitation link generated successfully.");
    formWindow.value = "form-2";
  } catch (error) {
    handleInviteError(error);
  } finally {
    isLoading.value = false;
  }
};

const sendEmailInvite = async () => {
  isLoading.value = true;
  try {
    await namespacesStore.sendEmailInvitation(getInvitePayload());
    snackbar.showSuccess("Invitation email sent successfully.");
    update();
    resetFields();
  } catch (error) {
    handleInviteError(error);
  } finally {
    isLoading.value = false;
  }
};

defineExpose({ emailError, formWindow, invitationLink });
</script>
