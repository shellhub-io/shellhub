<template>
  <fragment>
    <v-tooltip
      v-if="addUser"
      bottom
      :disabled="hasAuthorization"
    >
      <template #activator="{ on }">
        <div v-on="on">
          <v-btn
            :disabled="!hasAuthorization"
            class="mr-2"
            color="primary"
            data-test="add-btn"
            @click="setShowDialog()"
          >
            Add Member
          </v-btn>
        </div>
      </template>

      <span>
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-tooltip
      v-else
      bottom
      :disabled="hasAuthorization"
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-list-item-title data-test="edit-list">
            Edit
          </v-list-item-title>
        </span>

        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            left
            data-test="edit-icon"
            v-on="on"
          >
            mdi-pencil
          </v-icon>
        </span>
      </template>

      <span>
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="showDialog"
      max-width="450"
      @click:outside="close"
    >
      <v-card data-test="namespaceNewMember-dialog">
        <v-card-title class="headline primary">
          {{ addUser ? 'Add member to namespace' : 'Update member role' }}
        </v-card-title>

        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-text class="caption mb-0">
            <ValidationProvider
              v-slot="{ errors }"
              ref="providerUsername"
              name="Username"
              rules="required"
              vid="username"
            >
              <v-text-field
                v-model="memberLocal.username"
                :disabled="!addUser"
                label="Username"
                :error-messages="errors"
                require
                data-test="username-text"
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-text class="caption mb-0 pt-0">
            <ValidationProvider
              v-slot="{ errors }"
              ref="providerRole"
              vid="role"
              name="role"
              rules="required"
            >
              <v-row align="center">
                <v-col cols="12">
                  <v-select
                    v-model="memberLocal.selectedRole"
                    :items="items"
                    label="Role"
                    :error-messages="errors"
                    require
                    data-test="role-select"
                  />
                </v-col>
              </v-row>
            </ValidationProvider>
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn
              text
              data-test="dialogClose-btn"
              @click="close()"
            >
              Close
            </v-btn>

            <v-btn
              v-if="addUser"
              color="primary"
              text
              data-test="dialogAdd-btn"
              @click="passes(addMember)"
            >
              Add
            </v-btn>

            <v-btn
              v-else
              color="primary"
              text
              data-test="dialogEdit-btn"
              @click="passes(editMember)"
            >
              Edit
            </v-btn>
          </v-card-actions>
        </ValidationObserver>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

import hasPermission from '@/components/filter/permission';

export default {
  name: 'NamespaceNewMemberComponent',

  filters: { hasPermission },

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    member: {
      type: Object,
      required: false,
      default: Object,
    },

    addUser: {
      type: Boolean,
      required: true,
    },

    show: {
      type: Boolean,
      required: false,
      default: false,
    },
  },

  data() {
    return {
      username: '',
      selectedRole: '',
      memberLocal: [],
      items: ['administrator', 'operator', 'observer'],
    };
  },

  computed: {
    showDialog: {
      get() {
        return this.show && this.hasAuthorization;
      },

      set(value) {
        this.$emit('show', value);
      },
    },

    hasAuthorization() {
      const ownerID = this.$store.getters['namespaces/get'].owner;
      if (this.member.id === ownerID) {
        return false;
      }

      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        let action = '';
        if (this.addUser) action = 'addMember';
        else action = 'removeMember';

        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.namespace[action],
        ) && this.member.role !== role;
      }

      return false;
    },
  },

  async created() {
    await this.setLocalVariable();
  },

  async updated() {
    await this.setLocalVariable();
  },

  methods: {
    setLocalVariable() {
      if (this.addUser) {
        this.memberLocal = {
          id: '',
          username: '',
          selectedRole: '',
        };
      } else {
        this.memberLocal = { ...this.member, selectedRole: this.member.role };
      }
    },

    async addMember() {
      try {
        await this.$store.dispatch('namespaces/addUser', {
          username: this.memberLocal.username,
          tenant_id: this.$store.getters['auth/tenant'],
          role: this.memberLocal.selectedRole,
        });

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceNewMember);
        this.update();
      } catch (error) {
        if (error.response.status === 404) {
          this.$refs.obs.setErrors({
            username: 'The username doesn\'t exist.',
          });
        } else if (error.response.status === 409) {
          this.$refs.obs.setErrors({
            username: 'The username has already been added to namespace.',
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.namespaceNewMember);
        }
      }
    },

    async editMember() {
      try {
        await this.$store.dispatch('namespaces/editUser', {
          user_id: this.memberLocal.id,
          tenant_id: this.$store.getters['auth/tenant'],
          role: this.memberLocal.selectedRole,
        });

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceEditMember);
        this.update();
      } catch (error) {
        if (error.response.status === 400) {
          this.$refs.obs.setErrors({
            username: 'The user isn\'t linked to the namespace.',
          });
        } else if (error.response.status === 403) {
          this.$refs.obs.setErrors({
            role: 'You don\'t have permission to assign a role to the user.',
          });
        } else if (error.response.status === 404) {
          this.$refs.obs.setErrors({
            username: 'The username doesn\'t exist.',
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.namespaceEditMember);
        }
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    setShowDialog() {
      this.$emit('update:show', true);
    },

    close() {
      this.$emit('update:show', false);
      this.$refs.obs.reset();
    },
  },
};

</script>
