import Vue from 'vue'
import Vuex from 'vuex'
import * as mutations from './mutations.js'
import App from './App.vue'
import 'animate.css'
import 'bootstrap/dist/css/bootstrap.min.css'
import '@fortawesome/fontawesome-free/css/all.css'

Vue.config.productionTip = false

Vue.use(Vuex)

function getRoomIdFromQueryString() {
  const urlParams = new URLSearchParams(window.location.search);
  const roomId = urlParams.get('roomId');
  return roomId
}

const store = new Vuex.Store({
  state: {
    apiVersion: process.env.VUE_APP_API_VERSION,
    ws: null,
    connected: false,
    errorMessage: "",
    roomId: getRoomIdFromQueryString(),
    columns: [
      {
        id: "0",
        title: "Good",
        cardStyle: {
          backgroundColor: "bg-success",
        },
        groups: [{
          id: "default",
          columnId: "0",
          isEditable: false,
          title: "ungrouped cards",
          retroCards: [],
        }],
      },
      {
        id: "1",
        title: "Bad",
        cardStyle: {
          backgroundColor: "bg-danger",
        },
        groups: [{
          id: "default",
          columnId: "1",
          isEditable: false,
          title: "ungrouped cards",
          retroCards: [],
        }],
      },
      {
        id: "3",
        title: "Actions",
        cardStyle: {
          backgroundColor: "bg-primary",
        },
        groups: [{
          id: "default",
          columnId: "3",
          isEditable: false,
          title: "ungrouped cards",
          retroCards: [],
        }],
      },
    ],
  },
  mutations: mutations,
  getters: {
    connected: function (state) {
      return state.connected
    },
    errorMessage: function (state) {
      return state.errorMessage
    }
  },
})

new Vue({
  render: h => h(App),
  store: store,
}).$mount('#app')
