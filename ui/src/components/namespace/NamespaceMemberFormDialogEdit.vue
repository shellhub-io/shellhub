<template>
  <fragment>
    <v-list-item-icon class="mr-0">
      <v-icon
        left
        data-test="remove-icon"
        v-text="'mdi-pencil'"
      />
    </v-list-item-icon>

    <v-list-item-content>
      <v-list-item-title
        class="text-left"
        data-test="edit-title"
        v-text="'Edit'"
      />
    </v-list-item-content>

    <v-dialog
      v-model="showDialog"
      max-width="450"
      @click:outside="close"
    >
      <v-card data-test="namespaceNewMember-dialog">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'Update member role'"
        />

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
                :disabled="true"
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
              data-test="close-btn"
              @click="close()"
              v-text="'Close'"
            />

            <v-btn
              color="primary"
              text
              data-test="edit-btn"
              @click="passes(editMember)"
              v-text="'Edit'"
            />
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
        return this.show;
      },

      set(value) {
        this.$emit('show', value);
      },
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
      this.memberLocal = { ...this.member, selectedRole: this.member.role };
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

    close() {
      this.$emit('update:show', false);
      this.$refs.obs.reset();
    },
  },
};

</script>
