<template>
<fragment>

  <div class="d-flex pa-0 align-center">
  <h1>Device Feature</h1>
  <v-spacer/>
  <v-spacer/>
  <!-- <AddDevice/> -->
  <!-- <v-btn outlined @click="$store.dispatch('modals/showAddDevice', true)">Add Device</v-btn> -->
  </div>
    <v-card class="mt-2">
        <v-app-bar flat color="transparent">
        </v-app-bar>
        <v-divider></v-divider>
        <v-card-text>
          
          <div class="item">
            <div class="item-name">Uid: </div>
            <div class="item-description">{{device.uid}}</div>
          </div>

          <div class="item">
            <div class="item-name">Name: </div>
            <!-- <v-icon small left>mdi-file-edit</v-icon>  -->
            <div class="item-description">{{device.name}}</div>
          </div>

          <div class="item">
            <div class="item-name">Mac: </div>
            <div class="item-description">{{device.identity}}</div>
          </div>

          <div class="item"> 
            <div class="item-name">Operating System: </div>
            <div class="item-description">{{device.attributes}}</div>
          </div>

          <div class="item">
            <div class="item-name">Public Key: </div>
            <div>{{device.public_key}}</div>
          </div>

          <div class="item">
            <div class="item-name">Tenant Id: </div>
            <div class="item-description">{{device.tenant_id}}</div>
          </div>

          <div class="item">
            <div class="item-name">Last Seen: </div>
            <div class="item-description">{{device.last_seen}}</div>
          </div>

          <div class="item">
            <div class="item-name">Online: </div>
            <div class="item-description">{{device.online}}</div>
          </div>

          <div class="item">
            <div class="item-name">Namespace: </div>
            <div class="item-description">{{device.namespace}}</div>
          </div>

        </v-card-text>
        <v-snackbar v-model="copySnack" :timeout=3000>Device SSHID copied to clipboard</v-snackbar>
    </v-card>
</fragment>
</template>

<script>
// import TerminalDialog from "@/components/TerminalDialog.vue";
// import AddDevice from "@/components/AddDevice.vue";

export default {
  name: "DeviceFeature",


  components: {
    // TerminalDialog,
    // AddDevice

  },

  created() {
      this.uid = this.$route.params.id
      this.$store.dispatch("devices/get", this.uid);
  },

  computed: {
    device() {
      return this.$store.getters["devices/get"];
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
      uid: '',
      hostname: window.location.hostname,
      deviceIcon: {
        arch: "fl-archlinux",
        ubuntu: "fl-ubuntu"
      },
      copySnack: false,
      editName: "",
      headers: [
        // {
        //   text: "Online",
        //   value: "online",
        //   align: "center"
        // },
        {
          text: "Name",
          value: "name"
        },
        // {
        //   text: "Operating System",
        //   value: "attributes.pretty_name"
        // },

        // {
        //   text: "SSHID",
        //   value: "namespace",
        //   align: "center"
        // },
        // {
        //   text: "Actions",
        //   value: "actions",
        //   align: "center",
        //   sortable: false
        // }
      ]
    };
  }
};
</script>
<style scoped>
.merda {
  font-family: monospace;
}

.item {
  margin-bottom: 4px;
}
.item-name {
  font-size: 14px;
  color: #FFFFFF; 
  display:inline;
}

.item-description {
  font-size: 14px;
  display:inline;
}



</style>