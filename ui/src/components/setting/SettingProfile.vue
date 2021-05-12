<template>
  <fragment>
    <v-container>
      <v-row
        align="center"
        justify="center"
        class="mt-4"
      >
        <v-col
          sm="8"
        >
          <ValidationObserver
            ref="obs"
            v-slot="{ passes }"
          >
            <v-row>
              <v-col>
                <h3>
                  Account
                </h3>
              </v-col>

              <v-spacer />

              <v-col
                md="auto"
                class="ml-auto"
              >
                <v-btn
                  v-if="!editDataStatus"
                  outlined
                  @click="editDataStatus = !editDataStatus"
                >
                  Change Data
                </v-btn>

                <div
                  v-if="editDataStatus"
                >
                  <v-btn
                    class="mr-2"
                    outlined
                    @click="cancel('data')"
                  >
                    Cancel
                  </v-btn>

                  <v-btn
                    outlined
                    @click="passes(updateData)"
                  >
                    Save
                  </v-btn>
                </div>
              </v-col>
            </v-row>

            <div
              class="mt-4 pl-4 pr-4"
            >
              <ValidationProvider
                v-slot="{ errors }"
                ref="providerName"
                vid="name"
                name="Priority"
                rules="required"
              >
                <v-text-field
                  v-model="name"
                  label="Name"
                  :error-messages="errors"
                  required
                  :disabled="!editDataStatus"
                  data-test="name-text"
                />
              </ValidationProvider>

              <ValidationProvider
                v-slot="{ errors }"
                ref="providerUsername"
                vid="username"
                name="Priority"
                rules="required"
              >
                <v-text-field
                  v-model="username"
                  label="Username"
                  :error-messages="errors"
                  required
                  :disabled="!editDataStatus"
                  data-test="username-text"
                />
              </ValidationProvider>

              <ValidationProvider
                v-slot="{ errors }"
                ref="providerEmail"
                name="Priority"
                vid="email"
                rules="required|email"
              >
                <v-text-field
                  v-model="email"
                  class="mt-1 mb-4"
                  label="E-mail"
                  :error-messages="errors"
                  required
                  :disabled="!editDataStatus"
                  data-test="email-text"
                />
              </ValidationProvider>
            </div>
          </ValidationObserver>

          <v-divider class="mt-6" />
          <v-divider class="mb-6" />

          <ValidationObserver
            ref="pass"
            v-slot="{ passes }"
          >
            <v-row>
              <v-col>
                <h3>
                  Password
                </h3>
              </v-col>

              <v-spacer />

              <v-col
                md="auto"
                class="ml-auto"
              >
                <v-btn
                  v-if="!editPasswordStatus"
                  outlined
                  @click="editPasswordStatus = !editPasswordStatus"
                >
                  Change Password
                </v-btn>

                <div
                  v-if="editPasswordStatus"
                >
                  <v-btn
                    class="mr-2"
                    outlined
                    @click="cancel('password')"
                  >
                    Cancel
                  </v-btn>

                  <v-btn
                    outlined
                    @click="passes(updatePassword)"
                  >
                    Save
                  </v-btn>
                </div>
              </v-col>
            </v-row>

            <div
              class="mt-4 pl-4 pr-4"
            >
              <ValidationProvider
                v-slot="{ errors }"
                ref="providerCurrentPassword"
                name="Priority"
                rules="required"
                vid="currentPassword"
              >
                <v-text-field
                  v-model="currentPassword"
                  type="password"
                  label="Current password"
                  class="mb-4"
                  :error-messages="errors"
                  required
                  :disabled="!editPasswordStatus"
                  data-test="password-text"
                />
              </ValidationProvider>

              <ValidationProvider
                v-slot="{ errors }"
                ref="providerNewPassword"
                name="Priority"
                rules="required|password|comparePasswords:@currentPassword"
                vid="newPassword"
              >
                <v-text-field
                  v-model="newPassword"
                  type="password"
                  label="New password"
                  class="mb-4"
                  :error-messages="errors"
                  required
                  :disabled="!editPasswordStatus"
                  data-test="newPassword-text"
                />
              </ValidationProvider>

              <ValidationProvider
                v-slot="{ errors }"
                ref="providerConfirmPassword"
                rules="required|confirmed:newPassword"
                name="confirm"
              >
                <v-text-field
                  v-model="newPasswordConfirm"
                  label="Confirm new password"
                  type="password"
                  class="mb-4"
                  :error-messages="errors"
                  required
                  :disabled="!editPasswordStatus"
                  data-test="confirmNewPassword-text"
                />
              </ValidationProvider>
            </div>
          </ValidationObserver>
        </v-col>
      </v-row>
    </v-container>
  </fragment>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

export default {
  name: 'SettingProfile',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  data() {
    return {
      name: '',
      username: '',
      email: '',
      currentPassword: '',
      newPassword: '',
      newPasswordConfirm: '',
      editDataStatus: false,
      editPasswordStatus: false,
      show: false,
    };
  },

  computed: {
    tenant() {
      return this.$store.getters['auth/tenant'];
    },
  },

  created() {
    this.setData();
  },

  methods: {
    cancel(statusExit) {
      if (statusExit === 'data') {
        this.setData();
        this.$refs.obs.reset();
        this.editDataStatus = !this.editDataStatus;
      } else if (statusExit === 'password') {
        this.currentPassword = '';
        this.newPassword = '';
        this.newPasswordConfirm = '';
        this.$refs.pass.reset();
        this.editPasswordStatus = !this.editPasswordStatus;
      }
    },

    setData() {
      this.name = this.$store.getters['auth/currentName'];
      this.username = this.$store.getters['auth/currentUser'];
      this.email = this.$store.getters['auth/email'];
    },

    enableEdit(form) {
      if (form === 'data') {
        this.editDataStatus = !this.editDataStatus;
      } else if (form === 'password') {
        this.editPasswordStatus = !this.editPasswordStatus;
      }
    },

    async updateData() {
      const data = {
        id: this.$store.getters['auth/id'],
        name: this.name,
        username: this.username,
        email: this.email,
      };

      try {
        await this.$store.dispatch('users/put', data);
        this.$store.dispatch('auth/changeUserData', data);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.profileData);
        this.enableEdit('data');
      } catch (error) {
        if (error.response.status === 409) { // user data already exists
          error.response.data.forEach((item) => {
            if (item.Name === 'username') {
              this.$refs.obs.setErrors({
                username: item.Message,
              });
            }
            if (item.Name === 'email') {
              this.$refs.obs.setErrors({
                email: item.Message,
              });
            }
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      }
    },

    async updatePassword() {
      const data = {
        id: this.$store.getters['auth/id'],
        currentPassword: this.currentPassword,
        newPassword: this.newPassword,
      };

      try {
        await this.$store.dispatch('users/put', data);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.profilePassword);
        this.enableEdit('password');
      } catch (error) {
        if (error.response.status === 403) { // failed password
          this.$refs.pass.setErrors({
            currentPassword: ['Your password doesn\'t match'],
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      }
    },
  },

};
</script>
