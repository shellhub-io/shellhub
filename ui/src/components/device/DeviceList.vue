<template>
<fragment>

  <div class="d-flex pa-0 align-center">
    <h1>Devices</h1>
    <v-spacer/>
    <v-spacer/>
    <DeviceAdd/>
    <v-btn outlined @click="$store.dispatch('modals/showAddDevice', true)">Add Device</v-btn>
  </div>
    
  <v-card class="mt-2">
    <v-app-bar flat color="transparent"></v-app-bar>
    <v-divider></v-divider>

    <v-card-text class="pa-0">
      <v-data-table :headers="headers" :items=listDevices item-key="uid" :sort-by="['started_at']" :sort-desc="[true]" :items-per-page="10" :server-items-length='this.numberDevices' :options.sync='pagination'>

        <template v-slot:item.online="{ item }" >
          <v-icon color="success" v-if="item.online" @click.stop="detailsDevice(item)">check_circle</v-icon>
          <v-tooltip bottom v-else>
            <template #activator="{ on }">
              <v-icon v-on="on">check_circle</v-icon>
            </template>
            <span>last seen {{ item.last_seen | moment("from", "now") }}</span>
          </v-tooltip>
        </template>

        <template v-slot:item.uid="{ item }">
          <v-chip class="short" color="blue lighten-2">
            <span>{{ item.uid }}</span>
            <v-icon small dark right@click.stop v-clipboard="item.uid" v-clipboard:success="showCopySnack">mdi-content-copy none-text</v-icon>
          </v-chip>
        </template>

        <template v-slot:item.info.pretty_name="{ item }">
          <v-icon left>{{ deviceIcon[item.info.id] || 'fl-tux' }}</v-icon>
          {{ item.info.pretty_name }}
        </template>

        <template v-slot:item.identity.mac="{ item }">
          <code>{{ item.identity.mac }}</code>
        </template>

        <template v-slot:item.namespace="{ item }">
          <v-chip class="list-itens">
            {{ address(item) }}<v-icon small right @click.stop v-clipboard="() => address(item)" v-clipboard:success="showCopySnack">mdi-content-copy</v-icon>
          </v-chip>
        </template>

        <template v-slot:item.actions="{ item }">
          <v-icon class="icons" @click="detailsDevice(item)">info</v-icon>

          <TerminalDialog :uid="item.uid" v-if="item.online"></TerminalDialog>

          <v-icon @click="remove(item.uid)">delete</v-icon>
        </template>

      </v-data-table>
    </v-card-text>
    <v-snackbar v-model="copySnack" :timeout=3000>Device SSHID copied to clipboard</v-snackbar>
  </v-card>

</fragment>
</template>

<script>
import TerminalDialog from '@/components/terminal/TerminalDialog.vue';
import DeviceAdd from '@/components/device/DeviceAdd.vue';

export default {
  name: 'DeviceList',

  components: {
    TerminalDialog,
    DeviceAdd
  },

  data() {
    return {
      hostname: window.location.hostname,
      numberDevices: 0,
      listDevices: [],
      pagination: {},
      deviceIcon: {
        arch: 'fl-archlinux',
        ubuntu: 'fl-ubuntu'
      },
      copySnack: false,
      editName: '',
      headers: [
        {
          text: 'Online',
          value: 'online',
          align: 'center'
        },
        {
          text: 'Name',
          value: 'name'
        },
        {
          text: 'Operating System',
          value: 'info.pretty_name'
        },

        {
          text: 'SSHID',
          value: 'namespace',
          align: 'center'
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'center',
          sortable: false
        }
      ]
    };
  },

  watch: {
    pagination: {
      async handler () {
        const rowsPerPage = this.pagination.itemsPerPage
        const pageNumber = this.pagination.page

        this.data =  [{'per_page': rowsPerPage, 'page':pageNumber}]

        await this.$store.dispatch("devices/fetch", this.data[0]);
        this.listDevices = this.$store.getters["devices/list"];
        this.numberDevices = this.$store.getters["devices/getNumberDevices"]; 
      },
      deep: true
    },
  },

  methods: {
    detailsDevice(value){
      this.$router.push('/device/'+value.uid) 
    },

    address(item) {
      return `${item.namespace}.${item.name}@${this.hostname}`;
    },

    copy(device) {
      this.$clipboard(device.uid);
    },

    remove(uid) {
      if (confirm('Are you sure?')) {
        this.$store.dispatch('devices/remove', uid);
      }
    },

    showCopySnack() {
      this.copySnack = true;
    },

    save(item) {
      this.$store.dispatch('devices/rename', {
        uid: item.uid,
        name: this.editName
      });

      item.name = this.editName;
    }
  },
  
};
</script>
<style scoped>
.list-itens {
  font-family: monospace;
}

.icons{
  margin-right: 4px;
}

</style>