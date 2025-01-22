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
          <div class="d-flex justify-center align-center bg-whitea">
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

        <v-card-text>
          <p
            class="text-caption text-grey-lighten-4 mb-1"
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
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn
            data-test="close-btn"
            @click="close()"
          > Close </v-btn>

          <v-btn
            color="primary"
            data-test="invite-btn"
            @click="addMember()"
          >
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
import multiavatar from "@multiavatar/multiavatar/esm";
import hasPermission from "@/utils/permission";
import { useStore } from "@/store";
import { actions, authorizer } from "@/authorizer";
import {
  INotificationsError,
  INotificationsSuccess,
} from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";
import { envVariables } from "@/envVariables";

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

const getAvatar = (index: number) => multiavatar(`${Math.floor(Math.random() * (Number.MAX_SAFE_INTEGER - index + 1)) + index}`);

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
            handleError(error);
        }
      }
    }
  }
};

defineExpose({ emailError });
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
