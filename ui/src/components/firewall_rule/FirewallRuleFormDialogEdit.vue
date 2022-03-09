<template>
  <fragment>
    <v-list-item-icon class="mr-0">
      <v-icon
        left
        data-test="edit-icon"
        v-text="'edit'"
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
      max-width="520"
      @click:outside="close"
    >
      <v-card data-test="firewallRuleForm-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'Edit Firewall Rule'"
        />

        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-text>
            <v-row>
              <v-col>
                <v-select
                  v-model="ruleFirewallLocal.status"
                  :items="ruleStatus"
                  item-text="text"
                  item-value="type"
                  label="Rule status"
                  required
                />
              </v-col>

              <v-col>
                <ValidationProvider
                  v-slot="{ errors }"
                  name="Priority"
                  rules="required|integer"
                >
                  <v-text-field
                    v-model="ruleFirewallLocal.priority"
                    class="mb-2"
                    label="Rule priority"
                    type="number"
                    :error-messages="errors"
                    required
                    data-test="priority-field"
                  />
                </ValidationProvider>
              </v-col>

              <v-col>
                <ValidationProvider
                  v-slot="{ errors }"
                  name="Action"
                  rules="required"
                >
                  <v-select
                    v-model="ruleFirewallLocal.policy"
                    :items="state"
                    item-text="name"
                    item-value="id"
                    label="Rule policy"
                    :error-messages="errors"
                    required
                    data-test="action-field"
                  />
                </ValidationProvider>
              </v-col>
            </v-row>

            <v-row class="mt-1 mb-1 px-3">
              <v-select
                v-model="choiceIP"
                label="Source IP access restriction"
                :items="sourceIPFieldChoices"
                item-text="filterText"
                item-value="filterName"
                data-test="source_ip-field"
              />
            </v-row>

            <ValidationProvider
              v-if="choiceIP==='ipDetails'"
              v-slot="{ errors }"
              name="Source IP"
              rules="required"
            >
              <v-text-field
                v-model="ipField"
                label="Rule source IP"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <v-row class="mt-1 mb-1 px-3">
              <v-select
                v-model="choiceUsername"
                label="Device username access restriction"
                :items="usernameFieldChoices"
                item-text="filterText"
                item-value="filterName"
                data-test="username-field"
              />
            </v-row>

            <ValidationProvider
              v-if="choiceUsername==='username'"
              v-slot="{ errors }"
              name="Username"
              rules="required"
            >
              <v-text-field
                v-model="usernameField"
                label="Username access restriction"
                placeholder="Username used during the connection"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <v-row class="mt-1 mb-1 px-3">
              <v-select
                v-model="choiceFilter"
                label="Device access restriction"
                :items="filterFieldChoices"
                item-text="filterText"
                item-value="filterName"
                data-test="filter-field"
              />
            </v-row>

            <ValidationProvider
              v-if="choiceFilter==='hostname'"
              v-slot="{ errors }"
              name="Hostname"
              rules="required"
            >
              <v-text-field
                v-model="hostnameField"
                label="Device hostname access restriction"
                placeholder="Device hostname used during the connection"
                :error-messages="errors"
                data-test="hostname-field"
                required
              />
            </ValidationProvider>

            <ValidationProvider
              v-if="choiceFilter==='tags'"
              name="Tags"
              rules="required"
            >
              <v-select
                v-model="tagChoices"
                :items="tagNames"
                data-test="tags-selector"
                attach
                chips
                label="Rule device tag restriction"
                :rules="[validateLength]"
                :error-messages="errMsg"
                :menu-props="{ top: true, maxHeight: 150, offsetY: true }"
                multiple
                required
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions>
            <v-spacer />

            <v-btn
              text
              data-test="cancel-btn"
              @click="close"
              v-text="'Cancel'"
            />

            <v-btn
              text
              data-test="edit-btn"
              @click="passes(edit)"
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

