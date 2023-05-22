<template>
  <div>
    <v-list-item
      @click="showDialog = true"
      v-bind="$props"
      :disabled="notHasAuthorization"
    >
      <div class="d-flex align-center">
        <div data-test="namespace-edit-icon" class="mr-2">
          <v-icon> mdi-pencil </v-icon>
        </div>

        <v-list-item-title data-test="namespace-edit-title">
          Edit
        </v-list-item-title>
      </div>
    </v-list-item>

    <v-dialog max-width="450" v-model="showDialog">
      <v-card class="bg-v-theme-surface">
        <v-card-title class="text-h5 pa-4 bg-primary">
          Update member role
        </v-card-title>
        <v-divider />

        <v-card-text class="mt-4 mb-0 pb-1">
          <v-text-field
            v-model="memberLocal.username"
            :disabled="true"
            variant="underlined"
            label="Username"
            :error-messages="errorMessage"
            require
            data-test="username-text"
          />

          <v-row align="center">
            <v-col cols="12">
              <v-select
                v-model="memberLocal.selectedRole"
                :items="items"
                label="Role"
                variant="underlined"
                :error-messages="errorMessage"
                require
                data-test="role-select"
              />
            </v-col>
          </v-row>
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" data-test="close-btn" @click="close()">
            Close
          </v-btn>

          <v-btn
            color="primary"
            variant="text"
            data-test="edit-btn"
            @click="editMember()"
          >
            Edit
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from "vue";
import axios, { AxiosError } from "axios";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { IMember } from "../../interfaces/IMember";
import { useStore } from "../../store";
import handleError from "../../utils/handleError";

export default defineComponent({
  props: {
    member: {
      type: Object as () => IMember,
      required: false,
      default: {} as IMember,
    },
    show: {
      type: Boolean,
      required: false,
    },
    notHasAuthorization: {
      type: Boolean,
      default: false,
    },
    style: {
      type: [String, Object],
      default: undefined,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const store = useStore();
    const showDialog = ref(false);
    const memberLocal = ref({} as IMember);
    const errorMessage = ref("");

    const setLocalVariable = () => {
      memberLocal.value = { ...props.member, selectedRole: props.member.role };
    };

    onMounted(() => {
      setLocalVariable();
    });

    const close = () => {
      setLocalVariable();
      showDialog.value = false;
    };

    const update = () => {
      ctx.emit("update");
      close();
    };

    const editMember = async () => {
      try {
        await store.dispatch("namespaces/editUser", {
          user_id: memberLocal.value.id,
          tenant_id: store.getters["auth/tenant"],
          role: memberLocal.value.selectedRole,
        });

        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.namespaceEditMember,
        );
        update();
      } catch (error: unknown) {
        if (axios.isAxiosError(error)) {
          const axiosError = error as AxiosError;
          if (axiosError.response?.status === 400) {
            errorMessage.value = "The user isn't linked to the namespace.";
          } else if (axiosError.response?.status === 403) {
            errorMessage.value = "You don't have permission to assign a role to the user.";
          } else if (axiosError.response?.status === 404) {
            errorMessage.value = "The username doesn't exist.";
          }
        } else {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.namespaceEditMember,
          );
          handleError(error);
        }
      }
    };

    return {
      showDialog,
      items: ["administrator", "operator", "observer"],
      memberLocal,
      errorMessage,
      close,
      editMember,
    };
  },
});
</script>
