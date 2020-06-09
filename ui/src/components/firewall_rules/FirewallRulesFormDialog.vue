<template>
  <fragment>
    <v-btn
      v-if="createRule"
      outlined
      @click="dialog = !dialog"
    >
      Add Rule
    </v-btn>
    <v-tooltip
      v-else
      bottom
    >
      <template #activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          edit
        </v-icon>
      </template>
      <span>Edit</span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        <v-card-title
          v-if="createRule"
          class="headline grey lighten-2 text-center"
        >
          New Rule
        </v-card-title>
        <v-card-title
          v-else
          class="headline grey lighten-2 text-center"
        >
          Edit Rule
        </v-card-title>

        <v-card-text>
          <v-layout
            justify-space-between
            align-center
          >
            <v-flex>
              <v-card :elevation="0">
                <v-card-text class="v-label theme--light pl-0">
                  Active
                </v-card-text>
              </v-card>
            </v-flex>

            <v-flex
              xs2
            >
              <v-card
                :elevation="0"
              >
                <v-switch
                  v-model="ruleFirewallLocal.active"
                />
              </v-card>
            </v-flex>
          </v-layout>

          <v-text-field
            v-model="ruleFirewallLocal.priority"
            label="Priority"
            type="number"
          />

          <v-flex>
            <v-select
              v-model="ruleFirewallLocal.action"
              item-text="name"
              item-value="id"
              :items="state"
              label="Action"
            />
          </v-flex>

          <v-text-field
            v-model="ruleFirewallLocal.source_ip"
            label="Source IP"
          />

          <v-text-field
            v-model="ruleFirewallLocal.username"
            label="Username"
          />

          <v-text-field
            v-model="ruleFirewallLocal.hostname"
            label="Hostname"
          />
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            @click="dialog=!dialog"
          >
            Cancel
          </v-btn>

          <v-btn
            v-if="createRule"
            text
            @click="create()"
          >
            Create
          </v-btn>

          <v-btn
            v-else
            text
            @click="edit()"
          >
            Edit
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'FirewallEdit',

  props: {
    firewallRule: {
      type: Object,
      required: false,
      default: Object,
    },
    createRule: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
      state: [{
        id: 'allow',
        name: 'allow',
      },
      {
        id: 'deny',
        name: 'deny',
      }],
      ruleFirewallLocal: [],
    };
  },

  async created() {
    if (this.createRule) {
      this.ruleFirewallLocal = {
        active: false,
        priority: '',
        action: '',
        source_ip: '',
        username: '',
        hostname: '',
      };
    } else {
      this.ruleFirewallLocal = await { ...this.firewallRule };
    }
  },

  methods: {
    async create() {
      await this.$store.dispatch('firewallrules/post', this.ruleFirewallLocal);
      this.update();
    },

    async edit() {
      await this.$store.dispatch('firewallrules/put', this.ruleFirewallLocal);
      this.update();
    },

    update() {
      this.$emit('update');
      this.dialog = !this.dialog;
    },
  },
};
</script>