export default {
  name: 'FirewallRuleFormDialogEdit',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    firewallRule: {
      type: Object,
      required: false,
      default: Object,
    },

    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      choiceUsername: 'all',
      choiceFilter: 'all',
      choiceIP: 'all',
      validateLength: true,
      usernameField: '',
      hostnameField: '',
      ipField: '',
      tagChoices: [''],
      errMsg: '',
      sourceIPFieldChoices: [
        {
          filterName: 'all',
          filterText: 'Define source IP to all devices',
        },
        {
          filterName: 'ipDetails',
          filterText: 'Restrict source IP through a regexp',
        },
      ],
      filterFieldChoices: [
        {
          filterName: 'all',
          filterText: 'Define rule to all devices',
        },
        {
          filterName: 'hostname',
          filterText: 'Restrict rule with a regexp for hostname',
        },
        {
          filterName: 'tags',
          filterText: 'Restrict rule by device tags',
        },
      ],
      usernameFieldChoices: [
        {
          filterName: 'all',
          filterText: 'Define rule to all users',
        },
        {
          filterName: 'username',
          filterText: 'Restrict access using a regexp for username',
        },
      ],
      ruleStatus: [
        {
          type: 'active',
          text: 'Active',
        },
        {
          type: 'inactive',
          text: 'Inactive',
        },
      ],
      state: [{
        id: 'allow',
        name: 'allow',
      },
      {
        id: 'deny',
        name: 'deny',
      }],
      ruleFirewallLocal: {
        priority: 0,
        source_ip: '',
        filter: {},
        username: '',
        status: '',
        policy: '',
      },
    };
  },

  computed: {
    tagNames() {
      return this.$store.getters['tags/list'];
    },

    showDialog: {
      get() {
        return this.show;
      },

      set(value) {
        this.$emit('update:show', value);
      },
    },
  },

  watch: {
    showDialog(val) {
      if (val) {
        this.setLocalVariable();
      }
    },

    tagChoices(list) {
      if (list.length > 3) {
        this.validateLength = false;
        this.$nextTick(() => this.tagChoices.pop());
        this.errMsg = 'The maximum capacity has reached';
      } else if (list.length === 0) {
        this.validateLength = false;
        this.errMsg = 'You must choose at least one tag';
      } else if (list.length <= 2) {
        this.validateLength = true;
        this.errMsg = '';
      }
    },
  },

  methods: {
    selectRestriction() {
      if (this.choiceUsername === 'all') {
        this.ruleFirewallLocal = {
          ...this.ruleFirewallLocal,
          username: '.*',
        };
      } else if (this.choiceUsername === 'username') {
        this.ruleFirewallLocal = {
          ...this.ruleFirewallLocal,
          username: this.usernameField,
        };
      }

      let filter;

      if (this.choiceIP === 'all') {
        this.ruleFirewallLocal = {
          ...this.ruleFirewallLocal,
          source_ip: '.*',
        };
      } else if (this.choiceIP === 'ipDetails') {
        this.ruleFirewallLocal = {
          ...this.ruleFirewallLocal,
          source_ip: this.ipField,
        };
      }

      switch (this.choiceFilter) {
      case 'all': {
        filter = {
          hostname: '.*',
        };
        break;
      }
      case 'hostname': {
        filter = {
          hostname: this.hostnameField,
        };
        break;
      }
      case 'tags': {
        filter = {
          tags: this.tagChoices,
        };
        break;
      }
      default:
      }

      this.ruleFirewallLocal = {
        ...this.ruleFirewallLocal,
        filter,
      };
    },

    setLocalVariable() {
      let status = 'inactive';

      const {
        action, active,
        username, filter, ...fr
      } = this.firewallRule;

      if (fr.source_ip !== '.*') {
        this.choiceIP = 'ipDetails';
        this.ipField = fr.source_ip;
      } else {
        this.choiceIP = 'all';
        this.ipField = '';
      }

      if (username !== '.*') {
        this.choiceUsername = 'username';
        this.usernameField = username;
      } else {
        this.choiceUsername = 'all';
        this.usernameField = '';
      }

      if (!!filter.hostname && filter.hostname !== '.*') {
        this.choiceFilter = 'hostname';
        this.hostnameField = filter.hostname;
      } else if (filter.tags) {
        this.choiceFilter = 'tags';
        this.tagChoices = filter.tags;
      }

      if (active) {
        status = 'active';
      }

      let filtObj = {};

      if (this.choiceFilter === 'hostname') {
        filtObj = { hostname: this.hostnameField };
      } else if (this.choiceFilter === 'tags') {
        filtObj = { tags: this.tagChoices };
      }

      this.ruleFirewallLocal = {
        ...fr,
        username,
        filter: filtObj,
        status,
        policy: action,
      };
    },

    async edit() {
      this.selectRestriction();

      try {
        await this.$store.dispatch('firewallrules/put', this.ruleFirewallLocal);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.firewallRuleEditing);
        this.update();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.firewallRuleEditing);
      }
    },

    resetChoices() {
      this.choiceUsername = 'all';
      this.choiceFilter = 'all';
      this.choiceIP = 'all';
      this.validateLength = true;
      this.errMsg = '';
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.$emit('update:show', false);
      this.tagChoices = [''];
      this.resetChoices();
      this.$refs.obs.reset();
    },
  },
};
</script>
