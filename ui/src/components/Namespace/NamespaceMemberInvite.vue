<template>
  <div>
    <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorization()">
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

    <v-dialog v-model="dialog" max-width="450" @click:outside="close()">
      <v-card data-test="namespaceNewMember-dialog" class="bg-v-theme-surface">
        <v-card-title class="bg-primary"> Invite member </v-card-title>

        <v-card-text>
          <p class="text-caption text-grey-lighten-4 mb-1">
            If this email isn't associated with an existing account, we'll send  an email to sign-up.
          </p>
          <v-text-field
            v-model="email"
            label="Email"
            :error-messages="emailError"
            required
            variant="underlined"
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
            variant="underlined"
            data-test="role-select"
          />
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn data-test="close-btn" @click="close()"> Close </v-btn>

          <v-btn color="primary" data-test="invite-btn" @click="addMember()">
            Invite
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import hasPermission from "../../utils/permission";
import { useStore } from "../../store";
import { actions, authorizer } from "../../authorizer";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import handleError from "../../utils/handleError";

const items = ["administrator", "operator", "observer"];

const emit = defineEmits(["update"]);
const store = useStore();
const dialog = ref(false);

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
  setErrors: setSelectedRoleError,
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

const hasErrors = () => {
  if (selectedRole.value === "") {
    setSelectedRoleError("Select a role");
    return true;
  }

  if (email.value === "") {
    setEmailError("This field is required");
    return true;
  }

  return false;
};

const resetFields = () => {
  resetIdentifier();
  resetSelectedRole();
};

const close = () => {
  resetFields();
  dialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const addMember = async () => {
  if (!hasErrors()) {
    try {
      await store.dispatch("namespaces/addUser", {
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
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError;
        if (axiosError.response?.status === 409) {
          setEmailError(
            "This user is already a member of this namespace.",
          );
        } else {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.namespaceNewMember,
          );
          handleError(error);
        }
      }
    }
  }
};

defineExpose({ emailError });
</script>
