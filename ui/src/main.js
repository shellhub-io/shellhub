import Vue from 'vue';
import App from './App';
import router from './router';
import store from './store';
import vuetify from './plugins/vuetify';

Vue.config.productionTip = false;

Vue.use(require('vue-moment'));

new Vue({
  vuetify,
  router,
  store,
  render: (h) => h(App),
}).$mount('#app');
