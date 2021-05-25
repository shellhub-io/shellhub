<template>
  <fragment>
    <v-form>
      <ValidationObserver
        ref="data"
        v-slot="{ passes }"
      >
        <v-row>
          <v-col>
            <h3>
              Webhook
            </h3>
          </v-col>

          <v-spacer />

          <v-col
            md="auto"
            class="ml-auto"
          >
            <v-btn
              v-if="!editWebhook"
              outlined
              @click="editWebhook = !editWebhook"
            >
              Change webhook
            </v-btn>

            <div
              v-if="editWebhook"
            >
              <v-btn
                class="mr-2"
                outlined
                @click="cancel"
              >
                Cancel
              </v-btn>

              <v-btn
                outlined
                @click="passes(save)"
              >
                Save
              </v-btn>
            </div>
          </v-col>
        </v-row>
        <div class="pl-4 pr-4">
          <v-row>
            <v-col>
              <ValidationProvider
                v-slot="{ errors }"
                ref="provider-url"
                vid="url"
                name="Priority"
                rules="required"
              >
                <v-text-field
                  v-model="webhookUrlField"
                  data-test="field-url"
                  label="url"
                  :disabled="!editWebhook"
                  :error-messages="errors"
                  :messages="helperText"
                />
              </ValidationProvider>
            </v-col>
          </v-row>
          <div>
            <v-row>
              <div class="ml-3 mb-1">
                <v-checkbox
                  v-model="webhookStatus"
                  label="Enable webhook"
                />
              </div>
            </v-row>
            <v-row class="mb-2">
              <div class="ml-3 mb-6">
                <p>
                  Set a webhook to receive an alert when you try to connect
                  to a device and start up the agent.
                </p>
              </div>
            </v-row>
          </div>
        </div>
        <v-spacer />
      </ValidationObserver>
    </v-form>
  </fragment>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

export default {
  name: 'SettingWebhook',

  components: {
    ValidationObserver,
    ValidationProvider,
  },

  data() {
    return {
      enableWebhook: false,
      missingWh: false,
      editWebhook: false,
      webhookUrlField: '',
      helperText: 'Examples: (http://domain.com, https://webhookurl:8080)',
    };
  },

  computed: {
    webhookStatus: {
      get() {
        return this.$store.getters['namespaces/webhookActive'];
      },

      async set(value) {
        const data = {
          status: value,
          tenant_id: this.tenant,
        };
        try {
          await this.$store.dispatch('namespaces/updateWebhookStatus', data);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.updateWebhookStatus);
        } catch (error) {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      },
    },

    tenant() {
      return localStorage.getItem('tenant');
    },
  },

  created() {
    this.setWebhookData();
  },

  methods: {
    async updateWebhook() {
      try {
        await this.$store.dispatch('namespaces/updateWebhook', {
          tenant_id: this.tenant,
          url: this.webhookUrlField,
        });

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.updateWebhook);
      } catch (error) {
        if (error.response.status === 400) {
          error.response.data.forEach((item) => {
            if (item.Name === 'url') {
              this.$refs.data.setErrors({
                url: this.$errors.form.invalid(item.Name, item.Param, item.Extra),
              });
            }
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.updateWebhook);
        }
      }
    },

    setWebhookData() {
      this.webhookUrlField = this.$store.getters['namespaces/webhookUrl'];
      this.enableWebhook = this.$store.getters['namespaces/webhookActive'];
    },

    save() {
      this.updateWebhook();
    },

    cancel() {
      this.setWebhookData();
      this.$refs.data.reset();
      this.editWebhook = !this.editWebhook;
    },
  },
};

</script>
