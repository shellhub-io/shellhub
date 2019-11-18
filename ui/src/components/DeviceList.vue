<template>
<fragment>
    <v-card>
        <v-app-bar flat color="transparent">
            <v-toolbar-title>Device Fleet</v-toolbar-title>
        </v-app-bar>
        <v-divider></v-divider>
        <v-card-text class="pa-0">
            <v-data-table :headers="headers" :items="devices" item-key="uid" hide-default-footer>
                <template v-slot:item.online="{ item }">
                    <v-icon color="success" v-if="item.online">check_circle</v-icon>
                    <v-tooltip bottom v-else>
                        <template #activator="{ on }">
                            <v-icon v-on="on">check_circle</v-icon>
                        </template>
                        <span>last seen {{ item.last_seen | moment("from", "now") }}</span>
                    </v-tooltip>
                </template>

                <template v-slot:item.uid="{ item }">
                    <v-chip>
                        {{ item.uid }}
                        <v-icon small right @click v-clipboard="item.uid" v-clipboard:success="showCopySnack">mdi-content-copy</v-icon>
                    </v-chip>
                </template>

                <template v-slot:item.name="{ item }">
                    <v-edit-dialog :return-value="editName" large @open="editName = item.name" @save="save(item)">
                        <v-text-field slot="input" v-model="editName" label="Edit" single-line>
                        </v-text-field>
                        <v-icon small left>mdi-file-edit</v-icon>
                        {{ item.name }}
                    </v-edit-dialog>
                </template>

                <template v-slot:item.attributes.pretty_name="{ item }">
                    <v-icon left>{{ deviceIcon[item.attributes.id] || 'fl-tux' }}</v-icon>
                    {{ item.attributes.pretty_name }}
                </template>

                <template v-slot:item.identity.mac="{ item }">
                    <code>{{ item.identity.mac }}</code>
                </template>

                <template v-slot:item.namespace="{ item }">
                  <kbd>{{ address(item) }}<v-icon color="white" small right @click v-clipboard="() => address(item)" v-clipboard:success="showCopySnack">mdi-content-copy</v-icon>
                  </kbd>
                </template>

                <template v-slot:item.actions="{ item }">
                    <TerminalDialog :uid="item.uid" v-if="item.online"></TerminalDialog>

                    <v-icon @click="remove(item.uid)">
                        delete
                    </v-icon>
                </template>
            </v-data-table>
        </v-card-text>
        <v-snackbar v-model="copySnack" :timeout=3000>Device SSHID copied to clipboard</v-snackbar>
    </v-card>
</fragment>
</template>

<script>
import TerminalDialog from "@/components/TerminalDialog.vue";

export default {
  components: {
    TerminalDialog
  },

  created() {
    this.$store.dispatch("devices/fetch");
  },

  computed: {
    devices() {
      return this.$store.getters["devices/list"];
    }
  },

  methods: {
    address(item) {
      return `${item.namespace}.${item.name}@${this.hostname}`;
    },

    copy(device) {
      this.$clipboard(device.uid);
    },

    remove(uid) {
      if (confirm("Are you sure?")) {
        this.$store.dispatch("devices/remove", uid);
      }
    },

    showCopySnack() {
      this.copySnack = true;
    },

    save(item) {
      this.$store.dispatch("devices/rename", {
        uid: item.uid,
        name: this.editName
      });

      item.name = this.editName;
    }
  },

  data() {
    return {
      hostname: window.location.hostname,
      deviceIcon: {
        arch: "fl-archlinux",
        ubuntu: "fl-ubuntu"
      },
      copySnack: false,
      editName: "",
      headers: [
        {
          text: "Online",
          value: "online",
          align: "center"
        },
        {
          text: "Name",
          value: "name"
        },
        {
          text: "Operating System",
          value: "attributes.pretty_name"
        },

        {
          text: "MAC",
          value: "identity.mac",
          align: "center"
        },
        {
          text: "SSHID",
          value: "namespace",
          align: "center"
        },
        {
          text: "Actions",
          value: "actions",
          align: "center",
          sortable: false
        }
      ]
    };
  }
};
</script>
