<template>
  <v-form>
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
            ref="data"
            v-slot="{ passes }"
          >
            <div
              class="mt-6 pl-4 pr-4"
            >
              <ValidationProvider
                v-slot="{ errors }"
                ref="providerName"
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

              <v-btn
                v-if="!editDataStatus"
                class="mr-2"
                outlined
                @click="editDataStatus = !editDataStatus"
              >
                Change Data
              </v-btn>

              <div
                v-if="editDataStatus"
              >
                <v-btn
                  class="mr-2 mt-4"
                  outlined
                  @click="editDataStatus = !editDataStatus"
                >
                  Cancel
                </v-btn>

                <v-btn
                  class="mr-2 mt-4"
                  outlined
                  @click="passes(updateData)"
                >
                  Save
                </v-btn>
              </div>
            </div>
          </ValidationObserver>

          <ValidationObserver
            ref="pass"
            v-slot="{ passes }"
          >
            <v-divider class="mt-6" />
            <v-divider class="mb-6" />

            <div
              class="mt-6 pl-4 pr-4"
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

              <v-btn
                v-if="!editPasswordStatus"
                class="mr-2"
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
                  @click="editPasswordStatus = !editPasswordStatus"
                >
                  Cancel
                </v-btn>

                <v-btn
                  class="mr-2"
                  outlined
                  @click="passes(updatePassword)"
                >
                  Save
                </v-btn>
              </div>
            </div>
          </ValidationObserver>
        </v-col>
      </v-row>
    </v-container>
  </v-form>
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
    setData() {
      this.username = this.$store.getters['auth/currentUser'];
      this.email = this.$store.getters['auth/email'];
    },

    enableEdit() {
      if (this.editDataStatus) {
        this.editDataStatus = !this.editDataStatus;
      } else if (this.editPasswordStatus) {
        this.editPasswordStatus = !this.editPasswordStatus;
      }
    },

    async updateData() {
      const data = {
        username: this.username,
        email: this.email,
      };

      try {
        await this.$store.dispatch('users/put', data);
        this.$store.dispatch('auth/changeUserData', data);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.profileData);
        this.enableEdit();
      } catch (error) {
        if (error.response.status === 409) { // user data already exists
          error.response.data.forEach((item) => {
            if (item.Name === 'username') {
              this.$refs.data.setErrors({
                username: item.Message,
              });
            }
            if (item.Name === 'email') {
              this.$refs.data.setErrors({
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
        currentPassword: this.currentPassword,
        newPassword: this.newPassword,
      };

      try {
        await this.$store.dispatch('users/put', data);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.profilePassword);
        this.enableEdit();
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
