<template>
  <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorization()">
    <template v-slot:activator="{ props }">
      <div v-bind="props">
        <v-btn
          :disabled="!hasAuthorization()"
          color="primary"
          @click="dialog = !dialog"
        >
          Add Member
        </v-btn>
      </div>
    </template>
    <span> You don't have this kind of authorization. </span>
  </v-tooltip>

  <v-dialog v-model="dialog" max-width="450" @click:outside="close()">
    <v-card data-test="namespaceNewMember-dialog" class="bg-v-theme-surface">
      <v-card-title class="bg-primary"> Add member to namespace </v-card-title>

      <v-card-text>
        <v-text-field
          v-model="username"
          label="Username"
          :error-messages="usernameError"
          required
          variant="underlined"
          data-test="username-text"
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
        <v-btn text data-test="close-btn" @click="close()"> Close </v-btn>

        <v-btn color="primary" text data-test="add-btn" @click="addMember()">
          Add
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import hasPermission from "../../utils/permission";
import { useStore } from "../../store";
import { actions, authorizer } from "../../authorizer";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";

export default defineComponent({
  emits: ["update"],
  setup(props, ctx) {
    const store = useStore();
    const dialog = ref(false);

    const {
      value: username,
      errorMessage: usernameError,
      setErrors: setUsernameError,
      resetField: resetUsername,
    } = useField<string>("username", yup.string().required(), {
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

    const close = () => {
      username.value = "";
      selectedRole.value = "";
      setUsernameError("");
      setSelectedRoleError("");
      dialog.value = false;
    };

    const update = () => {
      ctx.emit("update");
      close();
    };

    const hasErrors = () => {
      if (selectedRole.value === "") {
        setSelectedRoleError("Select a role");
        return true;
      }

      if (username.value === "") {
        setUsernameError("This field is required");
        return true;
      }

      return false;
    };

    const resetFields = () => {
      resetUsername();
      resetSelectedRole();
    };

    const addMember = async () => {
      if (!hasErrors()) {
        try {
          await store.dispatch("namespaces/addUser", {
            username: username.value,
            tenant_id: store.getters["auth/tenant"],
            role: selectedRole.value,
          });

          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.namespaceNewMember,
          );
          update();
          resetFields();
        } catch (error: any) {
          if (error.response.status === 404) {
            setUsernameError("This username doesn't exist.");
          } else if (error.response.status === 409) {
            setUsernameError(
              "This user is already a member of this namespace.",
            );
          } else {
            store.dispatch(
              "snackbar/showSnackbarErrorAction",
              INotificationsError.namespaceNewMember,
            );
            throw new Error(error);
          }
        }
      }
    };

    return {
      items: ["administrator", "operator", "observer"],
      dialog,
      username,
      selectedRole,
      selectedRoleError,
      usernameError,
      hasAuthorization,
      addMember,
      close,
    };
  },
});
</script>
